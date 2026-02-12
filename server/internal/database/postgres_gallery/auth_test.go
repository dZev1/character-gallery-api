package postgres_gallery

import (
	"errors"
	"strings"
	"testing"

	"dZev1/character-gallery/models/auth"
	"github.com/DATA-DOG/go-sqlmock"
)

// ==================== ValidateAPIKey Tests ====================

func TestValidateAPIKey_Success(t *testing.T) {
	authStore, mock := setupMockAuthStore(t)

	key := createTestAPIKey()

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery(`SELECT EXISTS`).WithArgs(key.KeyHash).WillReturnRows(rows)

	valid, err := authStore.ValidateAPIKey(key.KeyHash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !valid {
		t.Fatal("expected key to be valid")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestValidateAPIKey_NotFound(t *testing.T) {
	authStore, mock := setupMockAuthStore(t)

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
	mock.ExpectQuery(`SELECT EXISTS`).WithArgs("nonexistent_hash").WillReturnRows(rows)

	valid, err := authStore.ValidateAPIKey("nonexistent_hash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if valid {
		t.Fatal("expected key to be invalid")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestValidateAPIKey_DBError(t *testing.T) {
	authStore, mock := setupMockAuthStore(t)

	dbErr := errors.New("database connection failed")
	mock.ExpectQuery(`SELECT EXISTS`).WithArgs("some_hash").WillReturnError(dbErr)

	_, err := authStore.ValidateAPIKey("some_hash")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// ==================== UpdateLastUsed Tests ====================

func TestUpdateLastUsed_Success(t *testing.T) {
	authStore, mock := setupMockAuthStore(t)

	key := createTestAPIKey()

	mock.ExpectExec(`UPDATE api_keys SET last_used = NOW\(\) WHERE key_hash`).WithArgs(key.KeyHash).WillReturnResult(sqlmock.NewResult(0, 1))

	err := authStore.UpdateLastUsed(key.KeyHash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateLastUsed_DBError(t *testing.T) {
	authStore, mock := setupMockAuthStore(t)

	dbErr := errors.New("database error")
	mock.ExpectExec(`UPDATE api_keys SET last_used = NOW\(\) WHERE key_hash`).WithArgs("some_hash").WillReturnError(dbErr)

	err := authStore.UpdateLastUsed("some_hash")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// ==================== CreateAPIKey Tests ====================

func TestCreateAPIKey_Success(t *testing.T) {
	authStore, mock := setupMockAuthStore(t)

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO api_keys`).WithArgs("test-key", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	rawKey, err := authStore.CreateAPIKey("test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rawKey == "" {
		t.Fatal("expected non-empty raw key")
	}

	if !strings.HasPrefix(rawKey, "dz_chars_") {
		t.Fatalf("expected key to have prefix 'dz_chars_', got: %s", rawKey)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateAPIKey_BeginError(t *testing.T) {
	authStore, mock := setupMockAuthStore(t)

	dbErr := errors.New("begin transaction failed")
	mock.ExpectBegin().WillReturnError(dbErr)

	_, err := authStore.CreateAPIKey("test-key")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateAPIKey_InsertError(t *testing.T) {
	authStore, mock := setupMockAuthStore(t)

	dbErr := errors.New("insert failed")
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO api_keys`).WithArgs("test-key", sqlmock.AnyArg()).WillReturnError(dbErr)
	mock.ExpectRollback()

	_, err := authStore.CreateAPIKey("test-key")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateAPIKey_CommitError(t *testing.T) {
	authStore, mock := setupMockAuthStore(t)

	dbErr := errors.New("commit failed")
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO api_keys`).WithArgs("test-key", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit().WillReturnError(dbErr)

	_, err := authStore.CreateAPIKey("test-key")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// ==================== GenerateAPIKey Tests ====================

func TestGenerateAPIKey(t *testing.T) {
	keyHash, rawKey, err := auth.GenerateAPIKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(rawKey, "dz_chars_") {
		t.Fatalf("expected prefix 'dz_chars_', got: %s", rawKey)
	}

	if len(keyHash) != 64 {
		t.Fatalf("expected SHA-256 hash (64 hex chars), got length: %d", len(keyHash))
	}

	// Verify hashing is consistent
	rehash := auth.HashAPIKey(rawKey)
	if rehash != keyHash {
		t.Fatal("hashing the raw key should produce the same hash")
	}
}

func TestHashAPIKey(t *testing.T) {
	// Same input should produce same hash
	hash1 := auth.HashAPIKey("test_key")
	hash2 := auth.HashAPIKey("test_key")

	if hash1 != hash2 {
		t.Fatal("hashing same input should produce same output")
	}

	// Different input should produce different hash
	hash3 := auth.HashAPIKey("different_key")
	if hash1 == hash3 {
		t.Fatal("different inputs should produce different hashes")
	}
}
