// SharePwd.io — Burn After Reading
// Zero-knowledge secret sharing. Client-side AES-256-GCM encryption, secrets self-destruct after viewing.
// Copyright (c) 2025-2026 Antonin HILY — CTO, Jizo AI

package handler

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
)

// verifyPoW checks that SHA-256(prefix + ":" + counter) has at least `difficulty` leading zero bits.
func verifyPoW(prefix string, counter uint64, difficulty uint8) bool {
	input := fmt.Sprintf("%s:%d", prefix, counter)
	hash := sha256.Sum256([]byte(input))
	return hasLeadingZeroBits(hash[:], difficulty)
}

func hasLeadingZeroBits(hash []byte, bits uint8) bool {
	fullBytes := bits / 8
	remaining := bits % 8

	for i := uint8(0); i < fullBytes; i++ {
		if hash[i] != 0 {
			return false
		}
	}

	if remaining > 0 {
		mask := byte(0xFF << (8 - remaining))
		if hash[fullBytes]&mask != 0 {
			return false
		}
	}

	return true
}

// behavioralProof is the compact proof sent by the frontend.
type behavioralProof struct {
	MovementCount int     `json:"mc"`
	MouseEntropy  float64 `json:"me"` // Shannon entropy * 1000
	VelVariance   float64 `json:"vv"`
	StraightRatio float64 `json:"sr"` // 0-100
	EventCount    int     `json:"ec"`
	TimeSpan      float64 `json:"ts"` // ms
	HasTouch      int     `json:"ht"` // 1 if touch device
	HasMouse      int     `json:"hm"` // 1 if mouse device
}

// scoreBehavioral scores behavioral proof. Returns (score 0-100, isTouch).
func scoreBehavioral(proofB64 string) (int, bool) {
	if proofB64 == "" {
		return 0, false
	}

	raw, err := base64.StdEncoding.DecodeString(proofB64)
	if err != nil {
		return 0, false
	}

	var p behavioralProof
	if err := json.Unmarshal(raw, &p); err != nil {
		return 0, false
	}

	score := 0
	entropy := p.MouseEntropy / 1000.0

	if p.MovementCount >= 3 {
		score += 20
	}
	if p.MovementCount >= 10 {
		score += 10
	}
	if entropy >= 1.0 {
		score += 15
	}
	if entropy >= 2.0 {
		score += 15
	}
	if p.StraightRatio < 80 {
		score += 15
	}
	if p.StraightRatio < 50 {
		score += 10
	}
	if p.VelVariance > 0 {
		score += 10
	}
	if p.TimeSpan >= 1000 {
		score += 5
	}

	isTouch := p.HasTouch == 1
	return score, isTouch
}

// envFingerprint is the environment fingerprint sent by the frontend.
type envFingerprint struct {
	Webdriver       bool    `json:"wd"`
	PluginCount     int     `json:"pc"`
	LanguageCount   int     `json:"lc"`
	ScreenWidth     int     `json:"sw"`
	ScreenHeight    int     `json:"sh"`
	ColorDepth      int     `json:"cd"`
	HardwareConcur  int     `json:"hc"`
	HasNotification bool    `json:"hn"`
	DeviceMemory    float64 `json:"dm"`
	PerfNavTiming   bool    `json:"pt"`
}

// scoreEnvFingerprint scores environment fingerprint. Starts at 50, deducts for red flags.
func scoreEnvFingerprint(fpB64 string) int {
	if fpB64 == "" {
		return 50
	}

	raw, err := base64.StdEncoding.DecodeString(fpB64)
	if err != nil {
		return 0
	}

	var fp envFingerprint
	if err := json.Unmarshal(raw, &fp); err != nil {
		return 0
	}

	score := 50

	if fp.Webdriver {
		score -= 50
	}
	if fp.PluginCount == 0 {
		score -= 20
	}
	if fp.LanguageCount == 0 {
		score -= 15
	}
	if fp.ScreenWidth == 0 && fp.ScreenHeight == 0 {
		score -= 20
	}
	if fp.ColorDepth < 15 && fp.ColorDepth > 0 {
		score -= 10
	}
	if fp.HardwareConcur == 0 {
		score -= 5
	}
	if fp.DeviceMemory == 0 {
		score -= 5
	}
	if !fp.PerfNavTiming {
		score -= 5
	}
	if !fp.HasNotification {
		score -= 5
	}

	return int(math.Max(0, float64(score)))
}
