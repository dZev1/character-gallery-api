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

	char, err := q.CreateCharacter(ctx, db.CreateCharacterParams{
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
	if err != nil {
		return nil, fmt.Errorf("create character: %w", err)
	}

	character.ID = characters.ID(char.ID)
	character.OwnerID = char.OwnerID
	character.Prestige = uint(char.Prestige)

	if _, err = q.CreateCustomization(ctx, db.CreateCustomizationParams{
		CharacterID: char.ID,
		Hair:        int16(character.Customization.Hair),
		Beard:       int16(character.Customization.Beard),
		SkinTone:    int16(character.Customization.SkinTone),
		HairColor:   int16(character.Customization.HairColor),
		Cloak:       int16(character.Customization.Cloak),
		AccentColor: int16(character.Customization.AccentColor),
	}); err != nil {
		return nil, fmt.Errorf("create customization: %w", err)
	}
	character.Customization.ID = character.ID

	if _, err = q.CreateStats(ctx, db.CreateStatsParams{
		CharacterID:  char.ID,
		Strength:     int16(character.Stats.Strength),
		Dexterity:    int16(character.Stats.Dexterity),
		Constitution: int16(character.Stats.Constitution),
		Intelligence: int16(character.Stats.Intelligence),
		Wisdom:       int16(character.Stats.Wisdom),
		Charisma:     int16(character.Stats.Charisma),
	}); err != nil {
		return nil, fmt.Errorf("create stats: %w", err)
	}
	character.Stats.ID = character.ID

	return character, nil
}

func (c *characterRepo) FindCharacter(ctx context.Context, exec db.DBTX, id characters.ID) (*characters.Character, error) {
	row, err := c.q(exec).SelectCharacterFull(ctx, int64(id))
	if err != nil {
		return nil, fmt.Errorf("character %d: %w", id, characters.ErrNotFound)
	}
	return toDomainCharacterFull(
		row.ID, row.OwnerID, row.OwnerUsername, row.Name, row.BodyType, row.Species, row.Class,
		row.Level, row.Xp, row.HpMax, row.HpCurrent, row.Prestige,
		row.Strength, row.Dexterity, row.Constitution, row.Intelligence, row.Wisdom, row.Charisma,
		row.Hair, row.Beard, row.SkinTone, row.HairColor, row.Cloak, row.AccentColor,
	), nil
}

func (c *characterRepo) FindAllCharacters(ctx context.Context, exec db.DBTX, limit, offset int) ([]characters.Character, uint64, error) {
	q := c.q(exec)

	rows, err := q.SelectAllCharactersFull(ctx, db.SelectAllCharactersFullParams{
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

	result := make([]characters.Character, len(rows))
	for i, row := range rows {
		result[i] = *toDomainCharacterFull(
			row.ID, row.OwnerID, row.OwnerUsername, row.Name, row.BodyType, row.Species, row.Class,
			row.Level, row.Xp, row.HpMax, row.HpCurrent, row.Prestige,
			row.Strength, row.Dexterity, row.Constitution, row.Intelligence, row.Wisdom, row.Charisma,
			row.Hair, row.Beard, row.SkinTone, row.HairColor, row.Cloak, row.AccentColor,
		)
	}
	return result, uint64(count), nil
}

func (c *characterRepo) FindCharactersByOwner(ctx context.Context, exec db.DBTX, ownerID uuid.UUID, limit, offset int) ([]characters.Character, uint64, error) {
	q := c.q(exec)

	rows, err := q.SelectCharactersByOwnerFull(ctx, db.SelectCharactersByOwnerFullParams{
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

	result := make([]characters.Character, len(rows))
	for i, row := range rows {
		result[i] = *toDomainCharacterFull(
			row.ID, row.OwnerID, row.OwnerUsername, row.Name, row.BodyType, row.Species, row.Class,
			row.Level, row.Xp, row.HpMax, row.HpCurrent, row.Prestige,
			row.Strength, row.Dexterity, row.Constitution, row.Intelligence, row.Wisdom, row.Charisma,
			row.Hair, row.Beard, row.SkinTone, row.HairColor, row.Cloak, row.AccentColor,
		)
	}
	return result, uint64(count), nil
}

func (c *characterRepo) CountCharactersByOwner(ctx context.Context, exec db.DBTX, ownerID uuid.UUID) (uint64, error) {
	count, err := c.q(exec).CountCharactersByOwner(ctx, ownerID)
	if err != nil {
		return 0, fmt.Errorf("count characters by owner %s: %w", ownerID, err)
	}
	return uint64(count), nil
}

func (c *characterRepo) UpdateCharacterHP(ctx context.Context, exec db.DBTX, id characters.ID, hpCurrent uint8) (*characters.Character, error) {
	_, err := c.q(exec).UpdateCharacterHP(ctx, db.UpdateCharacterHPParams{
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
	_, err := c.q(exec).HealCharacter(ctx, db.HealCharacterParams{
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
	_, err := c.q(exec).UpdateCharacterProgress(ctx, db.UpdateCharacterProgressParams{
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

	if _, err = q.PrestigeCharacter(ctx, db.PrestigeCharacterParams{
		ID:      int64(id),
		HpMax:   int32(newHpMax),
		OwnerID: dbChar.OwnerID,
	}); err != nil {
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

	if _, err = q.UpdateStats(ctx, db.UpdateStatsParams{
		CharacterID:  int64(character.ID),
		Strength:     int16(character.Stats.Strength),
		Dexterity:    int16(character.Stats.Dexterity),
		Constitution: int16(character.Stats.Constitution),
		Intelligence: int16(character.Stats.Intelligence),
		Wisdom:       int16(character.Stats.Wisdom),
		Charisma:     int16(character.Stats.Charisma),
	}); err != nil {
		return nil, fmt.Errorf("update stats for character %d: %w", character.ID, err)
	}

	if _, err = q.UpdateCustomization(ctx, db.UpdateCustomizationParams{
		CharacterID: int64(character.ID),
		Hair:        int16(character.Customization.Hair),
		Beard:       int16(character.Customization.Beard),
		SkinTone:    int16(character.Customization.SkinTone),
		HairColor:   int16(character.Customization.HairColor),
		Cloak:       int16(character.Customization.Cloak),
		AccentColor: int16(character.Customization.AccentColor),
	}); err != nil {
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

func (c *characterRepo) GetUserStatRoll(ctx context.Context, exec db.DBTX, userID uuid.UUID) (*characters.UserStatRoll, error) {
	row, err := c.q(exec).GetUserStatRoll(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user stat roll %s: %w", userID, err)
	}
	return toDomainStatRoll(row), nil
}

func (c *characterRepo) UpsertUserStatRoll(ctx context.Context, exec db.DBTX, userID uuid.UUID, roll *characters.UserStatRoll) (*characters.UserStatRoll, error) {
	rolls32 := make([]int32, len(roll.Rolls))
	for i, v := range roll.Rolls {
		rolls32[i] = int32(v)
	}

	row, err := c.q(exec).UpsertUserStatRoll(ctx, db.UpsertUserStatRollParams{
		UserID:      userID,
		Rolls:       rolls32,
		RerollsUsed: int32(roll.RerollsUsed),
	})
	if err != nil {
		return nil, fmt.Errorf("upsert user stat roll %s: %w", userID, err)
	}
	return toDomainStatRoll(row), nil
}

func (c *characterRepo) ConsumeUserStatRoll(ctx context.Context, exec db.DBTX, userID uuid.UUID) error {
	if _, err := c.q(exec).ConsumeUserStatRoll(ctx, userID); err != nil {
		return fmt.Errorf("consume user stat roll %s: %w", userID, err)
	}
	return nil
}

func (c *characterRepo) RecordUserCharacterDelete(ctx context.Context, exec db.DBTX, userID uuid.UUID) error {
	if _, err := c.q(exec).RecordUserCharacterDelete(ctx, userID); err != nil {
		return fmt.Errorf("record character delete %s: %w", userID, err)
	}
	return nil
}

func toDomainStatRoll(row *db.UserStatRoll) *characters.UserStatRoll {
	rolls := make([]int, len(row.Rolls))
	for i, v := range row.Rolls {
		rolls[i] = int(v)
	}
	res := &characters.UserStatRoll{
		Rolls:       rolls,
		RerollsUsed: int(row.RerollsUsed),
	}
	if row.LastCharacterDeletedAt.Valid {
		t := row.LastCharacterDeletedAt.Time
		res.LastCharacterDeletedAt = &t
	}
	return res
}

func toDomainCharacterFull(
	id int64, ownerID uuid.UUID, ownerUsername, name, bodyType, species, class string,
	level, xp, hpMax, hpCurrent, prestige int32,
	strength, dexterity, constitution, intelligence, wisdom, charisma int16,
	hair, beard, skinTone, hairColor, cloak, accentColor int16,
) *characters.Character {
	charID := characters.ID(id)
	return &characters.Character{
		ID:            charID,
		OwnerID:       ownerID,
		OwnerUsername: ownerUsername,
		Name:          name,
		BodyType:  characters.BodyType(bodyType),
		Species:   characters.Species(species),
		Class:     characters.Class(class),
		Level:     uint8(level),
		Xp:        uint64(xp),
		HpMax:     uint8(hpMax),
		HpCurrent: uint8(hpCurrent),
		Prestige:  uint(prestige),
		Stats: &characters.Stats{
			ID:           charID,
			Strength:     uint8(strength),
			Dexterity:    uint8(dexterity),
			Constitution: uint8(constitution),
			Intelligence: uint8(intelligence),
			Wisdom:       uint8(wisdom),
			Charisma:     uint8(charisma),
		},
		Customization: &characters.Customization{
			ID:          charID,
			Hair:        uint8(hair),
			Beard:       uint8(beard),
			SkinTone:    uint8(skinTone),
			HairColor:   uint8(hairColor),
			Cloak:       uint8(cloak),
			AccentColor: uint8(accentColor),
		},
	}
}
