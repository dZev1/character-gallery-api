package daily_run

import (
	"context"
	"dZev1/character-gallery/internal/cache"
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/enemies"
	"dZev1/character-gallery/internal/inventory"
	"dZev1/character-gallery/internal/items"
	"dZev1/character-gallery/internal/metrics"
	"dZev1/character-gallery/internal/postgres/db"
	"errors"
	"fmt"
	"log"
	"slices"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo      Repository
	charRepo  characters.Repository
	enemyRepo enemies.Repository
	invRepo   inventory.Repository
	itemRepo  items.Repository
	pool      *pgxpool.Pool
	charCache cache.Cache
	invCache  cache.Cache
}

func NewService(
	repo Repository,
	charRepo characters.Repository,
	enemyRepo enemies.Repository,
	invRepo inventory.Repository,
	itemRepo items.Repository,
	pool *pgxpool.Pool,
	charCache cache.Cache,
	invCache cache.Cache,
) *Service {
	return &Service{
		repo:      repo,
		charRepo:  charRepo,
		enemyRepo: enemyRepo,
		invRepo:   invRepo,
		itemRepo:  itemRepo,
		pool:      pool,
		charCache: charCache,
		invCache:  invCache,
	}
}

func TodayUTC() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func (s *Service) GetDailyView(ctx context.Context, userID uuid.UUID, tier uint8) (*DailyView, error) {
	if tier < 1 || tier > 5 {
		return nil, ErrInvalidTier
	}

	today := TodayUTC()
	activeRuns, err := s.repo.ListActiveDailyRunsByUser(ctx, s.pool, today, userID)
	if err != nil {
		return nil, err
	}
	userScores, err := s.repo.GetUserDailyScores(ctx, s.pool, today, userID)
	if err != nil {
		return nil, err
	}

	return &DailyView{
		Dungeon:    GenerateDungeon(today, tier),
		ActiveRuns: activeRuns,
		UserScores: userScores,
	}, nil
}

func (s *Service) GetLeaderboard(ctx context.Context, tier uint8, page, limit int) ([]DailyScore, uint64, error) {
	if tier < 1 || tier > 5 {
		return nil, 0, ErrInvalidTier
	}
	offset := (page - 1) * limit
	return s.repo.GetDailyLeaderboard(ctx, s.pool, TodayUTC(), tier, limit, offset)
}

