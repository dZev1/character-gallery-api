package postgres

import (
	"context"
	"testing"

	"dZev1/character-gallery/internal/auth"
	"dZev1/character-gallery/internal/postgres/db"
	"dZev1/character-gallery/internal/uuidv7"

	"github.com/google/uuid"
)

func newAuthRepo(t *testing.T) *authRepo {
	t.Helper()
	return NewAuthRepo(globalDB.pool)
}

func createTestUser(t *testing.T, q *db.Queries) *db.User {
	t.Helper()
	ctx := context.Background()
	user, err := q.CreateUser(ctx, db.CreateUserParams{
		ID:           uuidv7.Must(),
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

	user, err := repo.SaveUser(ctx, tt.Tx, uuidv7.Must(), "aragorn", "password_hash_123")
	if err != nil {
		t.Fatal(err)
	}
	if user.ID == (auth.UserID(uuid.Nil)) {
		t.Fatal("expected non-nil (generated) ID")
	}
	if user.Username != "aragorn" {
		t.Errorf("got username %q, want %q", user.Username, "aragorn")
	}
	if user.PasswordHash != "password_hash_123" {
		t.Errorf("got password_hash %q, want %q", user.PasswordHash, "password_hash_123")
	}
	if user.Role != "user" {
		t.Errorf("got role %q, want %q", user.Role, "user")
	}
	if user.CreatedAt == nil {
		t.Fatal("expected non-nil CreatedAt")
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
	if got.ID != auth.UserID(dbUser.ID) {
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

	_, err := repo.SaveUser(ctx, tt.Tx, uuidv7.Must(), "gimli", "axe_hash")
	if err != nil {
		t.Fatal(err)
	}

	_, err = repo.SaveUser(ctx, tt.Tx, uuidv7.Must(), "gimli", "different_hash")
	if err == nil {
		t.Fatal("expected error for duplicate username")
	}
}
