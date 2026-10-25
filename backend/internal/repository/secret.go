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

type SecretRepository struct {
	db *pgxpool.Pool
}

func NewSecretRepository(db *pgxpool.Pool) *SecretRepository {
	return &SecretRepository{db: db}
}

func (r *SecretRepository) Create(ctx context.Context, s *model.Secret) error {
	query := `INSERT INTO secrets (
		id, access_token, encrypted_data, iv, salt, max_views, 
		expires_at, burn_after_read, grace_until, creator_token,
		ip_hash, ua_hash, content_type
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.db.Exec(ctx, query,
		s.ID, s.AccessToken, s.EncryptedData, s.IV, s.Salt,
		s.MaxViews, s.ExpiresAt, s.BurnAfterRead, s.GraceUntil,
		s.CreatorToken, s.IPHash, s.UAHash, s.ContentType,
	)
	return err
}

func (r *SecretRepository) GetByAccessToken(ctx context.Context, token string) (*model.Secret, error) {
	query := `SELECT id, access_token, encrypted_data, iv, salt, max_views,
		current_views, expires_at, burn_after_read, grace_until,
		creator_token, ip_hash, ua_hash, content_type, is_expired,
		expired_at, created_at, updated_at
	FROM secrets WHERE access_token = $1`

	var s model.Secret
	var expiresAt, graceUntil, expiredAt sql.NullTime
	var salt, ipHash, uaHash sql.NullString
	var maxViews sql.NullInt32

	err := r.db.QueryRow(ctx, query, token).Scan(
		&s.ID, &s.AccessToken, &s.EncryptedData, &s.IV, &salt,
		&maxViews, &s.CurrentViews, &expiresAt, &s.BurnAfterRead,
		&graceUntil, &s.CreatorToken, &ipHash, &uaHash,
		&s.ContentType, &s.IsExpired, &expiredAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if salt.Valid {
		s.Salt = &salt.String
	}
	if maxViews.Valid {
		v := int(maxViews.Int32)
		s.MaxViews = &v
	}
	if expiresAt.Valid {
		s.ExpiresAt = &expiresAt.Time
	}
	if graceUntil.Valid {
		s.GraceUntil = &graceUntil.Time
	}
	if expiredAt.Valid {
		s.ExpiredAt = &expiredAt.Time
	}
	if ipHash.Valid {
		s.IPHash = &ipHash.String
	}
	if uaHash.Valid {
		s.UAHash = &uaHash.String
	}

	return &s, nil
}

func (r *SecretRepository) IncrementViews(ctx context.Context, id uuid.UUID) (int, error) {
	query := `UPDATE secrets SET current_views = current_views + 1, updated_at = NOW()
		WHERE id = $1 RETURNING current_views`
	var views int
	err := r.db.QueryRow(ctx, query, id).Scan(&views)
	return views, err
}


// AtomicReveal atomically increments views and returns the secret only if it hasn not exceeded max_views.
// Returns nil if the secret is expired or has reached max views.
func (r *SecretRepository) AtomicReveal(ctx context.Context, id uuid.UUID) (*model.Secret, error) {
	query := `
		UPDATE secrets 
		SET current_views = current_views + 1, updated_at = NOW()
		WHERE id = $1 
		AND is_expired = false
		AND (max_views IS NULL OR current_views < max_views)
		AND (expires_at IS NULL OR expires_at > NOW())
		RETURNING id, access_token, encrypted_data, iv, salt, max_views, current_views, 
		expires_at, burn_after_read, grace_until, creator_token, content_type, 
		is_expired, created_at, updated_at`

	var s model.Secret
	var expiresAt, graceUntil sql.NullTime
	var salt sql.NullString
	var maxViews sql.NullInt32

	err := r.db.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.AccessToken, &s.EncryptedData, &s.IV, &salt,
		&maxViews, &s.CurrentViews, &expiresAt, &s.BurnAfterRead,
		&graceUntil, &s.CreatorToken, &s.ContentType,
		&s.IsExpired, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if salt.Valid {
		s.Salt = &salt.String
	}
	if maxViews.Valid {
		v := int(maxViews.Int32)
		s.MaxViews = &v
	}
	if expiresAt.Valid {
		s.ExpiresAt = &expiresAt.Time
	}
	if graceUntil.Valid {
		s.GraceUntil = &graceUntil.Time
	}

	return &s, nil
}

func (r *SecretRepository) MarkExpired(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE secrets SET is_expired = true, expired_at = NOW(), 
		encrypted_data = '', updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *SecretRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM secrets WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// expiredSecretsPredicate selects the secrets whose lifetime is over. It is
// shared by DeleteExpired and ListFilesOfExpiredSecrets so the blobs that get
// deleted can never drift from the rows that get deleted. Column names are left
// unqualified on purpose: the files table carries none of them, so the
// predicate reads the same whether or not secrets is joined.
const expiredSecretsPredicate = `(expires_at IS NOT NULL AND expires_at < NOW() AND is_expired = true)
		OR (is_expired = true AND expired_at < NOW() - INTERVAL '24 hours')`

// ExpiredFileBlob locates a stored file and the chunks holding its content.
type ExpiredFileBlob struct {
	FileID     uuid.UUID
	SecretID   uuid.UUID
	StorageKey string
	ChunkCount int
}

// ListFilesOfExpiredSecrets returns the files attached to the secrets that
// DeleteExpired is about to remove. It has to run first: files.secret_id is
// ON DELETE CASCADE, so deleting the secret takes storage_key with it and
// leaves the objects with nothing pointing at them.
func (r *SecretRepository) ListFilesOfExpiredSecrets(ctx context.Context) ([]ExpiredFileBlob, error) {
	query := `SELECT f.id, f.secret_id, f.storage_key, f.chunk_count
		FROM files f
		JOIN secrets s ON s.id = f.secret_id
		WHERE ` + expiredSecretsPredicate

	return scanFileBlobs(r.db.Query(ctx, query))
}

// ListFilesPastDownloadWindow returns the stored files of secrets whose content
// is already gone and whose download window closed before cutoff.
//
// A secret is marked expired as soon as it is burned or its deadline passes,
// but its row is kept for the ordinary retention period. The file it carries
// only has to outlive that moment long enough for the reader to pull its
// chunks, which is what the grace period is for. Past that point the file has
// no reason to exist.
func (r *SecretRepository) ListFilesPastDownloadWindow(ctx context.Context, cutoff time.Time) ([]ExpiredFileBlob, error) {
	query := `SELECT f.id, f.secret_id, f.storage_key, f.chunk_count
		FROM files f
		JOIN secrets s ON s.id = f.secret_id
		WHERE s.is_expired = true
		  AND s.expired_at IS NOT NULL
		  AND s.expired_at < $1`

	return scanFileBlobs(r.db.Query(ctx, query, cutoff))
}

// scanFileBlobs drains a query returning the four ExpiredFileBlob columns, in
// the order both callers above select them.
func scanFileBlobs(rows pgx.Rows, err error) ([]ExpiredFileBlob, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blobs []ExpiredFileBlob
	for rows.Next() {
		var b ExpiredFileBlob
		if err := rows.Scan(&b.FileID, &b.SecretID, &b.StorageKey, &b.ChunkCount); err != nil {
			return nil, err
		}
		blobs = append(blobs, b)
	}
	return blobs, rows.Err()
}

// DeleteExpired removes the secrets whose lifetime is over. Secrets listed in
// keep are left in place for the next sweep: their objects could not be
// removed, and dropping the row now would strand them in storage for good.
func (r *SecretRepository) DeleteExpired(ctx context.Context, keep []uuid.UUID) (int64, error) {
	if keep == nil {
		// A nil slice encodes as SQL NULL, which would make the comparison
		// below NULL and delete nothing at all.
		keep = []uuid.UUID{}
	}

	query := `DELETE FROM secrets WHERE (` + expiredSecretsPredicate + `) AND NOT (id = ANY($1))`
	result, err := r.db.Exec(ctx, query, keep)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}


func (r *SecretRepository) MarkExpiredSecrets(ctx context.Context) (int64, error) {
	query := `UPDATE secrets SET is_expired = true, expired_at = NOW(), encrypted_data = '', updated_at = NOW()
		WHERE is_expired = false AND (
			(expires_at IS NOT NULL AND expires_at < NOW())
			OR (max_views IS NOT NULL AND current_views >= max_views)
		)`
	result, err := r.db.Exec(ctx, query)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (r *SecretRepository) GetByCreatorToken(ctx context.Context, accessToken string, creatorToken string) (*model.Secret, error) {
	query := `SELECT id FROM secrets WHERE access_token = $1 AND creator_token = $2`
	var id uuid.UUID
	err := r.db.QueryRow(ctx, query, accessToken, creatorToken).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &model.Secret{ID: id}, nil
}

// Ensure interface compliance
var _ time.Duration
