package inventory

import (
	"context"
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/items"
	"dZev1/character-gallery/internal/postgres/db"
)

type RepositoryParam struct {
	CharacterID characters.CharacterID
	ItemID      items.ItemID
	Quantity    uint8
}

type Repository interface {
	AddItemToCharacter(ctx context.Context, exec db.DBTX, params RepositoryParam) (*InventoryItem, error)
	RemoveItemFromCharacter(ctx context.Context, exec db.DBTX, param RepositoryParam) (*InventoryItem, error)
	GetCharacterInventory(ctx context.Context, exec db.DBTX, characterID characters.CharacterID) ([]InventoryItem, error)
}
