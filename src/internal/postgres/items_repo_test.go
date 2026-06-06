package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"dZev1/character-gallery/internal/postgres/db"
)

func TestItemRepo_CreateItem(t *testing.T) {
	q := newTxQueries(t)
	ctx := context.Background()

	item, err := q.CreateItem(ctx, db.CreateItemParams{
		Name:        "Health Potion",
		Type:        "potion",
		Description: "Restores 50 HP",
		Equippable:  false,
		Rarity:      1,
		HealAmount:  pgtype.Int4{Int32: 50, Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}

	if item.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if item.Name != "Health Potion" {
		t.Errorf("got name %q, want %q", item.Name, "Health Potion")
	}
}

func TestItemRepo_CreateItem_AllFields(t *testing.T) {
	q := newTxQueries(t)
	ctx := context.Background()

	item, err := q.CreateItem(ctx, db.CreateItemParams{
		Name:        "Flame Sword",
		Type:        "weapon",
		Description: "A sword that burns",
		Equippable:  true,
		Rarity:      4,
		Damage:      pgtype.Int4{Int32: 25, Valid: true},
		Defense:     pgtype.Int4{},
		HealAmount:  pgtype.Int4{},
		ManaCost:    pgtype.Int4{Int32: 5, Valid: true},
		Duration:    pgtype.Int4{Int32: 60, Valid: true},
		Cooldown:    pgtype.Int4{Int32: 10, Valid: true},
		Capacity:    pgtype.Int4{},
	})
	if err != nil {
		t.Fatal(err)
	}

	if item.Damage.Int32 != 25 {
		t.Errorf("got damage %d, want 25", item.Damage.Int32)
	}
	if !item.Equippable {
		t.Error("expected equippable to be true")
	}
}

func TestItemRepo_FindItem(t *testing.T) {
	q := newTxQueries(t)
	ctx := context.Background()
	created := createTestItem(t, q)

	got, err := q.FindItem(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != created.Name {
		t.Errorf("got name %q, want %q", got.Name, created.Name)
	}
}

func TestItemRepo_FindItem_NotFound(t *testing.T) {
	q := newTxQueries(t)
	ctx := context.Background()

	_, err := q.FindItem(ctx, 99999)
	if err == nil {
		t.Fatal("expected error for non-existent item")
	}
}

func TestItemRepo_FindAllItems(t *testing.T) {
	q := newTxQueries(t)
	ctx := context.Background()

	createTestItem(t, q)
	createTestItem(t, q)

	items, err := q.FindAllItems(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Errorf("got %d items, want 2", len(items))
	}
}

func TestItemRepo_FindAllItems_Empty(t *testing.T) {
	q := newTxQueries(t)
	ctx := context.Background()

	items, err := q.FindAllItems(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Errorf("expected no items, got %d", len(items))
	}
}

func TestItemRepo_SeedItems(t *testing.T) {
	q := newTxQueries(t)
	ctx := context.Background()

	params := []db.SeedItemsParams{
		{
			Name:        "Wooden Shield",
			Type:        "shield",
			Description: "A basic wooden shield",
			Equippable:  true,
			Rarity:      1,
			Defense:     pgtype.Int4{Int32: 5, Valid: true},
		},
		{
			Name:        "Silk Robe",
			Type:        "armor",
			Description: "Light and elegant",
			Equippable:  true,
			Rarity:      2,
			Defense:     pgtype.Int4{Int32: 2, Valid: true},
		},
	}

	n, err := q.SeedItems(ctx, params)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("expected 2 rows copied, got %d", n)
	}

	items, err := q.FindAllItems(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items after seed, got %d", len(items))
	}
}
