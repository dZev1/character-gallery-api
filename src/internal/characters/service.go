package characters

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

func (s *Service) Create(ctx context.Context, character *Character) (*Character, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	character, err = s.repo.SaveCharacter(ctx, tx, character)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	if err = s.cache.Set(ctx, fmt.Sprintf("%d", character.ID), character, 0); err != nil {
		log.Printf("cache set character %d: %v", character.ID, err)
	}

	return character, nil
}

func (s *Service) GetByID(ctx context.Context, id CharacterID) (*Character, error) {
	character := &Character{}
	err := s.cache.Get(ctx, fmt.Sprintf("%d", id), character)
	if err == nil {
		return character, nil
	}
	if !errors.Is(err, cache.ErrMiss) {
		log.Printf("cache get character %d: %v", id, err)
	}

	character, err = s.repo.FindCharacter(ctx, s.pool, id)
	if err != nil {
		return nil, err
	}

	if err = s.cache.Set(ctx, fmt.Sprintf("%d", id), character, 0); err != nil {
		log.Printf("cache set character %d: %v", id, err)
	}
	return character, nil
}

func (s *Service) GetAll(ctx context.Context, page int) ([]Character, uint64, error) {
	return s.repo.FindAllCharacters(ctx, s.pool, page)
}

func (s *Service) Update(ctx context.Context, character *Character) (*Character, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	updated, err := s.repo.UpdateCharacter(ctx, tx, character)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	if err = s.cache.Delete(ctx, fmt.Sprintf("%d", character.ID)); err != nil {
		log.Printf("cache delete character %d: %v", character.ID, err)
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id CharacterID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = s.repo.DeleteCharacter(ctx, tx, id)
	if err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	if err = s.cache.Delete(ctx, fmt.Sprintf("%d", id)); err != nil {
		log.Printf("cache delete character %d: %v", id, err)
	}

	return nil
}
