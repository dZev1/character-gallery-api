package postgres

import (
	"context"
	"testing"

	"dZev1/character-gallery/internal/postgres/db"
)

func TestAuthRepo_CreateAPIKey(t *testing.T) {
	tt := newTestTx(t)
	repo := newAuthRepo(t)
	ctx := context.Background()

	hash, err := repo.CreateAPIKey(ctx, tt.Tx, "a1b2c3d4e5f6", "test-key")
	if err != nil {
		t.Fatal(err)
	}
	if hash != "a1b2c3d4e5f6" {
		t.Errorf("got hash %q, want %q", hash, "a1b2c3d4e5f6")
	}

	var name string
	var isActive bool
	err = tt.Tx.QueryRow(ctx, "SELECT name, is_active FROM api_keys WHERE key_hash = $1", hash).Scan(&name, &isActive)
	if err != nil {
		t.Fatal(err)
	}
	if name != "test-key" {
		t.Errorf("got name %q, want %q", name, "test-key")
	}
	if !isActive {
		t.Error("expected new key to be active")
	}
}

func TestAuthRepo_CreateAPIKey_DuplicateHash(t *testing.T) {
	tt := newTestTx(t)
	repo := newAuthRepo(t)
	ctx := context.Background()

	_, err := repo.CreateAPIKey(ctx, tt.Tx, "duplicate-hash", "key-1")
	if err != nil {
		t.Fatal(err)
	}

	_, err = repo.CreateAPIKey(ctx, tt.Tx, "duplicate-hash", "key-2")
	if err == nil {
		t.Fatal("expected error for duplicate key_hash")
	}
}

func TestAuthRepo_ValidateAPIKey_Valid(t *testing.T) {
	tt := newTestTx(t)
	repo := newAuthRepo(t)
	ctx := context.Background()

	_, err := tt.Q.CreateAPIKey(ctx, db.CreateAPIKeyParams{
		KeyHash: "valid-key-hash",
		Name:    "valid-key",
	})
	if err != nil {
		t.Fatal(err)
	}

	exists, err := repo.ValidateAPIKey(ctx, tt.Tx, "valid-key-hash")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Error("expected valid key to be validated")
	}
}

func TestAuthRepo_ValidateAPIKey_NotFound(t *testing.T) {
	tt := newTestTx(t)
	repo := newAuthRepo(t)
	ctx := context.Background()

	exists, err := repo.ValidateAPIKey(ctx, tt.Tx, "non-existent-hash")
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("expected non-existent key to return false")
	}
}

func TestAuthRepo_ValidateAPIKey_Inactive(t *testing.T) {
	tt := newTestTx(t)
	repo := newAuthRepo(t)
	ctx := context.Background()

	_, err := tt.Q.CreateAPIKey(ctx, db.CreateAPIKeyParams{
		KeyHash: "inactive-key-hash",
		Name:    "inactive-key",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = tt.Tx.Exec(ctx, "UPDATE api_keys SET is_active = false WHERE key_hash = $1", "inactive-key-hash")
	if err != nil {
		t.Fatal(err)
	}

	exists, err := repo.ValidateAPIKey(ctx, tt.Tx, "inactive-key-hash")
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("expected inactive key to return false")
	}
}

func TestAuthRepo_UpdateLastUsed(t *testing.T) {
	tt := newTestTx(t)
	repo := newAuthRepo(t)
	ctx := context.Background()

	_, err := tt.Q.CreateAPIKey(ctx, db.CreateAPIKeyParams{
		KeyHash: "update-last-used-hash",
		Name:    "update-test",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = repo.UpdateLastUsed(ctx, tt.Tx, "update-last-used-hash")
	if err != nil {
		t.Fatal(err)
	}

	var lastUsed interface{}
	err = tt.Tx.QueryRow(ctx, "SELECT last_used_at FROM api_keys WHERE key_hash = $1", "update-last-used-hash").Scan(&lastUsed)
	if err != nil {
		t.Fatal(err)
	}
	if lastUsed == nil {
		t.Error("expected last_used_at to be set after UpdateLastUsed")
	}
}
