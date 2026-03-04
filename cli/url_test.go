// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

package main

import "testing"

func TestParseSharePwdURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		wantToken   string
		wantKey     string
		wantType    string
		wantServer  string
		wantErr     bool
	}{
		{
			name:       "text keyless",
			url:        "https://sharepwd.io/s/abc123#key456",
			wantToken:  "abc123",
			wantKey:    "key456",
			wantType:   "text",
			wantServer: "https://sharepwd.io",
		},
		{
			name:       "text passphrase",
			url:        "https://sharepwd.io/s/abc123",
			wantToken:  "abc123",
			wantKey:    "",
			wantType:   "text",
			wantServer: "https://sharepwd.io",
		},
		{
			name:       "file keyless",
			url:        "https://sharepwd.io/f/abc123#key456",
			wantToken:  "abc123",
			wantKey:    "key456",
			wantType:   "file",
			wantServer: "https://sharepwd.io",
		},
		{
			name:       "file passphrase",
			url:        "https://sharepwd.io/f/abc123",
			wantToken:  "abc123",
			wantKey:    "",
			wantType:   "file",
			wantServer: "https://sharepwd.io",
		},
		{
			name:       "custom server",
			url:        "https://secrets.mycompany.com/s/token123#keyABC",
			wantToken:  "token123",
			wantKey:    "keyABC",
			wantType:   "text",
			wantServer: "https://secrets.mycompany.com",
		},
		{
			name:       "long token and key",
			url:        "https://sharepwd.io/s/95a74126b9527f14c24665094ff48fc6#0UyarKd45ltGHFNoT_AuhwwhrAdFPbZ2h31SvMF2hPA",
			wantToken:  "95a74126b9527f14c24665094ff48fc6",
			wantKey:    "0UyarKd45ltGHFNoT_AuhwwhrAdFPbZ2h31SvMF2hPA",
			wantType:   "text",
			wantServer: "https://sharepwd.io",
		},
		{
			name:    "invalid prefix",
			url:     "https://sharepwd.io/x/abc123",
			wantErr: true,
		},
		{
			name:    "no path",
			url:     "https://sharepwd.io/",
			wantErr: true,
		},
		{
			name:    "missing token",
			url:     "https://sharepwd.io/s/",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseSharePwdURL(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if parsed.Token != tt.wantToken {
				t.Errorf("token: got %q, want %q", parsed.Token, tt.wantToken)
			}
			if parsed.KeyFragment != tt.wantKey {
				t.Errorf("key: got %q, want %q", parsed.KeyFragment, tt.wantKey)
			}
			if parsed.ContentType != tt.wantType {
				t.Errorf("type: got %q, want %q", parsed.ContentType, tt.wantType)
			}
			if parsed.Server != tt.wantServer {
				t.Errorf("server: got %q, want %q", parsed.Server, tt.wantServer)
			}
		})
	}
}

func TestBuildShareURL(t *testing.T) {
	tests := []struct {
		server      string
		token       string
		key         string
		contentType string
		want        string
	}{
		{"https://sharepwd.io", "abc123", "key456", "text", "https://sharepwd.io/s/abc123#key456"},
		{"https://sharepwd.io", "abc123", "", "text", "https://sharepwd.io/s/abc123"},
		{"https://sharepwd.io", "abc123", "key456", "file", "https://sharepwd.io/f/abc123#key456"},
		{"https://sharepwd.io/", "abc123", "key456", "text", "https://sharepwd.io/s/abc123#key456"},
	}

	for _, tt := range tests {
		got := BuildShareURL(tt.server, tt.token, tt.key, tt.contentType)
		if got != tt.want {
			t.Errorf("BuildShareURL(%q, %q, %q, %q) = %q, want %q",
				tt.server, tt.token, tt.key, tt.contentType, got, tt.want)
		}
	}
}
