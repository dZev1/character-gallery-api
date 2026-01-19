package postgres_gallery

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

func (cg *PostgresCharacterGallery) AddItemToCharacter(characterID characters.CharacterID, itemID inventory.ItemID, quantity uint8) error {
	query := `
		INSERT INTO inventory (character_id, item_id, quantity, is_equipped)
		VALUES ($1, $2, $3, FALSE)
		ON CONFLICT (character_id, item_id)
		DO UPDATE SET quantity = inventory.quantity + EXCLUDED.quantity;
	`

	_, err := cg.db.Exec(query, characterID, itemID, quantity)
	if err != nil {
		return err
	}

	return nil
}

func (cg *PostgresCharacterGallery) RemoveItemFromCharacter(characterID characters.CharacterID, itemID inventory.ItemID, quantity uint8) error {
	querySelect := `
		SELECT quantity FROM inventory
		WHERE character_id = $1 AND item_id = $2;
	`

	var currentQuantity uint8
	err := cg.db.QueryRow(querySelect, characterID, itemID).Scan(&currentQuantity)
	if err != nil {
		return err
	}

	if currentQuantity > quantity {
		queryUpdate := `
			UPDATE inventory
			SET quantity = quantity - $1
			WHERE character_id = $2 AND item_id = $3;
		`
		_, err = cg.db.Exec(queryUpdate, quantity, characterID, itemID)
		if err != nil {
			return err
		}
	} else {
		queryDelete := `
			DELETE FROM inventory
			WHERE character_id = $1 AND item_id = $2;
		`
		_, err = cg.db.Exec(queryDelete, characterID, itemID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cg *PostgresCharacterGallery) GetCharacterInventory(characterID characters.CharacterID) ([]inventory.Item, error) {
	return nil, nil
}
