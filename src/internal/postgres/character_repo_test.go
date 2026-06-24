package postgres

import (
	"context"
	"testing"

	"dZev1/character-gallery/internal/characters"
)

func TestCharacterRepo_SaveAndFindCharacter(t *testing.T) {
	tt := newTestTx(t)
	repo := newCharacterRepo(t)
	ctx := context.Background()

	char := &characters.Character{
		Name:     "Aragorn",
		BodyType: "type_a",
		Species:  "human",
		Class:    "ranger",
		Stats: &characters.Stats{
			Strength:     15,
			Dexterity:    14,
			Constitution: 13,
			Intelligence: 12,
			Wisdom:       11,
			Charisma:     10,
		},
		Customization: &characters.Customization{
			Hair:  5,
			Face:  3,
			Shirt: 12,
			Pants: 8,
			Shoes: 2,
		},
	}

	saved, err := repo.SaveCharacter(ctx, tt.Tx, char)
	if err != nil {
		t.Fatal(err)
	}
	if saved.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if saved.Stats.ID != saved.ID {
		t.Errorf("stats ID %d does not match character ID %d", saved.Stats.ID, saved.ID)
	}
	if saved.Customization.ID != saved.ID {
		t.Errorf("customization ID %d does not match character ID %d", saved.Customization.ID, saved.ID)
	}

	got, err := repo.FindCharacter(ctx, tt.Tx, saved.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Aragorn" {
		t.Errorf("got name %q, want %q", got.Name, "Aragorn")
	}
	if got.Stats.Strength != 15 {
		t.Errorf("got strength %d, want 15", got.Stats.Strength)
	}
	if got.Customization.Hair != 5 {
		t.Errorf("got hair %d, want 5", got.Customization.Hair)
	}
}

func TestCharacterRepo_FindCharacter_NotFound(t *testing.T) {
	tt := newTestTx(t)
	repo := newCharacterRepo(t)
	ctx := context.Background()

	_, err := repo.FindCharacter(ctx, tt.Tx, 99999)
	if err == nil {
		t.Fatal("expected error for non-existent character")
	}
}

func TestCharacterRepo_FindAllCharacters(t *testing.T) {
	tt := newTestTx(t)
	repo := newCharacterRepo(t)
	ctx := context.Background()

	base := &characters.Character{
		BodyType: "type_a",
		Species:  "human",
		Class:    "fighter",
		Stats:    &characters.Stats{Strength: 10, Dexterity: 10, Constitution: 10, Intelligence: 10, Wisdom: 10, Charisma: 10},
		Customization: &characters.Customization{Hair: 1, Face: 1, Shirt: 1, Pants: 1, Shoes: 1},
	}
	base.Name = "Hero One"
	if _, err := repo.SaveCharacter(ctx, tt.Tx, base); err != nil {
		t.Fatal(err)
	}
	base.Name = "Hero Two"
	if _, err := repo.SaveCharacter(ctx, tt.Tx, base); err != nil {
		t.Fatal(err)
	}

	chars, count, err := repo.FindAllCharacters(ctx, tt.Tx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("got count %d, want 2", count)
	}
	if len(chars) != 2 {
		t.Errorf("got %d characters, want 2", len(chars))
	}
}

