package postgres

import (
	"context"
	"dZev1/character-gallery/internal/items"
	"dZev1/character-gallery/internal/postgres/db"

	"github.com/jackc/pgx/v5/pgtype"
)

type itemRepo struct {
	db.Queries
}

func (i *itemRepo) FindAllItems(ctx context.Context) ([]items.Item, error) {
	allItems, err := i.FindAllItems(ctx)
	if err != nil {
		return nil, err
	}
	return allItems, nil
}

func (i *itemRepo) FindItem(ctx context.Context, itemID items.ItemID) (*items.Item, error) {
	itm, err := i.FindItem(ctx, itemID)
	if err != nil {
		return nil, err
	}
	return itm, nil
}

func (i *itemRepo) SaveItem(ctx context.Context, item *items.Item) (*items.Item, error) {
	saveItemParams := db.CreateItemParams{
		Name:        item.Name,
		Type:        string(item.Type),
		Description: item.Description,
		Equippable:  item.Equippable,
		Rarity:      int32(item.Rarity),
		Damage:      pgtype.Int4{Int32: int32(*item.Damage), Valid: item.Damage != nil},
		Defense:     pgtype.Int4{Int32: int32(*item.Defense), Valid: item.Defense != nil},
		HealAmount:  pgtype.Int4{Int32: int32(*item.HealAmount), Valid: item.HealAmount != nil},
		ManaCost:    pgtype.Int4{Int32: int32(*item.ManaCost), Valid: item.ManaCost != nil},
		Duration:    pgtype.Int4{Int32: int32(*item.Duration), Valid: item.Duration != nil},
		Cooldown:    pgtype.Int4{Int32: int32(*item.Cooldown), Valid: item.Cooldown != nil},
		Capacity:    pgtype.Int4{Int32: int32(*item.Capacity), Valid: item.Capacity != nil},
	}

	saveItem, err := i.Queries.CreateItem(ctx, saveItemParams)
	if err != nil {
		return nil, err
	}

	return &items.Item{
		ID:          items.ItemID(saveItem.ID),
		Name:        saveItem.Name,
		Type:        items.Type(saveItem.Type),
		Description: saveItem.Description,
		Equippable:  saveItem.Equippable,
		Rarity:      uint8(saveItem.Rarity),
		Damage:      toU64ptr(saveItem.Damage),
		Defense:     toU64ptr(saveItem.Defense),
		HealAmount:  toU64ptr(saveItem.HealAmount),
		ManaCost:    toU64ptr(saveItem.ManaCost),
		Duration:    toU64ptr(saveItem.Duration),
		Cooldown:    toU64ptr(saveItem.Cooldown),
		Capacity:    toU64ptr(saveItem.Capacity),
	}, nil
}

func (i *itemRepo) SeedItems(ctx context.Context, items []items.Item) error {
	seedItemParams := make([]db.SeedItemsParams, len(items))
	for idx, item := range items {
		seedItemParams[idx] = db.SeedItemsParams{
			Name:        item.Name,
			Type:        string(item.Type),
			Description: item.Description,
			Equippable:  item.Equippable,
			Rarity:      int32(item.Rarity),
			Damage:      pgtype.Int4{Int32: int32(*item.Damage), Valid: item.Damage != nil},
			Defense:     pgtype.Int4{Int32: int32(*item.Defense), Valid: item.Defense != nil},
			HealAmount:  pgtype.Int4{Int32: int32(*item.HealAmount), Valid: item.HealAmount != nil},
			ManaCost:    pgtype.Int4{Int32: int32(*item.ManaCost), Valid: item.ManaCost != nil},
			Duration:    pgtype.Int4{Int32: int32(*item.Duration), Valid: item.Duration != nil},
			Cooldown:    pgtype.Int4{Int32: int32(*item.Cooldown), Valid: item.Cooldown != nil},
			Capacity:    pgtype.Int4{Int32: int32(*item.Capacity), Valid: item.Capacity != nil},
		}
	}
	_, err := i.Queries.SeedItems(ctx, seedItemParams)

	if err != nil {
		return err
	}

	return nil
}

func toU64ptr(field pgtype.Int4) *uint64 {
	if field.Valid {
		v := uint64(field.Int32)
		return &v
	}
	return nil
}
