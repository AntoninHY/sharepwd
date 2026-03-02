package main

import "time"

// CreateSecretRequest mirrors the backend model.CreateSecretRequest.
type CreateSecretRequest struct {
	EncryptedData string  `json:"encrypted_data"`
	IV            string  `json:"iv"`
	Salt          *string `json:"salt,omitempty"`
	MaxViews      *int    `json:"max_views,omitempty"`
	ExpiresIn     *string `json:"expires_in,omitempty"`
	BurnAfterRead bool    `json:"burn_after_read"`
	ContentType   string  `json:"content_type"`
}

// CreateSecretResponse mirrors the backend model.CreateSecretResponse.
type CreateSecretResponse struct {
	AccessToken  string     `json:"access_token"`
	CreatorToken string     `json:"creator_token"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

// SecretMetadata mirrors the backend metadata + challenge nonce + PoW.
type SecretMetadata struct {
	AccessToken    string  `json:"access_token"`
	HasPassphrase  bool    `json:"has_passphrase"`
	ContentType    string  `json:"content_type"`
	MaxViews       *int    `json:"max_views,omitempty"`
	CurrentViews   int     `json:"current_views"`
	ExpiresAt      *string `json:"expires_at,omitempty"`
	BurnAfterRead  bool    `json:"burn_after_read"`
	IsExpired      bool    `json:"is_expired"`
	CreatedAt      string  `json:"created_at"`
	ChallengeNonce string  `json:"challenge_nonce"`
	PowChallenge   string  `json:"pow_challenge"`
	PowDifficulty  uint8   `json:"pow_difficulty"`
}

// RevealSecretRequest is sent to POST /v1/secrets/{token}/reveal.
type RevealSecretRequest struct {
	ChallengeNonce string `json:"challenge_nonce"`
	PowSolution    uint64 `json:"pow_solution,omitempty"`
}

// RevealSecretResponse contains the encrypted payload.
type RevealSecretResponse struct {
	EncryptedData string  `json:"encrypted_data"`
	IV            string  `json:"iv"`
	Salt          *string `json:"salt,omitempty"`
}

// DeleteSecretRequest is sent to DELETE /v1/secrets/{token}.
type DeleteSecretRequest struct {
	CreatorToken string `json:"creator_token"`
}

// APIErrorResponse is the standard error shape from the backend.
type APIErrorResponse struct {
	Error string `json:"error"`
}

// InitFileUploadRequest for chunked file upload init.
type InitFileUploadRequest struct {
	EncryptedName string  `json:"encrypted_name"`
	OriginalSize  int64   `json:"original_size"`
	ChunkCount    int     `json:"chunk_count"`
	MaxViews      *int    `json:"max_views,omitempty"`
	ExpiresIn     *string `json:"expires_in,omitempty"`
	BurnAfterRead bool    `json:"burn_after_read"`
	IV            string  `json:"iv"`
	Salt          *string `json:"salt,omitempty"`
}

// InitFileUploadResponse from the backend.
type InitFileUploadResponse struct {
	AccessToken  string `json:"access_token"`
	FileID       string `json:"file_id"`
	CreatorToken string `json:"creator_token"`
}

// FilePayload is the JSON structure stored as encrypted_data for inline files.
type FilePayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}
