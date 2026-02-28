package main

import (
	"bytes"
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

// RevealSecret sends a POST /v1/secrets/{token}/reveal request.
func (c *Client) RevealSecret(token, challengeNonce string) (*RevealSecretResponse, error) {
	body, err := json.Marshal(RevealSecretRequest{ChallengeNonce: challengeNonce})
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

// readError extracts a meaningful error from an HTTP response.
func (c *Client) readError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	var apiErr APIErrorResponse
	if json.Unmarshal(body, &apiErr) == nil && apiErr.Error != "" {
		return fmt.Errorf("API error %d: %s", resp.StatusCode, apiErr.Error)
	}

	return fmt.Errorf("API error %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
}
