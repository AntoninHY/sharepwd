package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

const (
	defaultServer = "https://sharepwd.io"

	exitUsage  = 1
	exitAPI    = 2
	exitCrypto = 3
	exitIO     = 4
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(exitUsage)
	}

	switch os.Args[1] {
	case "push":
		cmdPush(os.Args[2:])
	case "pull":
		cmdPull(os.Args[2:])
	case "delete":
		cmdDelete(os.Args[2:])
	case "version":
		cmdVersion()
	case "help", "-h", "--help":
		printUsage()
	default:
		errMsg("Unknown command: %s", os.Args[1])
		printUsage()
		os.Exit(exitUsage)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `%s

%s
  sharepwd <command> [flags]

%s
  push      Encrypt and share a secret
  pull      Retrieve and decrypt a secret
  delete    Delete a secret by creator token
  version   Show version information
  help      Show this help

%s
  sharepwd push "db_password=S3cureP@ss!" --burn --ttl 1h
  sharepwd push -f secret.pdf --burn --ttl 24h
  echo "secret" | sharepwd push --burn
  sharepwd pull https://sharepwd.io/s/abc123#key456
  sharepwd delete https://sharepwd.io/s/abc123 --creator-token def789

`,
		color(colorBold, "SharePwd CLI — Zero-knowledge secret sharing"),
		color(colorDim, "Usage:"),
		color(colorDim, "Commands:"),
		color(colorDim, "Examples:"),
	)
}

func cmdPush(args []string) {
	fs := flag.NewFlagSet("push", flag.ExitOnError)
	filePath := fs.String("f", "", "Path to file to encrypt and share")
	passphrase := fs.Bool("p", false, "Protect with passphrase (interactive prompt)")
	burn := fs.Bool("burn", false, "Burn after read")
	ttl := fs.String("ttl", "", "Time to live (e.g. 1h, 24h, 7d)")
	maxViews := fs.Int("max-views", 0, "Maximum number of views (0 = unlimited)")
	server := fs.String("server", defaultServer, "SharePwd server URL")
	jsonOutput := fs.Bool("json", false, "Output as JSON")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `%s

%s
  sharepwd push [flags] [secret text]

%s
`, color(colorBold, "sharepwd push — Encrypt and share a secret"),
			color(colorDim, "Usage:"),
			color(colorDim, "Flags:"))
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(exitUsage)
	}

	var plaintext []byte
	var contentType string

	switch {
	case *filePath != "":
		// File mode
		data, err := os.ReadFile(*filePath)
		if err != nil {
			errMsg("Cannot read file: %v", err)
			os.Exit(exitIO)
		}
		fileName := fileBaseName(*filePath)
		payload := FilePayload{
			Name: fileName,
			Data: base64.StdEncoding.EncodeToString(data),
		}
		jsonBytes, err := json.Marshal(payload)
		if err != nil {
			errMsg("Cannot encode file payload: %v", err)
			os.Exit(exitIO)
		}
		plaintext = jsonBytes
		contentType = "file"

	case fs.NArg() > 0:
		// Positional argument
		plaintext = []byte(strings.Join(fs.Args(), " "))
		contentType = "text"

	default:
		// Try stdin (only if piped)
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				errMsg("Cannot read stdin: %v", err)
				os.Exit(exitIO)
			}
			plaintext = data
			// Trim trailing newline from piped input
			plaintext = []byte(strings.TrimRight(string(plaintext), "\n"))
			contentType = "text"
		} else {
			errMsg("No input provided. Use: sharepwd push \"secret\" or echo \"secret\" | sharepwd push")
			os.Exit(exitUsage)
		}
	}

	if len(plaintext) == 0 {
		errMsg("Empty input")
		os.Exit(exitUsage)
	}

	var encryptedData, iv string
	var salt *string
	var keyFragment string

	if *passphrase {
		pass, err := readPassphraseConfirm()
		if err != nil {
			errMsg("%v", err)
			os.Exit(exitUsage)
		}
		result, err := EncryptWithPassphrase(plaintext, pass)
		if err != nil {
			errMsg("Encryption failed: %v", err)
			os.Exit(exitCrypto)
		}
		encryptedData = result.EncryptedData
		iv = result.IV
		salt = &result.Salt
	} else {
		result, err := EncryptKeyless(plaintext)
		if err != nil {
			errMsg("Encryption failed: %v", err)
			os.Exit(exitCrypto)
		}
		encryptedData = result.EncryptedData
		iv = result.IV
		keyFragment = result.Key
	}

	req := &CreateSecretRequest{
		EncryptedData: encryptedData,
		IV:            iv,
		Salt:          salt,
		BurnAfterRead: *burn,
		ContentType:   contentType,
	}

	if *ttl != "" {
		req.ExpiresIn = ttl
	}
	if *maxViews > 0 {
		req.MaxViews = maxViews
	}

	apiServer := resolveAPIServer(*server)
	client := NewClient(apiServer, Version)

	status("Uploading encrypted secret...")
	resp, err := client.CreateSecret(req)
	if err != nil {
		errMsg("API error: %v", err)
		os.Exit(exitAPI)
	}

	shareURL := BuildShareURL(*server, resp.AccessToken, keyFragment, contentType)

	if *jsonOutput {
		out := map[string]any{
			"url":           shareURL,
			"creator_token": resp.CreatorToken,
		}
		if resp.ExpiresAt != nil {
			out["expires_at"] = resp.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
		}
		printJSON(out)
	} else {
		success("Secret created")
		fmt.Fprintln(os.Stderr, color(colorDim, "Creator token: ")+resp.CreatorToken)
		if resp.ExpiresAt != nil {
			fmt.Fprintln(os.Stderr, color(colorDim, "Expires: ")+resp.ExpiresAt.Format("2006-01-02 15:04:05 UTC"))
		}
		// URL on stdout for piping (e.g., | pbcopy)
		fmt.Println(shareURL)
	}
}

