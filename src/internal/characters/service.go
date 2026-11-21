package characters

import (
	"context"
	"dZev1/character-gallery/internal/cache"
	"dZev1/character-gallery/internal/metrics"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/google/uuid"
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

func (s *Service) Create(ctx context.Context, character *Character, ownerID uuid.UUID) (*Character, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return nil, err
	}
	defer tx.Rollback(ctx)

	character, err = s.repo.SaveCharacter(ctx, tx, character, ownerID)
	if err != nil {
		log.Println("Error creating character:", err)
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		log.Println("Error committing transaction:", err)
		return nil, err
	}

	if err = s.cache.Set(ctx, fmt.Sprintf("%d", character.ID), character, 0); err != nil {
		log.Printf("cache set character %d: %v", character.ID, err)
	}

	return character, nil
}

func (s *Service) GetByID(ctx context.Context, id ID) (*Character, error) {
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
		log.Println("Error getting character:", err)
		return nil, err
	}

	if err = s.cache.Set(ctx, fmt.Sprintf("%d", id), character, 0); err != nil {
		log.Printf("cache set character %d: %v", id, err)
	}
	return character, nil
}

func (s *Service) GetAll(ctx context.Context, page, limit int) ([]Character, uint64, error) {
	offset := (page - 1) * limit
	return s.repo.FindAllCharacters(ctx, s.pool, limit, offset)
}

func (s *Service) GetByOwner(ctx context.Context, ownerID uuid.UUID, page, limit int) ([]Character, uint64, error) {
	offset := (page - 1) * limit
	return s.repo.FindCharactersByOwner(ctx, s.pool, ownerID, limit, offset)
}

func (s *Service) Heal(ctx context.Context, id ID, healAmount uint8) (*Character, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.Println("Error beginning transaction:", err)
		return nil, err
	}
	defer tx.Rollback(ctx)

	updated, err := s.repo.HealCharacter(ctx, tx, id, healAmount)
	if err != nil {
		log.Println("Error healing character:", err)
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		log.Println("Error committing transaction:", err)
		return nil, err
	}

	if s.cache != nil {
		if err = s.cache.Delete(ctx, fmt.Sprintf("%d", id)); err != nil {
			log.Printf("cache delete character %d: %v", id, err)
		}
	}

	return updated, nil
}

func (s *Service) Prestige(ctx context.Context, id ID, ownerID uuid.UUID) (*Character, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.Println("Error beginning transaction:", err)
		return nil, err
	}
	defer tx.Rollback(ctx)

	existing, err := s.repo.FindCharacter(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if ownerID != uuid.Nil && existing.OwnerID != ownerID {
		return nil, ErrNotFound
	}
	if existing.Level < 10 {
		return nil, ErrPrestigeRequirement
	}

	// Calculate new level-1 HpMax after prestige +1.
	existing.Prestige++
	newHpMax := existing.StartingHP()

	updated, err := s.repo.PrestigeCharacter(ctx, tx, id, ownerID, newHpMax)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		log.Println("Error committing transaction:", err)
		return nil, err
	}

	metrics.CharacterPrestigesTotal.Inc()

	if s.cache != nil {
		if err = s.cache.Delete(ctx, fmt.Sprintf("%d", id)); err != nil {
			log.Printf("cache delete character %d: %v", id, err)
		}
	}

	return updated, nil
}

func (s *Service) InvalidateCache(ctx context.Context, id ID) {
	if s.cache != nil {
		if err := s.cache.Delete(ctx, fmt.Sprintf("%d", id)); err != nil {
			log.Printf("cache delete character %d: %v", id, err)
		}
	}
}

func (s *Service) Update(ctx context.Context, character *Character, ownerID uuid.UUID) (*Character, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.Println("Error beginning transaction:", err)
		return nil, err
	}
	defer tx.Rollback(ctx)

	updated, err := s.repo.UpdateCharacter(ctx, tx, character, ownerID)
	if err != nil {
		log.Println("Error updating character:", err)
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		log.Println("Error commiting transaction:", err)
		return nil, err
	}

	if err = s.cache.Delete(ctx, fmt.Sprintf("%d", character.ID)); err != nil {
		log.Printf("cache delete character %d: %v", character.ID, err)
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id ID, ownerID uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.Println("Error beginning transaction:", err)
		return err
	}
	defer tx.Rollback(ctx)

	err = s.repo.DeleteCharacter(ctx, tx, id, ownerID)
	if err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		log.Println("Error committing transaction:", err)
		return err
	}

	if err = s.cache.Delete(ctx, fmt.Sprintf("%d", id)); err != nil {
		log.Printf("cache delete character %d: %v", id, err)
	}

	return nil
}
