// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func cmdAdmin(args []string) {
	if len(args) < 1 {
		printAdminUsage()
		os.Exit(exitUsage)
	}

	switch args[0] {
	case "keys":
		if len(args) < 2 {
			printAdminKeysUsage()
			os.Exit(exitUsage)
		}
		switch args[1] {
		case "create":
			cmdAdminKeysCreate(args[2:])
		case "list":
			cmdAdminKeysList(args[2:])
		case "revoke":
			cmdAdminKeysRevoke(args[2:])
		default:
			errMsg("Unknown subcommand: admin keys %s", args[1])
			printAdminKeysUsage()
			os.Exit(exitUsage)
		}
	default:
		errMsg("Unknown subcommand: admin %s", args[0])
		printAdminUsage()
		os.Exit(exitUsage)
	}
}

func cmdAdminKeysCreate(args []string) {
	fs := flag.NewFlagSet("admin keys create", flag.ExitOnError)
	name := fs.String("name", "", "Name for the API key (required)")
	rateLimit := fs.Int("rate-limit", 0, "Rate limit (requests/min, 0 = server default)")
	expiresAt := fs.String("expires-at", "", "Expiration time (RFC3339, e.g. 2026-12-31T23:59:59Z)")
	adminSecret := fs.String("admin-secret", "", "Admin secret (or SHAREPWD_ADMIN_SECRET env)")
	server := fs.String("server", defaultServer, "SharePwd server URL")
	jsonOutput := fs.Bool("json", false, "Output as JSON")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `%s

%s
  sharepwd admin keys create --name <name> [flags]

%s
`, color(colorBold, "sharepwd admin keys create — Create a new API key"),
			color(colorDim, "Usage:"),
			color(colorDim, "Flags:"))
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(exitUsage)
	}

	secret := resolveAuth(*adminSecret, "SHAREPWD_ADMIN_SECRET")
	if secret == "" {
		errMsg("--admin-secret or SHAREPWD_ADMIN_SECRET is required")
		os.Exit(exitUsage)
	}

	if *name == "" {
		errMsg("--name is required")
		os.Exit(exitUsage)
	}

	req := &AdminCreateAPIKeyRequest{Name: *name}
	if *rateLimit > 0 {
		req.RateLimit = rateLimit
	}
	if *expiresAt != "" {
		t, err := time.Parse(time.RFC3339, *expiresAt)
		if err != nil {
			errMsg("Invalid --expires-at format (expected RFC3339): %v", err)
			os.Exit(exitUsage)
		}
		req.ExpiresAt = &t
	}

	apiServer := resolveAPIServer(*server)
	client := NewClient(apiServer, Version)

	status("Creating API key...")
	resp, err := client.AdminCreateAPIKey(secret, req)
	if err != nil {
		errMsg("API error: %v", err)
		os.Exit(exitAPI)
	}

	if *jsonOutput {
		printJSON(resp)
	} else {
		success("API key created")
		fmt.Fprintln(os.Stderr, color(colorDim, "  ID:         ")+resp.ID)
		fmt.Fprintln(os.Stderr, color(colorDim, "  Name:       ")+resp.Name)
		fmt.Fprintln(os.Stderr, color(colorDim, "  Prefix:     ")+resp.KeyPrefix)
		fmt.Fprintln(os.Stderr, color(colorDim, "  Rate limit: ")+fmt.Sprintf("%d req/min", resp.RateLimit))
		// Raw key on stdout for piping (e.g., | pbcopy)
		fmt.Println(resp.Key)
	}
}

