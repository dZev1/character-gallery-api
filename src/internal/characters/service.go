package characters

import (
	"context"
	"dZev1/character-gallery/internal/cache"
	"dZev1/character-gallery/internal/metrics"
	"errors"
	"fmt"
	"log"
	"time"

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

	count, err := s.repo.CountCharactersByOwner(ctx, tx, ownerID)
	if err != nil {
		log.Println("Error counting user characters:", err)
		return nil, err
	}
	if count >= MaxCharactersPerOwner {
		return nil, ErrCharacterLimitReached
	}

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

func buildStatRollView(state *UserStatRoll, now time.Time) *StatRollView {
	view := &StatRollView{
		Rolls:          []int{},
		CanReroll:      true,
		CanCreate:      true,
		ClassModifiers: ClassModifiers,
	}
	if state == nil {
		return view
	}
	if len(state.Rolls) > 0 {
		view.Rolls = state.Rolls
	}
	if state.LastCharacterCreatedAt != nil {
		next := state.LastCharacterCreatedAt.Add(WeeklyCooldown)
		if now.Before(next) {
			view.CanCreate = false
			view.NextCreateAt = &next
			view.CanReroll = false
		}
	}
	if state.LastRerollAt != nil {
		next := state.LastRerollAt.Add(WeeklyCooldown)
		if now.Before(next) {
			view.CanReroll = false
			view.NextRerollAt = &next
		}
	}
	return view
}

func (s *Service) GetOrCreateStatRoll(ctx context.Context, userID uuid.UUID) (*StatRollView, error) {
	now := time.Now().UTC()

	state, err := s.repo.GetUserStatRoll(ctx, s.pool, userID)
	if err != nil {
		return nil, err
	}

	// If user is on weekly character creation cooldown, don't generate new rolls yet.
	if state != nil && state.LastCharacterCreatedAt != nil && now.Before(state.LastCharacterCreatedAt.Add(WeeklyCooldown)) {
		return buildStatRollView(state, now), nil
	}

	// If user already has an active 6-stat roll, return it unchanged.
	if state != nil && len(state.Rolls) == 6 {
		return buildStatRollView(state, now), nil
	}

	rolls, err := RollStatPool()
	if err != nil {
		return nil, fmt.Errorf("roll stat pool: %w", err)
	}

	nextState := &UserStatRoll{Rolls: rolls}
	if state != nil {
		nextState.LastRerollAt = state.LastRerollAt
		nextState.LastCharacterCreatedAt = state.LastCharacterCreatedAt
	}

	saved, err := s.repo.UpsertUserStatRoll(ctx, s.pool, userID, nextState)
	if err != nil {
		return nil, err
	}
	return buildStatRollView(saved, now), nil
}

func (s *Service) RerollStats(ctx context.Context, userID uuid.UUID) (*StatRollView, error) {
	now := time.Now().UTC()

	state, err := s.repo.GetUserStatRoll(ctx, s.pool, userID)
	if err != nil {
		return nil, err
	}
	if state != nil && state.LastCharacterCreatedAt != nil && now.Before(state.LastCharacterCreatedAt.Add(WeeklyCooldown)) {
		return nil, ErrCreateCooldown
	}
	if state != nil && state.LastRerollAt != nil && now.Before(state.LastRerollAt.Add(WeeklyCooldown)) {
		return nil, ErrRerollCooldown
	}

	rolls, err := RollStatPool()
	if err != nil {
		return nil, fmt.Errorf("roll stat pool: %w", err)
	}

	nextState := &UserStatRoll{
		Rolls:        rolls,
		LastRerollAt: &now,
	}
	if state != nil {
		nextState.LastCharacterCreatedAt = state.LastCharacterCreatedAt
	}

	saved, err := s.repo.UpsertUserStatRoll(ctx, s.pool, userID, nextState)
	if err != nil {
		return nil, err
	}
	return buildStatRollView(saved, now), nil
}

func (s *Service) CreateWithStatRoll(ctx context.Context, character *Character, ownerID uuid.UUID) (*Character, error) {
	now := time.Now().UTC()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	count, err := s.repo.CountCharactersByOwner(ctx, tx, ownerID)
	if err != nil {
		return nil, err
	}
	if count >= MaxCharactersPerOwner {
		return nil, ErrCharacterLimitReached
	}

	state, err := s.repo.GetUserStatRoll(ctx, tx, ownerID)
	if err != nil {
		return nil, err
	}
	if state != nil && state.LastCharacterCreatedAt != nil && now.Before(state.LastCharacterCreatedAt.Add(WeeklyCooldown)) {
		return nil, ErrCreateCooldown
	}
	if state == nil || len(state.Rolls) != 6 {
		return nil, ErrNoStatRoll
	}
	if !MatchesRollPool(character.Stats, state.Rolls) {
		return nil, ErrInvalidStatAllocation
	}

	character.Stats = ApplyClassModifiers(character.Class, character.Stats)
	character.Level = 1
	character.Xp = 0
	character.HpMax = character.StartingHP()
	character.HpCurrent = character.HpMax

	created, err := s.repo.SaveCharacter(ctx, tx, character, ownerID)
	if err != nil {
		return nil, err
	}

	if err = s.repo.ConsumeUserStatRoll(ctx, tx, ownerID); err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	if s.cache != nil {
		if err = s.cache.Set(ctx, fmt.Sprintf("%d", created.ID), created, 0); err != nil {
			log.Printf("cache set character %d: %v", created.ID, err)
		}
	}
	return created, nil
}
