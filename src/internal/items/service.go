package items

import (
	"dZev1/character-gallery/internal/cache"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo  Repository
	pool  *pgxpool.Pool
	cache *cache.Cache
}

func NewService(repo Repository, pool *pgxpool.Pool, cache *cache.Cache) *Service {
	return &Service{
		repo:  repo,
		pool:  pool,
		cache: cache,
	}
}
