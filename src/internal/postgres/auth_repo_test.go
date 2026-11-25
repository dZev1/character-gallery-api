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
		Email:        "testuser@example.com",
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

	user, err := repo.SaveUser(ctx, tt.Tx, uuidv7.Must(), "aragorn", "aragorn@gondor.me", "password_hash_123")
	if err != nil {
		t.Fatal(err)
	}
	if user.ID == (auth.UserID(uuid.Nil)) {
		t.Fatal("expected non-nil (generated) ID")
	}
	if user.Username != "aragorn" {
		t.Errorf("got username %q, want %q", user.Username, "aragorn")
	}
	if user.Email != "aragorn@gondor.me" {
		t.Errorf("got email %q, want %q", user.Email, "aragorn@gondor.me")
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
		t.Errorf("got id %s, want %s", got.ID, dbUser.ID)
	}
	if got.Username != dbUser.Username {
		t.Errorf("got username %q, want %q", got.Username, dbUser.Username)
	}
	if got.Email != dbUser.Email {
		t.Errorf("got email %q, want %q", got.Email, dbUser.Email)
	}
}

func TestAuthRepo_FindUserByEmail(t *testing.T) {
	tt := newTestTx(t)
	repo := newAuthRepo(t)
	ctx := context.Background()

	dbUser := createTestUser(t, tt.Q)

	got, err := repo.FindUserByEmail(ctx, tt.Tx, dbUser.Email)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != auth.UserID(dbUser.ID) {
		t.Errorf("got id %s, want %s", got.ID, dbUser.ID)
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

func TestAuthRepo_SaveUser_DuplicateUsernameOrEmail(t *testing.T) {
	tt := newTestTx(t)
	repo := newAuthRepo(t)
	ctx := context.Background()

	_, err := repo.SaveUser(ctx, tt.Tx, uuidv7.Must(), "gimli", "gimli@erebor.me", "axe_hash")
	if err != nil {
		t.Fatal(err)
	}

	_, err = repo.SaveUser(ctx, tt.Tx, uuidv7.Must(), "gimli", "other@erebor.me", "different_hash")
	if err == nil {
		t.Fatal("expected error for duplicate username")
	}

	_, err = repo.SaveUser(ctx, tt.Tx, uuidv7.Must(), "gimli2", "gimli@erebor.me", "different_hash")
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
}
