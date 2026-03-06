// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

package handler

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// generateHMACKey returns 32 cryptographically random bytes as a hex string.
func generateHMACKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return hex.EncodeToString(key), nil
}

// verifyProofHMAC verifies that signature == HMAC-SHA256(key, nonce+proofPayload).
// The key and signature are hex-encoded. Returns true if valid.
func verifyProofHMAC(hmacKeyHex, nonce, proofPayload, signature string) bool {
	keyBytes, err := hex.DecodeString(hmacKeyHex)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, keyBytes)
	mac.Write([]byte(nonce))
	mac.Write([]byte(proofPayload))
	expected := mac.Sum(nil)

	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}

	return hmac.Equal(expected, sigBytes)
}
