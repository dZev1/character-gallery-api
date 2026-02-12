package models

import (
	"dZev1/character-gallery/models/auth"
	"dZev1/character-gallery/models/characters"
	"dZev1/character-gallery/models/inventory"
)

type CharacterGallery interface {
	Create(character *characters.Character) error
	Close() error
	Get(id characters.CharacterID) (*characters.Character, error)
	GetAll(page, limit int) ([]characters.Character, error)
	Edit(character *characters.Character) error
	Remove(id characters.CharacterID) error

	CreateItem(item *inventory.Item) error
	SeedItems(items []inventory.Item) error
	DisplayPoolItems() ([]inventory.Item, error)
	DisplayItem(itemID inventory.ItemID) (*inventory.Item, error)
	AddItemToCharacter(characterID characters.CharacterID, itemID inventory.ItemID, quantity uint8) error
	RemoveItemFromCharacter(characterID characters.CharacterID, itemID inventory.ItemID, quantity uint8) error
	GetCharacterInventory(characterID characters.CharacterID) ([]inventory.InventoryItem, error)
	GetAuthStore() auth.AuthStore
}
