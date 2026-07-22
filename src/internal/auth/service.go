package auth

import (
	"dZev1/character-gallery/internal/cache"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo  Repository
	pool  *pgxpool.Pool
	cache cache.Cache
}
