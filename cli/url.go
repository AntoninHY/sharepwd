// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

package main

import (
	"fmt"
	"net/url"
	"strings"
)

// SharePwdURL holds parsed components of a SharePwd URL.
type SharePwdURL struct {
	Server      string // e.g. "https://sharepwd.io"
	Token       string // access token
	KeyFragment string // Base64URL key from URL fragment (empty if passphrase-protected)
	ContentType string // "text" or "file"
}

// ParseSharePwdURL extracts token, key fragment, and content type from a SharePwd URL.
// Supported formats:
//
//	https://sharepwd.io/s/{token}#{key}         (text, keyless)
//	https://sharepwd.io/s/{token}               (text, passphrase)
//	https://sharepwd.io/f/{token}#{key}         (file, keyless)
//	https://sharepwd.io/f/{token}               (file, passphrase)
func ParseSharePwdURL(rawURL string) (*SharePwdURL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	path := strings.TrimPrefix(u.Path, "/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 || parts[1] == "" {
		return nil, fmt.Errorf("invalid SharePwd URL: expected /s/{token} or /f/{token}")
	}

	prefix := parts[0]
	token := parts[1]

	var contentType string
	switch prefix {
	case "s":
		contentType = "text"
	case "f":
		contentType = "file"
	default:
		return nil, fmt.Errorf("invalid SharePwd URL: unknown path prefix /%s/", prefix)
	}

	server := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	return &SharePwdURL{
		Server:      server,
		Token:       token,
		KeyFragment: u.Fragment,
		ContentType: contentType,
	}, nil
}

// BuildShareURL constructs a full SharePwd URL.
func BuildShareURL(server, token, keyFragment, contentType string) string {
	prefix := "s"
	if contentType == "file" {
		prefix = "f"
	}

	base := fmt.Sprintf("%s/%s/%s", strings.TrimRight(server, "/"), prefix, token)
	if keyFragment != "" {
		return base + "#" + keyFragment
	}
	return base
}
