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
	_, existingCount, err := s.repo.FindAllItems(ctx, s.pool, 1, 0)
	if err == nil && existingCount > 0 {
		return nil
	}

	var items []Item
	if err := json.Unmarshal(itemPoolData, &items); err != nil {
		return fmt.Errorf("unmarshal item pool: %w", err)
	}

	for i := range items {
		items[i].ID = 0
		if !items[i].IsValid() {
			return fmt.Errorf("invalid item in pool: %s", items[i].Name)
		}
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err = s.repo.SeedItems(ctx, tx, items); err != nil {
		log.Println("Error seeding items:", err)
		return fmt.Errorf("seed items: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		log.Println("Error commiting transaction:", err)
		return fmt.Errorf("commit tx: %w", err)
	}

	log.Printf("seeded %d items from pool", len(items))
	return nil
}

func (s *Service) GetAll(ctx context.Context, page, limit int) ([]Item, uint64, error) {
	offset := (page - 1) * limit
	return s.repo.FindAllItems(ctx, s.pool, limit, offset)
}

func (s *Service) GetByID(ctx context.Context, id ID) (*Item, error) {
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
		log.Println("Error finding item:", err)
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
		log.Println("Error starting transaction:", err)
		return nil, err
	}
	defer tx.Rollback(ctx)

	item, err = s.repo.SaveItem(ctx, tx, item)
	if err != nil {
		log.Println("Error saving item:", err)
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		log.Println("Error commiting transaction:", err)
		return nil, err
	}

	if err = s.cache.Set(ctx, fmt.Sprintf("%d", item.ID), item, 0); err != nil {
		log.Printf("cache set item %d: %v", item.ID, err)
	}
	return item, nil
}

type BulkCreateResult struct {
	Created      []Item   `json:"created"`
	Skipped      []string `json:"skipped"`
	CreatedCount int      `json:"created_count"`
	SkippedCount int      `json:"skipped_count"`
}

func (s *Service) BulkCreate(ctx context.Context, list []Item) (*BulkCreateResult, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return nil, err
	}
	defer tx.Rollback(ctx)

	result := &BulkCreateResult{
		Created: make([]Item, 0, len(list)),
		Skipped: make([]string, 0),
	}

	for i := range list {
		it := list[i]
		it.ID = 0
		saved, inserted, err := s.repo.SaveItemIfNotExists(ctx, tx, &it)
		if err != nil {
			log.Println("Error bulk saving item:", err)
			return nil, err
		}
		if inserted {
			result.Created = append(result.Created, *saved)
		} else {
			result.Skipped = append(result.Skipped, fmt.Sprintf("%s (Rareza %d)", it.Name, it.Rarity))
		}
	}

	if err = tx.Commit(ctx); err != nil {
		log.Println("Error committing transaction:", err)
		return nil, err
	}

	result.CreatedCount = len(result.Created)
	result.SkippedCount = len(result.Skipped)
	return result, nil
}

func (s *Service) UpdateItem(ctx context.Context, item *Item) (*Item, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return nil, err
	}
	defer tx.Rollback(ctx)

	updated, err := s.repo.UpdateItem(ctx, tx, item)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		log.Println("Error committing transaction:", err)
		return nil, err
	}

	if s.cache != nil {
		if err = s.cache.Set(ctx, fmt.Sprintf("%d", updated.ID), updated, 0); err != nil {
			log.Printf("cache set item %d: %v", updated.ID, err)
		}
	}
	return updated, nil
}

func (s *Service) DeleteItem(ctx context.Context, id ID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer tx.Rollback(ctx)

	if err = s.repo.DeleteItem(ctx, tx, id); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		log.Println("Error committing transaction:", err)
		return err
	}

	if s.cache != nil {
		if err = s.cache.Delete(ctx, fmt.Sprintf("%d", id)); err != nil {
			log.Printf("cache delete item %d: %v", id, err)
		}
	}
	return nil
}
