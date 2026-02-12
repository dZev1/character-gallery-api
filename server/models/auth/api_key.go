package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"
)

type APIKey struct {
	ID          APIKeyID  `json:"id" db:"id"`
	KeyHash     string    `json:"key_hash" db:"key_hash"`
	Name        string    `json:"name" db:"name"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	LastUsedAt  time.Time `json:"last_used_at" db:"last_used_at"`
	IsActive    bool      `json:"is_active" db:"is_active"`
}

type APIKeyID uint64

func (id APIKeyID) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

func GenerateAPIKey() (keyHash string, rawKey string, err error) {
	bytes := make([]byte, 32)
	_, err = rand.Read(bytes)
	if err != nil {
		return "", "", err
	}

	rawKey = "dz_chars_" + hex.EncodeToString(bytes)

	hash := sha256.Sum256([]byte(rawKey))
	keyHash = hex.EncodeToString(hash[:])
	return keyHash, rawKey, nil
}

func HashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}