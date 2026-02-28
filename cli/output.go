package main

import (
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/term"
)

// ANSI color codes.
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

var isTTY bool

func init() {
	isTTY = term.IsTerminal(int(os.Stderr.Fd()))
}

func color(c, text string) string {
	if !isTTY {
		return text
	}
	return c + text + colorReset
}

// status prints an informational message to stderr.
func status(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, color(colorCyan, "→ ")+msg)
}

// success prints a success message to stderr.
func success(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, color(colorGreen, "✓ ")+msg)
}

// warn prints a warning message to stderr.
func warn(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, color(colorYellow, "⚠ ")+msg)
}

// errMsg prints an error message to stderr.
func errMsg(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, color(colorRed, "✗ ")+msg)
}

// printJSON writes v as JSON to stdout.
func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// readPassphrase reads a passphrase from /dev/tty without echo.
// Uses /dev/tty so it works even when stdin is piped.
func readPassphrase(prompt string) (string, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", fmt.Errorf("cannot open /dev/tty (are you running in a terminal?): %w", err)
	}
	defer tty.Close()

	fmt.Fprint(tty, prompt)
	pass, err := term.ReadPassword(int(tty.Fd()))
	fmt.Fprintln(tty) // newline after hidden input
	if err != nil {
		return "", fmt.Errorf("read passphrase: %w", err)
	}

	return string(pass), nil
}

// readPassphraseConfirm reads and confirms a passphrase.
func readPassphraseConfirm() (string, error) {
	pass, err := readPassphrase("Passphrase: ")
	if err != nil {
		return "", err
	}
	if pass == "" {
		return "", fmt.Errorf("passphrase cannot be empty")
	}

	confirm, err := readPassphrase("Confirm passphrase: ")
	if err != nil {
		return "", err
	}

	if pass != confirm {
		return "", fmt.Errorf("passphrases do not match")
	}

	return pass, nil
}
