package postgres

import (
	"context"
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/postgres/db"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func (c *characterRepo) SaveCharacter(ctx context.Context, exec db.DBTX, character *characters.Character, ownerID uuid.UUID) (*characters.Character, error) {
	q := c.q(exec)

	createCharacterParams := db.CreateCharacterParams{
		OwnerID:   ownerID,
		Name:      character.Name,
		BodyType:  string(character.BodyType),
		Species:   string(character.Species),
		Class:     string(character.Class),
		Level:     int32(character.Level),
		Xp:        int32(character.Xp),
		HpMax:     int32(character.HpMax),
		HpCurrent: int32(character.HpCurrent),
	}

	char, err := q.CreateCharacter(ctx, createCharacterParams)
	if err != nil {
		return nil, fmt.Errorf("create character: %w", err)
	}

	character.ID = characters.ID(char.ID)
	character.OwnerID = char.OwnerID
	character.Prestige = uint(char.Prestige)

	createCustomizationParams := db.CreateCustomizationParams{
		CharacterID: char.ID,
		Hair:        int16(character.Customization.Hair),
		Beard:       int16(character.Customization.Beard),
		SkinTone:    int16(character.Customization.SkinTone),
		HairColor:   int16(character.Customization.HairColor),
		Cloak:       int16(character.Customization.Cloak),
		AccentColor: int16(character.Customization.AccentColor),
	}

	_, err = q.CreateCustomization(ctx, createCustomizationParams)
	if err != nil {
		return nil, fmt.Errorf("create customization: %w", err)
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
		return nil, fmt.Errorf("create stats: %w", err)
	}

	character.Stats.ID = character.ID

	return character, nil
}

func (c *characterRepo) FindCharacter(ctx context.Context, exec db.DBTX, id characters.ID) (*characters.Character, error) {
	q := c.q(exec)

	dbChar, err := q.SelectCharacter(ctx, int64(id))
	if err != nil {
		return nil, fmt.Errorf("character %d: %w", id, characters.ErrNotFound)
	}

	dbStats, err := q.SelectCharacterStats(ctx, dbChar.ID)
	if err != nil {
		return nil, fmt.Errorf("stats for character %d: %w", id, err)
	}

	dbCust, err := q.SelectCharacterCustomization(ctx, dbChar.ID)
	if err != nil {
		return nil, fmt.Errorf("customization for character %d: %w", id, err)
	}

	return &characters.Character{
		ID:        characters.ID(dbChar.ID),
		OwnerID:   dbChar.OwnerID,
		Name:      dbChar.Name,
		BodyType:  characters.BodyType(dbChar.BodyType),
		Species:   characters.Species(dbChar.Species),
		Class:     characters.Class(dbChar.Class),
		Level:     uint8(dbChar.Level),
		Xp:        uint64(dbChar.Xp),
		HpMax:     uint8(dbChar.HpMax),
		HpCurrent: uint8(dbChar.HpCurrent),
		Prestige:  uint(dbChar.Prestige),
		Stats: &characters.Stats{
			ID:           characters.ID(dbStats.CharacterID),
			Strength:     uint8(dbStats.Strength),
			Dexterity:    uint8(dbStats.Dexterity),
			Constitution: uint8(dbStats.Constitution),
			Intelligence: uint8(dbStats.Intelligence),
			Wisdom:       uint8(dbStats.Wisdom),
			Charisma:     uint8(dbStats.Charisma),
		},
		Customization: &characters.Customization{
			ID:          characters.ID(dbCust.CharacterID),
			Hair:        uint8(dbCust.Hair),
			Beard:       uint8(dbCust.Beard),
			SkinTone:    uint8(dbCust.SkinTone),
			HairColor:   uint8(dbCust.HairColor),
			Cloak:       uint8(dbCust.Cloak),
			AccentColor: uint8(dbCust.AccentColor),
		},
	}, nil
}

