package postgres

import (
	"context"
	"testing"

	"dZev1/character-gallery/internal/postgres/db"
)

func TestCharacterRepo_CreateCharacter(t *testing.T) {
	q := newTxQueries(t)
	ctx := context.Background()

	char, err := q.CreateCharacter(ctx, db.CreateCharacterParams{
		Name:     "Aragorn",
		BodyType: "type_a",
		Species:  "human",
		Class:    "ranger",
	})
	if err != nil {
		t.Fatal(err)
	}

	if char.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if char.Name != "Aragorn" {
		t.Errorf("got name %q, want %q", char.Name, "Aragorn")
	}
	if char.Species != "human" {
		t.Errorf("got species %q, want %q", char.Species, "human")
	}
	if char.Class != "ranger" {
		t.Errorf("got class %q, want %q", char.Class, "ranger")
	}
	if char.BodyType != "type_a" {
		t.Errorf("got body_type %q, want %q", char.BodyType, "type_a")
	}
}

func TestCharacterRepo_SelectCharacter(t *testing.T) {
	q := newTxQueries(t)
	ctx := context.Background()
	char := createTestCharacter(t, q)

	got, err := q.SelectCharacter(ctx, char.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != char.Name {
		t.Errorf("got name %q, want %q", got.Name, char.Name)
	}
}

func TestCharacterRepo_SelectCharacter_NotFound(t *testing.T) {
	q := newTxQueries(t)
	ctx := context.Background()

	_, err := q.SelectCharacter(ctx, 99999)
	if err == nil {
		t.Fatal("expected error for non-existent character")
	}
}

func TestCharacterRepo_SelectAllCharacters(t *testing.T) {
	q := newTxQueries(t)
	ctx := context.Background()

	createTestCharacter(t, q)
	createTestCharacter(t, q)

	chars, err := q.SelectAllCharacters(ctx, db.SelectAllCharactersParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatal(err)
	}
	if len(chars) != 2 {
		t.Errorf("got %d characters, want 2", len(chars))
	}
}

func TestCharacterRepo_SelectAllCharacters_Pagination(t *testing.T) {
	q := newTxQueries(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		createTestCharacter(t, q)
	}

	page1, err := q.SelectAllCharacters(ctx, db.SelectAllCharactersParams{Limit: 2, Offset: 0})
	if err != nil {
		t.Fatal(err)
	}
	if len(page1) != 2 {
		t.Errorf("page 1: got %d characters, want 2", len(page1))
	}

	page2, err := q.SelectAllCharacters(ctx, db.SelectAllCharactersParams{Limit: 2, Offset: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(page2) != 2 {
		t.Errorf("page 2: got %d characters, want 2", len(page2))
	}
}

func TestCharacterRepo_SelectAllCharacters_Empty(t *testing.T) {
	q := newTxQueries(t)
	ctx := context.Background()

	chars, err := q.SelectAllCharacters(ctx, db.SelectAllCharactersParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatal(err)
	}
	if len(chars) != 0 {
		t.Errorf("expected no characters, got %d", len(chars))
	}
}

func TestCharacterRepo_UpdateCharacter(t *testing.T) {
	q := newTxQueries(t)
	ctx := context.Background()
	char := createTestCharacter(t, q)

	updated, err := q.UpdateCharacter(ctx, db.UpdateCharacterParams{
		ID:       char.ID,
		Name:     "Aragorn II",
		BodyType: "type_b",
		Species:  "elf",
		Class:    "warlock",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Aragorn II" {
		t.Errorf("got name %q, want %q", updated.Name, "Aragorn II")
	}
	if updated.Class != "warlock" {
		t.Errorf("got class %q, want %q", updated.Class, "warlock")
	}

	got, err := q.SelectCharacter(ctx, char.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Aragorn II" {
		t.Errorf("select after update: got name %q, want %q", got.Name, "Aragorn II")
	}
}

func TestCharacterRepo_DeleteCharacter(t *testing.T) {
	q := newTxQueries(t)
	ctx := context.Background()
	char := createTestCharacter(t, q)

	err := q.DeleteCharacter(ctx, char.ID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = q.SelectCharacter(ctx, char.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestCharacterRepo_CountAllCharacters(t *testing.T) {
	q := newTxQueries(t)
	ctx := context.Background()

	count, err := q.CountAllCharacters(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected 0, got %d", count)
	}

	createTestCharacter(t, q)
	createTestCharacter(t, q)

	count, err = q.CountAllCharacters(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("expected 2, got %d", count)
	}
}

func TestCharacterRepo_CreateAndSelectStats(t *testing.T) {
	q := newTxQueries(t)
	ctx := context.Background()
	char := createTestCharacter(t, q)

	stats, err := q.CreateStats(ctx, db.CreateStatsParams{
		CharacterID:  char.ID,
		Strength:     18,
		Dexterity:    12,
		Constitution: 14,
		Intelligence: 10,
		Wisdom:       8,
		Charisma:     15,
	})
	if err != nil {
		t.Fatal(err)
	}
	if stats.CharacterID != char.ID {
		t.Errorf("got character_id %d, want %d", stats.CharacterID, char.ID)
	}
	if stats.Strength != 18 {
		t.Errorf("got strength %d, want 18", stats.Strength)
	}

	got, err := q.SelectCharacterStats(ctx, char.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Strength != 18 {
		t.Errorf("select: got strength %d, want 18", got.Strength)
	}
}

func TestCharacterRepo_CreateAndSelectCustomization(t *testing.T) {
	q := newTxQueries(t)
	ctx := context.Background()
	char := createTestCharacter(t, q)

	cust, err := q.CreateCustomization(ctx, db.CreateCustomizationParams{
		CharacterID: char.ID,
		Hair:        10,
		Face:        5,
		Shirt:       20,
		Pants:       15,
		Shoes:       3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if cust.CharacterID != char.ID {
		t.Errorf("got character_id %d, want %d", cust.CharacterID, char.ID)
	}

	got, err := q.SelectCharacterCustomization(ctx, char.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Hair != 10 || got.Shoes != 3 {
		t.Errorf(" customization mismatch: %+v", got)
	}
}

func TestCharacterRepo_UpdateStats(t *testing.T) {
	q := newTxQueries(t)
	ctx := context.Background()
	char := createTestCharacterFull(t, q)

	updated, err := q.UpdateStats(ctx, db.UpdateStatsParams{
		CharacterID:  char.ID,
		Strength:     1,
		Dexterity:    1,
		Constitution: 1,
		Intelligence: 1,
		Wisdom:       1,
		Charisma:     1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Strength != 1 {
		t.Errorf("got strength %d, want 1", updated.Strength)
	}

	got, err := q.SelectCharacterStats(ctx, char.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Dexterity != 1 {
		t.Errorf("select after update: got dexterity %d, want 1", got.Dexterity)
	}
}

func TestCharacterRepo_UpdateCustomization(t *testing.T) {
	q := newTxQueries(t)
	ctx := context.Background()
	char := createTestCharacterFull(t, q)

	updated, err := q.UpdateCustomization(ctx, db.UpdateCustomizationParams{
		CharacterID: char.ID,
		Hair:        0,
		Face:        0,
		Shirt:       0,
		Pants:       0,
		Shoes:       0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Hair != 0 {
		t.Errorf("got hair %d, want 0", updated.Hair)
	}

	got, err := q.SelectCharacterCustomization(ctx, char.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Shirt != 0 {
		t.Errorf("select after update: got shirt %d, want 0", got.Shirt)
	}
}

func TestCharacterRepo_DeleteCharacter_Cascades(t *testing.T) {
	q := newTxQueries(t)
	ctx := context.Background()
	char := createTestCharacterFull(t, q)

	err := q.DeleteCharacter(ctx, char.ID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = q.SelectCharacterStats(ctx, char.ID)
	if err == nil {
		t.Error("expected error when selecting stats for deleted character")
	}

	_, err = q.SelectCharacterCustomization(ctx, char.ID)
	if err == nil {
		t.Error("expected error when selecting customization for deleted character")
	}
}
