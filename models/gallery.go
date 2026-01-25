package models

import (
	"github.com/dZev1/character-gallery/models/characters"
	"github.com/dZev1/character-gallery/models/inventory"
)

type CharacterGallery interface {
	Create(character *characters.Character) error
	Close() error
	Get(id characters.CharacterID) (*characters.Character, error)
	GetAll(page, limit int) ([]characters.Character, error)
	Edit(character *characters.Character) error
	Remove(id characters.CharacterID) error

	SeedItems() error
	AddItemToCharacter(characterID characters.CharacterID, itemID inventory.ItemID, quantity uint8) error
	RemoveItemFromCharacter(characterID characters.CharacterID, itemID inventory.ItemID, quantity uint8) error
	GetCharacterInventory(characterID characters.CharacterID) ([]inventory.InventoryItem, error)
}
