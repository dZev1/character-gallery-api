package auth

import (
	"context"
)

type Repository interface {
	ValidateAPIKey(ctx context.Context, keyHash string) (bool, error)
	UpdateLastUsed(ctx context.Context, keyHash string) error
	CreateAPIKey(ctx context.Context, name string) (string, error)
}
