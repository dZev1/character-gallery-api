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

func (cg *PostgresCharacterGallery) GetCharacterInventory(characterID characters.CharacterID) ([]inventory.Item, error) {
	return nil, nil
}
