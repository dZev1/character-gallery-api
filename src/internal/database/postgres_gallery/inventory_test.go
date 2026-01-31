package postgres_gallery

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dZev1/character-gallery/models/characters"
	"github.com/jmoiron/sqlx"
)

func TestGetCharacterInventory_MapsNestedItemFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	cg := &PostgresCharacterGallery{db: sqlxDB}

	characterID := characters.CharacterID(1)

	rows := sqlmock.NewRows([]string{
		"item.id",
		"item.name",
		"item.type",
		"item.description",
		"item.equippable",
		"item.rarity",
		"item.damage",
		"item.defense",
		"item.heal_amount",
		"item.mana_cost",
		"item.duration",
		"quantity",
		"is_equipped",
	}).AddRow(
		int64(1),
		"Master Sword",
		"weapon",
		"A legendary sword with immense power.",
		true,
		int64(5),
		int64(100),
		nil,
		nil,
		nil,
		nil,
		int64(2),
		false,
	)

	// Match only the stable parts of the query to avoid brittleness due to whitespace/formatting.
	pattern := regexp.MustCompile(`(?s)FROM\s+items\s+i.*JOIN\s+inventory\s+ci\s+ON\s+ci\.item_id\s*=\s*i\.id.*WHERE\s+ci\.character_id\s*=\s*\$1`).String()
	mock.ExpectQuery(pattern).
		WithArgs(characterID).
		WillReturnRows(rows)

	inv, err := cg.GetCharacterInventory(characterID)
	if err != nil {
		t.Fatalf("GetCharacterInventory(): %v", err)
	}
	if len(inv) != 1 {
		t.Fatalf("expected 1 inventory item, got %d", len(inv))
	}
	if inv[0].Item == nil {
		t.Fatalf("expected inv[0].Item to be non-nil")
	}
	if inv[0].Item.ID != 1 {
		t.Fatalf("expected item ID 1, got %d", inv[0].Item.ID)
	}
	if inv[0].Quantity != 2 {
		t.Fatalf("expected quantity 2, got %d", inv[0].Quantity)
	}
	if inv[0].IsEquipped {
		t.Fatalf("expected is_equipped false")
	}
	if inv[0].Item.Damage == nil || *inv[0].Item.Damage != 100 {
		t.Fatalf("expected damage 100, got %#v", inv[0].Item.Damage)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}
