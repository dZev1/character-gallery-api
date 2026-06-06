package postgres

import (
	"context"
	"testing"

	"dZev1/character-gallery/internal/postgres/db"
)

func TestAuthRepo_CreateAPIKey(t *testing.T) {
	q := newTxQueries(t)
	ctx := context.Background()

	key, err := q.CreateAPIKey(ctx, db.CreateAPIKeyParams{
		KeyHash: "a1b2c3d4e5f6",
		Name:    "test-key",
	})
	if err != nil {
		t.Fatal(err)
	}

	if key.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if key.KeyHash != "a1b2c3d4e5f6" {
		t.Errorf("got hash %q, want %q", key.KeyHash, "a1b2c3d4e5f6")
	}
	if key.Name != "test-key" {
		t.Errorf("got name %q, want %q", key.Name, "test-key")
	}
	if !key.IsActive {
		t.Error("expected new key to be active")
	}
}

func TestAuthRepo_CreateAPIKey_DuplicateHash(t *testing.T) {
	q := newTxQueries(t)
	ctx := context.Background()

	_, err := q.CreateAPIKey(ctx, db.CreateAPIKeyParams{
		KeyHash: "duplicate-hash",
		Name:    "key-1",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = q.CreateAPIKey(ctx, db.CreateAPIKeyParams{
		KeyHash: "duplicate-hash",
		Name:    "key-2",
	})
	if err == nil {
		t.Fatal("expected error for duplicate key_hash")
	}
}

func TestAuthRepo_ValidateAPIKey_Valid(t *testing.T) {
	q := newTxQueries(t)
	ctx := context.Background()

	_, err := q.CreateAPIKey(ctx, db.CreateAPIKeyParams{
		KeyHash: "valid-key-hash",
		Name:    "valid-key",
	})
	if err != nil {
		t.Fatal(err)
	}

	exists, err := q.ValidateAPIKey(ctx, "valid-key-hash")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Error("expected valid key to be validated")
	}
}

func TestAuthRepo_ValidateAPIKey_NotFound(t *testing.T) {
	q := newTxQueries(t)
	ctx := context.Background()

	exists, err := q.ValidateAPIKey(ctx, "non-existent-hash")
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("expected non-existent key to return false")
	}
}

func TestAuthRepo_ValidateAPIKey_Inactive(t *testing.T) {
	tt := newTestTx(t)
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

	exists, err := tt.Q.ValidateAPIKey(ctx, "inactive-key-hash")
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("expected inactive key to return false")
	}
}

func TestAuthRepo_UpdateLastUsed(t *testing.T) {
	tt := newTestTx(t)
	ctx := context.Background()

	key, err := tt.Q.CreateAPIKey(ctx, db.CreateAPIKeyParams{
		KeyHash: "update-last-used-hash",
		Name:    "update-test",
	})
	if err != nil {
		t.Fatal(err)
	}

	if key.LastUsedAt.Valid {
		t.Log("new key has last_used_at set (expected to be null)")
	}

	err = tt.Q.UpdateLastUsed(ctx, "update-last-used-hash")
	if err != nil {
		t.Fatal(err)
	}

	var lastUsed interface{}
	err = tt.Tx.QueryRow(ctx, "SELECT last_used_at FROM api_keys WHERE id = $1", key.ID).Scan(&lastUsed)
	if err != nil {
		t.Fatal(err)
	}
	if lastUsed == nil {
		t.Error("expected last_used_at to be set after UpdateLastUsed")
	}
}
