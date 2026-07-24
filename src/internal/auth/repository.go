package auth

import (
	"context"
	"dZev1/character-gallery/internal/postgres/db"
)

type Repository interface {
	SaveUser(ctx context.Context, exec db.DBTX, username, password string) (*User, error)
	FindUserByUsername(ctx context.Context, exec db.DBTX, username string) (*User, error)
}
