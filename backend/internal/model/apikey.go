// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

package model

import (
	"time"

	"github.com/google/uuid"
)

type APIKey struct {
	ID         uuid.UUID  `json:"id"`
	KeyPrefix  string     `json:"key_prefix"`
	KeyHash    string     `json:"-"`
	Name       string     `json:"name"`
	RateLimit  int        `json:"rate_limit"`
	IsActive   bool       `json:"is_active"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

type CreateAPIKeyRequest struct {
	Name      string     `json:"name" validate:"required"`
	RateLimit *int       `json:"rate_limit,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type CreateAPIKeyResponse struct {
	ID        uuid.UUID `json:"id"`
	Key       string    `json:"key"`
	KeyPrefix string    `json:"key_prefix"`
	Name      string    `json:"name"`
	RateLimit int       `json:"rate_limit"`
}
