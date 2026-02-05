package postgres_gallery

import (
	"fmt"

	"github.com/dZev1/character-gallery/models/characters"
	"github.com/dZev1/character-gallery/models/inventory"
)

func (cg *PostgresCharacterGallery) SeedItems(items []inventory.Item) error {
	tx, err := cg.db.Beginx()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedInitializeTransaction, err)
	}
	defer tx.Rollback()

	for _, item := range items {
		err := cg.seedItemPool(tx, &item)
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
		fmt.Println("Error de comienzo de transacción")
		return fmt.Errorf("%w: %w", ErrFailedInitializeTransaction, err)
	}
	defer tx.Rollback()

	err = insertIntoCharacterInventory(tx, characterID, itemID, quantity)
	if err != nil {
		fmt.Println("Error al insertar en el inventario del personaje")
		return err
	}

	if err = tx.Commit(); err != nil {
		fmt.Println("Error al commitear!")
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
		SELECT
			i.id          AS "item.id",
			i.name        AS "item.name",
			i.type        AS "item.type",
			i.description AS "item.description",
			i.equippable  AS "item.equippable",
			i.rarity      AS "item.rarity",
			i.damage      AS "item.damage",
			i.defense     AS "item.defense",
			i.heal_amount AS "item.heal_amount",
			i.mana_cost   AS "item.mana_cost",
			i.duration    AS "item.duration",
			ci.quantity,
			ci.is_equipped
		FROM items i
		JOIN inventory ci ON ci.item_id = i.id
		WHERE ci.character_id = $1
		ORDER BY i.id
	`

	var characterInventory []inventory.InventoryItem
	err := cg.db.Select(&characterInventory, query, characterID)
	if err != nil {
		fmt.Println("Error al seleccionar el inventario del personaje")
		return nil, fmt.Errorf("%w: %w", ErrFailedSelectCharacterInventory, err)
	}
	return characterInventory, nil
}

func (cg *PostgresCharacterGallery) DisplayPoolItems() ([]inventory.Item, error) {
	query := `
		SELECT *
		FROM items
		ORDER BY id;
	`

	var items []inventory.Item
	err := cg.db.Select(&items, query)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve items from pool: %v", err)
	}

	return items, nil
}

func (cg *PostgresCharacterGallery) DisplayItem(itemID inventory.ItemID) (*inventory.Item, error) {
	query := `
		SELECT *
		FROM items
		WHERE id = $1; 
	`
	item := &inventory.Item{}
	err := cg.db.Get(item, query, itemID)

	if err != nil {
		return nil, fmt.Errorf("could not retrieve item from item pool: %v", err)
	}

	return item, nil
}

func (cg *PostgresCharacterGallery) CreateItem(item *inventory.Item) error {
	tx, err := cg.db.Beginx()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedInitializeTransaction, err)
	}
	defer tx.Rollback()

	err = cg.insertIntoItemPool(tx, item)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%w: %w", ErrFailedCommitTransaction, err)
	}

	return nil
}
