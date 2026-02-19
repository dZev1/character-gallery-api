package postgres_gallery

import (
	"fmt"

	"dZev1/character-gallery/models/characters"
	"dZev1/character-gallery/models/inventory"

	"github.com/jmoiron/sqlx"
)

/*
 *
 * Character related queries
 *
 */

func (cg *PostgresCharacterGallery) insertBaseCharacter(tx *sqlx.Tx, character *characters.Character) error {
	query := `
		INSERT INTO characters (name, body_type, species, class)
		VALUES (:name, :body_type, :species, :class) RETURNING id
	`

	stmt, err := tx.PrepareNamed(query)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCouldNotInsert, err)
	}
	defer stmt.Close()

	err = stmt.Get(&character.ID, character)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCouldNotInsert, err)
	}

	return nil
}

func (cg *PostgresCharacterGallery) insertStats(tx *sqlx.Tx, stats *characters.Stats) error {
	query := `
		INSERT INTO stats (id, strength, dexterity, constitution, intelligence, wisdom, charisma)
		VALUES(:id, :strength, :dexterity, :constitution, :intelligence, :wisdom, :charisma)
	`

	_, err := tx.NamedExec(query, stats)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCouldNotInsert, err)
	}
	return nil
}

func (cg *PostgresCharacterGallery) insertCustomization(tx *sqlx.Tx, customization *characters.Customization) error {
	query := `
		INSERT INTO customizations (id, hair, face, shirt, pants, shoes)
		VALUES(:id, :hair, :face, :shirt, :pants, :shoes)
	`
	_, err := tx.NamedExec(query, customization)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCouldNotInsert, err)
	}
	return nil
}

func (cg *PostgresCharacterGallery) getBaseCharacter(id characters.CharacterID) (*characters.Character, error) {
	character := &characters.Character{}
	query := `
		SELECT * FROM characters
		WHERE id=$1
	`

	err := cg.db.Get(character, query, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCouldNotGet, err)
	}
	return character, nil
}

func (cg *PostgresCharacterGallery) getCustomizationByID(id characters.CharacterID) (*characters.Customization, error) {
	customization := &characters.Customization{}
	query := `
			SELECT * FROM customizations
			WHERE id = $1
		`

	err := cg.db.Get(customization, query, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCouldNotGet, err)
	}

	return customization, nil
}

func (cg *PostgresCharacterGallery) getStatsByID(id characters.CharacterID) (*characters.Stats, error) {
	stats := &characters.Stats{}
	query := `
			SELECT * FROM stats
			WHERE id = $1
		`

	err := cg.db.Get(stats, query, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCouldNotGet, err)
	}

	return stats, nil
}

func (cg *PostgresCharacterGallery) updateBaseCharacters(tx *sqlx.Tx, character *characters.Character) error {
	query := `
		UPDATE characters
		SET name = :name,
			body_type = :body_type,
			species = :species,
			class = :class
		WHERE id = :id
	`

	_, err := tx.NamedExec(query, character)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCouldNotFind, err)
	}

	return nil
}

func (cg *PostgresCharacterGallery) updateCustomization(tx *sqlx.Tx, customization *characters.Customization) error {
	query := `
		UPDATE customizations
		SET hair = :hair,
			face = :face,
			shirt = :shirt,
			pants = :pants,
			shoes = :shoes
		WHERE id = :id
	`

	_, err := tx.NamedExec(query, customization)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCouldNotFind, err)
	}

	return nil
}

func (cg *PostgresCharacterGallery) updateStats(tx *sqlx.Tx, stats *characters.Stats) error {
	query := `
		UPDATE stats
		SET strength = :strength,
			dexterity = :dexterity,
			constitution = :constitution,
			intelligence = :intelligence,
			wisdom = :wisdom,
			charisma = :charisma
		WHERE id = :id
	`

	_, err := tx.NamedExec(query, stats)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCouldNotFind, err)
	}

	return nil
}

/*
 *
 * Item related queries
 *
 */

