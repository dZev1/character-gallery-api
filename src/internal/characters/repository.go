package characters

import "context"

type Repository interface {
	SaveCharacter(ctx context.Context, character *Character) error
	FindCharacter(ctx context.Context, id CharacterID) (*Character, error)
	FindAllCharacters(ctx context.Context, page int) ([]Character, uint64, error)
	UpdateCharacter(ctx context.Context, character *Character) error
	DeleteCharacter(ctx context.Context, id CharacterID) error
}
