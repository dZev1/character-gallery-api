package characters

import (
	"context"
	"dZev1/character-gallery/internal/postgres/db"

	"github.com/google/uuid"
)

type Repository interface {
	SaveCharacter(ctx context.Context, exec db.DBTX, character *Character, ownerID uuid.UUID) (*Character, error)
	FindCharacter(ctx context.Context, exec db.DBTX, id ID) (*Character, error)
	FindAllCharacters(ctx context.Context, exec db.DBTX, limit, offset int) ([]Character, uint64, error)
	FindCharactersByOwner(ctx context.Context, exec db.DBTX, ownerID uuid.UUID, limit, offset int) ([]Character, uint64, error)
	UpdateCharacter(ctx context.Context, exec db.DBTX, character *Character, ownerID uuid.UUID) (*Character, error)
	UpdateCharacterHP(ctx context.Context, exec db.DBTX, id ID, hpCurrent uint8) (*Character, error)
	HealCharacter(ctx context.Context, exec db.DBTX, id ID, healAmount uint8) (*Character, error)
	UpdateCharacterProgress(ctx context.Context, exec db.DBTX, id ID, level uint8, xp uint64, hpMax, hpCurrent uint8) (*Character, error)
	PrestigeCharacter(ctx context.Context, exec db.DBTX, id ID, ownerID uuid.UUID, newHpMax uint8) (*Character, error)
	DeleteCharacter(ctx context.Context, exec db.DBTX, id ID, ownerID uuid.UUID) error
}
