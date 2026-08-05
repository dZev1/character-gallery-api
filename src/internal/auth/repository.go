package auth

import (
	"context"
	"dZev1/character-gallery/internal/postgres/db"

	"github.com/google/uuid"
)

type Repository interface {
	SaveUser(ctx context.Context, exec db.DBTX, id uuid.UUID, username, password string) (*User, error)
	FindUserByUsername(ctx context.Context, exec db.DBTX, username string) (*User, error)
}
