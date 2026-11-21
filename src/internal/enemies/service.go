package enemies

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"dZev1/character-gallery/internal/cache"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed enemy_pool.json
var enemyPoolData []byte

type Service struct {
	repo  Repository
	pool  *pgxpool.Pool
	cache cache.Cache
}

func NewService(repo Repository, pool *pgxpool.Pool, cache cache.Cache) *Service {
	return &Service{
		repo:  repo,
		pool:  pool,
		cache: cache,
	}
}

func (s *Service) SeedEnemies(ctx context.Context) error {
	_, existingCount, err := s.repo.FindAllEnemies(ctx, s.pool, 1, 0)
	if err == nil && existingCount > 0 {
		return nil
	}

	var list []Enemy
	if err := json.Unmarshal(enemyPoolData, &list); err != nil {
		return fmt.Errorf("unmarshal enemy pool: %w", err)
	}

	for i := range list {
		list[i].ID = 0
		if !list[i].IsValid() {
			return fmt.Errorf("invalid enemy in pool: %s", list[i].Name)
		}
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err = s.repo.SeedEnemies(ctx, tx, list); err != nil {
		log.Println("Error seeding enemies:", err)
		return fmt.Errorf("seed enemies: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		log.Println("Error committing transaction:", err)
		return fmt.Errorf("commit tx: %w", err)
	}

	log.Printf("seeded %d enemies from pool", len(list))
	return nil
}

func (s *Service) GetAll(ctx context.Context, page, limit int) ([]Enemy, uint64, error) {
	offset := (page - 1) * limit
	return s.repo.FindAllEnemies(ctx, s.pool, limit, offset)
}

func (s *Service) GetByTier(ctx context.Context, tier uint8) ([]Enemy, error) {
	return s.repo.FindEnemiesByTier(ctx, s.pool, tier)
}

func (s *Service) GetByID(ctx context.Context, id ID) (*Enemy, error) {
	enemy := &Enemy{}
	key := fmt.Sprintf("%d", id)

	if s.cache != nil {
		err := s.cache.Get(ctx, key, enemy)
		if err == nil {
			return enemy, nil
		}
		if !errors.Is(err, cache.ErrMiss) {
			log.Printf("cache get enemy %d: %v", id, err)
		}
	}

	enemy, err := s.repo.FindEnemy(ctx, s.pool, id)
	if err != nil {
		log.Println("Error finding enemy:", err)
		return nil, err
	}

	if s.cache != nil {
		if err = s.cache.Set(ctx, key, enemy, 0); err != nil {
			log.Printf("cache set enemy %d: %v", id, err)
		}
	}

	return enemy, nil
}

func (s *Service) Create(ctx context.Context, enemy *Enemy) (*Enemy, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return nil, err
	}
	defer tx.Rollback(ctx)

	enemy, err = s.repo.SaveEnemy(ctx, tx, enemy)
	if err != nil {
		log.Println("Error saving enemy:", err)
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		log.Println("Error committing transaction:", err)
		return nil, err
	}

	if s.cache != nil {
		if err = s.cache.Set(ctx, fmt.Sprintf("%d", enemy.ID), enemy, 0); err != nil {
			log.Printf("cache set enemy %d: %v", enemy.ID, err)
		}
	}
	return enemy, nil
}

func (s *Service) Update(ctx context.Context, enemy *Enemy) (*Enemy, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return nil, err
	}
	defer tx.Rollback(ctx)

	updated, err := s.repo.UpdateEnemy(ctx, tx, enemy)
	if err != nil {
		log.Println("Error updating enemy:", err)
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		log.Println("Error committing transaction:", err)
		return nil, err
	}

	if s.cache != nil {
		if err = s.cache.Delete(ctx, fmt.Sprintf("%d", enemy.ID)); err != nil {
			log.Printf("cache delete enemy %d: %v", enemy.ID, err)
		}
	}
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id ID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer tx.Rollback(ctx)

	if err = s.repo.DeleteEnemy(ctx, tx, id); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		log.Println("Error committing transaction:", err)
		return err
	}

	if s.cache != nil {
		if err = s.cache.Delete(ctx, fmt.Sprintf("%d", id)); err != nil {
			log.Printf("cache delete enemy %d: %v", id, err)
		}
	}
	return nil
}
