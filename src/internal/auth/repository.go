package auth

import (
	"context"
	"dZev1/character-gallery/internal/postgres/db"
)

type Repository interface {
	ValidateAPIKey(ctx context.Context, exec db.DBTX, keyHash string) (bool, error)
	UpdateLastUsed(ctx context.Context, exec db.DBTX, keyHash string) error
	CreateAPIKey(ctx context.Context, exec db.DBTX, name string) (string, error)
}