func (cg *PostgresCharacterGallery) seedItemPool(tx *sqlx.Tx, item *inventory.Item) error {
	query := `
	INSERT INTO items (id, name, type, description, equippable, rarity, damage, defense, heal_amount, mana_cost, duration, cooldown, capacity)
	VALUES (:id, :name, :type, :description, :equippable, :rarity, :damage, :defense, :heal_amount, :mana_cost, :duration, :cooldown, :capacity)
	ON CONFLICT (id) DO UPDATE SET
		name = EXCLUDED.name,
		type = EXCLUDED.type,
		description = EXCLUDED.description,
		equippable = EXCLUDED.equippable,
		rarity = EXCLUDED.rarity,
		damage = EXCLUDED.damage,
		defense = EXCLUDED.defense,
		heal_amount = EXCLUDED.heal_amount,
		mana_cost = EXCLUDED.mana_cost,
		duration = EXCLUDED.duration,
		cooldown = EXCLUDED.cooldown;
	`

	_, err := tx.NamedExec(query, item)

	if err != nil {
		fmt.Println("NO PUDE CARGAR NADA MACHO")
		return fmt.Errorf("could not add item to database: %v", err)
	}

	return nil
}

func (cg *PostgresCharacterGallery) insertIntoItemPool(tx *sqlx.Tx, item *inventory.Item) error {
	query := `
	INSERT INTO items (name, type, description, equippable, rarity, damage, defense, heal_amount, mana_cost, duration, capacity)
	VALUES (:name, :type, :description, :equippable, :rarity, :damage, :defense, :heal_amount, :mana_cost, :duration, :capacity)
	RETURNING id;
	`

	stmt, err := tx.PrepareNamed(query)
	if err != nil {
		return fmt.Errorf("could not prepare statement: %w", err)
	}
	err = stmt.Get(&item.ID, item)
	if err != nil {
		return fmt.Errorf("could not insert item (duplicate?): %w", err)
	}

	return nil
}

func insertIntoCharacterInventory(tx *sqlx.Tx, characterID characters.CharacterID, itemID inventory.ItemID, quantity uint8) error {
	selectQuery := `
		SELECT * FROM inventory WHERE item_id = $1 AND character_id = $2;
	`
	rows, err := tx.Query(selectQuery, itemID, characterID)
	if err != nil {
		return err
	}

	// If the item already exists for the character, update the quantity
	if rows.Next() {
		rows.Close()
		updateQuery := `
			UPDATE inventory
			SET quantity = quantity + $1
			WHERE character_id = $2 AND item_id = $3;
		`
		_, err = tx.Exec(updateQuery, quantity, characterID, itemID)
		if err != nil {
			return err
		}
		return nil
	}
	rows.Close()

	query := `
		INSERT INTO inventory (character_id, item_id, quantity, is_equipped)
		VALUES ($1, $2, $3, FALSE);
	`

	_, err = tx.Exec(query, characterID, itemID, quantity)
	if err != nil {
		return err
	}
	return nil
}

func (cg *PostgresCharacterGallery) selectCurrentQuantity(characterID characters.CharacterID, itemID inventory.ItemID) (uint8, error) {
	querySelect := `
		SELECT quantity FROM inventory
		WHERE character_id = $1 AND item_id = $2;
	`

	var currentQuantity uint8
	err := cg.db.QueryRow(querySelect, characterID, itemID).Scan(&currentQuantity)
	if err != nil {
		return 0, err
	}
	return currentQuantity, nil
}

func updateItemQuantity(tx *sqlx.Tx, quantity uint8, characterID characters.CharacterID, itemID inventory.ItemID) error {
	queryUpdate := `
			UPDATE inventory
			SET quantity = quantity - $1
			WHERE character_id = $2 AND item_id = $3;
		`
	_, err := tx.Exec(queryUpdate, quantity, characterID, itemID)
	if err != nil {
		return err
	}
	return nil
}

func deleteItemFromCharacter(tx *sqlx.Tx, characterID characters.CharacterID, itemID inventory.ItemID) error {
	queryDelete := `
			DELETE FROM inventory
			WHERE character_id = $1 AND item_id = $2;
		`
	_, err := tx.Exec(queryDelete, characterID, itemID)
	if err != nil {
		return err
	}
	return nil
}
