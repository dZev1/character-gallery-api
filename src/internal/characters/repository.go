package characters

import (
	"context"
	"dZev1/character-gallery/internal/postgres/db"
)

type Repository interface {
	SaveCharacter(ctx context.Context, exec db.DBTX, character *Character) (*Character, error)
	FindCharacter(ctx context.Context, exec db.DBTX, id CharacterID) (*Character, error)
	FindAllCharacters(ctx context.Context, exec db.DBTX, page int) ([]Character, uint64, error)
	UpdateCharacter(ctx context.Context, exec db.DBTX, character *Character) error
	DeleteCharacter(ctx context.Context, exec db.DBTX, id CharacterID) error
}
