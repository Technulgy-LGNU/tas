package util

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateSessionToken Generate a random session token (e.g., 32-byte token)
func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32) // 32 bytes = 64 characters when hex-encoded
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
