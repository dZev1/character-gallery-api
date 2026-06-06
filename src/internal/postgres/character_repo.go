package postgres

import (
	"context"
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/postgres/db"
)

type characterRepo struct {
	db.Queries
}

func (c *characterRepo) SaveCharacter(ctx context.Context, character *characters.Character) error {
	//TODO implement me
	panic("implement me")
}

func (c *characterRepo) FindCharacter(ctx context.Context, id characters.CharacterID) (*characters.Character, error) {
	//TODO implement me
	panic("implement me")
}

func (c *characterRepo) FindAllCharacters(ctx context.Context, page int) ([]characters.Character, uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (c *characterRepo) UpdateCharacter(ctx context.Context, character *characters.Character) error {
	//TODO implement me
	panic("implement me")
}

func (c *characterRepo) DeleteCharacter(ctx context.Context, id characters.CharacterID) error {
	//TODO implement me
	panic("implement me")
}
