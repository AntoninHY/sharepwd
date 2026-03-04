// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	pbkdf2Iterations = 600_000
	aesKeyLength     = 32 // 256 bits
	ivLength         = 12 // 96 bits, standard GCM nonce
	saltLength       = 16 // 128 bits
)

// EncryptResult holds keyless encryption output.
type EncryptResult struct {
	EncryptedData string // Base64 standard
	IV            string // Base64 standard
	Key           string // Base64URL (no padding) — goes in URL fragment
}

// EncryptWithPassphraseResult holds passphrase-based encryption output.
type EncryptWithPassphraseResult struct {
	EncryptedData string // Base64 standard
	IV            string // Base64 standard
	Salt          string // Base64 standard
}

// EncryptKeyless encrypts plaintext with a random AES-256-GCM key.
// Compatible with frontend encryptText().
func EncryptKeyless(plaintext []byte) (*EncryptResult, error) {
	key := make([]byte, aesKeyLength)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}

	iv := make([]byte, ivLength)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("generate iv: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}

	ciphertext := gcm.Seal(nil, iv, plaintext, nil)

	return &EncryptResult{
		EncryptedData: base64.StdEncoding.EncodeToString(ciphertext),
		IV:            base64.StdEncoding.EncodeToString(iv),
		Key:           base64.RawURLEncoding.EncodeToString(key),
	}, nil
}

// DecryptKeyless decrypts data encrypted with EncryptKeyless.
// Compatible with frontend decryptText().
func DecryptKeyless(encryptedDataB64, ivB64, keyB64URL string) ([]byte, error) {
	key, err := base64.RawURLEncoding.DecodeString(keyB64URL)
	if err != nil {
		return nil, fmt.Errorf("decode key: %w", err)
	}

	iv, err := base64.StdEncoding.DecodeString(ivB64)
	if err != nil {
		return nil, fmt.Errorf("decode iv: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedDataB64)
	if err != nil {
		return nil, fmt.Errorf("decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}

	plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	return plaintext, nil
}

// EncryptWithPassphrase encrypts plaintext using PBKDF2-derived key.
// Compatible with frontend encryptWithPassphrase().
func EncryptWithPassphrase(plaintext []byte, passphrase string) (*EncryptWithPassphraseResult, error) {
	salt := make([]byte, saltLength)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("generate salt: %w", err)
	}

	key := pbkdf2.Key([]byte(passphrase), salt, pbkdf2Iterations, aesKeyLength, sha256.New)

	iv := make([]byte, ivLength)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("generate iv: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}

	ciphertext := gcm.Seal(nil, iv, plaintext, nil)

	return &EncryptWithPassphraseResult{
		EncryptedData: base64.StdEncoding.EncodeToString(ciphertext),
		IV:            base64.StdEncoding.EncodeToString(iv),
		Salt:          base64.StdEncoding.EncodeToString(salt),
	}, nil
}

// DecryptWithPassphrase decrypts data encrypted with EncryptWithPassphrase.
// Compatible with frontend decryptWithPassphrase().
func DecryptWithPassphrase(encryptedDataB64, ivB64, saltB64, passphrase string) ([]byte, error) {
	salt, err := base64.StdEncoding.DecodeString(saltB64)
	if err != nil {
		return nil, fmt.Errorf("decode salt: %w", err)
	}

	key := pbkdf2.Key([]byte(passphrase), salt, pbkdf2Iterations, aesKeyLength, sha256.New)

	iv, err := base64.StdEncoding.DecodeString(ivB64)
	if err != nil {
		return nil, fmt.Errorf("decode iv: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedDataB64)
	if err != nil {
		return nil, fmt.Errorf("decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}

	plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	return plaintext, nil
}
