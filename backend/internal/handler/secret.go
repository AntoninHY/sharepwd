// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/AntoninHY/sharepwd/internal/config"
	"github.com/AntoninHY/sharepwd/internal/middleware"
	"github.com/AntoninHY/sharepwd/internal/model"
	"github.com/AntoninHY/sharepwd/internal/service"
)

type SecretHandler struct {
	service *service.SecretService
	nonces  *nonceStore
	cfg     *config.Config
}

type nonceEntry struct {
	ExpiresAt  time.Time
	IssuedAt   time.Time // Layer 1: grace period server-side
	IPHash     string    // Layer 3: IP binding
	PowPrefix  string    // Layer 2: PoW challenge prefix
	Difficulty uint8     // Layer 2: PoW difficulty
}

type nonceStore struct {
	mu      sync.RWMutex
	store   map[string]*nonceEntry
	ipCount map[string]int
	cfg     *config.Config
}

func newNonceStore(cfg *config.Config) *nonceStore {
	ns := &nonceStore{
		store:   make(map[string]*nonceEntry),
		ipCount: make(map[string]int),
		cfg:     cfg,
	}
	go ns.cleanup()
	return ns
}

const maxNonces = 100000

func (ns *nonceStore) generate(ipHash string) (string, *nonceEntry, error) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if len(ns.store) >= maxNonces {
		return "", nil, fmt.Errorf("nonce store is full")
	}

	if ns.ipCount[ipHash] >= ns.cfg.MaxNoncesPerIP {
		return "", nil, fmt.Errorf("too many active nonces for this IP")
	}

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", nil, err
	}
	nonce := hex.EncodeToString(b)

	pb := make([]byte, 16)
	if _, err := rand.Read(pb); err != nil {
		return "", nil, err
	}
	powPrefix := hex.EncodeToString(pb)

	now := time.Now()
	entry := &nonceEntry{
		ExpiresAt:  now.Add(ns.cfg.ChallengeTTL),
		IssuedAt:   now,
		IPHash:     ipHash,
		PowPrefix:  powPrefix,
		Difficulty: ns.cfg.PowDifficulty,
	}

	ns.store[nonce] = entry
	ns.ipCount[ipHash]++

	return nonce, entry, nil
}

func (ns *nonceStore) validate(nonce string, ipHash string) (*nonceEntry, string) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	entry, ok := ns.store[nonce]
	if !ok {
		return nil, "invalid challenge nonce"
	}

	// Consume nonce (single-use)
	delete(ns.store, nonce)
	ns.ipCount[entry.IPHash]--
	if ns.ipCount[entry.IPHash] <= 0 {
		delete(ns.ipCount, entry.IPHash)
	}

	if time.Now().After(entry.ExpiresAt) {
		return nil, "challenge nonce expired"
	}

	if entry.IPHash != ipHash {
		return nil, "IP mismatch on challenge nonce"
	}

	return entry, ""
}

func (ns *nonceStore) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		ns.mu.Lock()
		now := time.Now()
		for k, entry := range ns.store {
			if now.After(entry.ExpiresAt) {
				ns.ipCount[entry.IPHash]--
				if ns.ipCount[entry.IPHash] <= 0 {
					delete(ns.ipCount, entry.IPHash)
				}
				delete(ns.store, k)
			}
		}
		ns.mu.Unlock()
	}
}

func NewSecretHandler(svc *service.SecretService, cfg *config.Config) *SecretHandler {
	return &SecretHandler{
		service: svc,
		nonces:  newNonceStore(cfg),
		cfg:     cfg,
	}
}

func hashIP(ip string) string {
	h := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(h[:])
}

