// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

package model

import (
	"time"

	"github.com/google/uuid"
)

type ContentType string

const (
	ContentTypeText ContentType = "text"
	ContentTypeFile ContentType = "file"
)

type Secret struct {
	ID            uuid.UUID   `json:"id"`
	AccessToken   string      `json:"access_token"`
	EncryptedData string      `json:"encrypted_data"`
	IV            string      `json:"iv"`
	Salt          *string     `json:"salt,omitempty"`
	MaxViews      *int        `json:"max_views,omitempty"`
	CurrentViews  int         `json:"current_views"`
	ExpiresAt     *time.Time  `json:"expires_at,omitempty"`
	BurnAfterRead bool        `json:"burn_after_read"`
	GraceUntil    *time.Time  `json:"grace_until,omitempty"`
	CreatorToken  string      `json:"-"`
	IPHash        *string     `json:"-"`
	UAHash        *string     `json:"-"`
	ContentType   ContentType `json:"content_type"`
	IsExpired     bool        `json:"is_expired"`
	ExpiredAt     *time.Time  `json:"expired_at,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

type CreateSecretRequest struct {
	EncryptedData string  `json:"encrypted_data" validate:"required"`
	IV            string  `json:"iv" validate:"required"`
	Salt          *string `json:"salt,omitempty"`
	MaxViews      *int    `json:"max_views,omitempty"`
	ExpiresIn     *string `json:"expires_in,omitempty"`
	BurnAfterRead bool    `json:"burn_after_read"`
	ContentType   string  `json:"content_type"`
}

type SecretMetadata struct {
	AccessToken   string      `json:"access_token"`
	HasPassphrase bool        `json:"has_passphrase"`
	ContentType   ContentType `json:"content_type"`
	MaxViews      *int        `json:"max_views,omitempty"`
	CurrentViews  int         `json:"current_views"`
	ExpiresAt     *time.Time  `json:"expires_at,omitempty"`
	BurnAfterRead bool        `json:"burn_after_read"`
	IsExpired     bool        `json:"is_expired"`
	CreatedAt     time.Time   `json:"created_at"`
}

type RevealSecretRequest struct {
	ChallengeNonce  string `json:"challenge_nonce"`
	PowSolution     uint64 `json:"pow_solution,omitempty"`
	BehavioralProof string `json:"behavioral_proof,omitempty"`
	EnvFingerprint  string `json:"env_fingerprint,omitempty"`
}

type RevealSecretResponse struct {
	EncryptedData string  `json:"encrypted_data"`
	IV            string  `json:"iv"`
	Salt          *string `json:"salt,omitempty"`
}

type CreateSecretResponse struct {
	AccessToken  string     `json:"access_token"`
	CreatorToken string     `json:"creator_token"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

type DeleteSecretRequest struct {
	CreatorToken string `json:"creator_token"`
}
