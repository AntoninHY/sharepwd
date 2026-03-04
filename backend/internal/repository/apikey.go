// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/AntoninHY/sharepwd/internal/model"
)

type APIKeyRepository struct {
	db *pgxpool.Pool
}

func NewAPIKeyRepository(db *pgxpool.Pool) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) Create(ctx context.Context, k *model.APIKey) error {
	query := `INSERT INTO api_keys (id, key_prefix, key_hash, name, rate_limit, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query,
		k.ID, k.KeyPrefix, k.KeyHash, k.Name, k.RateLimit, k.ExpiresAt,
	)
	return err
}

func (r *APIKeyRepository) GetByKeyHash(ctx context.Context, keyHash string) (*model.APIKey, error) {
	query := `SELECT id, key_prefix, key_hash, name, rate_limit, is_active,
		last_used_at, created_at, expires_at
	FROM api_keys WHERE key_hash = $1 AND is_active = true`

	var k model.APIKey
	var lastUsed, expiresAt sql.NullTime
	err := r.db.QueryRow(ctx, query, keyHash).Scan(
		&k.ID, &k.KeyPrefix, &k.KeyHash, &k.Name, &k.RateLimit,
		&k.IsActive, &lastUsed, &k.CreatedAt, &expiresAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if lastUsed.Valid {
		k.LastUsedAt = &lastUsed.Time
	}
	if expiresAt.Valid {
		k.ExpiresAt = &expiresAt.Time
	}
	return &k, nil
}

func (r *APIKeyRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *APIKeyRepository) List(ctx context.Context) ([]model.APIKey, error) {
	query := `SELECT id, key_prefix, name, rate_limit, is_active,
		last_used_at, created_at, expires_at
	FROM api_keys ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []model.APIKey
	for rows.Next() {
		var k model.APIKey
		var lastUsed, expiresAt sql.NullTime
		if err := rows.Scan(
			&k.ID, &k.KeyPrefix, &k.Name, &k.RateLimit, &k.IsActive,
			&lastUsed, &k.CreatedAt, &expiresAt,
		); err != nil {
			return nil, err
		}
		if lastUsed.Valid {
			k.LastUsedAt = &lastUsed.Time
		}
		if expiresAt.Valid {
			k.ExpiresAt = &expiresAt.Time
		}
		keys = append(keys, k)
	}
	return keys, nil
}

func (r *APIKeyRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE api_keys SET is_active = false WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

var _ time.Duration
