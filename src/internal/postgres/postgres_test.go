package postgres

import (
	"context"
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/inventory"
	"dZev1/character-gallery/internal/items"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"dZev1/character-gallery/internal/postgres/db"
)

const (
	IRON_SWORD   string = "IRON_SWORD"
	WOODEN_SWORD string = "WOODEN_SWORD"
)

type testDB struct {
	pool *pgxpool.Pool
}

var globalDB *testDB

func TestMain(m *testing.M) {
	dsn := os.Getenv("POSTGRES_TEST_DATABASE_URL")
	if dsn == "" {
		fmt.Println("SKIP: POSTGRES_TEST_DATABASE_URL not set")
		os.Exit(0)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect: %v\n", err)
		os.Exit(1)
	}

	if _, err := pool.Exec(ctx, `
		DROP TABLE IF EXISTS customizations CASCADE;
		DROP TABLE IF EXISTS stats CASCADE;
		DROP TABLE IF EXISTS inventory CASCADE;
		DROP TABLE IF EXISTS items CASCADE;
		DROP TABLE IF EXISTS characters CASCADE;
		DROP TABLE IF EXISTS users CASCADE;
	`); err != nil {
		fmt.Fprintf(os.Stderr, "failed to drop tables: %v\n", err)
		os.Exit(1)
	}

	schema, err := os.ReadFile("../../database/schema/schema.sql")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read schema: %v\n", err)
		os.Exit(1)
	}
	if _, err := pool.Exec(ctx, string(schema)); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run schema: %v\n", err)
		os.Exit(1)
	}

	globalDB = &testDB{pool: pool}

	code := m.Run()

	pool.Close()
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
	tx, err := globalDB.pool.Begin(ctx)
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

func newCharacterRepo(t *testing.T) *characterRepo {
	t.Helper()
	return NewCharacterRepo(globalDB.pool)
}

func newItemRepo(t *testing.T) *itemRepo {
	t.Helper()
	return NewItemRepo(globalDB.pool)
}

func newInventoryRepo(t *testing.T) *inventoryRepo {
	t.Helper()
	return NewInventoryRepo(globalDB.pool)
}

func domainCharacter(char *db.Character, stats *db.Stat, cust *db.Customization) *characters.Character {
	return &characters.Character{
		ID:       characters.CharacterID(char.ID),
		Name:     char.Name,
		BodyType: characters.BodyType(char.BodyType),
		Species:  characters.Species(char.Species),
		Class:    characters.Class(char.Class),
		Level:    uint8(char.Level),
		Xp:       uint64(char.Xp),
		HpMax:    uint8(char.HpMax),
		HpCurrent: uint8(char.HpCurrent),
		Stats: &characters.Stats{
			ID:           characters.CharacterID(stats.CharacterID),
			Strength:     uint8(stats.Strength),
			Dexterity:    uint8(stats.Dexterity),
			Constitution: uint8(stats.Constitution),
			Intelligence: uint8(stats.Intelligence),
			Wisdom:       uint8(stats.Wisdom),
			Charisma:     uint8(stats.Charisma),
		},
		Customization: &characters.Customization{
			ID:    characters.CharacterID(cust.CharacterID),
			Hair:  uint8(cust.Hair),
			Face:  uint8(cust.Face),
			Shirt: uint8(cust.Shirt),
			Pants: uint8(cust.Pants),
			Shoes: uint8(cust.Shoes),
		},
	}
}

func domainItem(it *db.Item) *items.Item {
	return &items.Item{
		ID:          items.ItemID(it.ID),
		Name:        it.Name,
		Type:        items.Type(it.Type),
		Description: it.Description,
		Equippable:  it.Equippable,
		Rarity:      uint8(it.Rarity),
		Damage:      toU64ptr(it.Damage),
		Defense:     toU64ptr(it.Defense),
		HealAmount:  toU64ptr(it.HealAmount),
		ManaCost:    toU64ptr(it.ManaCost),
		Duration:    toU64ptr(it.Duration),
		Cooldown:    toU64ptr(it.Cooldown),
		Capacity:    toU64ptr(it.Capacity),
	}
}

func domainInventoryParam(charID int64, itemID int64, quantity int32) inventory.RepositoryParam {
	return inventory.RepositoryParam{
		CharacterID: characters.CharacterID(charID),
		ItemID:      items.ItemID(itemID),
		Quantity:    uint8(quantity),
	}
}
