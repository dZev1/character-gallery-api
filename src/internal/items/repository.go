package items

import "context"

type Repository interface {
	FindAllItems(ctx context.Context) ([]Item, error)
	FindItem(ctx context.Context, itemID ItemID) (*Item, error)
	SaveItem(ctx context.Context, item *Item) *Item
	SeedItems(ctx context.Context, items []Item) error
}
