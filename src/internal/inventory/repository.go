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
}
