package postgres

import (
	"context"
	"dZev1/character-gallery/internal/items"
	"dZev1/character-gallery/internal/postgres/db"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ items.Repository = (*itemRepo)(nil)

type itemRepo struct {
	pool *pgxpool.Pool
}

func NewItemRepo(pool *pgxpool.Pool) *itemRepo {
	return &itemRepo{pool: pool}
}

func (i *itemRepo) q(exec db.DBTX) *db.Queries {
	return db.New(exec)
}

func (i *itemRepo) FindAllItems(ctx context.Context, exec db.DBTX) ([]items.Item, uint64, error) {
	q := i.q(exec)

	allItems, err := q.FindAllItems(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list items: %w", err)
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

	return result, uint64(len(result)), nil
}

func (i *itemRepo) FindItem(ctx context.Context, exec db.DBTX, itemID items.ItemID) (*items.Item, error) {
	q := i.q(exec)

	itm, err := q.FindItem(ctx, int64(itemID))
	if err != nil {
		return nil, fmt.Errorf("item %d: %w", itemID, items.ErrNotFound)
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
	q := i.q(exec)

	saveItemParams := db.CreateItemParams{
		Name:        item.Name,
		Type:        string(item.Type),
		Description: item.Description,
		Equippable:  item.Equippable,
		Rarity:      int32(item.Rarity),
		Damage:      toInt4(item.Damage),
		Defense:     toInt4(item.Defense),
		HealAmount:  toInt4(item.HealAmount),
		ManaCost:    toInt4(item.ManaCost),
		Duration:    toInt4(item.Duration),
		Cooldown:    toInt4(item.Cooldown),
		Capacity:    toInt4(item.Capacity),
	}

	saveItem, err := q.CreateItem(ctx, saveItemParams)
	if err != nil {
		return nil, fmt.Errorf("create item: %w", err)
	}

	item.ID = items.ItemID(saveItem.ID)

	return item, nil
}

func (i *itemRepo) SeedItems(ctx context.Context, exec db.DBTX, items []items.Item) error {
	q := i.q(exec)

	seedItemParams := make([]db.SeedItemsParams, len(items))
	for idx, item := range items {
		seedItemParams[idx] = db.SeedItemsParams{
			Name:        item.Name,
			Type:        string(item.Type),
			Description: item.Description,
			Equippable:  item.Equippable,
			Rarity:      int32(item.Rarity),
			Damage:      toInt4(item.Damage),
			Defense:     toInt4(item.Defense),
			HealAmount:  toInt4(item.HealAmount),
			ManaCost:    toInt4(item.ManaCost),
			Duration:    toInt4(item.Duration),
			Cooldown:    toInt4(item.Cooldown),
			Capacity:    toInt4(item.Capacity),
		}
	}

	_, err := q.SeedItems(ctx, seedItemParams)
	if err != nil {
		return fmt.Errorf("seed items: %w", err)
	}

	return nil
}

func toInt4(ptr *uint64) pgtype.Int4 {
	if ptr != nil {
		return pgtype.Int4{Int32: int32(*ptr), Valid: true}
	}
	return pgtype.Int4{}
}

func toU64ptr(field pgtype.Int4) *uint64 {
	if field.Valid {
		v := uint64(field.Int32)
		return &v
	}
	return nil
}