func (c *characterRepo) FindAllCharacters(ctx context.Context, exec db.DBTX, limit, offset int) ([]characters.Character, uint64, error) {
	q := c.q(exec)

	dbChars, err := q.SelectAllCharacters(ctx, db.SelectAllCharactersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list characters: %w", err)
	}

	count, err := q.CountAllCharacters(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count characters: %w", err)
	}

	result := make([]characters.Character, len(dbChars))
	for i, dbChar := range dbChars {
		dbStats, err := q.SelectCharacterStats(ctx, dbChar.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("stats for character %d: %w", dbChar.ID, err)
		}

		dbCust, err := q.SelectCharacterCustomization(ctx, dbChar.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("customization for character %d: %w", dbChar.ID, err)
		}

		result[i] = characters.Character{
			ID:        characters.ID(dbChar.ID),
			OwnerID:   dbChar.OwnerID,
			Name:      dbChar.Name,
			BodyType:  characters.BodyType(dbChar.BodyType),
			Species:   characters.Species(dbChar.Species),
			Class:     characters.Class(dbChar.Class),
			Level:     uint8(dbChar.Level),
			Xp:        uint64(dbChar.Xp),
			HpMax:     uint8(dbChar.HpMax),
			HpCurrent: uint8(dbChar.HpCurrent),
			Prestige:  uint(dbChar.Prestige),
			Stats: &characters.Stats{
				ID:           characters.ID(dbStats.CharacterID),
				Strength:     uint8(dbStats.Strength),
				Dexterity:    uint8(dbStats.Dexterity),
				Constitution: uint8(dbStats.Constitution),
				Intelligence: uint8(dbStats.Intelligence),
				Wisdom:       uint8(dbStats.Wisdom),
				Charisma:     uint8(dbStats.Charisma),
			},
			Customization: &characters.Customization{
				ID:          characters.ID(dbCust.CharacterID),
				Hair:        uint8(dbCust.Hair),
				Beard:       uint8(dbCust.Beard),
				SkinTone:    uint8(dbCust.SkinTone),
				HairColor:   uint8(dbCust.HairColor),
				Cloak:       uint8(dbCust.Cloak),
				AccentColor: uint8(dbCust.AccentColor),
			},
		}
	}

	return result, uint64(count), nil
}

