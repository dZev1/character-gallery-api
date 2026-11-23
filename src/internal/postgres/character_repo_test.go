package postgres

import (
	"context"
	"errors"
	"testing"

	"dZev1/character-gallery/internal/characters"

	"github.com/google/uuid"
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
			Hair:        5,
			Beard:       3,
			SkinTone:    4,
			HairColor:   8,
			Cloak:       2,
			AccentColor: 6,
		},
	}

	saved, err := repo.SaveCharacter(ctx, tt.Tx, char, testOwnerID(t, tt.Q))
	if err != nil {
		t.Fatal(err)
	}
	if saved.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if saved.OwnerID == (uuid.UUID{}) {
		t.Fatal("expected owner ID to be set")
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
	owner := testOwnerID(t, tt.Q)

	base := &characters.Character{
		BodyType:      "type_a",
		Species:       "human",
		Class:         "fighter",
		Stats:         &characters.Stats{Strength: 10, Dexterity: 10, Constitution: 10, Intelligence: 10, Wisdom: 10, Charisma: 10},
		Customization: &characters.Customization{Hair: 1, Beard: 1, SkinTone: 1, HairColor: 1, Cloak: 1, AccentColor: 1},
	}
	base.Name = "Hero One"
	if _, err := repo.SaveCharacter(ctx, tt.Tx, base, owner); err != nil {
		t.Fatal(err)
	}
	base.Name = "Hero Two"
	if _, err := repo.SaveCharacter(ctx, tt.Tx, base, owner); err != nil {
		t.Fatal(err)
	}

	chars, count, err := repo.FindAllCharacters(ctx, tt.Tx, 20, 0)
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
	owner := testOwnerID(t, tt.Q)

	base := &characters.Character{
		BodyType:      "type_a",
		Species:       "human",
		Class:         "fighter",
		Stats:         &characters.Stats{Strength: 10, Dexterity: 10, Constitution: 10, Intelligence: 10, Wisdom: 10, Charisma: 10},
		Customization: &characters.Customization{Hair: 1, Beard: 1, SkinTone: 1, HairColor: 1, Cloak: 1, AccentColor: 1},
	}

	for i := 0; i < 5; i++ {
		base.Name = "Hero"
		if _, err := repo.SaveCharacter(ctx, tt.Tx, base, owner); err != nil {
			t.Fatal(err)
		}
	}

	page1, count, err := repo.FindAllCharacters(ctx, tt.Tx, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if count != 5 {
		t.Errorf("got count %d, want 5", count)
	}
	if len(page1) != 5 {
		t.Errorf("page 1: got %d characters, want 5", len(page1))
	}

	page2, _, err := repo.FindAllCharacters(ctx, tt.Tx, 20, 20)
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

	chars, count, err := repo.FindAllCharacters(ctx, tt.Tx, 20, 0)
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

	base := &characters.Character{
		Name:          "Aragorn",
		BodyType:      "type_a",
		Species:       "human",
		Class:         "ranger",
		Stats:         &characters.Stats{Strength: 15, Dexterity: 14, Constitution: 13, Intelligence: 12, Wisdom: 11, Charisma: 10},
		Customization: &characters.Customization{Hair: 5, Beard: 3, SkinTone: 4, HairColor: 8, Cloak: 2, AccentColor: 6},
	}
	saved, err := repo.SaveCharacter(ctx, tt.Tx, base, testOwnerID(t, tt.Q))
	if err != nil {
		t.Fatal(err)
	}

	saved.Name = "Aragorn II"
	saved.Species = "elf"
	saved.Stats.Strength = 18

	got, err := repo.UpdateCharacter(ctx, tt.Tx, saved, saved.OwnerID)
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
		Name:          "ToDelete",
		BodyType:      "type_a",
		Species:       "human",
		Class:         "fighter",
		Stats:         &characters.Stats{Strength: 10, Dexterity: 10, Constitution: 10, Intelligence: 10, Wisdom: 10, Charisma: 10},
		Customization: &characters.Customization{Hair: 1, Beard: 1, SkinTone: 1, HairColor: 1, Cloak: 1, AccentColor: 1},
	}
	owner := testOwnerID(t, tt.Q)
	saved, err := repo.SaveCharacter(ctx, tt.Tx, char, owner)
	if err != nil {
		t.Fatal(err)
	}

	err = repo.DeleteCharacter(ctx, tt.Tx, saved.ID, owner)
	if err != nil {
		t.Fatal(err)
	}

	_, err = repo.FindCharacter(ctx, tt.Tx, saved.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestCharacterRepo_DeleteCharacter_ForbiddenOwner(t *testing.T) {
	tt := newTestTx(t)
	repo := newCharacterRepo(t)
	ctx := context.Background()

	owner := testOwnerID(t, tt.Q)
	other := testOwnerID(t, tt.Q)

	char := &characters.Character{
		Name:          "Private",
		BodyType:      "type_a",
		Species:       "human",
		Class:         "fighter",
		Stats:         &characters.Stats{Strength: 10, Dexterity: 10, Constitution: 10, Intelligence: 10, Wisdom: 10, Charisma: 10},
		Customization: &characters.Customization{Hair: 1, Beard: 1, SkinTone: 1, HairColor: 1, Cloak: 1, AccentColor: 1},
	}
	saved, err := repo.SaveCharacter(ctx, tt.Tx, char, owner)
	if err != nil {
		t.Fatal(err)
	}

	err = repo.DeleteCharacter(ctx, tt.Tx, saved.ID, other)
	if err == nil || !errors.Is(err, characters.ErrNotFound) {
		t.Fatalf("expected ErrNotFound when deleting another user's character, got %v", err)
	}

	if _, err := repo.FindCharacter(ctx, tt.Tx, saved.ID); err != nil {
		t.Fatalf("character should still exist, got %v", err)
	}
}

func TestCharacterRepo_UpdateCharacter_ForbiddenOwner(t *testing.T) {
	tt := newTestTx(t)
	repo := newCharacterRepo(t)
	ctx := context.Background()

	owner := testOwnerID(t, tt.Q)
	other := testOwnerID(t, tt.Q)

	char := &characters.Character{
		Name:          "Private",
		BodyType:      "type_a",
		Species:       "human",
		Class:         "fighter",
		Stats:         &characters.Stats{Strength: 10, Dexterity: 10, Constitution: 10, Intelligence: 10, Wisdom: 10, Charisma: 10},
		Customization: &characters.Customization{Hair: 1, Beard: 1, SkinTone: 1, HairColor: 1, Cloak: 1, AccentColor: 1},
	}
	saved, err := repo.SaveCharacter(ctx, tt.Tx, char, owner)
	if err != nil {
		t.Fatal(err)
	}

	saved.Name = "Hijacked"
	_, err = repo.UpdateCharacter(ctx, tt.Tx, saved, other)
	if err == nil || !errors.Is(err, characters.ErrNotFound) {
		t.Fatalf("expected ErrNotFound when updating another user's character, got %v", err)
	}
}

func TestCharacterRepo_AdminBypassesOwnership(t *testing.T) {
	tt := newTestTx(t)
	repo := newCharacterRepo(t)
	ctx := context.Background()

	owner := testOwnerID(t, tt.Q)
	var admin uuid.UUID = uuid.Nil // admin bypass sentinel (handlers pass uuid.Nil for admin role)

	char := &characters.Character{
		Name:          "OthersChar",
		BodyType:      "type_a",
		Species:       "human",
		Class:         "fighter",
		Stats:         &characters.Stats{Strength: 10, Dexterity: 10, Constitution: 10, Intelligence: 10, Wisdom: 10, Charisma: 10},
		Customization: &characters.Customization{Hair: 1, Beard: 1, SkinTone: 1, HairColor: 1, Cloak: 1, AccentColor: 1},
	}
	saved, err := repo.SaveCharacter(ctx, tt.Tx, char, owner)
	if err != nil {
		t.Fatal(err)
	}

	saved.Name = "AdminEdited"
	got, err := repo.UpdateCharacter(ctx, tt.Tx, saved, admin)
	if err != nil {
		t.Fatalf("admin should be able to update, got %v", err)
	}
	if got.Name != "AdminEdited" {
		t.Errorf("got name %q, want %q", got.Name, "AdminEdited")
	}

	err = repo.DeleteCharacter(ctx, tt.Tx, saved.ID, admin)
	if err != nil {
		t.Fatalf("admin should be able to delete, got %v", err)
	}
	if _, err := repo.FindCharacter(ctx, tt.Tx, saved.ID); err == nil {
		t.Fatal("expected character to be deleted by admin")
	}
}

func TestCharacterRepo_DeleteCharacter_Cascades(t *testing.T) {
	tt := newTestTx(t)
	repo := newCharacterRepo(t)
	ctx := context.Background()

	char := &characters.Character{
		Name:          "CascadeTest",
		BodyType:      "type_a",
		Species:       "human",
		Class:         "fighter",
		Stats:         &characters.Stats{Strength: 10, Dexterity: 10, Constitution: 10, Intelligence: 10, Wisdom: 10, Charisma: 10},
		Customization: &characters.Customization{Hair: 1, Beard: 1, SkinTone: 1, HairColor: 1, Cloak: 1, AccentColor: 1},
	}
	owner := testOwnerID(t, tt.Q)
	saved, err := repo.SaveCharacter(ctx, tt.Tx, char, owner)
	if err != nil {
		t.Fatal(err)
	}

	err = repo.DeleteCharacter(ctx, tt.Tx, saved.ID, owner)
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

func TestCharacterRepo_HealProgressAndPrestige(t *testing.T) {
	tt := newTestTx(t)
	repo := newCharacterRepo(t)
	ctx := context.Background()

	owner := testOwnerID(t, tt.Q)
	char := &characters.Character{
		Name:          "PaladinHero",
		BodyType:      "type_a",
		Species:       "human",
		Class:         "paladin",
		Level:         9,
		Xp:            900,
		HpMax:         40,
		HpCurrent:     10,
		Stats:         &characters.Stats{Strength: 15, Dexterity: 12, Constitution: 14, Intelligence: 10, Wisdom: 13, Charisma: 16},
		Customization: &characters.Customization{Hair: 1, Beard: 1, SkinTone: 1, HairColor: 1, Cloak: 1, AccentColor: 1},
	}

	saved, err := repo.SaveCharacter(ctx, tt.Tx, char, owner)
	if err != nil {
		t.Fatal(err)
	}

	// Heal +15 -> 25 HP
	healed, err := repo.HealCharacter(ctx, tt.Tx, saved.ID, 15)
	if err != nil {
		t.Fatal(err)
	}
	if healed.HpCurrent != 25 {
		t.Fatalf("got hp_current %d, want 25", healed.HpCurrent)
	}

	// Heal beyond HpMax clamps to 40
	healed, err = repo.HealCharacter(ctx, tt.Tx, saved.ID, 50)
	if err != nil {
		t.Fatal(err)
	}
	if healed.HpCurrent != 40 {
		t.Fatalf("got hp_current %d, want 40 (clamped to HpMax)", healed.HpCurrent)
	}

	// Prestige before level 10 fails
	if _, err := repo.PrestigeCharacter(ctx, tt.Tx, saved.ID, owner, 20); !errors.Is(err, characters.ErrPrestigeRequirement) {
		t.Fatalf("expected ErrPrestigeRequirement at level 9, got %v", err)
	}

	// Level up to 10 via UpdateCharacterProgress
	lvl10, err := repo.UpdateCharacterProgress(ctx, tt.Tx, saved.ID, 10, 1200, 45, 30)
	if err != nil {
		t.Fatal(err)
	}
	if lvl10.Level != 10 || lvl10.HpMax != 45 || lvl10.HpCurrent != 30 {
		t.Fatalf("unexpected progress state: %+v", lvl10)
	}

	// Prestige at level 10 resets level to 1, increments prestige, and boosts all stats by +1
	prestiged, err := repo.PrestigeCharacter(ctx, tt.Tx, saved.ID, owner, 22)
	if err != nil {
		t.Fatal(err)
	}
	if prestiged.Level != 1 || prestiged.Prestige != 1 || prestiged.HpMax != 22 || prestiged.HpCurrent != 22 {
		t.Fatalf("unexpected prestiged character: %+v", prestiged)
	}
	if prestiged.Stats.Strength != 16 || prestiged.Stats.Charisma != 17 {
		t.Fatalf("expected stats boosted by +1, got %+v", prestiged.Stats)
	}
}
