package server

import (
	"fmt"
	"log/slog"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jizo-hr/sharepwd/internal/config"
	"github.com/jizo-hr/sharepwd/internal/handler"
	"github.com/jizo-hr/sharepwd/internal/middleware"
	"github.com/jizo-hr/sharepwd/internal/repository"
	"github.com/jizo-hr/sharepwd/internal/service"
	"github.com/jizo-hr/sharepwd/internal/storage"
)

func NewRouter(db *pgxpool.Pool, cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.SecurityHeaders(cfg.CORSOrigins))
	r.Use(middleware.RateLimit(cfg.RateLimitPublic))

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
	secretH := handler.NewSecretHandler(secretSvc, cfg)
	apiKeyH := handler.NewAPIKeyHandler(apiKeyRepo)
	fileH := handler.NewFileHandler(secretSvc, fileRepo, store, cfg)

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
