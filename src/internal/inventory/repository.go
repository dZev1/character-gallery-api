package inventory

import (
	"context"
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/items"
	"dZev1/character-gallery/internal/postgres/db"
	"errors"

	"github.com/google/uuid"
)

var ErrInsufficientQuantity = errors.New("cannot remove more items than you have")
var ErrItemNotEquippable = errors.New("item is not equippable")
var ErrItemNotInInventory = errors.New("item is not in character inventory")
var ErrItemNotConsumable = errors.New("item is not a healing consumable")

type RepositoryParam struct {
	CharacterID characters.ID
	OwnerID     uuid.UUID
	ItemID      items.ID
	Quantity    uint8
}

type Repository interface {
	AddItemToCharacter(ctx context.Context, exec db.DBTX, params RepositoryParam) (*InventoryItem, error)
	RemoveItemFromCharacter(ctx context.Context, exec db.DBTX, param RepositoryParam) (*InventoryItem, error)
	GetCharacterInventory(ctx context.Context, exec db.DBTX, characterID characters.ID) ([]InventoryItem, error)
	GetCharacterEquippedItems(ctx context.Context, exec db.DBTX, characterID characters.ID) ([]InventoryItem, error)
	SetItemEquipped(ctx context.Context, exec db.DBTX, characterID characters.ID, ownerID uuid.UUID, itemID items.ID, equip bool) (*InventoryItem, error)
}
