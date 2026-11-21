package inventory

import (
	"context"
	"dZev1/character-gallery/internal/cache"
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/items"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo      Repository
	charRepo  characters.Repository
	pool      *pgxpool.Pool
	cache     cache.Cache
	charCache cache.Cache
}

func NewService(repo Repository, pool *pgxpool.Pool, cache cache.Cache) *Service {
	return &Service{
		repo:  repo,
		pool:  pool,
		cache: cache,
	}
}

// WithCharacterStore attaches the character repository and cache so inventory
// actions like UseItem and EquipItem can update/read character state atomically.
func (s *Service) WithCharacterStore(charRepo characters.Repository, charCache cache.Cache) *Service {
	s.charRepo = charRepo
	s.charCache = charCache
	return s
}

type EquipResult struct {
	CharacterID   characters.ID `json:"id"`
	Name          string        `json:"name"`
	AC            int           `json:"ac"`
	HpCurrent     uint8         `json:"hp_current"`
	HpMax         uint8         `json:"hp_max"`
	Level         uint8         `json:"level"`
	Xp            uint64        `json:"xp"`
	EquippedItems []items.ID    `json:"equipped_items"`
}

func (s *Service) EquipItem(ctx context.Context, characterID characters.ID, ownerID uuid.UUID, itemID items.ID, equip bool) (*EquipResult, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := s.repo.SetItemEquipped(ctx, tx, characterID, ownerID, itemID, equip); err != nil {
		return nil, err
	}

	equipped, err := s.repo.GetCharacterEquippedItems(ctx, tx, characterID)
	if err != nil {
		return nil, err
	}

	var result *EquipResult
	if s.charRepo != nil {
		char, err := s.charRepo.FindCharacter(ctx, tx, characterID)
		if err != nil {
			return nil, err
		}
		totalDef := 0
		equippedIDs := make([]items.ID, len(equipped))
		for i, eq := range equipped {
			equippedIDs[i] = eq.Item.ID
			if eq.Item.Defense != nil {
				totalDef += int(*eq.Item.Defense)
			}
		}
		result = &EquipResult{
			CharacterID:   char.ID,
			Name:          char.Name,
			AC:            char.ArmorClass(totalDef),
			HpCurrent:     char.HpCurrent,
			HpMax:         char.HpMax,
			Level:         char.Level,
			Xp:            char.Xp,
			EquippedItems: equippedIDs,
		}
	} else {
		equippedIDs := make([]items.ID, len(equipped))
		for i, eq := range equipped {
			equippedIDs[i] = eq.Item.ID
		}
		result = &EquipResult{
			CharacterID:   characterID,
			EquippedItems: equippedIDs,
		}
	}

	if err = tx.Commit(ctx); err != nil {
		log.Println("Error committing transaction:", err)
		return nil, err
	}

	if s.cache != nil {
		if err = s.cache.Delete(ctx, fmt.Sprintf("%d", characterID)); err != nil {
			log.Printf("cache delete inventory for character %d: %v", characterID, err)
		}
	}

	return result, nil
}

func (s *Service) GetEquippedItems(ctx context.Context, characterID characters.ID) ([]InventoryItem, error) {
	return s.repo.GetCharacterEquippedItems(ctx, s.pool, characterID)
}

func (s *Service) UseItem(ctx context.Context, characterID characters.ID, ownerID uuid.UUID, itemID items.ID) (*characters.Character, error) {
	if s.charRepo == nil {
		return nil, errors.New("character repository not configured on inventory service")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return nil, err
	}
	defer tx.Rollback(ctx)

	removed, err := s.repo.RemoveItemFromCharacter(ctx, tx, RepositoryParam{
		CharacterID: characterID,
		OwnerID:     ownerID,
		ItemID:      itemID,
		Quantity:    1,
	})
	if err != nil {
		return nil, err
	}

	if removed.Item == nil || removed.Item.HealAmount == nil || *removed.Item.HealAmount == 0 {
		return nil, ErrItemNotConsumable
	}

	healAmt := *removed.Item.HealAmount
	if healAmt > 255 {
		healAmt = 255
	}

	updatedChar, err := s.charRepo.HealCharacter(ctx, tx, characterID, uint8(healAmt))
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		log.Println("Error committing transaction:", err)
		return nil, err
	}

	if s.cache != nil {
		if err = s.cache.Delete(ctx, fmt.Sprintf("%d", characterID)); err != nil {
			log.Printf("cache delete inventory for character %d: %v", characterID, err)
		}
	}
	if s.charCache != nil {
		if err = s.charCache.Delete(ctx, fmt.Sprintf("%d", characterID)); err != nil {
			log.Printf("cache delete character %d: %v", characterID, err)
		}
	}

	return updatedChar, nil
}

func (s *Service) AddItem(ctx context.Context, characterID characters.ID, ownerID uuid.UUID, itemID items.ID, quantity uint8) (*InventoryItem, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return nil, err
	}
	defer tx.Rollback(ctx)

	repoParam := RepositoryParam{
		CharacterID: characterID,
		OwnerID:     ownerID,
		ItemID:      itemID,
		Quantity:    quantity,
	}

	invItem, err := s.repo.AddItemToCharacter(ctx, tx, repoParam)
	if err != nil {
		log.Println("Error adding item to character:", err)
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		log.Println("Error committing transaction:", err)
		return nil, err
	}

	if err = s.cache.Delete(ctx, fmt.Sprintf("%d", characterID)); err != nil {
		log.Printf("cache delete inventory for character %d: %v", characterID, err)
	}

	return invItem, nil
}

func (s *Service) DeleteItem(ctx context.Context, characterID characters.ID, ownerID uuid.UUID, itemID items.ID, quantity uint8) (*InventoryItem, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return nil, err
	}
	defer tx.Rollback(ctx)

	repoParam := RepositoryParam{
		CharacterID: characterID,
		OwnerID:     ownerID,
		ItemID:      itemID,
		Quantity:    quantity,
	}

	item, err := s.repo.RemoveItemFromCharacter(ctx, tx, repoParam)
	if err != nil {
		log.Println("Error removing item from character:", err)
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		log.Println("Error committing transaction:", err)
		return nil, err
	}

	if err = s.cache.Delete(ctx, fmt.Sprintf("%d", characterID)); err != nil {
		log.Printf("cache delete inventory for character %d: %v", characterID, err)
	}

	return item, nil
}

func (s *Service) GetByCharacterID(ctx context.Context, characterID characters.ID) ([]InventoryItem, error) {
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
		log.Println("Error getting inventory for character:", err)
		return nil, err
	}

	if err = s.cache.Set(ctx, key, inv, 0); err != nil {
		log.Printf("cache set inventory for character %d: %v", characterID, err)
	}
	return inv, nil
}
