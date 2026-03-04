package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/AntoninHY/sharepwd/internal/config"
	"github.com/AntoninHY/sharepwd/internal/repository"
)

type CleanupService struct {
	secretRepo *repository.SecretRepository
	config     *config.Config
}

func NewCleanupService(secretRepo *repository.SecretRepository, cfg *config.Config) *CleanupService {
	return &CleanupService{secretRepo: secretRepo, config: cfg}
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

	deleted, err := s.secretRepo.DeleteExpired(ctx)
	if err != nil {
		slog.Error("failed to delete expired secrets", "error", err)
		return
	}

	if marked > 0 || deleted > 0 {
		slog.Info("cleanup completed", "marked", marked, "deleted", deleted)
	}
}
