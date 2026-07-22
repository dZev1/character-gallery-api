package postgres

import (
	"context"
	"testing"

	"dZev1/character-gallery/internal/postgres/db"
)

func newAuthRepo(t *testing.T) *authRepo {
	t.Helper()
	return NewAuthRepo(globalDB.pool)
}

func createTestUser(t *testing.T, q *db.Queries) *db.User {
	t.Helper()
	ctx := context.Background()
	user, err := q.CreateUser(ctx, db.CreateUserParams{
		Username:     "testuser",
		PasswordHash: "test_hash",
	})
	if err != nil {
		t.Fatal(err)
	}
	return user
}

func TestAuthRepo_SaveUser(t *testing.T) {
	tt := newTestTx(t)
	repo := newAuthRepo(t)
	ctx := context.Background()

	user, err := repo.SaveUser(ctx, tt.Tx, "aragorn", "password_hash_123")
	if err != nil {
		t.Fatal(err)
	}
	if user.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if user.Username != "aragorn" {
		t.Errorf("got username %q, want %q", user.Username, "aragorn")
	}
	if user.PasswordHash != "password_hash_123" {
		t.Errorf("got password_hash %q, want %q", user.PasswordHash, "password_hash_123")
	}
	if user.CreatedAt == nil {
		t.Fatal("expected non-nil CreatedAt")
	}
}

func TestAuthRepo_FindUserByID(t *testing.T) {
	tt := newTestTx(t)
	repo := newAuthRepo(t)
	ctx := context.Background()

	dbUser := createTestUser(t, tt.Q)

	got, err := repo.FindUserByID(ctx, tt.Tx, uint64(dbUser.ID))
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != uint64(dbUser.ID) {
		t.Errorf("got id %d, want %d", got.ID, dbUser.ID)
	}
	if got.Username != dbUser.Username {
		t.Errorf("got username %q, want %q", got.Username, dbUser.Username)
	}
	if got.PasswordHash != dbUser.PasswordHash {
		t.Errorf("got password_hash %q, want %q", got.PasswordHash, dbUser.PasswordHash)
	}
}

func TestAuthRepo_FindUserByID_NotFound(t *testing.T) {
	tt := newTestTx(t)
	repo := newAuthRepo(t)
	ctx := context.Background()

	_, err := repo.FindUserByID(ctx, tt.Tx, 99999)
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}

func TestAuthRepo_FindUserByUsername(t *testing.T) {
	tt := newTestTx(t)
	repo := newAuthRepo(t)
	ctx := context.Background()

	dbUser := createTestUser(t, tt.Q)

	got, err := repo.FindUserByUsername(ctx, tt.Tx, dbUser.Username)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != uint64(dbUser.ID) {
		t.Errorf("got id %d, want %d", got.ID, dbUser.ID)
	}
	if got.Username != dbUser.Username {
		t.Errorf("got username %q, want %q", got.Username, dbUser.Username)
	}
}

func TestAuthRepo_FindUserByUsername_NotFound(t *testing.T) {
	tt := newTestTx(t)
	repo := newAuthRepo(t)
	ctx := context.Background()

	_, err := repo.FindUserByUsername(ctx, tt.Tx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent username")
	}
}

func TestAuthRepo_SaveUser_DuplicateUsername(t *testing.T) {
	tt := newTestTx(t)
	repo := newAuthRepo(t)
	ctx := context.Background()

	_, err := repo.SaveUser(ctx, tt.Tx, "gimli", "axe_hash")
	if err != nil {
		t.Fatal(err)
	}

	_, err = repo.SaveUser(ctx, tt.Tx, "gimli", "different_hash")
	if err == nil {
		t.Fatal("expected error for duplicate username")
	}
}

func TestAuthRepo_SaveAndRoundTrip(t *testing.T) {
	tt := newTestTx(t)
	repo := newAuthRepo(t)
	ctx := context.Background()

	saved, err := repo.SaveUser(ctx, tt.Tx, "frodo", "ring_hash")
	if err != nil {
		t.Fatal(err)
	}

	byID, err := repo.FindUserByID(ctx, tt.Tx, uint64(saved.ID))
	if err != nil {
		t.Fatal(err)
	}
	if byID.Username != "frodo" {
		t.Errorf("FindUserByID: got username %q, want %q", byID.Username, "frodo")
	}

	byName, err := repo.FindUserByUsername(ctx, tt.Tx, "frodo")
	if err != nil {
		t.Fatal(err)
	}
	if byName.ID != saved.ID {
		t.Errorf("FindUserByUsername: got id %d, want %d", byName.ID, saved.ID)
	}
}
