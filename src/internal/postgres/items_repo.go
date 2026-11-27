package postgres

import (
	"context"
	"dZev1/character-gallery/internal/items"
	"dZev1/character-gallery/internal/postgres/db"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
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

func (i *itemRepo) FindAllItems(ctx context.Context, exec db.DBTX, limit, offset int) ([]items.Item, uint64, error) {
	q := i.q(exec)

	rows, err := q.FindAllItems(ctx, db.FindAllItemsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list items: %w", err)
	}

	total, err := q.CountAllItems(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list items: %w", err)
	}

	return toDomainItems(rows), uint64(total), nil
}

func (i *itemRepo) FindItem(ctx context.Context, exec db.DBTX, itemID items.ID) (*items.Item, error) {
	row, err := i.q(exec).FindItem(ctx, int64(itemID))
	if err != nil {
		return nil, fmt.Errorf("item %d: %w", itemID, items.ErrNotFound)
	}
	return toDomainItem(row), nil
}

func (i *itemRepo) FindItemsByRarityRange(ctx context.Context, exec db.DBTX, minRarity, maxRarity uint8) ([]items.Item, error) {
	rows, err := i.q(exec).FindItemsByRarityRange(ctx, db.FindItemsByRarityRangeParams{
		Rarity:   int32(minRarity),
		Rarity_2: int32(maxRarity),
	})
	if err != nil {
		return nil, fmt.Errorf("find items by rarity %d..%d: %w", minRarity, maxRarity, err)
	}
	return toDomainItems(rows), nil
}

func (i *itemRepo) FindRandomItemByRarityRange(ctx context.Context, exec db.DBTX, minRarity, maxRarity uint8) (*items.Item, error) {
	row, err := i.q(exec).FindRandomItemByRarityRange(ctx, db.FindRandomItemByRarityRangeParams{
		Rarity:   int32(minRarity),
		Rarity_2: int32(maxRarity),
	})
	if err != nil {
		return nil, fmt.Errorf("random item by rarity %d..%d: %w", minRarity, maxRarity, items.ErrNotFound)
	}
	return toDomainItem(row), nil
}

func (i *itemRepo) SaveItem(ctx context.Context, exec db.DBTX, item *items.Item) (*items.Item, error) {
	saved, err := i.q(exec).CreateItem(ctx, db.CreateItemParams{
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
		ImageUrl:    toText(item.ImageURL),
	})
	if err != nil {
		return nil, fmt.Errorf("create item: %w", err)
	}
	item.ID = items.ID(saved.ID)
	return item, nil
}

func (i *itemRepo) SaveItemIfNotExists(ctx context.Context, exec db.DBTX, item *items.Item) (*items.Item, bool, error) {
	saved, err := i.q(exec).CreateItemIfNotExists(ctx, db.CreateItemIfNotExistsParams{
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
		ImageUrl:    toText(item.ImageURL),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("create item %q: %w", item.Name, err)
	}
	return toDomainItem(saved), true, nil
}

func (i *itemRepo) UpdateItem(ctx context.Context, exec db.DBTX, item *items.Item) (*items.Item, error) {
	updated, err := i.q(exec).UpdateItem(ctx, db.UpdateItemParams{
		ID:          int64(item.ID),
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
		ImageUrl:    toText(item.ImageURL),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("update item %d: %w", item.ID, items.ErrNotFound)
		}
		return nil, fmt.Errorf("update item %d: %w", item.ID, err)
	}
	return toDomainItem(updated), nil
}

func (i *itemRepo) DeleteItem(ctx context.Context, exec db.DBTX, itemID items.ID) error {
	rows, err := i.q(exec).DeleteItem(ctx, int64(itemID))
	if err != nil {
		return fmt.Errorf("delete item %d: %w", itemID, err)
	}
	if rows == 0 {
		return fmt.Errorf("delete item %d: %w", itemID, items.ErrNotFound)
	}
	return nil
}

func (i *itemRepo) SeedItems(ctx context.Context, exec db.DBTX, list []items.Item) error {
	params := make([]db.SeedItemsParams, len(list))
	for idx, item := range list {
		params[idx] = db.SeedItemsParams{
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
			ImageUrl:    toText(item.ImageURL),
		}
	}
	if _, err := i.q(exec).SeedItems(ctx, params); err != nil {
		return fmt.Errorf("seed items: %w", err)
	}
	return nil
}

func toDomainItem(row *db.Item) *items.Item {
	return &items.Item{
		ID:          items.ID(row.ID),
		Name:        row.Name,
		Type:        items.Type(row.Type),
		Description: row.Description,
		Equippable:  row.Equippable,
		Rarity:      uint8(row.Rarity),
		Damage:      toU64ptr(row.Damage),
		Defense:     toU64ptr(row.Defense),
		HealAmount:  toU64ptr(row.HealAmount),
		ManaCost:    toU64ptr(row.ManaCost),
		Duration:    toU64ptr(row.Duration),
		Cooldown:    toU64ptr(row.Cooldown),
		Capacity:    toU64ptr(row.Capacity),
		ImageURL:    fromText(row.ImageUrl),
	}
}

func toDomainItems(rows []*db.Item) []items.Item {
	result := make([]items.Item, len(rows))
	for idx, row := range rows {
		result[idx] = *toDomainItem(row)
	}
	return result
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

func toText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func fromText(field pgtype.Text) string {
	if field.Valid {
		return field.String
	}
	return ""
}
