package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/jizo-hr/sharepwd/internal/model"
	"github.com/jizo-hr/sharepwd/internal/repository"
)

// AdminHandler exposes admin-only endpoints for managing API keys.
type AdminHandler struct {
	repo *repository.APIKeyRepository
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(repo *repository.APIKeyRepository) *AdminHandler {
	return &AdminHandler{repo: repo}
}

// CreateAPIKey generates a new API key and persists it.
// The raw key is returned only once in the response; only its hash is stored.
func (h *AdminHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	var req model.CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	// Generate random API key
	rawKey := make([]byte, 32)
	if _, err := rand.Read(rawKey); err != nil {
		slog.Error("failed to generate random key", "error", err)
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