func (h *SecretHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 512*1024)

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

	ip := extractIP(r)
	ipH := hashIP(ip)

	nonce, entry, err := h.nonces.generate(ipH)
	if err != nil {
		slog.Warn("nonce generation failed", "error", err, "ip_hash", ipH[:8])
		writeError(w, http.StatusTooManyRequests, "too many requests, please wait")
		return
	}

	type metaWithChallenge struct {
		*model.SecretMetadata
		ChallengeNonce string `json:"challenge_nonce"`
		PowChallenge   string `json:"pow_challenge"`
		PowDifficulty  uint8  `json:"pow_difficulty"`
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metaWithChallenge{
		SecretMetadata: meta,
		ChallengeNonce: nonce,
		PowChallenge:   entry.PowPrefix,
		PowDifficulty:  entry.Difficulty,
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

	ip := extractIP(r)
	ipH := hashIP(ip)

	// Check if request comes from an API key (skip behavioral/env checks)
	hasAPIKey := middleware.GetAPIKey(r.Context()) != nil

	// Layer 3: Validate nonce (single-use, IP-bound, not expired)
	entry, errMsg := h.nonces.validate(req.ChallengeNonce, ipH)
	if entry == nil {
		slog.Warn("nonce validation failed", "reason", errMsg, "ip_hash", ipH[:8])
		writeError(w, http.StatusForbidden, "invalid or expired challenge nonce")
		return
	}

	// Layer 1: Grace period (server-enforced)
	elapsed := time.Since(entry.IssuedAt)
	if elapsed < h.cfg.ChallengeMinSolveTime {
		slog.Warn("grace period not met",
			"elapsed_ms", elapsed.Milliseconds(),
			"required_ms", h.cfg.ChallengeMinSolveTime.Milliseconds(),
			"ip_hash", ipH[:8],
		)
		writeError(w, http.StatusForbidden, "challenge solved too quickly")
		return
	}

	// Layer 2: Proof-of-Work verification
	if req.PowSolution == 0 && h.cfg.DefenseStrictMode {
		writeError(w, http.StatusForbidden, "proof of work required")
		return
	}
	if req.PowSolution > 0 {
		if !verifyPoW(entry.PowPrefix, req.PowSolution, entry.Difficulty) {
			slog.Warn("PoW verification failed", "ip_hash", ipH[:8])
			writeError(w, http.StatusForbidden, "invalid proof of work")
			return
		}
	} else if !h.cfg.DefenseStrictMode {
		slog.Info("PoW skipped (non-strict mode)", "ip_hash", ipH[:8])
	}

	// Layer 4: Behavioral analysis (skip for API key holders)
	if !hasAPIKey {
		if req.BehavioralProof == "" && h.cfg.DefenseStrictMode {
			writeError(w, http.StatusForbidden, "behavioral proof required")
			return
		}
		if req.BehavioralProof != "" {
			score, isTouch := scoreBehavioral(req.BehavioralProof)
			threshold := h.cfg.BehavioralMinScore
			if isTouch {
				threshold = 10
			}
			if score < threshold {
				slog.Warn("behavioral score too low",
					"score", score,
					"threshold", threshold,
					"is_touch", isTouch,
					"ip_hash", ipH[:8],
				)
				if h.cfg.DefenseStrictMode {
					writeError(w, http.StatusForbidden, "behavioral verification failed")
					return
				}
			}
		} else if !h.cfg.DefenseStrictMode {
			slog.Info("behavioral proof skipped (non-strict mode)", "ip_hash", ipH[:8])
		}
	}

	// Layer 5: Environment fingerprint (skip for API key holders)
	if !hasAPIKey {
		if req.EnvFingerprint == "" && h.cfg.DefenseStrictMode {
			writeError(w, http.StatusForbidden, "environment fingerprint required")
			return
		}
		if req.EnvFingerprint != "" {
			envScore := scoreEnvFingerprint(req.EnvFingerprint)
			if envScore < h.cfg.EnvMinScore {
				slog.Warn("env fingerprint score too low",
					"score", envScore,
					"threshold", h.cfg.EnvMinScore,
					"ip_hash", ipH[:8],
				)
				if h.cfg.DefenseStrictMode {
					writeError(w, http.StatusForbidden, "environment verification failed")
					return
				}
			}
		} else if !h.cfg.DefenseStrictMode {
			slog.Info("env fingerprint skipped (non-strict mode)", "ip_hash", ipH[:8])
		}
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
