// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/AntoninHY/sharepwd/internal/model"
	"github.com/AntoninHY/sharepwd/internal/repository"
)

type APIKeyHandler struct {
	repo *repository.APIKeyRepository
}

func NewAPIKeyHandler(repo *repository.APIKeyRepository) *APIKeyHandler {
	return &APIKeyHandler{repo: repo}
}

func (h *APIKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	rawKey := make([]byte, 32)
	if _, err := rand.Read(rawKey); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate key")
		return
	}

	keyStr := "spwd_" + hex.EncodeToString(rawKey)
	prefix := keyStr[:9]

	h2 := sha256.Sum256([]byte(keyStr))
	keyHash := hex.EncodeToString(h2[:])

	rateLimit := 60
	if req.RateLimit != nil {
		rateLimit = *req.RateLimit
	}

	apiKey := &model.APIKey{
		ID:        uuid.New(),
		KeyPrefix: prefix,
		KeyHash:   keyHash,
		Name:      req.Name,
		RateLimit: rateLimit,
		IsActive:  true,
		ExpiresAt: req.ExpiresAt,
	}

	if err := h.repo.Create(r.Context(), apiKey); err != nil {
		slog.Error("failed to create api key", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create api key")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model.CreateAPIKeyResponse{
		ID:        apiKey.ID,
		Key:       keyStr,
		KeyPrefix: prefix,
		Name:      apiKey.Name,
		RateLimit: apiKey.RateLimit,
	})
}

func (h *APIKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	keys, err := h.repo.List(r.Context())
	if err != nil {
		slog.Error("failed to list api keys", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list api keys")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}

func (h *APIKeyHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.repo.Revoke(r.Context(), id); err != nil {
		slog.Error("failed to revoke api key", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to revoke api key")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