func cmdAdminKeysList(args []string) {
	fs := flag.NewFlagSet("admin keys list", flag.ExitOnError)
	apiKey := fs.String("api-key", "", "API key for authentication (or SHAREPWD_API_KEY env)")
	server := fs.String("server", defaultServer, "SharePwd server URL")
	jsonOutput := fs.Bool("json", false, "Output as JSON to stdout")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `%s

%s
  sharepwd admin keys list [flags]

%s
`, color(colorBold, "sharepwd admin keys list — List all API keys"),
			color(colorDim, "Usage:"),
			color(colorDim, "Flags:"))
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(exitUsage)
	}

	key := resolveAuth(*apiKey, "SHAREPWD_API_KEY")
	if key == "" {
		errMsg("--api-key or SHAREPWD_API_KEY is required")
		os.Exit(exitUsage)
	}

	apiServer := resolveAPIServer(*server)
	client := NewClient(apiServer, Version)

	status("Fetching API keys...")
	keys, err := client.ListAPIKeys(key)
	if err != nil {
		errMsg("API error: %v", err)
		os.Exit(exitAPI)
	}

	if *jsonOutput {
		printJSON(keys)
		return
	}

	if len(keys) == 0 {
		fmt.Fprintln(os.Stderr, "No API keys found.")
		return
	}

	// Table header
	fmt.Fprintf(os.Stderr, "\n%-38s %-12s %-20s %6s %-8s %-20s\n",
		"ID", "PREFIX", "NAME", "RATE", "ACTIVE", "CREATED")
	fmt.Fprintf(os.Stderr, "%s\n", "────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────")

	for _, k := range keys {
		active := "yes"
		if !k.IsActive {
			active = "no"
		}
		fmt.Fprintf(os.Stderr, "%-38s %-12s %-20s %6d %-8s %-20s\n",
			k.ID,
			k.KeyPrefix,
			truncate(k.Name, 20),
			k.RateLimit,
			active,
			k.CreatedAt.Format("2006-01-02 15:04"),
		)
	}
	fmt.Fprintln(os.Stderr)
}

func cmdAdminKeysRevoke(args []string) {
	fs := flag.NewFlagSet("admin keys revoke", flag.ExitOnError)
	id := fs.String("id", "", "API key ID to revoke (required)")
	apiKey := fs.String("api-key", "", "API key for authentication (or SHAREPWD_API_KEY env)")
	server := fs.String("server", defaultServer, "SharePwd server URL")
	jsonOutput := fs.Bool("json", false, "Output as JSON")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `%s

%s
  sharepwd admin keys revoke --id <uuid> [flags]

%s
`, color(colorBold, "sharepwd admin keys revoke — Revoke an API key"),
			color(colorDim, "Usage:"),
			color(colorDim, "Flags:"))
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(exitUsage)
	}

	key := resolveAuth(*apiKey, "SHAREPWD_API_KEY")
	if key == "" {
		errMsg("--api-key or SHAREPWD_API_KEY is required")
		os.Exit(exitUsage)
	}

	if *id == "" {
		errMsg("--id is required")
		os.Exit(exitUsage)
	}

	apiServer := resolveAPIServer(*server)
	client := NewClient(apiServer, Version)

	status("Revoking API key...")
	if err := client.RevokeAPIKey(key, *id); err != nil {
		errMsg("API error: %v", err)
		os.Exit(exitAPI)
	}

	if *jsonOutput {
		printJSON(map[string]string{"status": "revoked", "id": *id})
	} else {
		success("API key %s revoked", *id)
	}
}

// resolveAuth returns flagValue if set, otherwise falls back to the env var.
func resolveAuth(flagValue, envVar string) string {
	if flagValue != "" {
		return flagValue
	}
	return os.Getenv(envVar)
}

// truncate shortens s to maxLen, appending "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func printAdminUsage() {
	fmt.Fprintf(os.Stderr, `%s

%s
  sharepwd admin <subcommand>

%s
  keys      Manage API keys (create, list, revoke)

`, color(colorBold, "sharepwd admin — Admin operations"),
		color(colorDim, "Usage:"),
		color(colorDim, "Subcommands:"),
	)
}

func printAdminKeysUsage() {
	fmt.Fprintf(os.Stderr, `%s

%s
  sharepwd admin keys <command>

%s
  create    Create a new API key
  list      List all API keys
  revoke    Revoke an API key

%s
  sharepwd admin keys create --name "ci-pipeline" --admin-secret $SECRET
  sharepwd admin keys list --api-key $KEY
  sharepwd admin keys revoke --id <uuid> --api-key $KEY

`, color(colorBold, "sharepwd admin keys — Manage API keys"),
		color(colorDim, "Usage:"),
		color(colorDim, "Commands:"),
		color(colorDim, "Examples:"),
	)
}
