package items

import (
	"context"
	"dZev1/character-gallery/internal/postgres/db"
)

type Repository interface {
	FindAllItems(ctx context.Context, exec db.DBTX, limit, offset int) ([]Item, uint64, error)
	FindItem(ctx context.Context, exec db.DBTX, itemID ID) (*Item, error)
	FindItemsByRarityRange(ctx context.Context, exec db.DBTX, minRarity, maxRarity uint8) ([]Item, error)
	FindRandomItemByRarityRange(ctx context.Context, exec db.DBTX, minRarity, maxRarity uint8) (*Item, error)
	SaveItem(ctx context.Context, exec db.DBTX, item *Item) (*Item, error)
	UpdateItem(ctx context.Context, exec db.DBTX, item *Item) (*Item, error)
	DeleteItem(ctx context.Context, exec db.DBTX, itemID ID) error
	SeedItems(ctx context.Context, exec db.DBTX, items []Item) error
}
