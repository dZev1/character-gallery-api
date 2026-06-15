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

func (i *itemRepo) FindAllItems(ctx context.Context, exec db.DBTX) ([]items.Item, error) {
	q := db.New(exec)
	allItems, err := q.FindAllItems(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]items.Item, len(allItems))
	for idx, it := range allItems {
		result[idx] = items.Item{
			ID:          items.ItemID(it.ID),
			Name:        it.Name,
			Type:        items.Type(it.Type),
			Description: it.Description,
			Equippable:  it.Equippable,
			Rarity:      uint8(it.Rarity),
			Damage:      toU64ptr(it.Damage),
			Defense:     toU64ptr(it.Defense),
			HealAmount:  toU64ptr(it.HealAmount),
			ManaCost:    toU64ptr(it.ManaCost),
			Duration:    toU64ptr(it.Duration),
			Cooldown:    toU64ptr(it.Cooldown),
			Capacity:    toU64ptr(it.Capacity),
		}
	}

	return result, nil
}

func (i *itemRepo) FindItem(ctx context.Context, exec db.DBTX, itemID items.ItemID) (*items.Item, error) {
	q := db.New(exec)
	itm, err := q.FindItem(ctx, int64(itemID))
	if err != nil {
		return nil, err
	}

	return &items.Item{
		ID:          items.ItemID(itm.ID),
		Name:        itm.Name,
		Type:        items.Type(itm.Type),
		Description: itm.Description,
		Equippable:  itm.Equippable,
		Rarity:      uint8(itm.Rarity),
		Damage:      toU64ptr(itm.Damage),
		Defense:     toU64ptr(itm.Defense),
		HealAmount:  toU64ptr(itm.HealAmount),
		ManaCost:    toU64ptr(itm.ManaCost),
		Duration:    toU64ptr(itm.Duration),
		Cooldown:    toU64ptr(itm.Cooldown),
		Capacity:    toU64ptr(itm.Capacity),
	}, nil
}

func (i *itemRepo) SaveItem(ctx context.Context, exec db.DBTX, item *items.Item) (*items.Item, error) {
	q := db.New(exec)

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

	saveItem, err := q.CreateItem(ctx, saveItemParams)
	if err != nil {
		return nil, err
	}

	item.ID = items.ItemID(saveItem.ID)

	return item, nil
}

func (i *itemRepo) SeedItems(ctx context.Context, exec db.DBTX, items []items.Item) error {
	q := db.New(exec)

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

	_, err := q.SeedItems(ctx, seedItemParams)
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
