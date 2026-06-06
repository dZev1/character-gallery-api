package inventory

import (
	"context"
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/items"
)

type Repository interface {
	AddItemToCharacter(ctx context.Context, characterID characters.CharacterID, itemID items.ItemID, quantity uint8) (*InventoryItem, error)
	RemoveItemFromCharacter(ctx context.Context, characterID characters.CharacterID, itemID items.ItemID, quantity uint8) error
	GetCharacterInventory(ctx context.Context, characterID characters.CharacterID) ([]InventoryItem, error)
}
