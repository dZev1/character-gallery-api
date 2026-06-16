package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/items"
	"dZev1/character-gallery/internal/postgres/db"
)

func TestInventoryRepo_AddItemToCharacter(t *testing.T) {
	tt := newTestTx(t)
	repo := newInventoryRepo(t)
	ctx := context.Background()
	q := tt.Q

	char := createTestCharacter(t, q)
	item := createTestItem(t, q, IRON_SWORD)

	inv, err := repo.AddItemToCharacter(ctx, tt.Tx, domainInventoryParam(char.ID, item.ID, 3))
	if err != nil {
		t.Fatal(err)
	}

	if inv.Quantity != 3 {
		t.Errorf("got quantity %d, want 3", inv.Quantity)
	}
	if inv.IsEquipped {
		t.Error("expected is_equipped to be false")
	}
}

func TestInventoryRepo_AddItemToCharacter_Upsert(t *testing.T) {
	tt := newTestTx(t)
	repo := newInventoryRepo(t)
	ctx := context.Background()
	q := tt.Q

	char := createTestCharacter(t, q)
	item := createTestItem(t, q, IRON_SWORD)

	_, err := q.AddItemToCharacter(ctx, db.AddItemToCharacterParams{
		CharacterID: char.ID,
		ItemID:      item.ID,
		Quantity:    2,
	})
	if err != nil {
		t.Fatal(err)
	}

	inv, err := repo.AddItemToCharacter(ctx, tt.Tx, domainInventoryParam(char.ID, item.ID, 3))
	if err != nil {
		t.Fatal(err)
	}
	if inv.Quantity != 5 {
		t.Errorf("expected quantity 2+3=5, got %d", inv.Quantity)
	}
}

func TestInventoryRepo_GetCharacterInventory_Empty(t *testing.T) {
	tt := newTestTx(t)
	repo := newInventoryRepo(t)
	ctx := context.Background()
	q := tt.Q

	char := createTestCharacter(t, q)

	items, err := repo.GetCharacterInventory(ctx, tt.Tx, characters.CharacterID(char.ID))
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Errorf("expected no items, got %d", len(items))
	}
}

