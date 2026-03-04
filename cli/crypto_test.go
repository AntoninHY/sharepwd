// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"golang.org/x/crypto/pbkdf2"
)

func TestEncryptDecryptKeyless(t *testing.T) {
	plaintext := []byte("Hello, SharePwd!")

	result, err := EncryptKeyless(plaintext)
	if err != nil {
		t.Fatalf("EncryptKeyless: %v", err)
	}

	if result.EncryptedData == "" || result.IV == "" || result.Key == "" {
		t.Fatal("EncryptKeyless returned empty fields")
	}

	decrypted, err := DecryptKeyless(result.EncryptedData, result.IV, result.Key)
	if err != nil {
		t.Fatalf("DecryptKeyless: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Fatalf("got %q, want %q", decrypted, plaintext)
	}
}

func TestEncryptDecryptPassphrase(t *testing.T) {
	plaintext := []byte("Secret with passphrase!")
	passphrase := "my-strong-pass"

	result, err := EncryptWithPassphrase(plaintext, passphrase)
	if err != nil {
		t.Fatalf("EncryptWithPassphrase: %v", err)
	}

	if result.EncryptedData == "" || result.IV == "" || result.Salt == "" {
		t.Fatal("EncryptWithPassphrase returned empty fields")
	}

	decrypted, err := DecryptWithPassphrase(result.EncryptedData, result.IV, result.Salt, passphrase)
	if err != nil {
		t.Fatalf("DecryptWithPassphrase: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Fatalf("got %q, want %q", decrypted, plaintext)
	}
}

func TestDecryptPassphraseWrongKey(t *testing.T) {
	plaintext := []byte("Wrong passphrase test")

	result, err := EncryptWithPassphrase(plaintext, "correct-pass")
	if err != nil {
		t.Fatalf("EncryptWithPassphrase: %v", err)
	}

	_, err = DecryptWithPassphrase(result.EncryptedData, result.IV, result.Salt, "wrong-pass")
	if err == nil {
		t.Fatal("expected error with wrong passphrase")
	}
}

func TestBase64Encoding(t *testing.T) {
	// Verify our encoding matches what the frontend produces
	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"1 byte", []byte{0x41}},
		{"2 bytes", []byte{0x41, 0x42}},
		{"3 bytes", []byte{0x41, 0x42, 0x43}},
		{"4 bytes", []byte{0x41, 0x42, 0x43, 0x44}},
		{"32 bytes key", make([]byte, 32)},
		{"12 bytes iv", make([]byte, 12)},
		{"16 bytes salt", make([]byte, 16)},
	}

	for _, tt := range tests {
		t.Run(tt.name+" StdEncoding", func(t *testing.T) {
			encoded := base64.StdEncoding.EncodeToString(tt.data)
			decoded, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if len(decoded) != len(tt.data) {
				t.Fatalf("length mismatch: got %d, want %d", len(decoded), len(tt.data))
			}
		})

		t.Run(tt.name+" RawURLEncoding", func(t *testing.T) {
			encoded := base64.RawURLEncoding.EncodeToString(tt.data)
			decoded, err := base64.RawURLEncoding.DecodeString(encoded)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if len(decoded) != len(tt.data) {
				t.Fatalf("length mismatch: got %d, want %d", len(decoded), len(tt.data))
			}
		})
	}
}

// TestCrossCompatKeyless verifies that data encrypted with our Go code
// uses the same format as the frontend (AES-256-GCM, base64 standard for
// data/iv, base64url for key).
func TestCrossCompatKeyless(t *testing.T) {
	plaintext := []byte("Cross-compat test 🔐")

	result, err := EncryptKeyless(plaintext)
	if err != nil {
		t.Fatalf("EncryptKeyless: %v", err)
	}

	// Manually verify the format
	keyBytes, err := base64.RawURLEncoding.DecodeString(result.Key)
	if err != nil {
		t.Fatalf("decode key: %v", err)
	}
	if len(keyBytes) != aesKeyLength {
		t.Fatalf("key length: got %d, want %d", len(keyBytes), aesKeyLength)
	}

	ivBytes, err := base64.StdEncoding.DecodeString(result.IV)
	if err != nil {
		t.Fatalf("decode iv: %v", err)
	}
	if len(ivBytes) != ivLength {
		t.Fatalf("iv length: got %d, want %d", len(ivBytes), ivLength)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(result.EncryptedData)
	if err != nil {
		t.Fatalf("decode ciphertext: %v", err)
	}

	// AES-GCM ciphertext = plaintext length + 16 bytes auth tag
	expectedLen := len(plaintext) + 16
	if len(ciphertext) != expectedLen {
		t.Fatalf("ciphertext length: got %d, want %d", len(ciphertext), expectedLen)
	}
}

// TestCrossCompatPassphrase verifies the PBKDF2 + AES-GCM format.
func TestCrossCompatPassphrase(t *testing.T) {
	plaintext := []byte("Passphrase cross-compat")
	passphrase := "test-passphrase"

	result, err := EncryptWithPassphrase(plaintext, passphrase)
	if err != nil {
		t.Fatalf("EncryptWithPassphrase: %v", err)
	}

	saltBytes, err := base64.StdEncoding.DecodeString(result.Salt)
	if err != nil {
		t.Fatalf("decode salt: %v", err)
	}
	if len(saltBytes) != saltLength {
		t.Fatalf("salt length: got %d, want %d", len(saltBytes), saltLength)
	}

	// Verify PBKDF2 derivation manually
	derivedKey := pbkdf2.Key([]byte(passphrase), saltBytes, pbkdf2Iterations, aesKeyLength, sha256.New)

	ivBytes, err := base64.StdEncoding.DecodeString(result.IV)
	if err != nil {
		t.Fatalf("decode iv: %v", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(result.EncryptedData)
	if err != nil {
		t.Fatalf("decode ciphertext: %v", err)
	}

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		t.Fatalf("aes: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("gcm: %v", err)
	}

	decrypted, err := gcm.Open(nil, ivBytes, ciphertext, nil)
	if err != nil {
		t.Fatalf("manual decrypt: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Fatalf("got %q, want %q", decrypted, plaintext)
	}
}
