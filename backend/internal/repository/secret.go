package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jizo-hr/sharepwd/internal/model"
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

func (r *SecretRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM secrets WHERE 
		(expires_at IS NOT NULL AND expires_at < NOW() AND is_expired = true)
		OR (is_expired = true AND expired_at < NOW() - INTERVAL '24 hours')`
	result, err := r.db.Exec(ctx, query)
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
