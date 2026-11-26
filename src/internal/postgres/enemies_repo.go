package postgres

import (
	"context"
	"dZev1/character-gallery/internal/enemies"
	"dZev1/character-gallery/internal/postgres/db"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ enemies.Repository = (*enemyRepo)(nil)

type enemyRepo struct {
	pool *pgxpool.Pool
}

func NewEnemyRepo(pool *pgxpool.Pool) *enemyRepo {
	return &enemyRepo{pool: pool}
}

func (r *enemyRepo) q(exec db.DBTX) *db.Queries {
	return db.New(exec)
}

func (r *enemyRepo) FindAllEnemies(ctx context.Context, exec db.DBTX, limit, offset int) ([]enemies.Enemy, uint64, error) {
	q := r.q(exec)

	rows, err := q.FindAllEnemies(ctx, db.FindAllEnemiesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list enemies: %w", err)
	}

	total, err := q.CountAllEnemies(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count enemies: %w", err)
	}

	result := make([]enemies.Enemy, len(rows))
	for i, row := range rows {
		result[i] = *toDomainEnemy(row)
	}

	return result, uint64(total), nil
}

func (r *enemyRepo) FindEnemy(ctx context.Context, exec db.DBTX, enemyID enemies.ID) (*enemies.Enemy, error) {
	q := r.q(exec)

	row, err := q.FindEnemy(ctx, int64(enemyID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("enemy %d: %w", enemyID, enemies.ErrNotFound)
		}
		return nil, fmt.Errorf("find enemy %d: %w", enemyID, err)
	}

	return toDomainEnemy(row), nil
}

func (r *enemyRepo) FindEnemiesByTier(ctx context.Context, exec db.DBTX, tier uint8) ([]enemies.Enemy, error) {
	q := r.q(exec)

	rows, err := q.FindEnemiesByTier(ctx, int16(tier))
	if err != nil {
		return nil, fmt.Errorf("find enemies by tier %d: %w", tier, err)
	}

	result := make([]enemies.Enemy, len(rows))
	for i, row := range rows {
		result[i] = *toDomainEnemy(row)
	}
	return result, nil
}

func (r *enemyRepo) FindEnemiesByTierAndType(ctx context.Context, exec db.DBTX, tier uint8, enemyType enemies.Type) ([]enemies.Enemy, error) {
	q := r.q(exec)

	rows, err := q.FindEnemiesByTierAndType(ctx, db.FindEnemiesByTierAndTypeParams{
		Tier: int16(tier),
		Type: string(enemyType),
	})
	if err != nil {
		return nil, fmt.Errorf("find enemies by tier %d and type %s: %w", tier, enemyType, err)
	}

	result := make([]enemies.Enemy, len(rows))
	for i, row := range rows {
		result[i] = *toDomainEnemy(row)
	}
	return result, nil
}

func (r *enemyRepo) SaveEnemy(ctx context.Context, exec db.DBTX, enemy *enemies.Enemy) (*enemies.Enemy, error) {
	q := r.q(exec)

	saved, err := q.CreateEnemy(ctx, db.CreateEnemyParams{
		Name:        enemy.Name,
		Type:        string(enemy.Type),
		Tier:        int16(enemy.Tier),
		Description: enemy.Description,
		HpMax:       int32(enemy.HpMax),
		Ac:          int32(enemy.ArmorClass),
		Damage:      int32(enemy.Damage),
		XpReward:    int32(enemy.XPReward),
		ScoreReward: int32(enemy.ScoreReward),
		ImageUrl:    toText(enemy.ImageURL),
	})
	if err != nil {
		return nil, fmt.Errorf("create enemy: %w", err)
	}

	enemy.ID = enemies.ID(saved.ID)
	return enemy, nil
}

func (r *enemyRepo) UpdateEnemy(ctx context.Context, exec db.DBTX, enemy *enemies.Enemy) (*enemies.Enemy, error) {
	q := r.q(exec)

	updated, err := q.UpdateEnemy(ctx, db.UpdateEnemyParams{
		ID:          int64(enemy.ID),
		Name:        enemy.Name,
		Type:        string(enemy.Type),
		Tier:        int16(enemy.Tier),
		Description: enemy.Description,
		HpMax:       int32(enemy.HpMax),
		Ac:          int32(enemy.ArmorClass),
		Damage:      int32(enemy.Damage),
		XpReward:    int32(enemy.XPReward),
		ScoreReward: int32(enemy.ScoreReward),
		ImageUrl:    toText(enemy.ImageURL),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("update enemy %d: %w", enemy.ID, enemies.ErrNotFound)
		}
		return nil, fmt.Errorf("update enemy %d: %w", enemy.ID, err)
	}

	return toDomainEnemy(updated), nil
}

func (r *enemyRepo) DeleteEnemy(ctx context.Context, exec db.DBTX, enemyID enemies.ID) error {
	q := r.q(exec)

	rows, err := q.DeleteEnemy(ctx, int64(enemyID))
	if err != nil {
		return fmt.Errorf("delete enemy %d: %w", enemyID, err)
	}
	if rows == 0 {
		return fmt.Errorf("delete enemy %d: %w", enemyID, enemies.ErrNotFound)
	}
	return nil
}

func (r *enemyRepo) SeedEnemies(ctx context.Context, exec db.DBTX, list []enemies.Enemy) error {
	q := r.q(exec)

	params := make([]db.SeedEnemiesParams, len(list))
	for i, e := range list {
		params[i] = db.SeedEnemiesParams{
			Name:        e.Name,
			Type:        string(e.Type),
			Tier:        int16(e.Tier),
			Description: e.Description,
			HpMax:       int32(e.HpMax),
			Ac:          int32(e.ArmorClass),
			Damage:      int32(e.Damage),
			XpReward:    int32(e.XPReward),
			ScoreReward: int32(e.ScoreReward),
			ImageUrl:    toText(e.ImageURL),
		}
	}

	if _, err := q.SeedEnemies(ctx, params); err != nil {
		return fmt.Errorf("seed enemies: %w", err)
	}
	return nil
}

func toDomainEnemy(row *db.Enemy) *enemies.Enemy {
	return &enemies.Enemy{
		ID:          enemies.ID(row.ID),
		Name:        row.Name,
		Type:        enemies.Type(row.Type),
		Tier:        uint8(row.Tier),
		Description: row.Description,
		HpMax:       uint(row.HpMax),
		ArmorClass:  uint8(row.Ac),
		Damage:      uint(row.Damage),
		XPReward:    uint(row.XpReward),
		ScoreReward: uint(row.ScoreReward),
		ImageURL:    fromText(row.ImageUrl),
	}
}
