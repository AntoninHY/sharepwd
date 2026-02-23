package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jizo-hr/sharepwd/internal/config"
	"github.com/jizo-hr/sharepwd/internal/model"
	"github.com/jizo-hr/sharepwd/internal/repository"
)

type SecretService struct {
	repo   *repository.SecretRepository
	config *config.Config
}

func NewSecretService(repo *repository.SecretRepository, cfg *config.Config) *SecretService {
	return &SecretService{repo: repo, config: cfg}
}

func (s *SecretService) Create(ctx context.Context, req *model.CreateSecretRequest, ip, ua string) (*model.CreateSecretResponse, error) {
	if int64(len(req.EncryptedData)) > s.config.MaxTextSize*2 {
		return nil, fmt.Errorf("encrypted data exceeds maximum size")
	}

	accessToken, err := generateToken(16)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	creatorTokenRaw, err := generateToken(32)
	if err != nil {
		return nil, fmt.Errorf("generate creator token: %w", err)
	}
	creatorTokenHash := hashSHA256(creatorTokenRaw)

	secret := &model.Secret{
		ID:            uuid.New(),
		AccessToken:   accessToken,
		EncryptedData: req.EncryptedData,
		IV:            req.IV,
		Salt:          req.Salt,
		BurnAfterRead: req.BurnAfterRead,
		CreatorToken:  creatorTokenHash,
		ContentType:   model.ContentTypeText,
		MaxViews:      req.MaxViews,
	}

	if req.ContentType == "file" {
		secret.ContentType = model.ContentTypeFile
	}

	if req.ExpiresIn != nil {
		d, err := time.ParseDuration(*req.ExpiresIn)
		if err != nil {
			return nil, fmt.Errorf("invalid expires_in duration: %w", err)
		}
		t := time.Now().Add(d)
		secret.ExpiresAt = &t
	}

	if secret.BurnAfterRead {
		g := time.Now().Add(s.config.GracePeriod)
		secret.GraceUntil = &g
	}

	ipHash := hashSHA256(ip)
	uaHash := hashSHA256(ua)
	secret.IPHash = &ipHash
	secret.UAHash = &uaHash

	if err := s.repo.Create(ctx, secret); err != nil {
		return nil, fmt.Errorf("create secret: %w", err)
	}

	slog.Info("secret created",
		"access_token", accessToken,
		"content_type", secret.ContentType,
		"burn_after_read", secret.BurnAfterRead,
	)

	return &model.CreateSecretResponse{
		AccessToken:  accessToken,
		CreatorToken: creatorTokenRaw,
		ExpiresAt:    secret.ExpiresAt,
	}, nil
}

func (s *SecretService) GetMetadata(ctx context.Context, token string) (*model.SecretMetadata, error) {
	secret, err := s.repo.GetByAccessToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if secret == nil {
		return nil, nil
	}

	return &model.SecretMetadata{
		AccessToken:   secret.AccessToken,
		HasPassphrase: secret.Salt != nil,
		ContentType:   secret.ContentType,
		MaxViews:      secret.MaxViews,
		CurrentViews:  secret.CurrentViews,
		ExpiresAt:     secret.ExpiresAt,
		BurnAfterRead: secret.BurnAfterRead,
		IsExpired:     secret.IsExpired,
		CreatedAt:     secret.CreatedAt,
	}, nil
}

func (s *SecretService) Reveal(ctx context.Context, token string) (*model.RevealSecretResponse, error) {
	secret, err := s.repo.GetByAccessToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("get secret: %w", err)
	}
	if secret == nil {
		return nil, nil
	}
	if secret.IsExpired {
		return nil, fmt.Errorf("secret has expired")
	}

	if secret.ExpiresAt != nil && time.Now().After(*secret.ExpiresAt) {
		_ = s.repo.MarkExpired(ctx, secret.ID)
		return nil, fmt.Errorf("secret has expired")
	}

	if secret.MaxViews != nil && secret.CurrentViews >= *secret.MaxViews {
		_ = s.repo.MarkExpired(ctx, secret.ID)
		return nil, fmt.Errorf("secret has reached max views")
	}

	views, err := s.repo.IncrementViews(ctx, secret.ID)
	if err != nil {
		return nil, fmt.Errorf("increment views: %w", err)
	}

	shouldBurn := false
	if secret.BurnAfterRead {
		if secret.GraceUntil != nil && time.Now().After(*secret.GraceUntil) {
			shouldBurn = true
		} else if secret.GraceUntil == nil {
			shouldBurn = true
		}
	}

	if secret.MaxViews != nil && views >= *secret.MaxViews {
		shouldBurn = true
	}

	if shouldBurn {
		if err := s.repo.MarkExpired(ctx, secret.ID); err != nil {
			slog.Error("failed to mark secret expired", "error", err)
		}
	}

	return &model.RevealSecretResponse{
		EncryptedData: secret.EncryptedData,
		IV:            secret.IV,
		Salt:          secret.Salt,
	}, nil
}

func (s *SecretService) Delete(ctx context.Context, accessToken, creatorToken string) error {
	creatorHash := hashSHA256(creatorToken)
	secret, err := s.repo.GetByCreatorToken(ctx, accessToken, creatorHash)
	if err != nil {
		return fmt.Errorf("get secret by creator token: %w", err)
	}
	if secret == nil {
		return fmt.Errorf("secret not found or invalid creator token")
	}
	return s.repo.Delete(ctx, secret.ID)
}

func generateToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashSHA256(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
