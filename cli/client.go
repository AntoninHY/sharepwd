// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is the HTTP client for the SharePwd API.
type Client struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string
}

// NewClient creates a new API client.
func NewClient(baseURL, version string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		userAgent: fmt.Sprintf("SharePwd-CLI/%s", version),
	}
}

// CreateSecret sends a POST /v1/secrets request.
func (c *Client) CreateSecret(req *CreateSecretRequest) (*CreateSecretResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, c.baseURL+"/v1/secrets", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, c.readError(resp)
	}

	var result CreateSecretResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &result, nil
}

// GetMetadata sends a GET /v1/secrets/{token} request.
func (c *Client) GetMetadata(token string) (*SecretMetadata, error) {
	httpReq, err := http.NewRequest(http.MethodGet, c.baseURL+"/v1/secrets/"+token, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.readError(resp)
	}

	var result SecretMetadata
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &result, nil
}

// solvePoW finds a counter such that SHA-256(prefix + ":" + counter) has N leading zero bits.
func solvePoW(prefix string, difficulty uint8) uint64 {
	for counter := uint64(1); ; counter++ {
		input := fmt.Sprintf("%s:%d", prefix, counter)
		hash := sha256.Sum256([]byte(input))
		if hasLeadingZeroBits(hash[:], difficulty) {
			return counter
		}
	}
}

func hasLeadingZeroBits(hash []byte, bits uint8) bool {
	fullBytes := bits / 8
	remaining := bits % 8

	for i := uint8(0); i < fullBytes; i++ {
		if hash[i] != 0 {
			return false
		}
	}
	if remaining > 0 {
		mask := byte(0xFF << (8 - remaining))
		if hash[fullBytes]&mask != 0 {
			return false
		}
	}
	return true
}

// RevealSecret sends a POST /v1/secrets/{token}/reveal request.
// It solves the PoW challenge between GetMetadata and Reveal.
func (c *Client) RevealSecret(token string, meta *SecretMetadata) (*RevealSecretResponse, error) {
	req := RevealSecretRequest{ChallengeNonce: meta.ChallengeNonce}

	// Solve Proof-of-Work if challenge is present
	if meta.PowChallenge != "" && meta.PowDifficulty > 0 {
		req.PowSolution = solvePoW(meta.PowChallenge, meta.PowDifficulty)
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, c.baseURL+"/v1/secrets/"+token+"/reveal", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.readError(resp)
	}

	var result RevealSecretResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &result, nil
}

// DeleteSecret sends a DELETE /v1/secrets/{token} request.
func (c *Client) DeleteSecret(token, creatorToken string) error {
	body, err := json.Marshal(DeleteSecretRequest{CreatorToken: creatorToken})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodDelete, c.baseURL+"/v1/secrets/"+token, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return c.readError(resp)
	}
	return nil
}

// AdminCreateAPIKey sends a POST /v1/admin/api-keys request.
func (c *Client) AdminCreateAPIKey(adminSecret string, req *AdminCreateAPIKeyRequest) (*AdminCreateAPIKeyResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, c.baseURL+"/v1/admin/api-keys", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", c.userAgent)
	httpReq.Header.Set("X-Admin-Secret", adminSecret)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, c.readError(resp)
	}

	var result AdminCreateAPIKeyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &result, nil
}

// ListAPIKeys sends a GET /v1/api-keys request.
func (c *Client) ListAPIKeys(apiKey string) ([]APIKeyInfo, error) {
	httpReq, err := http.NewRequest(http.MethodGet, c.baseURL+"/v1/api-keys", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("User-Agent", c.userAgent)
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.readError(resp)
	}

	var result []APIKeyInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return result, nil
}

// RevokeAPIKey sends a DELETE /v1/api-keys/{id} request.
func (c *Client) RevokeAPIKey(apiKey, id string) error {
	httpReq, err := http.NewRequest(http.MethodDelete, c.baseURL+"/v1/api-keys/"+id, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("User-Agent", c.userAgent)
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return c.readError(resp)
	}
	return nil
}

// readError extracts a meaningful error from an HTTP response.
func (c *Client) readError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	var apiErr APIErrorResponse
	if json.Unmarshal(body, &apiErr) == nil && apiErr.Error != "" {
		return fmt.Errorf("API error %d: %s", resp.StatusCode, apiErr.Error)
	}

	return fmt.Errorf("API error %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
}
