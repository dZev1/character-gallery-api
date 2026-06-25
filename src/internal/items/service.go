package items

import (
	"context"
	"dZev1/character-gallery/internal/cache"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

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

func (s *Service) GetAll(ctx context.Context) ([]Item, uint64, error) {
	return s.repo.FindAllItems(ctx, s.pool)
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
