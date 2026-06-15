package postgres

import (
	"context"
	"dZev1/character-gallery/internal/items"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"

	"dZev1/character-gallery/internal/postgres/db"
)

const (
	IRON_SWORD   string = "IRON_SWORD"
	WOODEN_SWORD string = "WOODEN_SWORD"
)

type testDB struct {
	conn *pgx.Conn
}

var globalDB *testDB

func TestMain(m *testing.M) {
	dsn := os.Getenv("POSTGRES_TEST_DATABASE_URL")
	if dsn == "" {
		fmt.Println("SKIP: POSTGRES_TEST_DATABASE_URL not set")
		os.Exit(0)
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect: %v\n", err)
		os.Exit(1)
	}

	schema, err := os.ReadFile("../../database/schema/schema.sql")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read schema: %v\n", err)
		os.Exit(1)
	}
	if _, err := conn.Exec(ctx, string(schema)); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run schema: %v\n", err)
		os.Exit(1)
	}

	globalDB = &testDB{conn: conn}

	code := m.Run()

	conn.Close(ctx)
	os.Exit(code)
}

type testTx struct {
	Q  *db.Queries
	Tx pgx.Tx
}

func newTxQueries(t *testing.T) *db.Queries {
	t.Helper()
	return newTestTx(t).Q
}

func newTestTx(t *testing.T) *testTx {
	t.Helper()
	ctx := context.Background()
	tx, err := globalDB.conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = tx.Rollback(ctx)
	})
	return &testTx{
		Q:  db.New(tx),
		Tx: tx,
	}
}

func createTestCharacter(t *testing.T, q *db.Queries) *db.Character {
	t.Helper()
	ctx := context.Background()
	char, err := q.CreateCharacter(ctx, db.CreateCharacterParams{
		Name:     "Test Hero",
		BodyType: "type_a",
		Species:  "human",
		Class:    "fighter",
	})
	if err != nil {
		t.Fatal(err)
	}
	return char
}

func createTestCharacterFull(t *testing.T, q *db.Queries) *db.Character {
	t.Helper()
	char := createTestCharacter(t, q)
	ctx := context.Background()

	_, err := q.CreateStats(ctx, db.CreateStatsParams{
		CharacterID:  char.ID,
		Strength:     15,
		Dexterity:    14,
		Constitution: 13,
		Intelligence: 12,
		Wisdom:       11,
		Charisma:     10,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = q.CreateCustomization(ctx, db.CreateCustomizationParams{
		CharacterID: char.ID,
		Hair:        5,
		Face:        3,
		Shirt:       12,
		Pants:       8,
		Shoes:       2,
	})
	if err != nil {
		t.Fatal(err)
	}

	return char
}

func createTestItem(t *testing.T, q *db.Queries, it string) *db.Item {
	t.Helper()
	ctx := context.Background()

	var params db.CreateItemParams

	switch it {
	case IRON_SWORD:
		params = db.CreateItemParams{
			Name:        "Iron Sword",
			Type:        "weapon",
			Description: "A sturdy iron sword",
			Equippable:  true,
			Rarity:      2,
		}
	case WOODEN_SWORD:
		params = db.CreateItemParams{
			Name:        "Wooden Sword",
			Type:        "weapon",
			Description: "A stuffy wood sword", // Nota: copiaste la descripción de la iron sword
			Equippable:  true,
			Rarity:      3,
		}
	default:
		params = db.CreateItemParams{
			Name:        "Spell of Wisdom",
			Type:        string(items.Scroll),
			Description: "A magical scroll",
			Equippable:  true,
			Rarity:      3,
		}
	}

	item, err := q.CreateItem(ctx, params)
	if err != nil {
		t.Fatal(err)
	}

	return item
}
