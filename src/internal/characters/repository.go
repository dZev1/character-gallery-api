package characters

import (
	"context"
	"dZev1/character-gallery/internal/postgres/db"

	"github.com/google/uuid"
)

type Repository interface {
	SaveCharacter(ctx context.Context, exec db.DBTX, character *Character, ownerID uuid.UUID) (*Character, error)
	FindCharacter(ctx context.Context, exec db.DBTX, id CharacterID) (*Character, error)
	FindAllCharacters(ctx context.Context, exec db.DBTX, limit, offset int) ([]Character, uint64, error)
	UpdateCharacter(ctx context.Context, exec db.DBTX, character *Character, ownerID uuid.UUID) (*Character, error)
	DeleteCharacter(ctx context.Context, exec db.DBTX, id CharacterID, ownerID uuid.UUID) error
}
