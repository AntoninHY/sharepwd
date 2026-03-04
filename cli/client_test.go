// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientCreateSecret(t *testing.T) {
	expires := time.Now().Add(time.Hour)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/secrets" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("missing Content-Type")
		}
		if r.Header.Get("User-Agent") == "" {
			t.Error("missing User-Agent")
		}

		var req CreateSecretRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.EncryptedData == "" || req.IV == "" {
			t.Error("missing encrypted_data or iv")
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(CreateSecretResponse{
			AccessToken:  "test-token",
			CreatorToken: "test-creator",
			ExpiresAt:    &expires,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test")
	resp, err := client.CreateSecret(&CreateSecretRequest{
		EncryptedData: "encrypted",
		IV:            "iv-data",
		BurnAfterRead: true,
		ContentType:   "text",
	})
	if err != nil {
		t.Fatalf("CreateSecret: %v", err)
	}
	if resp.AccessToken != "test-token" {
		t.Errorf("access_token: got %q", resp.AccessToken)
	}
	if resp.CreatorToken != "test-creator" {
		t.Errorf("creator_token: got %q", resp.CreatorToken)
	}
}

func TestClientGetMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/secrets/my-token" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode(SecretMetadata{
			AccessToken:    "my-token",
			HasPassphrase:  false,
			ContentType:    "text",
			BurnAfterRead:  true,
			ChallengeNonce: "nonce123",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test")
	meta, err := client.GetMetadata("my-token")
	if err != nil {
		t.Fatalf("GetMetadata: %v", err)
	}
	if meta.ChallengeNonce != "nonce123" {
		t.Errorf("nonce: got %q", meta.ChallengeNonce)
	}
	if !meta.BurnAfterRead {
		t.Error("expected burn_after_read")
	}
}

func TestClientRevealSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/secrets/my-token/reveal" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		var req RevealSecretRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.ChallengeNonce != "nonce123" {
			t.Errorf("nonce: got %q", req.ChallengeNonce)
		}
		json.NewEncoder(w).Encode(RevealSecretResponse{
			EncryptedData: "encrypted-data",
			IV:            "iv-data",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test")
	meta := &SecretMetadata{ChallengeNonce: "nonce123"}
	resp, err := client.RevealSecret("my-token", meta)
	if err != nil {
		t.Fatalf("RevealSecret: %v", err)
	}
	if resp.EncryptedData != "encrypted-data" {
		t.Errorf("encrypted_data: got %q", resp.EncryptedData)
	}
}

func TestClientDeleteSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v1/secrets/my-token" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		var req DeleteSecretRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.CreatorToken != "creator-123" {
			t.Errorf("creator_token: got %q", req.CreatorToken)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test")
	if err := client.DeleteSecret("my-token", "creator-123"); err != nil {
		t.Fatalf("DeleteSecret: %v", err)
	}
}

func TestClientAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(APIErrorResponse{Error: "secret not found"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test")
	_, err := client.GetMetadata("nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	if got := err.Error(); got != "API error 404: secret not found" {
		t.Errorf("error: got %q", got)
	}
}
