// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/AntoninHY/sharepwd/internal/config"
	"github.com/AntoninHY/sharepwd/internal/handler"
	"github.com/AntoninHY/sharepwd/internal/middleware"
	"github.com/AntoninHY/sharepwd/internal/repository"
	"github.com/AntoninHY/sharepwd/internal/service"
	"github.com/AntoninHY/sharepwd/internal/storage"
)

func NewRouter(db *pgxpool.Pool, cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.SecurityHeaders(cfg.CORSOrigins))
	r.Use(middleware.RateLimit(cfg.RateLimitPublic))

	// Redis client
	rdsOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		slog.Error("failed to parse REDIS_URL", "error", err)
		panic(fmt.Sprintf("failed to parse REDIS_URL: %v", err))
	}
	rdb := redis.NewClient(rdsOpts)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		slog.Error("failed to connect to Redis", "error", err)
		panic(fmt.Sprintf("failed to connect to Redis: %v", err))
	}
	slog.Info("connected to Redis", "addr", rdsOpts.Addr)

	// Repositories
	secretRepo := repository.NewSecretRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	fileRepo := repository.NewFileRepository(db)

	// Services
	secretSvc := service.NewSecretService(secretRepo, cfg)

	// Storage backend
	store, err := initStorage(cfg)
	if err != nil {
		slog.Error("failed to initialize storage backend", "error", err)
		panic(fmt.Sprintf("failed to initialize storage backend: %v", err))
	}

	// Handlers
	healthH := handler.NewHealthHandler()
	secretH := handler.NewSecretHandler(secretSvc, rdb, cfg)
	apiKeyH := handler.NewAPIKeyHandler(apiKeyRepo)
	fileH := handler.NewFileHandler(secretSvc, fileRepo, store, cfg)
	adminH := handler.NewAdminHandler(apiKeyRepo)

	// Routes
	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", healthH.Health)

		// File upload/download routes (registered before wildcard {token} routes)
		r.Post("/secrets/file", fileH.InitUpload)
		r.Put("/secrets/file/{id}/chunk/{n}", fileH.UploadChunk)
		r.Post("/secrets/file/{id}/complete", fileH.CompleteUpload)
		r.Get("/secrets/file/{id}/chunk/{n}", fileH.DownloadChunk)

		// Public secret routes
		r.Post("/secrets", secretH.Create)
		r.With(middleware.BotDetect, middleware.RateLimit(cfg.MetadataRateLimit)).Get("/secrets/{token}", secretH.GetMetadata)
		r.With(middleware.BotDetect).Post("/secrets/{token}/reveal", secretH.Reveal)
		r.Delete("/secrets/{token}", secretH.Delete)

		// Authenticated routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.APIKeyAuth(apiKeyRepo))
			r.Post("/api-keys", apiKeyH.Create)
			r.Get("/api-keys", apiKeyH.List)
			r.Delete("/api-keys/{id}", apiKeyH.Revoke)
		})

		// Admin routes (only if ADMIN_SECRET is configured)
		if cfg.AdminSecret != "" {
			r.Group(func(r chi.Router) {
				r.Use(middleware.AdminAuth(cfg.AdminSecret))
				r.Use(middleware.RateLimit(3))
				r.Post("/admin/api-keys", adminH.CreateAPIKey)
			})
			slog.Info("admin endpoint enabled", "path", "/v1/admin/api-keys")
		}
	})

	return r
}

func initStorage(cfg *config.Config) (storage.Storage, error) {
	switch cfg.StorageBackend {
	case "local":
		slog.Info("using local filesystem storage", "path", cfg.LocalStorePath)
		return storage.NewLocalStorage(cfg.LocalStorePath)
	case "s3":
		slog.Info("using S3/MinIO storage", "endpoint", cfg.S3Endpoint, "bucket", cfg.S3Bucket)
		return storage.NewS3Storage(
			cfg.S3Endpoint,
			cfg.S3AccessKey,
			cfg.S3SecretKey,
			cfg.S3Bucket,
			cfg.S3UseSSL,
			cfg.S3Region,
		)
	default:
		return nil, fmt.Errorf("unsupported storage backend: %s", cfg.StorageBackend)
	}
}
