package postgres

import (
	"context"
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/inventory"
	"dZev1/character-gallery/internal/items"
	"dZev1/character-gallery/internal/postgres/db"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var _ inventory.Repository = (*inventoryRepo)(nil)

type inventoryRepo struct {
	pool *pgxpool.Pool
}

func NewInventoryRepo(pool *pgxpool.Pool) *inventoryRepo {
	return &inventoryRepo{pool: pool}
}

func (i *inventoryRepo) q(exec db.DBTX) *db.Queries {
	return db.New(exec)
}

func (i *inventoryRepo) AddItemToCharacter(ctx context.Context, exec db.DBTX, param inventory.RepositoryParam) (*inventory.InventoryItem, error) {
	q := i.q(exec)

	item, err := checkItem(ctx, q, param.ItemID)
	if err != nil {
		return nil, err
	}

	if err := checkCharacter(ctx, q, param.CharacterID); err != nil {
		return nil, err
	}

	params := db.AddItemToCharacterParams{
		CharacterID: int64(param.CharacterID),
		ItemID:      int64(param.ItemID),
		Quantity:    int32(param.Quantity),
	}

	row, err := q.AddItemToCharacter(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("add item %d to character %d: %w", param.ItemID, param.CharacterID, err)
	}

	return &inventory.InventoryItem{
		Item:       item,
		Quantity:   uint8(row.Quantity),
		IsEquipped: row.IsEquipped,
	}, nil

}

func (i *inventoryRepo) RemoveItemFromCharacter(ctx context.Context, exec db.DBTX, param inventory.RepositoryParam) (*inventory.InventoryItem, error) {
	q := i.q(exec)

	item, err := checkItem(ctx, q, param.ItemID)
	if err != nil {
		return nil, err
	}

	if err := checkCharacter(ctx, q, param.CharacterID); err != nil {
		return nil, err
	}

	params := db.RemoveItemFromCharacterParams{
		CharacterID: int64(param.CharacterID),
		ItemID:      int64(param.ItemID),
		Quantity:    int32(param.Quantity),
	}

	row, err := q.RemoveItemFromCharacter(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("remove item %d from character %d: %w", param.ItemID, param.CharacterID, err)
	}
	if row == nil {
		return nil, errors.New("cannot remove more items than you have")
	}
	return &inventory.InventoryItem{
		Item:       item,
		Quantity:   uint8(row.Quantity),
		IsEquipped: row.IsEquipped,
	}, nil
}

func (i *inventoryRepo) GetCharacterInventory(ctx context.Context, exec db.DBTX, characterID characters.CharacterID) ([]inventory.InventoryItem, error) {
	q := i.q(exec)

	var itemSlice []inventory.InventoryItem

	rows, err := q.GetCharacterInventory(ctx, int64(characterID))
	if err != nil {
		return nil, fmt.Errorf("get inventory for character %d: %w", characterID, err)
	}
	for _, row := range rows {
		item := &items.Item{
			ID:          items.ItemID(row.ItemID),
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
		}

		itemSlice = append(itemSlice, inventory.InventoryItem{
			Item:       item,
			Quantity:   uint8(row.Quantity),
			IsEquipped: row.IsEquipped,
		})
	}
	return itemSlice, nil
}

func checkCharacter(ctx context.Context, q *db.Queries, charID characters.CharacterID) error {
	_, err := q.SelectCharacter(ctx, int64(charID))
	if err != nil {
		return fmt.Errorf("character %d: %w", charID, characters.ErrNotFound)
	}
	return nil
}

func checkItem(ctx context.Context, q *db.Queries, itemID items.ItemID) (*items.Item, error) {
	itemRow, err := q.FindItem(ctx, int64(itemID))
	if err != nil {
		return nil, fmt.Errorf("item %d: %w", itemID, items.ErrNotFound)
	}
	item := &items.Item{
		ID:          itemID,
		Name:        itemRow.Name,
		Type:        items.Type(itemRow.Type),
		Description: itemRow.Description,
		Equippable:  itemRow.Equippable,
		Rarity:      uint8(itemRow.Rarity),
		Damage:      toU64ptr(itemRow.Damage),
		Defense:     toU64ptr(itemRow.Defense),
		HealAmount:  toU64ptr(itemRow.HealAmount),
		ManaCost:    toU64ptr(itemRow.ManaCost),
		Duration:    toU64ptr(itemRow.Duration),
		Cooldown:    toU64ptr(itemRow.Cooldown),
		Capacity:    toU64ptr(itemRow.Capacity),
	}
	return item, nil
}
