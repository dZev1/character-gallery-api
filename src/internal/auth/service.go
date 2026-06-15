package auth

import "github.com/jackc/pgx/v5/pgxpool"

type Service struct {
	repo Repository
	pool *pgxpool.Pool
}

func NewService(repo Repository, pool *pgxpool.Pool) *Service {
	return &Service{
		repo: repo,
		pool: pool,
	}
}
