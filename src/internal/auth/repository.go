package auth

import (
	"context"
	"dZev1/character-gallery/internal/postgres/db"

	"github.com/google/uuid"
)

type Repository interface {
	SaveUser(ctx context.Context, exec db.DBTX, id uuid.UUID, username, email, password string) (*User, error)
	FindUserByUsername(ctx context.Context, exec db.DBTX, username string) (*User, error)
	FindUserByEmail(ctx context.Context, exec db.DBTX, email string) (*User, error)
}
