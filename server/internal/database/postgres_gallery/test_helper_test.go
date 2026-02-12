package postgres_gallery

import (
	"testing"
	"time"

	"dZev1/character-gallery/models/auth"
	"dZev1/character-gallery/models/characters"
	"dZev1/character-gallery/models/inventory"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func setupMockDB(t *testing.T) (*PostgresCharacterGallery, sqlmock.Sqlmock) {
	t.Helper()

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	gallery := &PostgresCharacterGallery{db: sqlxDB, AuthStore: NewAuthStore(sqlxDB)}

	return gallery, mock
}

func createTestCharacter() *characters.Character {
	return &characters.Character{
		Name:     "TestHero",
		BodyType: characters.TypeA,
		Species:  characters.Human,
		Class:    characters.Fighter,
		Stats: &characters.Stats{
			Strength:     15,
			Dexterity:    12,
			Constitution: 14,
			Intelligence: 10,
			Wisdom:       8,
			Charisma:     11,
		},
		Customization: &characters.Customization{
			Hair:  1,
			Face:  2,
			Shirt: 3,
			Pants: 4,
			Shoes: 5,
		},
	}
}

func createTestItem() *inventory.Item {
	damage := uint64(50)
	return &inventory.Item{
		Name:        "Test Sword",
		Type:        inventory.Weapon,
		Description: "A test weapon",
		Equippable:  true,
		Rarity:      3,
		Damage:      &damage,
	}
}

func createTestAPIKey() *auth.APIKey {
	return &auth.APIKey{
		ID:         1,
		Name:       "Test API Key",
		KeyHash:    "testhash123",
		CreatedAt:  time.Now(),
		LastUsedAt: time.Now(),
		IsActive:   true,
	}
}

func setupMockAuthStore(t *testing.T) (*PGAuthStore, sqlmock.Sqlmock) {
	t.Helper()

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	authStore := &PGAuthStore{db: sqlxDB}

	return authStore, mock
}
