package items

import (
	"context"
	"dZev1/character-gallery/internal/postgres/db"
)

type Repository interface {
	FindAllItems(ctx context.Context, exec db.DBTX, limit, offset int) ([]Item, uint64, error)
	FindItem(ctx context.Context, exec db.DBTX, itemID ItemID) (*Item, error)
	SaveItem(ctx context.Context, exec db.DBTX, item *Item) (*Item, error)
	SeedItems(ctx context.Context, exec db.DBTX, items []Item) error
}