func TestInventoryRepo_GetCharacterInventory(t *testing.T) {
	tt := newTestTx(t)
	repo := newInventoryRepo(t)
	ctx := context.Background()
	q := tt.Q

	char := createTestCharacter(t, q)
	item1 := createTestItem(t, q, IRON_SWORD)

	item2, err := q.CreateItem(ctx, db.CreateItemParams{
		Name:        "Leather Armor",
		Type:        "armor",
		Description: "Light armor",
		Equippable:  true,
		Rarity:      1,
		Defense:     defaultInt4(5),
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = item2

	_, err = q.AddItemToCharacter(ctx, db.AddItemToCharacterParams{
		CharacterID: char.ID, ItemID: item1.ID, Quantity: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = q.AddItemToCharacter(ctx, db.AddItemToCharacterParams{
		CharacterID: char.ID, ItemID: item2.ID, Quantity: 2,
	})
	if err != nil {
		t.Fatal(err)
	}

	inventoryItems, err := repo.GetCharacterInventory(ctx, tt.Tx, characters.CharacterID(char.ID))
	if err != nil {
		t.Fatal(err)
	}
	if len(inventoryItems) != 2 {
		t.Fatalf("expected 2 items, got %d", len(inventoryItems))
	}

	found := false
	for _, it := range inventoryItems {
		if it.Item.ID == items.ItemID(item1.ID) {
			found = true
			if it.Quantity != 1 {
				t.Errorf("item %d: got quantity %d, want 1", item1.ID, it.Quantity)
			}
		}
	}
	if !found {
		t.Errorf("item %d not found in inventory", item1.ID)
	}
}

func TestInventoryRepo_GetCharacterInventory_OtherCharacter(t *testing.T) {
	tt := newTestTx(t)
	repo := newInventoryRepo(t)
	ctx := context.Background()
	q := tt.Q

	char1 := createTestCharacter(t, q)
	char2 := createTestCharacter(t, q)
	item := createTestItem(t, q, IRON_SWORD)

	_, err := q.AddItemToCharacter(ctx, db.AddItemToCharacterParams{
		CharacterID: char1.ID, ItemID: item.ID, Quantity: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	inventoryItems, err := repo.GetCharacterInventory(ctx, tt.Tx, characters.CharacterID(char2.ID))
	if err != nil {
		t.Fatal(err)
	}
	if len(inventoryItems) != 0 {
		t.Errorf("expected 0 items for other character, got %d", len(inventoryItems))
	}
}

func TestInventoryRepo_RemoveItemFromCharacter_Partial(t *testing.T) {
	tt := newTestTx(t)
	repo := newInventoryRepo(t)
	ctx := context.Background()
	q := tt.Q

	char := createTestCharacter(t, q)
	item := createTestItem(t, q, IRON_SWORD)

	_, err := q.AddItemToCharacter(ctx, db.AddItemToCharacterParams{
		CharacterID: char.ID, ItemID: item.ID, Quantity: 5,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = repo.RemoveItemFromCharacter(ctx, tt.Tx, domainInventoryParam(char.ID, item.ID, 3))
	if err != nil {
		t.Fatal(err)
	}

	inventoryItems, err := repo.GetCharacterInventory(ctx, tt.Tx, characters.CharacterID(char.ID))
	if err != nil {
		t.Fatal(err)
	}
	if len(inventoryItems) != 1 {
		t.Fatalf("expected 1 item, got %d", len(inventoryItems))
	}
	if inventoryItems[0].Quantity != 2 {
		t.Errorf("expected quantity 2, got %d", inventoryItems[0].Quantity)
	}
}

func TestInventoryRepo_RemoveItemFromCharacter_Full(t *testing.T) {
	tt := newTestTx(t)
	repo := newInventoryRepo(t)
	ctx := context.Background()
	q := tt.Q

	char := createTestCharacter(t, q)
	item := createTestItem(t, q, IRON_SWORD)

	_, err := q.AddItemToCharacter(ctx, db.AddItemToCharacterParams{
		CharacterID: char.ID, ItemID: item.ID, Quantity: 2,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = repo.RemoveItemFromCharacter(ctx, tt.Tx, domainInventoryParam(char.ID, item.ID, 2))
	if err != nil {
		t.Fatal(err)
	}

	inventoryItems, err := repo.GetCharacterInventory(ctx, tt.Tx, characters.CharacterID(char.ID))
	if err != nil {
		t.Fatal(err)
	}
	if len(inventoryItems) != 0 {
		t.Errorf("expected 0 items after full removal, got %d", len(inventoryItems))
	}
}

func TestInventoryRepo_RemoveMoreThanOwned(t *testing.T) {
	tt := newTestTx(t)
	repo := newInventoryRepo(t)
	ctx := context.Background()
	q := tt.Q

	char := createTestCharacter(t, q)
	item := createTestItem(t, q, IRON_SWORD)

	_, err := q.AddItemToCharacter(ctx, db.AddItemToCharacterParams{
		CharacterID: char.ID, ItemID: item.ID, Quantity: 2,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = repo.RemoveItemFromCharacter(ctx, tt.Tx, domainInventoryParam(char.ID, item.ID, 5))
	if err == nil {
		t.Fatal("expected error when removing more than owned, got nil")
	}

	inventoryItems, err := repo.GetCharacterInventory(ctx, tt.Tx, characters.CharacterID(char.ID))
	if err != nil {
		t.Fatal(err)
	}
	if len(inventoryItems) != 1 {
		t.Fatalf("expected 1 item (unchanged), got %d", len(inventoryItems))
	}
	if inventoryItems[0].Quantity != 2 {
		t.Errorf("expected quantity 2, got %d", inventoryItems[0].Quantity)
	}
}

func TestInventoryRepo_DeleteCharacter_CascadesToInventory(t *testing.T) {
	tt := newTestTx(t)
	repo := newInventoryRepo(t)
	ctx := context.Background()
	q := tt.Q

	char := createTestCharacter(t, q)
	item := createTestItem(t, q, IRON_SWORD)

	_, err := q.AddItemToCharacter(ctx, db.AddItemToCharacterParams{
		CharacterID: char.ID, ItemID: item.ID, Quantity: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = q.DeleteCharacter(ctx, char.ID)
	if err != nil {
		t.Fatal(err)
	}

	inventoryItems, err := repo.GetCharacterInventory(ctx, tt.Tx, characters.CharacterID(char.ID))
	if err != nil {
		t.Fatal(err)
	}
	if len(inventoryItems) != 0 {
		t.Errorf("expected 0 inventory items after character delete, got %d", len(inventoryItems))
	}
}

func defaultInt4(n int32) pgtype.Int4 {
	return pgtype.Int4{Int32: n, Valid: true}
}
