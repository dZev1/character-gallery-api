package postgres_gallery

import (
	"errors"
	"testing"

	"dZev1/character-gallery/models/characters"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGet_Success(t *testing.T) {
	gallery, mock := setupMockDB(t)

	charID := characters.CharacterID(1)

	// Mock base character query
	charRows := sqlmock.NewRows([]string{"id", "name", "body_type", "species", "class"}).
		AddRow(1, "TestHero", "type_a", "human", "fighter")
	mock.ExpectQuery(`SELECT \* FROM characters`).
		WithArgs(charID).
		WillReturnRows(charRows)

	// Mock stats query
	statsRows := sqlmock.NewRows([]string{"id", "strength", "dexterity", "constitution", "intelligence", "wisdom", "charisma"}).
		AddRow(1, 15, 12, 14, 10, 8, 11)
	mock.ExpectQuery(`SELECT \* FROM stats`).
		WithArgs(charID).
		WillReturnRows(statsRows)

	// Mock customization query
	custRows := sqlmock.NewRows([]string{"id", "hair", "face", "shirt", "pants", "shoes"}).
		AddRow(1, 1, 2, 3, 4, 5)
	mock.ExpectQuery(`SELECT \* FROM customizations`).
		WithArgs(charID).
		WillReturnRows(custRows)

	char, err := gallery.Get(charID)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if char == nil {
		t.Fatal("expected character, got nil")
	}
	if char.Name != "TestHero" {
		t.Errorf("expected name TestHero, got %s", char.Name)
	}
	if char.Stats.Strength != 15 {
		t.Errorf("expected strength 15, got %d", char.Stats.Strength)
	}
	if char.Customization.Hair != 1 {
		t.Errorf("expected hair 1, got %d", char.Customization.Hair)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGet_NotFound(t *testing.T) {
	gallery, mock := setupMockDB(t)

	charID := characters.CharacterID(999)

	mock.ExpectQuery(`SELECT \* FROM characters`).
		WithArgs(charID).
		WillReturnError(errors.New("no rows"))

	char, err := gallery.Get(charID)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if char != nil {
		t.Error("expected nil character")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetAll_Success(t *testing.T) {
	gallery, mock := setupMockDB(t)

	rows := sqlmock.NewRows([]string{
		"id", "name", "body_type", "species", "class",
		"stats.strength", "stats.dexterity", "stats.constitution",
		"stats.intelligence", "stats.wisdom", "stats.charisma",
		"customization.hair", "customization.face", "customization.shirt",
		"customization.pants", "customization.shoes",
	}).
		AddRow(1, "Hero1", "type_a", "human", "fighter", 15, 12, 14, 10, 8, 11, 1, 2, 3, 4, 5).
		AddRow(2, "Hero2", "type_b", "elf", "wizard", 8, 14, 10, 18, 12, 14, 2, 3, 4, 5, 6)

	mock.ExpectQuery(`SELECT`).
		WithArgs(0, 20).
		WillReturnRows(rows)

	chars, err := gallery.GetAll(0, 20)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(chars) != 2 {
		t.Errorf("expected 2 characters, got %d", len(chars))
	}
	if chars[0].Name != "Hero1" {
		t.Errorf("expected Hero1, got %s", chars[0].Name)
	}
	if chars[1].Name != "Hero2" {
		t.Errorf("expected Hero2, got %s", chars[1].Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetAll_Empty(t *testing.T) {
	gallery, mock := setupMockDB(t)

	rows := sqlmock.NewRows([]string{
		"id", "name", "body_type", "species", "class",
		"stats.strength", "stats.dexterity", "stats.constitution",
		"stats.intelligence", "stats.wisdom", "stats.charisma",
		"customization.hair", "customization.face", "customization.shirt",
		"customization.pants", "customization.shoes",
	})

	mock.ExpectQuery(`SELECT`).
		WithArgs(0, 20).
		WillReturnRows(rows)

	chars, err := gallery.GetAll(0, 20)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(chars) != 0 {
		t.Errorf("expected 0 characters, got %d", len(chars))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRemove_Success(t *testing.T) {
	gallery, mock := setupMockDB(t)

	charID := characters.CharacterID(1)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM characters`).
		WithArgs(charID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := gallery.Remove(charID)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRemove_NotFound(t *testing.T) {
	gallery, mock := setupMockDB(t)

	charID := characters.CharacterID(999)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM characters`).
		WithArgs(charID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := gallery.Remove(charID)

	if err == nil {
		t.Error("expected error for non-existent character")
	}
	if !errors.Is(err, ErrCouldNotFind) {
		t.Errorf("expected ErrCouldNotFind, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRemove_TransactionError(t *testing.T) {
	gallery, mock := setupMockDB(t)

	charID := characters.CharacterID(1)

	mock.ExpectBegin().WillReturnError(errors.New("tx error"))

	err := gallery.Remove(charID)

	if err == nil {
		t.Error("expected error")
	}
	if !errors.Is(err, ErrFailedInitializeTransaction) {
		t.Errorf("expected ErrFailedInitializeTransaction, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
