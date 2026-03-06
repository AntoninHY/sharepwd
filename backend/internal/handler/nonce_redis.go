// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/AntoninHY/sharepwd/internal/config"
)

// nonceEntry represents a single challenge nonce with all associated metadata.
type nonceEntry struct {
	ExpiresAt  time.Time `json:"expires_at"`
	IssuedAt   time.Time `json:"issued_at"`
	IPHash     string    `json:"ip_hash"`
	PowPrefix  string    `json:"pow_prefix"`
	Difficulty uint8     `json:"difficulty"`
	HMACKey    string    `json:"hmac_key"`
}

const (
	nonceKeyPrefix = "nonce:"
	ipKeyPrefix    = "nonce_ip:"
)

// nonceStore provides Redis-backed nonce storage for anti-bot challenges.
type nonceStore struct {
	rdb *redis.Client
	cfg *config.Config
}

func newNonceStore(rdb *redis.Client, cfg *config.Config) *nonceStore {
	return &nonceStore{
		rdb: rdb,
		cfg: cfg,
	}
}

func nonceKey(nonce string) string {
	return nonceKeyPrefix + nonce
}

func ipKey(ipHash string) string {
	return ipKeyPrefix + ipHash
}

// generateScript atomically checks IP limit, stores nonce, and increments IP counter.
// KEYS[1] = nonce key, KEYS[2] = IP counter key
// ARGV[1] = max nonces per IP, ARGV[2] = TTL in milliseconds, ARGV[3] = JSON nonce entry
var generateScript = redis.NewScript(`
	local ip_count = tonumber(redis.call("GET", KEYS[2]) or "0")
	if ip_count >= tonumber(ARGV[1]) then
		return -1
	end
	local ok = redis.call("SET", KEYS[1], ARGV[3], "NX", "PX", ARGV[2])
	if not ok then
		return -2
	end
	local new_count = redis.call("INCR", KEYS[2])
	if new_count == 1 then
		redis.call("PEXPIRE", KEYS[2], ARGV[2])
	end
	return 1
`)

func (ns *nonceStore) generate(ipHash string) (string, *nonceEntry, error) {
	ctx := context.Background()
	ttl := ns.cfg.ChallengeTTL
	ttlMs := int64(ttl.Milliseconds())

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	nonce := hex.EncodeToString(b)

	pb := make([]byte, 16)
	if _, err := rand.Read(pb); err != nil {
		return "", nil, fmt.Errorf("failed to generate pow prefix: %w", err)
	}
	powPrefix := hex.EncodeToString(pb)

	hmacKey, err := generateHMACKey()
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate HMAC key: %w", err)
	}

	now := time.Now()
	entry := &nonceEntry{
		ExpiresAt:  now.Add(ttl),
		IssuedAt:   now,
		IPHash:     ipHash,
		PowPrefix:  powPrefix,
		Difficulty: ns.cfg.PowDifficulty,
		HMACKey:    hmacKey,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal nonce entry: %w", err)
	}

	nk := nonceKey(nonce)
	ipk := ipKey(ipHash)

	result, err := generateScript.Run(ctx, ns.rdb, []string{nk, ipk}, ns.cfg.MaxNoncesPerIP, ttlMs, string(data)).Int64()
	if err != nil {
		return "", nil, fmt.Errorf("redis: generate script failed: %w", err)
	}

	switch result {
	case 1:
		// Success
	case -1:
		return "", nil, fmt.Errorf("too many active nonces for this IP")
	case -2:
		// Extremely unlikely collision; retry once
		b2 := make([]byte, 16)
		if _, err := rand.Read(b2); err != nil {
			return "", nil, fmt.Errorf("failed to generate nonce (retry): %w", err)
		}
		nonce = hex.EncodeToString(b2)
		nk = nonceKey(nonce)
		result2, err := generateScript.Run(ctx, ns.rdb, []string{nk, ipk}, ns.cfg.MaxNoncesPerIP, ttlMs, string(data)).Int64()
		if err != nil {
			return "", nil, fmt.Errorf("redis: generate script retry failed: %w", err)
		}
		if result2 != 1 {
			return "", nil, fmt.Errorf("nonce generation failed after retry (code: %d)", result2)
		}
	default:
		return "", nil, fmt.Errorf("redis: unexpected generate result: %d", result)
	}

	return nonce, entry, nil
}

// validateScript atomically gets and deletes a nonce (single-use consumption).
var validateScript = redis.NewScript(`
	local data = redis.call("GET", KEYS[1])
	if not data then
		return false
	end
	redis.call("DEL", KEYS[1])
	return data
`)

// decrIPScript decrements the IP counter and removes the key if it reaches zero.
var decrIPScript = redis.NewScript(`
	local count = redis.call("DECR", KEYS[1])
	if count <= 0 then
		redis.call("DEL", KEYS[1])
	end
	return count
`)

func (ns *nonceStore) validate(nonce string, ipHash string) (*nonceEntry, string) {
	ctx := context.Background()
	nk := nonceKey(nonce)

	result, err := validateScript.Run(ctx, ns.rdb, []string{nk}).Result()
	if err == redis.Nil || result == nil {
		return nil, "invalid challenge nonce"
	}
	if err != nil {
		slog.Error("redis: failed to validate nonce", "error", err)
		return nil, "invalid challenge nonce"
	}

	data, ok := result.(string)
	if !ok || data == "" {
		return nil, "invalid challenge nonce"
	}

	var entry nonceEntry
	if err := json.Unmarshal([]byte(data), &entry); err != nil {
		slog.Error("redis: failed to unmarshal nonce entry", "error", err)
		return nil, "invalid challenge nonce"
	}

	// Decrement IP counter (nonce consumed regardless of validation outcome)
	ipk := ipKey(entry.IPHash)
	if err := decrIPScript.Run(ctx, ns.rdb, []string{ipk}).Err(); err != nil && err != redis.Nil {
		slog.Error("redis: failed to decrement IP counter", "error", err, "ip_key", ipk)
	}

	if time.Now().After(entry.ExpiresAt) {
		return nil, "challenge nonce expired"
	}

	if entry.IPHash != ipHash {
		return nil, "IP mismatch on challenge nonce"
	}

	return &entry, ""
}
