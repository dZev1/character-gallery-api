package postgres_gallery

import (
	"fmt"

	"github.com/dZev1/character-gallery/models/characters"
	"github.com/dZev1/character-gallery/models/inventory"
)

func (cg *PostgresCharacterGallery) SeedItems() error {
	tx, err := cg.db.Beginx()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedInitializeTransaction, err)
	}
	defer tx.Rollback()

	for _, item := range inventory.Items {
		err := cg.insertItemIntoPool(tx, &item)
		if err != nil {
			fmt.Println("Error inserting item:", err)
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%w: %w", ErrFailedCommitTransaction, err)
	}

	return nil
}

func (cg *PostgresCharacterGallery) AddItemToCharacter(characterID characters.CharacterID, itemID inventory.ItemID, quantity uint8) error {
	tx, err := cg.db.Beginx()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedInitializeTransaction, err)
	}
	defer tx.Rollback()

	err = insertIntoCharacterInventory(tx, characterID, itemID, quantity)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%w: %w", ErrFailedCommitTransaction, err)
	}

	return nil
}

func (cg *PostgresCharacterGallery) RemoveItemFromCharacter(characterID characters.CharacterID, itemID inventory.ItemID, quantity uint8) error {
	tx, err := cg.db.Beginx()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedInitializeTransaction, err)
	}
	defer tx.Rollback()

	currentQuantity, err := cg.selectCurrentQuantity(characterID, itemID)
	if err != nil {
		return err
	}

	if currentQuantity > quantity {
		err = updateItemQuantity(tx, quantity, characterID, itemID)
		if err != nil {
			return err
		}
	} else {
		err = deleteItemFromCharacter(tx, characterID, itemID)
		if err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%w: %w", ErrFailedCommitTransaction, err)
	}

	return nil
}

func (cg *PostgresCharacterGallery) GetCharacterInventory(characterID characters.CharacterID) ([]inventory.InventoryItem, error) {
	query := `
		SELECT i.id, i.name, i.type, i.description, i.equippable, i.rarity, i.damage, i.defense, i.heal_amount, i.mana_cost, i.duration
		, ci.quantity, ci.is_equipped
		FROM character_inventory ci
		JOIN items i ON ci.item_id = i.id
		WHERE ci.character_id = $1
	`

	var characterInventory []inventory.InventoryItem
	err := cg.db.Select(&characterInventory, query, characterID)
	if err != nil {
		
		return nil, fmt.Errorf("%w: %w", ErrFailedSelectCharacterInventory, err)
	}
	return characterInventory, nil
}