func cmdPull(args []string) {
	fs := flag.NewFlagSet("pull", flag.ExitOnError)
	passphrase := fs.Bool("p", false, "Decrypt with passphrase (interactive prompt)")
	output := fs.String("o", "", "Output file path (for file secrets)")
	server := fs.String("server", "", "Override API server URL")
	jsonOutput := fs.Bool("json", false, "Output as JSON")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `%s

%s
  sharepwd pull [flags] <url>

%s
`, color(colorBold, "sharepwd pull — Retrieve and decrypt a secret"),
			color(colorDim, "Usage:"),
			color(colorDim, "Flags:"))
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(exitUsage)
	}

	if fs.NArg() < 1 {
		errMsg("URL is required. Usage: sharepwd pull <url>")
		os.Exit(exitUsage)
	}

	parsed, err := ParseSharePwdURL(fs.Arg(0))
	if err != nil {
		errMsg("Invalid URL: %v", err)
		os.Exit(exitUsage)
	}

	apiServer := resolveAPIServer(*server)
	if apiServer == "" {
		apiServer = resolveAPIServer(parsed.Server)
	}
	client := NewClient(apiServer, Version)

	status("Fetching secret metadata...")
	meta, err := client.GetMetadata(parsed.Token)
	if err != nil {
		errMsg("API error: %v", err)
		os.Exit(exitAPI)
	}

	if meta.IsExpired {
		errMsg("This secret has expired")
		os.Exit(exitAPI)
	}

	if meta.BurnAfterRead {
		warn("This secret will be destroyed after viewing")
	}

	// Determine decryption key
	var decryptKey string
	var usePassphrase bool

	if meta.HasPassphrase || *passphrase {
		usePassphrase = true
	} else if parsed.KeyFragment != "" {
		decryptKey = parsed.KeyFragment
	} else {
		errMsg("No encryption key found in URL and secret is not passphrase-protected")
		os.Exit(exitCrypto)
	}

	var pass string
	if usePassphrase {
		pass, err = readPassphrase("Passphrase: ")
		if err != nil {
			errMsg("%v", err)
			os.Exit(exitUsage)
		}
		if pass == "" {
			errMsg("Passphrase cannot be empty")
			os.Exit(exitUsage)
		}
	}

	status("Revealing secret...")
	revealed, err := client.RevealSecret(parsed.Token, meta.ChallengeNonce)
	if err != nil {
		errMsg("API error: %v", err)
		os.Exit(exitAPI)
	}

	// Decrypt
	var plaintext []byte
	if usePassphrase {
		if revealed.Salt == nil {
			errMsg("Missing salt for passphrase decryption")
			os.Exit(exitCrypto)
		}
		plaintext, err = DecryptWithPassphrase(revealed.EncryptedData, revealed.IV, *revealed.Salt, pass)
	} else {
		plaintext, err = DecryptKeyless(revealed.EncryptedData, revealed.IV, decryptKey)
	}
	if err != nil {
		errMsg("Decryption failed: %v", err)
		os.Exit(exitCrypto)
	}

	// Handle content type
	contentType := meta.ContentType
	if contentType == "" {
		contentType = parsed.ContentType
	}

	if contentType == "file" {
		var fp FilePayload
		if err := json.Unmarshal(plaintext, &fp); err != nil {
			errMsg("Cannot parse file payload: %v", err)
			os.Exit(exitCrypto)
		}

		fileData, err := base64.StdEncoding.DecodeString(fp.Data)
		if err != nil {
			errMsg("Cannot decode file data: %v", err)
			os.Exit(exitCrypto)
		}

		outPath := fp.Name
		if *output != "" {
			outPath = *output
		}

		if *jsonOutput {
			printJSON(map[string]any{
				"filename": fp.Name,
				"size":     len(fileData),
				"saved_to": outPath,
			})
		}

		if err := os.WriteFile(outPath, fileData, 0600); err != nil {
			errMsg("Cannot write file: %v", err)
			os.Exit(exitIO)
		}

		if !*jsonOutput {
			success("File saved: %s (%d bytes)", outPath, len(fileData))
		}
	} else {
		if *jsonOutput {
			printJSON(map[string]any{
				"content": string(plaintext),
			})
		} else {
			// Plain text on stdout for piping
			fmt.Print(string(plaintext))
			// Ensure trailing newline if not already present
			if len(plaintext) > 0 && plaintext[len(plaintext)-1] != '\n' {
				fmt.Println()
			}
		}
	}
}

