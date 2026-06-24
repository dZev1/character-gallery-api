package postgres

import (
	"context"
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/postgres/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

var _ characters.Repository = (*characterRepo)(nil)

type characterRepo struct {
	pool *pgxpool.Pool
}

func NewCharacterRepo(pool *pgxpool.Pool) *characterRepo {
	return &characterRepo{pool: pool}
}

func (c *characterRepo) q(exec db.DBTX) *db.Queries {
	return db.New(exec)
}

func (c *characterRepo) SaveCharacter(ctx context.Context, exec db.DBTX, character *characters.Character) (*characters.Character, error) {
	q := c.q(exec)

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
	q := c.q(exec)

	dbChar, err := q.SelectCharacter(ctx, int64(id))
	if err != nil {
		return nil, err
	}

	dbStats, err := q.SelectCharacterStats(ctx, dbChar.ID)
	if err != nil {
		return nil, err
	}

	dbCust, err := q.SelectCharacterCustomization(ctx, dbChar.ID)
	if err != nil {
		return nil, err
	}

	return &characters.Character{
		ID:       characters.CharacterID(dbChar.ID),
		Name:     dbChar.Name,
		BodyType: characters.BodyType(dbChar.BodyType),
		Species:  characters.Species(dbChar.Species),
		Class:    characters.Class(dbChar.Class),
		Stats: &characters.Stats{
			ID:           characters.CharacterID(dbStats.CharacterID),
			Strength:     uint8(dbStats.Strength),
			Dexterity:    uint8(dbStats.Dexterity),
			Constitution: uint8(dbStats.Constitution),
			Intelligence: uint8(dbStats.Intelligence),
			Wisdom:       uint8(dbStats.Wisdom),
			Charisma:     uint8(dbStats.Charisma),
		},
		Customization: &characters.Customization{
			ID:    characters.CharacterID(dbCust.CharacterID),
			Hair:  uint8(dbCust.Hair),
			Face:  uint8(dbCust.Face),
			Shirt: uint8(dbCust.Shirt),
			Pants: uint8(dbCust.Pants),
			Shoes: uint8(dbCust.Shoes),
		},
	}, nil
}

func (c *characterRepo) FindAllCharacters(ctx context.Context, exec db.DBTX, page int) ([]characters.Character, uint64, error) {
	q := c.q(exec)

	limit := int32(20)
	offset := int32((page - 1) * int(limit))

	dbChars, err := q.SelectAllCharacters(ctx, db.SelectAllCharactersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, 0, err
	}

	count, err := q.CountAllCharacters(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]characters.Character, len(dbChars))
	for i, dbChar := range dbChars {
		result[i] = characters.Character{
			ID:       characters.CharacterID(dbChar.ID),
			Name:     dbChar.Name,
			BodyType: characters.BodyType(dbChar.BodyType),
			Species:  characters.Species(dbChar.Species),
			Class:    characters.Class(dbChar.Class),
		}
	}

	return result, uint64(count), nil
}

func (c *characterRepo) UpdateCharacter(ctx context.Context, exec db.DBTX, character *characters.Character) (*characters.Character, error) {
	q := c.q(exec)

	_, err := q.UpdateCharacter(ctx, db.UpdateCharacterParams{
		ID:       int64(character.ID),
		Name:     character.Name,
		BodyType: string(character.BodyType),
		Species:  string(character.Species),
		Class:    string(character.Class),
	})
	if err != nil {
		return nil, err
	}

	_, err = q.UpdateStats(ctx, db.UpdateStatsParams{
		CharacterID:  int64(character.ID),
		Strength:     int16(character.Stats.Strength),
		Dexterity:    int16(character.Stats.Dexterity),
		Constitution: int16(character.Stats.Constitution),
		Intelligence: int16(character.Stats.Intelligence),
		Wisdom:       int16(character.Stats.Wisdom),
		Charisma:     int16(character.Stats.Charisma),
	})
	if err != nil {
		return nil, err
	}

	_, err = q.UpdateCustomization(ctx, db.UpdateCustomizationParams{
		CharacterID: int64(character.ID),
		Hair:        int16(character.Customization.Hair),
		Face:        int16(character.Customization.Face),
		Shirt:       int16(character.Customization.Shirt),
		Pants:       int16(character.Customization.Pants),
		Shoes:       int16(character.Customization.Shoes),
	})
	if err != nil {
		return nil, err
	}

	return c.FindCharacter(ctx, exec, character.ID)
}

func (c *characterRepo) DeleteCharacter(ctx context.Context, exec db.DBTX, id characters.CharacterID) error {
	q := c.q(exec)
	return q.DeleteCharacter(ctx, int64(id))
}
