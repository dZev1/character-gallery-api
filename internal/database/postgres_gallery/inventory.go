package database

import (
	"github.com/dZev1/character-gallery/models/characters"
	"github.com/dZev1/character-gallery/models/inventory"
)

func (cg *PostgresCharacterGallery) SeedItems() error {
	for _, item := range inventory.Items {
		err := cg.insertItem(&item)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cg *PostgresCharacterGallery) AddItemToCharacter(characterID characters.CharacterID, itemID inventory.ItemID) error {
	return nil
}

func (cg *PostgresCharacterGallery) RemoveItemFromCharacter(characterID characters.CharacterID, itemID inventory.ItemID) error {
	return nil
}

func (cg *PostgresCharacterGallery) GetCharacterInventory(characterID characters.CharacterID) ([]inventory.Item, error) {
	return nil, nil
}
