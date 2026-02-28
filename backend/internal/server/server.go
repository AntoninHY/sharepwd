package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jizo-hr/sharepwd/internal/config"
	"github.com/jizo-hr/sharepwd/internal/repository"
	"github.com/jizo-hr/sharepwd/internal/service"
)

type Server struct {
	httpServer *http.Server
	cleanup    *service.CleanupService
	db         *pgxpool.Pool
}

func New(db *pgxpool.Pool, cfg *config.Config) *Server {
	router := NewRouter(db, cfg)

	secretRepo := repository.NewSecretRepository(db)
	cleanupSvc := service.NewCleanupService(secretRepo, cfg)

	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.ListenAddr,
			Handler:      router,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 60 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		cleanup: cleanupSvc,
		db:      db,
	}
}

func (s *Server) Start(ctx context.Context) error {
	go s.cleanup.Start(ctx)

	slog.Info("server starting", "addr", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("server shutting down")
	return s.httpServer.Shutdown(ctx)
}