func TestCharacterRepo_FindAllCharacters_Pagination(t *testing.T) {
	tt := newTestTx(t)
	repo := newCharacterRepo(t)
	ctx := context.Background()

	base := &characters.Character{
		BodyType: "type_a",
		Species:  "human",
		Class:    "fighter",
		Stats:    &characters.Stats{Strength: 10, Dexterity: 10, Constitution: 10, Intelligence: 10, Wisdom: 10, Charisma: 10},
		Customization: &characters.Customization{Hair: 1, Face: 1, Shirt: 1, Pants: 1, Shoes: 1},
	}

	for i := 0; i < 5; i++ {
		base.Name = "Hero"
		if _, err := repo.SaveCharacter(ctx, tt.Tx, base); err != nil {
			t.Fatal(err)
		}
	}

	page1, count, err := repo.FindAllCharacters(ctx, tt.Tx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if count != 5 {
		t.Errorf("got count %d, want 5", count)
	}
	if len(page1) != 5 {
		t.Errorf("page 1: got %d characters, want 5", len(page1))
	}

	page2, _, err := repo.FindAllCharacters(ctx, tt.Tx, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(page2) != 0 {
		t.Errorf("page 2: got %d characters, want 0", len(page2))
	}
}

func TestCharacterRepo_FindAllCharacters_Empty(t *testing.T) {
	tt := newTestTx(t)
	repo := newCharacterRepo(t)
	ctx := context.Background()

	chars, count, err := repo.FindAllCharacters(ctx, tt.Tx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}
	if len(chars) != 0 {
		t.Errorf("expected no characters, got %d", len(chars))
	}
}

func TestCharacterRepo_UpdateCharacter(t *testing.T) {
	tt := newTestTx(t)
	repo := newCharacterRepo(t)
	ctx := context.Background()

	char := &characters.Character{
		Name:     "Aragorn",
		BodyType: "type_a",
		Species:  "human",
		Class:    "ranger",
		Stats:    &characters.Stats{Strength: 15, Dexterity: 14, Constitution: 13, Intelligence: 12, Wisdom: 11, Charisma: 10},
		Customization: &characters.Customization{Hair: 5, Face: 3, Shirt: 12, Pants: 8, Shoes: 2},
	}
	saved, err := repo.SaveCharacter(ctx, tt.Tx, char)
	if err != nil {
		t.Fatal(err)
	}

	saved.Name = "Aragorn II"
	saved.Species = "elf"
	saved.Stats.Strength = 18

	got, err := repo.UpdateCharacter(ctx, tt.Tx, saved)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Aragorn II" {
		t.Errorf("got name %q, want %q", got.Name, "Aragorn II")
	}
	if got.Stats.Strength != 18 {
		t.Errorf("got strength %d, want 18", got.Stats.Strength)
	}
}

func TestCharacterRepo_DeleteCharacter(t *testing.T) {
	tt := newTestTx(t)
	repo := newCharacterRepo(t)
	ctx := context.Background()

	char := &characters.Character{
		Name:     "ToDelete",
		BodyType: "type_a",
		Species:  "human",
		Class:    "fighter",
		Stats:    &characters.Stats{Strength: 10, Dexterity: 10, Constitution: 10, Intelligence: 10, Wisdom: 10, Charisma: 10},
		Customization: &characters.Customization{Hair: 1, Face: 1, Shirt: 1, Pants: 1, Shoes: 1},
	}
	saved, err := repo.SaveCharacter(ctx, tt.Tx, char)
	if err != nil {
		t.Fatal(err)
	}

	err = repo.DeleteCharacter(ctx, tt.Tx, saved.ID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = repo.FindCharacter(ctx, tt.Tx, saved.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestCharacterRepo_DeleteCharacter_Cascades(t *testing.T) {
	tt := newTestTx(t)
	repo := newCharacterRepo(t)
	ctx := context.Background()

	char := &characters.Character{
		Name:     "CascadeTest",
		BodyType: "type_a",
		Species:  "human",
		Class:    "fighter",
		Stats:    &characters.Stats{Strength: 10, Dexterity: 10, Constitution: 10, Intelligence: 10, Wisdom: 10, Charisma: 10},
		Customization: &characters.Customization{Hair: 1, Face: 1, Shirt: 1, Pants: 1, Shoes: 1},
	}
	saved, err := repo.SaveCharacter(ctx, tt.Tx, char)
	if err != nil {
		t.Fatal(err)
	}

	err = repo.DeleteCharacter(ctx, tt.Tx, saved.ID)
	if err != nil {
		t.Fatal(err)
	}

	q := tt.Q
	if _, err := q.SelectCharacterStats(ctx, int64(saved.ID)); err == nil {
		t.Error("expected error when selecting stats for deleted character")
	}
	if _, err := q.SelectCharacterCustomization(ctx, int64(saved.ID)); err == nil {
		t.Error("expected error when selecting customization for deleted character")
	}
}
