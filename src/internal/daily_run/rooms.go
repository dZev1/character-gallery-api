package daily_run

import (
	"context"
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/dice"
	"dZev1/character-gallery/internal/enemies"
	"dZev1/character-gallery/internal/inventory"
	"dZev1/character-gallery/internal/items"
	"dZev1/character-gallery/internal/postgres/db"
	"fmt"
	"time"
)

const (
	restHealPercent   = 30
	restRoomScore     = 100
	chestSuccessScore = 150
	chestFailScore    = 50
)

func (s *Service) resolveRoomEntry(
	ctx context.Context,
	exec db.DBTX,
	run *DailyRun,
	char *characters.Character,
	node GraphNode,
) (*characters.Character, *RunStateResponse, error) {
	resp := &RunStateResponse{
		CurrentNode: node,
		NextNodeIDs: node.NextNodeIDs,
	}

	switch node.Type {
	case NodeRest:
		updatedChar, logLine, err := s.enterRestRoom(ctx, exec, run, char)
		if err != nil {
			return nil, nil, err
		}
		resp.EventLog = []string{logLine}
		return updatedChar, resp, nil

	case NodeChest:
		return s.enterChestRoom(ctx, exec, run, char, node, resp)

	default:
		enemy, logLine, err := s.enterCombatRoom(ctx, exec, run, node)
		if err != nil {
			return nil, nil, err
		}
		resp.CurrentEnemy = enemy
		resp.EventLog = []string{logLine}
		return char, resp, nil
	}
}

func (s *Service) enterRestRoom(
	ctx context.Context,
	exec db.DBTX,
	run *DailyRun,
	char *characters.Character,
) (*characters.Character, string, error) {
	healAmt := uint8(max(1, (int(char.HpMax)*restHealPercent)/100))
	updatedChar, err := s.charRepo.HealCharacter(ctx, exec, char.ID, healAmt)
	if err != nil {
		return nil, "", err
	}

	run.Score += restRoomScore
	run.Status = InProgress
	run.CurrentEnemyID = nil
	run.CurrentEnemyHP = nil

	msg := fmt.Sprintf("Rested by the campfire and recovered up to %d HP (now %d/%d HP).", healAmt, updatedChar.HpCurrent, updatedChar.HpMax)
	return updatedChar, msg, nil
}

func (s *Service) enterChestRoom(
	ctx context.Context,
	exec db.DBTX,
	run *DailyRun,
	char *characters.Character,
	node GraphNode,
	resp *RunStateResponse,
) (*characters.Character, *RunStateResponse, error) {
	check, err := rollSkillCheck(char, node.CheckStat, node.CheckDC)
	if err != nil {
		return nil, nil, err
	}

	resp.CheckRoll = check
	run.Status = InProgress
	run.CurrentEnemyID = nil
	run.CurrentEnemyHP = nil

	if check.Success {
		run.Score += chestSuccessScore
		loot, err := s.grantTierLoot(ctx, exec, char, run.Tier)
		if err == nil && loot != nil {
			resp.LootItem = loot
			resp.EventLog = []string{fmt.Sprintf("Skill check (%s DC %d) succeeded with %d! Found %s.", check.Stat, check.DC, check.Total, loot.Name)}
		} else {
			resp.EventLog = []string{fmt.Sprintf("Skill check (%s DC %d) succeeded with %d!", check.Stat, check.DC, check.Total)}
		}
		return char, resp, nil
	}

	trapDmg := int(run.Tier)*3 + 2
	nextHP := max(1, int(char.HpCurrent)-trapDmg)
	updatedChar, err := s.charRepo.UpdateCharacterHP(ctx, exec, char.ID, uint8(nextHP))
	if err != nil {
		return nil, nil, err
	}

	run.Score += chestFailScore
	resp.EventLog = []string{fmt.Sprintf("Skill check (%s DC %d) failed with %d! Triggered a trap for %d damage (%d/%d HP).", check.Stat, check.DC, check.Total, trapDmg, updatedChar.HpCurrent, updatedChar.HpMax)}
	return updatedChar, resp, nil
}

