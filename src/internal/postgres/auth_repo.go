package postgres

import (
	"context"
	"dZev1/character-gallery/internal/auth"
	"dZev1/character-gallery/internal/postgres/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

var _ auth.Repository = (*authRepo)(nil)

type authRepo struct {
	pool *pgxpool.Pool
}

func NewAuthRepo(pool *pgxpool.Pool) *authRepo {
	return &authRepo{pool: pool}
}

func (a *authRepo) q(exec db.DBTX) *db.Queries {
	return db.New(exec)
}

func (a *authRepo) ValidateAPIKey(ctx context.Context, exec db.DBTX, keyHash string) (bool, error) {
	q := a.q(exec)

	exists, err := q.ValidateAPIKey(ctx, keyHash)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (a *authRepo) UpdateLastUsed(ctx context.Context, exec db.DBTX, keyHash string) error {
	q := a.q(exec)

	err := q.UpdateLastUsed(ctx, keyHash)
	if err != nil {
		return err
	}
	return nil
}

func (a *authRepo) CreateAPIKey(ctx context.Context, exec db.DBTX, keyHash, name string) (string, error) {
	q := a.q(exec)

	createParams := db.CreateAPIKeyParams{
		KeyHash: keyHash,
		Name:    name,
	}

	row, err := q.CreateAPIKey(ctx, createParams)
	if err != nil {
		return "", err
	}
	return row.KeyHash, nil
}
