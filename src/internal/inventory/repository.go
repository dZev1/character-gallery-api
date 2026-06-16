package inventory

import (
	"context"
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/items"
	"dZev1/character-gallery/internal/postgres/db"
)

type Repository interface {
	AddItemToCharacter(ctx context.Context, exec db.DBTX, characterID characters.CharacterID, itemID items.ItemID, quantity uint8) (*InventoryItem, error)
	RemoveItemFromCharacter(ctx context.Context, exec db.DBTX, characterID characters.CharacterID, itemID items.ItemID, quantity uint8) (*InventoryItem, error)
	GetCharacterInventory(ctx context.Context, exec db.DBTX, characterID characters.CharacterID) ([]InventoryItem, error)
}
