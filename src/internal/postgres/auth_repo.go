package postgres

import (
	"context"
	"dZev1/character-gallery/internal/auth"
	"dZev1/character-gallery/internal/postgres/db"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ auth.Repository = (*authRepo)(nil)

type authRepo struct {
	pool *pgxpool.Pool
}

func NewAuthRepo(pool *pgxpool.Pool) *authRepo {
	return &authRepo{pool: pool}
}

func (c *authRepo) q(exec db.DBTX) *db.Queries {
	return db.New(exec)
}

func (a *authRepo) SaveUser(ctx context.Context, exec db.DBTX, id uuid.UUID, username, email, passwordHash string) (*auth.User, error) {
	q := a.q(exec)

	usr, err := q.CreateUser(ctx, db.CreateUserParams{
		ID:           id,
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return toDomainUser(usr), nil
}

func (a *authRepo) FindUserByUsername(ctx context.Context, exec db.DBTX, username string) (*auth.User, error) {
	q := a.q(exec)

	usr, err := q.SelectUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("cannot find user '%s': %w", username, err)
	}

	return toDomainUser(usr), nil
}

func (a *authRepo) FindUserByEmail(ctx context.Context, exec db.DBTX, email string) (*auth.User, error) {
	q := a.q(exec)

	usr, err := q.SelectUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("cannot find user by email: %w", err)
	}

	return toDomainUser(usr), nil
}

func toDomainUser(usr *db.User) *auth.User {
	return &auth.User{
		ID:           auth.UserID(usr.ID),
		Username:     usr.Username,
		Email:        usr.Email,
		PasswordHash: usr.PasswordHash,
		Role:         usr.Role,
		CreatedAt:    &usr.CreatedAt.Time,
	}
}
