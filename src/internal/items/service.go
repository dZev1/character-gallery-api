package items

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

//go:embed item_pool.json
var itemPoolData []byte

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

func (s *Service) SeedItems(ctx context.Context) error {
	var items []Item
	if err := json.Unmarshal(itemPoolData, &items); err != nil {
		return fmt.Errorf("unmarshal item pool: %w", err)
	}

	for i := range items {
		items[i].ID = 0
		if !items[i].Validate() {
			return fmt.Errorf("invalid item in pool: %s", items[i].Name)
		}
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.repo.SeedItems(ctx, tx, items); err != nil {
		return fmt.Errorf("seed items: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	log.Printf("seeded %d items from pool", len(items))
	return nil
}

func (s *Service) GetAll(ctx context.Context, page, limit int) ([]Item, uint64, error) {
	offset := (page - 1) * limit
	return s.repo.FindAllItems(ctx, s.pool, limit, offset)
}

func (s *Service) GetByID(ctx context.Context, id ItemID) (*Item, error) {
	item := &Item{}
	key := fmt.Sprintf("%d", id)

	err := s.cache.Get(ctx, key, item)
	if err == nil {
		return item, nil
	}
	if !errors.Is(err, cache.ErrMiss) {
		log.Printf("cache get item %d: %v", id, err)
	}

	item, err = s.repo.FindItem(ctx, s.pool, id)
	if err != nil {
		return nil, err
	}

	if err = s.cache.Set(ctx, key, item, 0); err != nil {
		log.Printf("cache set item %d: %v", id, err)
	}

	return item, nil
}

func (s *Service) CreateItem(ctx context.Context, item *Item) (*Item, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	item, err = s.repo.SaveItem(ctx, tx, item)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	if err = s.cache.Set(ctx, fmt.Sprintf("%d", item.ID), item, 0); err != nil {
		log.Printf("cache set item %d: %v", item.ID, err)
	}
	return item, nil
}
