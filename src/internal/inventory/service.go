package inventory

import (
	"context"
	"dZev1/character-gallery/internal/cache"
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/items"
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

func (s *Service) AddItem(ctx context.Context, characterID characters.CharacterID, itemID items.ItemID, quantity uint8) (*InventoryItem, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	repoParam := RepositoryParam{
		CharacterID: characterID,
		ItemID:      itemID,
		Quantity:    quantity,
	}

	invItem, err := s.repo.AddItemToCharacter(ctx, tx, repoParam)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	if err = s.cache.Delete(ctx, fmt.Sprintf("%d", characterID)); err != nil {
		log.Printf("cache delete inventory for character %d: %v", characterID, err)
	}

	return invItem, nil
}

func (s *Service) DeleteItem(ctx context.Context, characterID characters.CharacterID, itemID items.ItemID, quantity uint8) (*InventoryItem, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	repoParam := RepositoryParam{
		CharacterID: characterID,
		ItemID:      itemID,
		Quantity:    quantity,
	}

	item, err := s.repo.RemoveItemFromCharacter(ctx, tx, repoParam)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	if err = s.cache.Delete(ctx, fmt.Sprintf("%d", characterID)); err != nil {
		log.Printf("cache delete inventory for character %d: %v", characterID, err)
	}

	return item, nil
}

func (s *Service) GetByCharacterID(ctx context.Context, characterID characters.CharacterID) ([]InventoryItem, error) {
	key := fmt.Sprintf("%d", characterID)

	var inv []InventoryItem
	err := s.cache.Get(ctx, key, &inv)
	if err == nil {
		return inv, nil
	}
	if !errors.Is(err, cache.ErrMiss) {
		log.Printf("cache get inventory for character %d: %v", characterID, err)
	}

	inv, err = s.repo.GetCharacterInventory(ctx, s.pool, characterID)
	if err != nil {
		return nil, err
	}

	if err = s.cache.Set(ctx, key, inv, 0); err != nil {
		log.Printf("cache set inventory for character %d: %v", characterID, err)
	}
	return inv, nil
}
