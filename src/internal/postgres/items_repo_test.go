package postgres

import (
	"context"
	"testing"

	"dZev1/character-gallery/internal/items"
)

func TestItemRepo_SaveAndFindItem(t *testing.T) {
	tt := newTestTx(t)
	repo := newItemRepo(t)
	ctx := context.Background()

	item := &items.Item{
		Name:        "Health Potion",
		Type:        items.Potion,
		Description: "Restores 50 HP",
		Equippable:  false,
		Rarity:      1,
		HealAmount:  ptr[uint64](50),
	}

	saved, err := repo.SaveItem(ctx, tt.Tx, item)
	if err != nil {
		t.Fatal(err)
	}
	if saved.ID == 0 {
		t.Fatal("expected non-zero ID")
	}

	got, err := repo.FindItem(ctx, tt.Tx, saved.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Health Potion" {
		t.Errorf("got name %q, want %q", got.Name, "Health Potion")
	}
	if got.HealAmount == nil || *got.HealAmount != 50 {
		t.Errorf("got heal_amount %v, want 50", got.HealAmount)
	}
}

func TestItemRepo_SaveItem_AllFields(t *testing.T) {
	tt := newTestTx(t)
	repo := newItemRepo(t)
	ctx := context.Background()

	item := &items.Item{
		Name:        "Flame Sword",
		Type:        items.Weapon,
		Description: "A sword that burns",
		Equippable:  true,
		Rarity:      4,
		Damage:      ptr[uint64](25),
		ManaCost:    ptr[uint64](5),
		Duration:    ptr[uint64](60),
		Cooldown:    ptr[uint64](10),
	}

	saved, err := repo.SaveItem(ctx, tt.Tx, item)
	if err != nil {
		t.Fatal(err)
	}

	got, err := repo.FindItem(ctx, tt.Tx, saved.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Damage == nil || *got.Damage != 25 {
		t.Errorf("got damage %v, want 25", got.Damage)
	}
	if !got.Equippable {
		t.Error("expected equippable to be true")
	}
}

func TestItemRepo_FindItem_NotFound(t *testing.T) {
	tt := newTestTx(t)
	repo := newItemRepo(t)
	ctx := context.Background()

	_, err := repo.FindItem(ctx, tt.Tx, 99999)
	if err == nil {
		t.Fatal("expected error for non-existent item")
	}
}

func TestItemRepo_FindAllItems(t *testing.T) {
	tt := newTestTx(t)
	repo := newItemRepo(t)
	ctx := context.Background()

	if _, err := repo.SaveItem(ctx, tt.Tx, &items.Item{
		Name: "Iron Sword", Type: items.Weapon, Description: "A sturdy iron sword",
		Equippable: true, Rarity: 2,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.SaveItem(ctx, tt.Tx, &items.Item{
		Name: "Wooden Sword", Type: items.Weapon, Description: "A wooden sword",
		Equippable: true, Rarity: 3,
	}); err != nil {
		t.Fatal(err)
	}

	all, err := repo.FindAllItems(ctx, tt.Tx)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Errorf("got %d items, want 2", len(all))
	}
}

func TestItemRepo_FindAllItems_Empty(t *testing.T) {
	tt := newTestTx(t)
	repo := newItemRepo(t)
	ctx := context.Background()

	all, err := repo.FindAllItems(ctx, tt.Tx)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 0 {
		t.Errorf("expected no items, got %d", len(all))
	}
}

func TestItemRepo_SeedItems(t *testing.T) {
	tt := newTestTx(t)
	repo := newItemRepo(t)
	ctx := context.Background()

	seeds := []items.Item{
		{
			Name: "Wooden Shield", Type: items.Shield, Description: "A basic wooden shield",
			Equippable: true, Rarity: 1, Defense: ptr[uint64](5),
		},
		{
			Name: "Silk Robe", Type: items.Armor, Description: "Light and elegant",
			Equippable: true, Rarity: 2, Defense: ptr[uint64](2),
		},
	}

	if err := repo.SeedItems(ctx, tt.Tx, seeds); err != nil {
		t.Fatal(err)
	}

	all, err := repo.FindAllItems(ctx, tt.Tx)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 items after seed, got %d", len(all))
	}
}

func ptr[T any](v T) *T {
	return &v
}
