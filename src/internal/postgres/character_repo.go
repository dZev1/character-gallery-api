package postgres

import (
	"context"
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/postgres/db"
)

type characterRepo struct {
	db.Queries
}

func (c *characterRepo) SaveCharacter(ctx context.Context, exec db.DBTX, character *characters.Character) (*characters.Character, error) {
	q := db.New(exec)

	createCharacterParams := db.CreateCharacterParams{
		Name:     character.Name,
		BodyType: string(character.BodyType),
		Species:  string(character.Species),
		Class:    string(character.Class),
	}

	char, err := q.CreateCharacter(ctx, createCharacterParams)
	if err != nil {
		return nil, err
	}

	character.ID = characters.CharacterID(char.ID)

	createCustomizationParams := db.CreateCustomizationParams{
		CharacterID: char.ID,
		Hair:        int16(character.Customization.Hair),
		Face:        int16(character.Customization.Face),
		Shirt:       int16(character.Customization.Shirt),
		Pants:       int16(character.Customization.Pants),
		Shoes:       int16(character.Customization.Shoes),
	}

	_, err = q.CreateCustomization(ctx, createCustomizationParams)
	if err != nil {
		return nil, err
	}

	character.Customization.ID = character.ID

	createStatParams := db.CreateStatsParams{
		CharacterID:  char.ID,
		Strength:     int16(character.Stats.Strength),
		Dexterity:    int16(character.Stats.Dexterity),
		Constitution: int16(character.Stats.Constitution),
		Intelligence: int16(character.Stats.Intelligence),
		Wisdom:       int16(character.Stats.Wisdom),
		Charisma:     int16(character.Stats.Charisma),
	}

	_, err = q.CreateStats(ctx, createStatParams)
	if err != nil {
		return nil, err
	}

	character.Stats.ID = character.ID

	return character, nil
}

func (c *characterRepo) FindCharacter(ctx context.Context, exec db.DBTX, id characters.CharacterID) (*characters.Character, error) {
	char, err := c.FindCharacter(ctx, exec, id)
	if err != nil {
		return nil, err
	}
	return char, nil
}

func (c *characterRepo) FindAllCharacters(ctx context.Context, exec db.DBTX, page int) ([]characters.Character, uint64, error) {
	chars, count, err := c.FindAllCharacters(ctx, exec, page)
	if err != nil {
		return nil, -1, err
	}
	return chars, count, nil
}

func (c *characterRepo) UpdateCharacter(ctx context.Context, exec db.DBTX, character *characters.Character) error {
	//TODO implement me
	panic("implement me")
}

func (c *characterRepo) DeleteCharacter(ctx context.Context, exec db.DBTX, id characters.CharacterID) error {
	//TODO implement me
	panic("implement me")
}
