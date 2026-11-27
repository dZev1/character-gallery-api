package enemies

import (
	"context"
	"dZev1/character-gallery/internal/postgres/db"
)

type Repository interface {
	FindAllEnemies(ctx context.Context, exec db.DBTX, limit, offset int) ([]Enemy, uint64, error)
	FindEnemy(ctx context.Context, exec db.DBTX, enemyID ID) (*Enemy, error)
	FindEnemiesByTier(ctx context.Context, exec db.DBTX, tier uint8) ([]Enemy, error)
	FindEnemiesByTierAndType(ctx context.Context, exec db.DBTX, tier uint8, enemyType Type) ([]Enemy, error)
	SaveEnemy(ctx context.Context, exec db.DBTX, enemy *Enemy) (*Enemy, error)
	SaveEnemyIfNotExists(ctx context.Context, exec db.DBTX, enemy *Enemy) (*Enemy, bool, error)
	UpdateEnemy(ctx context.Context, exec db.DBTX, enemy *Enemy) (*Enemy, error)
	DeleteEnemy(ctx context.Context, exec db.DBTX, enemyID ID) error
	SeedEnemies(ctx context.Context, exec db.DBTX, enemies []Enemy) error
}
