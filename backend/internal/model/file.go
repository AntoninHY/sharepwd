// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

package model

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID             uuid.UUID `json:"id"`
	SecretID       uuid.UUID `json:"secret_id"`
	EncryptedName  string    `json:"encrypted_name"`
	FileSize       int64     `json:"file_size"`
	OriginalSize   int64     `json:"original_size"`
	StorageKey     string    `json:"-"`
	StorageBackend string    `json:"-"`
	ChunkCount     int       `json:"chunk_count"`
	UploadComplete bool      `json:"upload_complete"`
	CreatedAt      time.Time `json:"created_at"`
}

type InitFileUploadRequest struct {
	EncryptedName string  `json:"encrypted_name" validate:"required"`
	OriginalSize  int64   `json:"original_size" validate:"required"`
	ChunkCount    int     `json:"chunk_count" validate:"required"`
	MaxViews      *int    `json:"max_views,omitempty"`
	ExpiresIn     *string `json:"expires_in,omitempty"`
	BurnAfterRead bool    `json:"burn_after_read"`
	IV            string  `json:"iv" validate:"required"`
	Salt          *string `json:"salt,omitempty"`
}

type InitFileUploadResponse struct {
	SecretAccessToken string `json:"access_token"`
	FileID            string `json:"file_id"`
	CreatorToken      string `json:"creator_token"`
}

type FileMetadata struct {
	EncryptedName string `json:"encrypted_name"`
	OriginalSize  int64  `json:"original_size"`
	ChunkCount    int    `json:"chunk_count"`
}