func (c *characterRepo) FindCharactersByOwner(ctx context.Context, exec db.DBTX, ownerID uuid.UUID, limit, offset int) ([]characters.Character, uint64, error) {
	q := c.q(exec)

	dbChars, err := q.SelectCharactersByOwner(ctx, db.SelectCharactersByOwnerParams{
		OwnerID: ownerID,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list characters by owner %s: %w", ownerID, err)
	}

	count, err := q.CountCharactersByOwner(ctx, ownerID)
	if err != nil {
		return nil, 0, fmt.Errorf("count characters by owner %s: %w", ownerID, err)
	}

	result := make([]characters.Character, len(dbChars))
	for i, dbChar := range dbChars {
		char, err := c.FindCharacter(ctx, exec, characters.ID(dbChar.ID))
		if err != nil {
			return nil, 0, err
		}
		result[i] = *char
	}

	return result, uint64(count), nil
}

func (c *characterRepo) CountCharactersByOwner(ctx context.Context, exec db.DBTX, ownerID uuid.UUID) (uint64, error) {
	q := c.q(exec)
	count, err := q.CountCharactersByOwner(ctx, ownerID)
	if err != nil {
		return 0, fmt.Errorf("count characters by owner %s: %w", ownerID, err)
	}
	return uint64(count), nil
}

func (c *characterRepo) UpdateCharacterHP(ctx context.Context, exec db.DBTX, id characters.ID, hpCurrent uint8) (*characters.Character, error) {
	q := c.q(exec)

	_, err := q.UpdateCharacterHP(ctx, db.UpdateCharacterHPParams{
		ID:        int64(id),
		HpCurrent: int32(hpCurrent),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("update character hp %d: %w", id, characters.ErrNotFound)
		}
		return nil, fmt.Errorf("update character hp %d: %w", id, err)
	}

	return c.FindCharacter(ctx, exec, id)
}

func (c *characterRepo) HealCharacter(ctx context.Context, exec db.DBTX, id characters.ID, healAmount uint8) (*characters.Character, error) {
	q := c.q(exec)

	_, err := q.HealCharacter(ctx, db.HealCharacterParams{
		ID:        int64(id),
		HpCurrent: int32(healAmount),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("heal character %d: %w", id, characters.ErrNotFound)
		}
		return nil, fmt.Errorf("heal character %d: %w", id, err)
	}

	return c.FindCharacter(ctx, exec, id)
}

func (c *characterRepo) UpdateCharacterProgress(ctx context.Context, exec db.DBTX, id characters.ID, level uint8, xp uint64, hpMax, hpCurrent uint8) (*characters.Character, error) {
	q := c.q(exec)

	_, err := q.UpdateCharacterProgress(ctx, db.UpdateCharacterProgressParams{
		ID:        int64(id),
		Level:     int32(level),
		Xp:        int32(xp),
		HpMax:     int32(hpMax),
		HpCurrent: int32(hpCurrent),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("update character progress %d: %w", id, characters.ErrNotFound)
		}
		return nil, fmt.Errorf("update character progress %d: %w", id, err)
	}

	return c.FindCharacter(ctx, exec, id)
}

func (c *characterRepo) PrestigeCharacter(ctx context.Context, exec db.DBTX, id characters.ID, ownerID uuid.UUID, newHpMax uint8) (*characters.Character, error) {
	q := c.q(exec)

	dbChar, err := q.SelectCharacter(ctx, int64(id))
	if err != nil {
		return nil, fmt.Errorf("character %d: %w", id, characters.ErrNotFound)
	}
	if ownerID != uuid.Nil && dbChar.OwnerID != ownerID {
		return nil, fmt.Errorf("character %d: %w", id, characters.ErrNotFound)
	}
	if dbChar.Level < 10 {
		return nil, characters.ErrPrestigeRequirement
	}

	_, err = q.PrestigeCharacter(ctx, db.PrestigeCharacterParams{
		ID:      int64(id),
		HpMax:   int32(newHpMax),
		OwnerID: dbChar.OwnerID,
	})
	if err != nil {
		return nil, fmt.Errorf("prestige character %d: %w", id, err)
	}

	if _, err := q.BoostAllStatsOnPrestige(ctx, int64(id)); err != nil {
		return nil, fmt.Errorf("boost stats on prestige for character %d: %w", id, err)
	}

	return c.FindCharacter(ctx, exec, id)
}

func (c *characterRepo) UpdateCharacter(ctx context.Context, exec db.DBTX, character *characters.Character, ownerID uuid.UUID) (*characters.Character, error) {
	q := c.q(exec)

	var err error
	if ownerID == uuid.Nil {
		_, err = q.UpdateCharacter(ctx, db.UpdateCharacterParams{
			ID:        int64(character.ID),
			Name:      character.Name,
			BodyType:  string(character.BodyType),
			Species:   string(character.Species),
			Class:     string(character.Class),
			Level:     int32(character.Level),
			Xp:        int32(character.Xp),
			HpMax:     int32(character.HpMax),
			HpCurrent: int32(character.HpCurrent),
		})
	} else {
		_, err = q.UpdateCharacterOwned(ctx, db.UpdateCharacterOwnedParams{
			ID:        int64(character.ID),
			OwnerID:   ownerID,
			Name:      character.Name,
			BodyType:  string(character.BodyType),
			Species:   string(character.Species),
			Class:     string(character.Class),
			Level:     int32(character.Level),
			Xp:        int32(character.Xp),
			HpMax:     int32(character.HpMax),
			HpCurrent: int32(character.HpCurrent),
		})
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("update character %d: %w", character.ID, characters.ErrNotFound)
		}
		return nil, fmt.Errorf("update character %d: %w", character.ID, err)
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
		return nil, fmt.Errorf("update stats for character %d: %w", character.ID, err)
	}

	_, err = q.UpdateCustomization(ctx, db.UpdateCustomizationParams{
		CharacterID: int64(character.ID),
		Hair:        int16(character.Customization.Hair),
		Beard:       int16(character.Customization.Beard),
		SkinTone:    int16(character.Customization.SkinTone),
		HairColor:   int16(character.Customization.HairColor),
		Cloak:       int16(character.Customization.Cloak),
		AccentColor: int16(character.Customization.AccentColor),
	})
	if err != nil {
		return nil, fmt.Errorf("update customization for character %d: %w", character.ID, err)
	}

	return c.FindCharacter(ctx, exec, character.ID)
}

func (c *characterRepo) DeleteCharacter(ctx context.Context, exec db.DBTX, id characters.ID, ownerID uuid.UUID) error {
	q := c.q(exec)

	var rows int64
	var err error
	if ownerID == uuid.Nil {
		rows, err = q.DeleteCharacter(ctx, int64(id))
	} else {
		rows, err = q.DeleteCharacterOwned(ctx, db.DeleteCharacterOwnedParams{
			ID:      int64(id),
			OwnerID: ownerID,
		})
	}
	if err != nil {
		return fmt.Errorf("delete character %d: %w", id, err)
	}
	if rows == 0 {
		return fmt.Errorf("delete character %d: %w", id, characters.ErrNotFound)
	}
	return nil
}