func cmdDelete(args []string) {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	creatorToken := fs.String("creator-token", "", "Creator token (required)")
	server := fs.String("server", "", "Override API server URL")
	jsonOutput := fs.Bool("json", false, "Output as JSON")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `%s

%s
  sharepwd delete [flags] <url>

%s
`, color(colorBold, "sharepwd delete — Delete a secret"),
			color(colorDim, "Usage:"),
			color(colorDim, "Flags:"))
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(exitUsage)
	}

	if fs.NArg() < 1 {
		errMsg("URL is required. Usage: sharepwd delete <url> --creator-token <token>")
		os.Exit(exitUsage)
	}

	if *creatorToken == "" {
		errMsg("--creator-token is required")
		os.Exit(exitUsage)
	}

	parsed, err := ParseSharePwdURL(fs.Arg(0))
	if err != nil {
		errMsg("Invalid URL: %v", err)
		os.Exit(exitUsage)
	}

	apiServer := resolveAPIServer(*server)
	if apiServer == "" {
		apiServer = resolveAPIServer(parsed.Server)
	}
	client := NewClient(apiServer, Version)

	status("Deleting secret...")
	if err := client.DeleteSecret(parsed.Token, *creatorToken); err != nil {
		errMsg("API error: %v", err)
		os.Exit(exitAPI)
	}

	if *jsonOutput {
		printJSON(map[string]string{"status": "deleted"})
	} else {
		success("Secret deleted")
	}
}

func cmdVersion() {
	fmt.Printf("sharepwd %s\n", Version)
	fmt.Printf("  commit:  %s\n", Commit)
	fmt.Printf("  built:   %s\n", BuildDate)
}

// resolveAPIServer converts a frontend server URL to its API equivalent.
// For sharepwd.io, the API runs on the same host.
// If server is empty, returns empty (caller should use parsed URL server).
func resolveAPIServer(server string) string {
	if server == "" {
		return ""
	}
	// The API is served from the same origin on a different path prefix (/v1/).
	// No transformation needed — the client adds /v1/... paths.
	return strings.TrimRight(server, "/")
}

// fileBaseName returns the base file name from a path.
func fileBaseName(path string) string {
	// Handle both unix and windows paths
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[i+1:]
		}
	}
	return path
}
