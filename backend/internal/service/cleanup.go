// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/AntoninHY/sharepwd/internal/config"
	"github.com/AntoninHY/sharepwd/internal/repository"
	"github.com/AntoninHY/sharepwd/internal/storage"
)

type CleanupService struct {
	secretRepo *repository.SecretRepository
	store      storage.Storage
	config     *config.Config
}

func NewCleanupService(secretRepo *repository.SecretRepository, store storage.Storage, cfg *config.Config) *CleanupService {
	return &CleanupService{secretRepo: secretRepo, store: store, config: cfg}
}

func (s *CleanupService) Start(ctx context.Context) {
	ticker := time.NewTicker(s.config.CleanupInterval)
	defer ticker.Stop()

	slog.Info("cleanup worker started", "interval", s.config.CleanupInterval)

	for {
		select {
		case <-ctx.Done():
			slog.Info("cleanup worker stopped")
			return
		case <-ticker.C:
			s.run(ctx)
		}
	}
}

func (s *CleanupService) run(ctx context.Context) {
	marked, err := s.secretRepo.MarkExpiredSecrets(ctx)
	if err != nil {
		slog.Error("failed to mark expired secrets", "error", err)
		return
	}

	kept, purged, err := s.purgeBlobs(ctx)
	if err != nil {
		// Without the storage keys, deleting the rows now would strand every
		// one of these objects for good. The sweep gives up and retries later.
		slog.Error("skipping secret deletion, stored files could not be listed", "error", err)
		return
	}

	deleted, err := s.secretRepo.DeleteExpired(ctx, kept)
	if err != nil {
		slog.Error("failed to delete expired secrets", "error", err)
		return
	}

	if marked > 0 || deleted > 0 || purged > 0 {
		slog.Info("cleanup completed",
			"marked", marked, "deleted", deleted,
			"blobs_purged", purged, "blobs_retried_next_sweep", len(kept))
	}
}

// purgeBlobs removes the stored files of the secrets the sweep is about to
// delete, and returns the secrets whose objects could not be removed so their
// rows survive and the next sweep tries again. An error means nothing could be
// listed, and the caller must then delete no rows at all.
//
// One failing object must not stop the sweep, so every file is attempted.
// Storage keys embed the secret access token and are therefore never logged;
// the secret id is logged instead.
func (s *CleanupService) purgeBlobs(ctx context.Context) ([]uuid.UUID, int, error) {
	blobs, err := s.secretRepo.ListFilesOfExpiredSecrets(ctx)
	if err != nil {
		return nil, 0, err
	}

	var kept []uuid.UUID
	purged := 0

	for _, b := range blobs {
		if err := deleteFileBlobs(ctx, s.store, b.StorageKey, b.ChunkCount); err != nil {
			slog.Error("failed to delete stored file of expired secret",
				"secret_id", b.SecretID, "error", err)
			kept = append(kept, b.SecretID)
			continue
		}
		purged++
	}

	return kept, purged, nil
}
