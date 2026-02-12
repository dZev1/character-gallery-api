package postgres_gallery

import (
	"errors"
	"testing"

	"dZev1/character-gallery/models/characters"
	"dZev1/character-gallery/models/inventory"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestDisplayPoolItems_Success(t *testing.T) {
	gallery, mock := setupMockDB(t)

	rows := sqlmock.NewRows([]string{
		"id", "name", "type", "description", "equippable", "rarity",
		"damage", "defense", "heal_amount", "mana_cost", "duration", "cooldown", "capacity",
	}).
		AddRow(1, "Sword", "weapon", "A sharp sword", true, 3, 50, nil, nil, nil, nil, nil, nil).
		AddRow(2, "Healing Potion", "potion", "Restores health", false, 1, nil, nil, 60, nil, nil, nil, nil)

	mock.ExpectQuery(`SELECT \*`).WillReturnRows(rows)

	items, err := gallery.DisplayPoolItems()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
	if items[0].Name != "Sword" {
		t.Errorf("expected Sword, got %s", items[0].Name)
	}
	if items[1].Name != "Healing Potion" {
		t.Errorf("expected Healing Potion, got %s", items[1].Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestDisplayPoolItems_Empty(t *testing.T) {
	gallery, mock := setupMockDB(t)

	rows := sqlmock.NewRows([]string{
		"id", "name", "type", "description", "equippable", "rarity",
		"damage", "defense", "heal_amount", "mana_cost", "duration", "cooldown", "capacity",
	})

	mock.ExpectQuery(`SELECT \*`).WillReturnRows(rows)

	items, err := gallery.DisplayPoolItems()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestDisplayPoolItems_Error(t *testing.T) {
	gallery, mock := setupMockDB(t)

	mock.ExpectQuery(`SELECT \*`).WillReturnError(errors.New("db error"))

	items, err := gallery.DisplayPoolItems()

	if err == nil {
		t.Error("expected error, got nil")
	}
	if items != nil {
		t.Error("expected nil items on error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestDisplayItem_Success(t *testing.T) {
	gallery, mock := setupMockDB(t)

	itemID := inventory.ItemID(1)

	rows := sqlmock.NewRows([]string{
		"id", "name", "type", "description", "equippable", "rarity",
		"damage", "defense", "heal_amount", "mana_cost", "duration", "cooldown", "capacity",
	}).AddRow(1, "Sword", "weapon", "A sharp sword", true, 3, 50, nil, nil, nil, nil, nil, nil)

	mock.ExpectQuery(`SELECT \*`).WithArgs(itemID).WillReturnRows(rows)

	item, err := gallery.DisplayItem(itemID)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if item == nil {
		t.Fatal("expected item, got nil")
	}
	if item.Name != "Sword" {
		t.Errorf("expected Sword, got %s", item.Name)
	}
	if item.Type != inventory.Weapon {
		t.Errorf("expected weapon type, got %s", item.Type)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestDisplayItem_NotFound(t *testing.T) {
	gallery, mock := setupMockDB(t)

	itemID := inventory.ItemID(999)

	mock.ExpectQuery(`SELECT \*`).WithArgs(itemID).WillReturnError(errors.New("no rows"))

	item, err := gallery.DisplayItem(itemID)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if item != nil {
		t.Error("expected nil item")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetCharacterInventory_Success(t *testing.T) {
	gallery, mock := setupMockDB(t)

	charID := characters.CharacterID(1)

	rows := sqlmock.NewRows([]string{
		"item.id", "item.name", "item.type", "item.description", "item.equippable", "item.rarity",
		"item.damage", "item.defense", "item.heal_amount", "item.mana_cost", "item.duration",
		"quantity", "is_equipped",
	}).
		AddRow(1, "Sword", "weapon", "A sharp sword", true, 3, 50, nil, nil, nil, nil, 1, true).
		AddRow(2, "Potion", "potion", "Heals", false, 1, nil, nil, 60, nil, nil, 5, false)

	mock.ExpectQuery(`SELECT`).WithArgs(charID).WillReturnRows(rows)

	inv, err := gallery.GetCharacterInventory(charID)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(inv) != 2 {
		t.Errorf("expected 2 items, got %d", len(inv))
	}
	if inv[0].Quantity != 1 {
		t.Errorf("expected quantity 1, got %d", inv[0].Quantity)
	}
	if !inv[0].IsEquipped {
		t.Error("expected first item to be equipped")
	}
	if inv[1].Quantity != 5 {
		t.Errorf("expected quantity 5, got %d", inv[1].Quantity)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetCharacterInventory_Empty(t *testing.T) {
	gallery, mock := setupMockDB(t)

	charID := characters.CharacterID(1)

	rows := sqlmock.NewRows([]string{
		"item.id", "item.name", "item.type", "item.description", "item.equippable", "item.rarity",
		"item.damage", "item.defense", "item.heal_amount", "item.mana_cost", "item.duration",
		"quantity", "is_equipped",
	})

	mock.ExpectQuery(`SELECT`).WithArgs(charID).WillReturnRows(rows)

	inv, err := gallery.GetCharacterInventory(charID)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(inv) != 0 {
		t.Errorf("expected 0 items, got %d", len(inv))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetCharacterInventory_Error(t *testing.T) {
	gallery, mock := setupMockDB(t)

	charID := characters.CharacterID(1)

	mock.ExpectQuery(`SELECT`).WithArgs(charID).WillReturnError(errors.New("db error"))

	inv, err := gallery.GetCharacterInventory(charID)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if !errors.Is(err, ErrFailedSelectCharacterInventory) {
		t.Errorf("expected ErrFailedSelectCharacterInventory, got %v", err)
	}
	if inv != nil {
		t.Error("expected nil inventory on error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestAddItemToCharacter_Success(t *testing.T) {
	gallery, mock := setupMockDB(t)

	charID := characters.CharacterID(1)
	itemID := inventory.ItemID(1)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM inventory`).
		WithArgs(itemID, charID).
		WillReturnRows(sqlmock.NewRows([]string{"character_id", "item_id", "quantity", "is_equipped"}))
	mock.ExpectExec(`INSERT INTO inventory`).
		WithArgs(charID, itemID, uint8(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := gallery.AddItemToCharacter(charID, itemID, 1)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestAddItemToCharacter_UpdateExisting(t *testing.T) {
	gallery, mock := setupMockDB(t)

	charID := characters.CharacterID(1)
	itemID := inventory.ItemID(1)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM inventory`).
		WithArgs(itemID, charID).
		WillReturnRows(sqlmock.NewRows([]string{"character_id", "item_id", "quantity", "is_equipped"}).
			AddRow(1, 1, 2, false))
	mock.ExpectExec(`UPDATE inventory`).
		WithArgs(uint8(3), charID, itemID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := gallery.AddItemToCharacter(charID, itemID, 3)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestAddItemToCharacter_TransactionError(t *testing.T) {
	gallery, mock := setupMockDB(t)

	charID := characters.CharacterID(1)
	itemID := inventory.ItemID(1)

	mock.ExpectBegin().WillReturnError(errors.New("tx error"))

	err := gallery.AddItemToCharacter(charID, itemID, 1)

	if err == nil {
		t.Error("expected error")
	}
	if !errors.Is(err, ErrFailedInitializeTransaction) {
		t.Errorf("expected ErrFailedInitializeTransaction, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRemoveItemFromCharacter_ReduceQuantity(t *testing.T) {
	gallery, mock := setupMockDB(t)

	charID := characters.CharacterID(1)
	itemID := inventory.ItemID(1)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT quantity FROM inventory`).
		WithArgs(charID, itemID).
		WillReturnRows(sqlmock.NewRows([]string{"quantity"}).AddRow(5))
	mock.ExpectExec(`UPDATE inventory`).
		WithArgs(uint8(2), charID, itemID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := gallery.RemoveItemFromCharacter(charID, itemID, 2)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRemoveItemFromCharacter_DeleteItem(t *testing.T) {
	gallery, mock := setupMockDB(t)

	charID := characters.CharacterID(1)
	itemID := inventory.ItemID(1)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT quantity FROM inventory`).
		WithArgs(charID, itemID).
		WillReturnRows(sqlmock.NewRows([]string{"quantity"}).AddRow(2))
	mock.ExpectExec(`DELETE FROM inventory`).
		WithArgs(charID, itemID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := gallery.RemoveItemFromCharacter(charID, itemID, 5)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRemoveItemFromCharacter_TransactionError(t *testing.T) {
	gallery, mock := setupMockDB(t)

	charID := characters.CharacterID(1)
	itemID := inventory.ItemID(1)

	mock.ExpectBegin().WillReturnError(errors.New("tx error"))

	err := gallery.RemoveItemFromCharacter(charID, itemID, 1)

	if err == nil {
		t.Error("expected error")
	}
	if !errors.Is(err, ErrFailedInitializeTransaction) {
		t.Errorf("expected ErrFailedInitializeTransaction, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