func (s *Service) enterCombatRoom(
	ctx context.Context,
	exec db.DBTX,
	run *DailyRun,
	node GraphNode,
) (*enemies.Enemy, string, error) {
	enemy, err := s.pickEnemyForRoom(ctx, exec, run.RunDate, run.Tier, node)
	if err != nil {
		return nil, "", err
	}

	enemyID := enemy.ID
	enemyHP := enemy.HpMax
	run.Status = InCombat
	run.CurrentEnemyID = &enemyID
	run.CurrentEnemyHP = &enemyHP

	msg := fmt.Sprintf("Entered Floor %d/%d: %s (%d HP, AC %d) blocks your path!", node.Floor, run.TotalFloors, enemy.Name, enemy.HpMax, enemy.ArmorClass)
	return enemy, msg, nil
}

func (s *Service) pickEnemyForRoom(ctx context.Context, exec db.DBTX, runDate time.Time, tier uint8, node GraphNode) (*enemies.Enemy, error) {
	enemyType := enemyTypeForNode(node.Type)

	pool, err := s.enemyRepo.FindEnemiesByTierAndType(ctx, exec, tier, enemyType)
	if err != nil {
		return nil, err
	}
	if len(pool) == 0 {
		pool, err = s.enemyRepo.FindEnemiesByTier(ctx, exec, tier)
		if err != nil {
			return nil, err
		}
		if len(pool) == 0 {
			return nil, ErrNoEnemiesConfigured
		}
	}

	idx := DeterministicEnemyIndex(runDate, tier, node.ID, len(pool))
	chosen := pool[idx]
	return &chosen, nil
}

func (s *Service) grantTierLoot(ctx context.Context, exec db.DBTX, char *characters.Character, tier uint8) (*items.Item, error) {
	minRarity := tier
	maxRarity := min(uint8(5), tier+1)
	if tier == 5 {
		minRarity = 5
		maxRarity = 5
	}

	loot, err := s.itemRepo.FindRandomItemByRarityRange(ctx, exec, minRarity, maxRarity)
	if err != nil {
		loot, err = s.itemRepo.FindRandomItemByRarityRange(ctx, exec, 1, maxRarity)
		if err != nil {
			return nil, err
		}
	}

	_, err = s.invRepo.AddItemToCharacter(ctx, exec, inventory.RepositoryParam{
		CharacterID: char.ID,
		OwnerID:     char.OwnerID,
		ItemID:      loot.ID,
		Quantity:    1,
	})
	if err != nil {
		return nil, err
	}
	return loot, nil
}

func enemyTypeForNode(t NodeType) enemies.Type {
	switch t {
	case NodeBoss:
		return enemies.Boss
	case NodeMiniBoss:
		return enemies.MiniBoss
	case NodeElite:
		return enemies.Elite
	default:
		return enemies.Normal
	}
}

func rollSkillCheck(char *characters.Character, statName string, dc int) (*SkillCheckResult, error) {
	roll, err := dice.RollD20()
	if err != nil {
		return nil, err
	}
	mod := characters.StatMod(statValueByName(char, statName))
	total := roll + mod
	return &SkillCheckResult{
		Stat:     statName,
		DC:       dc,
		Roll:     roll,
		Modifier: mod,
		Total:    total,
		Success:  roll == 20 || (roll != 1 && total >= dc),
	}, nil
}

func statValueByName(char *characters.Character, statName string) uint8 {
	if char.Stats == nil {
		return 10
	}
	switch statName {
	case "strength":
		return char.Stats.Strength
	case "dexterity":
		return char.Stats.Dexterity
	case "constitution":
		return char.Stats.Constitution
	case "intelligence":
		return char.Stats.Intelligence
	case "wisdom":
		return char.Stats.Wisdom
	case "charisma":
		return char.Stats.Charisma
	default:
		return 10
	}
}
