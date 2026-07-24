package postgres

import (
	"context"
	"dZev1/character-gallery/internal/auth"
	"dZev1/character-gallery/internal/postgres/db"
	"fmt"

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

func (a *authRepo) SaveUser(ctx context.Context, exec db.DBTX, username, passwordHash string) (*auth.User, error) {
	q := a.q(exec)

	createParams := db.CreateUserParams{
		Username:     username,
		PasswordHash: passwordHash,
	}

	usr, err := q.CreateUser(ctx, createParams)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &auth.User{
		ID:           auth.UserID(usr.ID),
		Username:     usr.Username,
		PasswordHash: usr.PasswordHash,
		CreatedAt:    &usr.CreatedAt.Time,
	}, nil
}

func (a *authRepo) FindUserByUsername(ctx context.Context, exec db.DBTX, username string) (*auth.User, error) {
	q := a.q(exec)

	usr, err := q.SelectUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("cannot find user '%s': %w", username, err)
	}

	return &auth.User{
		ID:           auth.UserID(usr.ID),
		Username:     username,
		PasswordHash: usr.PasswordHash,
		CreatedAt:    &usr.CreatedAt.Time,
	}, nil
}
