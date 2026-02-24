package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"crypto/rand"
	"encoding/hex"

	"github.com/go-chi/chi/v5"
	"github.com/jizo-hr/sharepwd/internal/model"
	"github.com/jizo-hr/sharepwd/internal/service"
)

type SecretHandler struct {
	service *service.SecretService
	nonces  *nonceStore
}

type nonceStore struct {
	mu     sync.RWMutex
	store  map[string]time.Time
}

func newNonceStore() *nonceStore {
	ns := &nonceStore{store: make(map[string]time.Time)}
	go ns.cleanup()
	return ns
}

const maxNonces = 100000

func (ns *nonceStore) generate() (string, error) {
	ns.mu.Lock()
	if len(ns.store) >= maxNonces {
		ns.mu.Unlock()
		return "", fmt.Errorf("nonce store is full")
	}
	ns.mu.Unlock()

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	nonce := hex.EncodeToString(b)
	ns.mu.Lock()
	ns.store[nonce] = time.Now().Add(10 * time.Minute)
	ns.mu.Unlock()
	return nonce, nil
}

func (ns *nonceStore) validate(nonce string) bool {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	exp, ok := ns.store[nonce]
	if !ok {
		return false
	}
	delete(ns.store, nonce)
	return time.Now().Before(exp)
}

func (ns *nonceStore) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		ns.mu.Lock()
		now := time.Now()
		for k, exp := range ns.store {
			if now.After(exp) {
				delete(ns.store, k)
			}
		}
		ns.mu.Unlock()
	}
}

func NewSecretHandler(svc *service.SecretService) *SecretHandler {
	return &SecretHandler{
		service: svc,
		nonces:  newNonceStore(),
	}
}

func (h *SecretHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 512*1024) // 512KB max for text secrets

	var req model.CreateSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.EncryptedData == "" || req.IV == "" {
		writeError(w, http.StatusBadRequest, "encrypted_data and iv are required")
		return
	}

	ip := extractIP(r)
	ua := r.UserAgent()

	resp, err := h.service.Create(r.Context(), &req, ip, ua)
	if err != nil {
		slog.Error("failed to create secret", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create secret")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *SecretHandler) GetMetadata(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "token is required")
		return
	}

	meta, err := h.service.GetMetadata(r.Context(), token)
	if err != nil {
		slog.Error("failed to get secret metadata", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if meta == nil {
		writeError(w, http.StatusNotFound, "secret not found")
		return
	}

	nonce, err := h.nonces.generate()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	type metaWithNonce struct {
		*model.SecretMetadata
		ChallengeNonce string `json:"challenge_nonce"`
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metaWithNonce{
		SecretMetadata: meta,
		ChallengeNonce: nonce,
	})
}

func (h *SecretHandler) Reveal(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "token is required")
		return
	}

	var req model.RevealSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if !h.nonces.validate(req.ChallengeNonce) {
		writeError(w, http.StatusForbidden, "invalid or expired challenge nonce")
		return
	}

	resp, err := h.service.Reveal(r.Context(), token)
	if err != nil {
		if strings.Contains(err.Error(), "expired") || strings.Contains(err.Error(), "max views") {
			writeError(w, http.StatusGone, err.Error())
			return
		}
		slog.Error("failed to reveal secret", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if resp == nil {
		writeError(w, http.StatusNotFound, "secret not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *SecretHandler) Delete(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "token is required")
		return
	}

	var req model.DeleteSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.CreatorToken == "" {
		writeError(w, http.StatusBadRequest, "creator_token is required")
		return
	}

	if err := h.service.Delete(r.Context(), token, req.CreatorToken); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "secret not found or invalid creator token")
			return
		}
		slog.Error("failed to delete secret", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	idx := strings.LastIndex(r.RemoteAddr, ":")
	if idx == -1 {
		return r.RemoteAddr
	}
	return r.RemoteAddr[:idx]
}
