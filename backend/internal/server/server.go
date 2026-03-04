// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/AntoninHY/sharepwd/internal/config"
	"github.com/AntoninHY/sharepwd/internal/repository"
	"github.com/AntoninHY/sharepwd/internal/service"
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