func (s *Service) StartOrResumeRun(ctx context.Context, userID uuid.UUID, isAdmin bool, charID characters.ID) (*RunStateResponse, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	char, err := s.findAuthorizedCharacter(ctx, tx, charID, userID, isAdmin)
	if err != nil {
		return nil, err
	}

	today := TodayUTC()
	tier := char.Tier()
	dungeon := GenerateDungeon(today, tier)

	existing, err := s.repo.GetDailyRun(ctx, tx, today, tier, charID)
	if err == nil {
		if existing.Status == Victory || existing.Status == Defeated {
			return nil, ErrRunAlreadyFinished
		}
		resp, err := s.buildStateResponse(ctx, tx, existing, char, dungeon)
		if err != nil {
			return nil, err
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return resp, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	if char.HpCurrent == 0 {
		return nil, ErrCharacterDead
	}

	startNode, _ := dungeon.NodeByID(dungeon.StartNodeID)
	enemy, err := s.pickEnemyForRoom(ctx, tx, today, tier, startNode)
	if err != nil {
		return nil, err
	}

	enemyID := enemy.ID
	enemyHP := enemy.HpMax
	savedRun, err := s.repo.CreateDailyRun(ctx, tx, &DailyRun{
		RunDate:        today,
		Tier:           tier,
		UserID:         char.OwnerID,
		CharacterID:    char.ID,
		TotalFloors:    dungeon.TotalFloors,
		CurrentFloor:   1,
		CurrentNodeID:  startNode.ID,
		VisitedNodes:   []string{startNode.ID},
		Status:         InCombat,
		CurrentEnemyID: &enemyID,
		CurrentEnemyHP: &enemyHP,
	})
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	metrics.DailyRunsStartedTotal.WithLabelValues(strconv.Itoa(int(tier))).Inc()

	return &RunStateResponse{
		Run:          savedRun,
		Character:    char,
		CurrentNode:  startNode,
		CurrentEnemy: enemy,
		NextNodeIDs:  startNode.NextNodeIDs,
		EventLog:     []string{fmt.Sprintf("Entered Tier %d Daily Dungeon (Floor 1/%d). %s appears!", tier, dungeon.TotalFloors, enemy.Name)},
	}, nil
}

func (s *Service) Move(ctx context.Context, userID uuid.UUID, isAdmin bool, charID characters.ID, targetNodeID string) (*RunStateResponse, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	char, err := s.findAuthorizedCharacter(ctx, tx, charID, userID, isAdmin)
	if err != nil {
		return nil, err
	}

	run, err := s.repo.GetActiveDailyRunForCharacter(ctx, tx, TodayUTC(), charID)
	if err != nil {
		return nil, err
	}
	if run.Status == InCombat {
		return nil, ErrRunInCombat
	}

	nextNode, err := validateNextNode(run, targetNodeID)
	if err != nil {
		return nil, err
	}

	run.CurrentFloor = nextNode.Floor
	run.CurrentNodeID = nextNode.ID
	run.VisitedNodes = append(run.VisitedNodes, nextNode.ID)

	updatedChar, resp, err := s.resolveRoomEntry(ctx, tx, run, char, nextNode)
	if err != nil {
		return nil, err
	}

	updatedRun, err := s.repo.UpdateDailyRunState(ctx, tx, run)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.invalidateCaches(ctx, updatedChar.ID)
	resp.Run = updatedRun
	resp.Character = updatedChar
	return resp, nil
}

func (s *Service) CombatAction(
	ctx context.Context,
	userID uuid.UUID,
	isAdmin bool,
	charID characters.ID,
	action ActionType,
	itemID *items.ID,
) (*RunStateResponse, error) {
	if !action.IsValid() {
		return nil, ErrInvalidAction
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	char, err := s.findAuthorizedCharacter(ctx, tx, charID, userID, isAdmin)
	if err != nil {
		return nil, err
	}

	run, err := s.repo.GetActiveDailyRunForCharacter(ctx, tx, TodayUTC(), charID)
	if err != nil {
		return nil, err
	}
	if run.Status != InCombat || run.CurrentEnemyID == nil || run.CurrentEnemyHP == nil {
		return nil, ErrRunNotInCombat
	}

	enemy, err := s.enemyRepo.FindEnemy(ctx, tx, *run.CurrentEnemyID)
	if err != nil {
		return nil, err
	}
	equipped, err := s.invRepo.GetCharacterEquippedItems(ctx, tx, char.ID)
	if err != nil {
		return nil, err
	}

	weaponDmg, gearDef := equippedBonuses(equipped)
	dungeon := GenerateDungeon(run.RunDate, run.Tier)
	currentNode, _ := dungeon.NodeByID(run.CurrentNodeID)

	cc := &combatContext{
		run:         run,
		char:        char,
		enemy:       enemy,
		currentNode: currentNode,
		weaponDmg:   weaponDmg,
		playerAC:    char.ArmorClass(gearDef),
		primaryMod:  characters.StatMod(char.PrimaryStatValue()),
	}

	resp, err := s.resolveCombatTurn(ctx, tx, cc, action, itemID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.invalidateCaches(ctx, char.ID)
	return resp, nil
}

func (s *Service) findAuthorizedCharacter(ctx context.Context, exec db.DBTX, charID characters.ID, userID uuid.UUID, isAdmin bool) (*characters.Character, error) {
	char, err := s.charRepo.FindCharacter(ctx, exec, charID)
	if err != nil {
		return nil, err
	}
	if !isAdmin && char.OwnerID != userID {
		return nil, characters.ErrNotFound
	}
	return char, nil
}

func (s *Service) buildStateResponse(ctx context.Context, exec db.DBTX, run *DailyRun, char *characters.Character, dungeon *DungeonGraph) (*RunStateResponse, error) {
	node, _ := dungeon.NodeByID(run.CurrentNodeID)
	resp := &RunStateResponse{
		Run:         run,
		Character:   char,
		CurrentNode: node,
		NextNodeIDs: node.NextNodeIDs,
	}
	if run.CurrentEnemyID != nil {
		enemy, err := s.enemyRepo.FindEnemy(ctx, exec, *run.CurrentEnemyID)
		if err == nil {
			resp.CurrentEnemy = enemy
		}
	}
	return resp, nil
}

func validateNextNode(run *DailyRun, targetNodeID string) (GraphNode, error) {
	dungeon := GenerateDungeon(run.RunDate, run.Tier)
	currentNode, ok := dungeon.NodeByID(run.CurrentNodeID)
	if !ok || !slices.Contains(currentNode.NextNodeIDs, targetNodeID) {
		return GraphNode{}, ErrInvalidMove
	}
	nextNode, ok := dungeon.NodeByID(targetNodeID)
	if !ok {
		return GraphNode{}, ErrInvalidMove
	}
	return nextNode, nil
}

func (s *Service) invalidateCaches(ctx context.Context, charID characters.ID) {
	key := fmt.Sprintf("%d", charID)
	if s.charCache != nil {
		if err := s.charCache.Delete(ctx, key); err != nil {
			log.Printf("cache delete character %d: %v", charID, err)
		}
	}
	if s.invCache != nil {
		if err := s.invCache.Delete(ctx, key); err != nil {
			log.Printf("cache delete inventory %d: %v", charID, err)
		}
	}
}
