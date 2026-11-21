package daily_run

import (
	"context"
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/dice"
	"dZev1/character-gallery/internal/enemies"
	"dZev1/character-gallery/internal/inventory"
	"dZev1/character-gallery/internal/items"
	"dZev1/character-gallery/internal/metrics"
	"dZev1/character-gallery/internal/postgres/db"
	"fmt"
	"strconv"
)

const (
	heavyAttackAccuracyPenalty = 3
	normalDropChancePercent    = 35
	roomClearBaseScore         = 100
	bossCompletionScoreBonus   = 500
	maxCharacterLevel          = 10
	maxCharacterHP             = 250
)

type combatContext struct {
	run         *DailyRun
	char        *characters.Character
	enemy       *enemies.Enemy
	currentNode GraphNode
	weaponDmg   int
	playerAC    int
	primaryMod  int
}

func (s *Service) resolveCombatTurn(
	ctx context.Context,
	exec db.DBTX,
	cc *combatContext,
	action ActionType,
	itemID *items.ID,
) (*RunStateResponse, error) {
	cc.run.TurnsTaken++
	enemyHP := int(*cc.run.CurrentEnemyHP)
	turnRes := &CombatTurnResult{Action: action}
	var logs []string

	playerLog, err := s.applyPlayerAction(ctx, exec, cc, action, itemID, &enemyHP, turnRes)
	if err != nil {
		return nil, err
	}
	logs = append(logs, playerLog)

	resp := &RunStateResponse{
		CurrentNode: cc.currentNode,
		NextNodeIDs: cc.currentNode.NextNodeIDs,
		TurnResult:  turnRes,
	}

	if enemyHP <= 0 {
		return s.handleEnemyDefeated(ctx, exec, cc, resp, logs)
	}

	enemyLog, defeated, err := s.applyEnemyCounterAttack(ctx, exec, cc, enemyHP, turnRes, resp)
	if err != nil {
		return nil, err
	}
	logs = append(logs, enemyLog)

	if defeated {
		updatedRun, err := s.repo.UpdateDailyRunState(ctx, exec, cc.run)
		if err != nil {
			return nil, err
		}
		resp.Run = updatedRun
		resp.Character = cc.char
		resp.CurrentEnemy = cc.enemy
		resp.EventLog = logs
		return resp, nil
	}

	remainingHP := uint(enemyHP)
	updatedRun, err := s.repo.UpdateDailyRunEnemyHP(ctx, exec, cc.run.ID, &remainingHP, cc.run.TurnsTaken)
	if err != nil {
		return nil, err
	}

	resp.Run = updatedRun
	resp.Character = cc.char
	resp.CurrentEnemy = cc.enemy
	resp.EventLog = logs
	return resp, nil
}

func (s *Service) applyPlayerAction(
	ctx context.Context,
	exec db.DBTX,
	cc *combatContext,
	action ActionType,
	itemID *items.ID,
	enemyHP *int,
	turnRes *CombatTurnResult,
) (string, error) {
	if action == ActionUseItem {
		return s.useCombatConsumable(ctx, exec, cc, itemID, turnRes)
	}
	return resolvePlayerAttack(cc, action, enemyHP, turnRes)
}

func (s *Service) useCombatConsumable(
	ctx context.Context,
	exec db.DBTX,
	cc *combatContext,
	itemID *items.ID,
	turnRes *CombatTurnResult,
) (string, error) {
	if itemID == nil || *itemID == 0 {
		return "", inventory.ErrItemNotConsumable
	}

	removed, err := s.invRepo.RemoveItemFromCharacter(ctx, exec, inventory.RepositoryParam{
		CharacterID: cc.char.ID,
		OwnerID:     cc.char.OwnerID,
		ItemID:      *itemID,
		Quantity:    1,
	})
	if err != nil {
		return "", err
	}
	if removed.Item == nil || removed.Item.HealAmount == nil || *removed.Item.HealAmount == 0 {
		return "", inventory.ErrItemNotConsumable
	}

	healAmt := min(255, int(*removed.Item.HealAmount))
	beforeHP := int(cc.char.HpCurrent)
	healedChar, err := s.charRepo.HealCharacter(ctx, exec, cc.char.ID, uint8(healAmt))
	if err != nil {
		return "", err
	}

	cc.char = healedChar
	turnRes.HealedHP = int(healedChar.HpCurrent) - beforeHP
	metrics.CombatActionsTotal.WithLabelValues(string(ActionUseItem), "heal").Inc()
	return fmt.Sprintf("Used %s and recovered %d HP (%d/%d HP).", removed.Item.Name, turnRes.HealedHP, healedChar.HpCurrent, healedChar.HpMax), nil
}

func resolvePlayerAttack(cc *combatContext, action ActionType, enemyHP *int, turnRes *CombatTurnResult) (string, error) {
	roll, err := dice.RollD20()
	if err != nil {
		return "", err
	}

	atkMod := cc.primaryMod
	if action == ActionHeavyAttack {
		atkMod -= heavyAttackAccuracyPenalty
	}

	totalAtk := roll + atkMod
	isCrit := roll == 20
	isHit := roll != 1 && (isCrit || totalAtk >= int(cc.enemy.ArmorClass))

	turnRes.PlayerAttackRoll = roll
	turnRes.PlayerAttackTotal = totalAtk
	turnRes.PlayerHit = isHit
	turnRes.PlayerCritical = isCrit

	if !isHit {
		metrics.CombatActionsTotal.WithLabelValues(string(action), "miss").Inc()
		return fmt.Sprintf("Rolled %d (%d vs AC %d) and missed %s.", roll, totalAtk, cc.enemy.ArmorClass, cc.enemy.Name), nil
	}

	dmg, err := rollPlayerDamage(action, cc.weaponDmg, cc.primaryMod, isCrit)
	if err != nil {
		return "", err
	}

	*enemyHP = max(0, *enemyHP-dmg)
	turnRes.PlayerDamage = dmg

	if isCrit {
		metrics.CombatActionsTotal.WithLabelValues(string(action), "critical").Inc()
		return fmt.Sprintf("CRITICAL HIT (nat 20)! Dealt %d damage to %s.", dmg, cc.enemy.Name), nil
	}
	metrics.CombatActionsTotal.WithLabelValues(string(action), "hit").Inc()
	return fmt.Sprintf("Rolled %d (+%d = %d vs AC %d) and dealt %d damage to %s.", roll, atkMod, totalAtk, cc.enemy.ArmorClass, dmg, cc.enemy.Name), nil
}

func rollPlayerDamage(action ActionType, weaponDmg, primaryMod int, isCrit bool) (int, error) {
	baseDie, err := dice.RollN(8)
	if err != nil {
		return 0, err
	}
	dmg := baseDie + weaponDmg + primaryMod
	if action == ActionHeavyAttack {
		bonusDie, err := dice.RollN(6)
		if err != nil {
			return 0, err
		}
		dmg += bonusDie + 2
	}
	if isCrit {
		dmg *= 2
	}
	return max(1, dmg), nil
}

func (s *Service) handleEnemyDefeated(
	ctx context.Context,
	exec db.DBTX,
	cc *combatContext,
	resp *RunStateResponse,
	logs []string,
) (*RunStateResponse, error) {
	resp.TurnResult.EnemyDefeated = true
	cc.run.CurrentEnemyID = nil
	cc.run.CurrentEnemyHP = nil
	cc.run.Score += roomClearBaseScore + cc.enemy.ScoreReward
	logs = append(logs, fmt.Sprintf("%s was defeated! (+%d XP, +%d Score)", cc.enemy.Name, cc.enemy.XPReward, roomClearBaseScore+cc.enemy.ScoreReward))

	if shouldDropLoot(cc.currentNode.Type) {
		if loot, err := s.grantTierLoot(ctx, exec, cc.char, cc.run.Tier); err == nil && loot != nil {
			resp.LootItem = loot
			logs = append(logs, fmt.Sprintf("Looted %s!", loot.Name))
		}
	}

	isFinalBoss := cc.currentNode.Type == NodeBoss || cc.run.CurrentFloor >= cc.run.TotalFloors
	if isFinalBoss {
		victoryLog, err := s.completeRunWithVictory(ctx, exec, cc, resp)
		if err != nil {
			return nil, err
		}
		logs = append(logs, victoryLog)
	} else {
		cc.run.Status = InProgress
		updatedChar, err := s.charRepo.UpdateCharacterProgress(ctx, exec, cc.char.ID, cc.char.Level, cc.char.Xp+uint64(cc.enemy.XPReward), cc.char.HpMax, cc.char.HpCurrent)
		if err != nil {
			return nil, err
		}
		cc.char = updatedChar
	}

	updatedRun, err := s.repo.UpdateDailyRunState(ctx, exec, cc.run)
	if err != nil {
		return nil, err
	}

	resp.Run = updatedRun
	resp.Character = cc.char
	resp.EventLog = logs
	return resp, nil
}

func (s *Service) completeRunWithVictory(
	ctx context.Context,
	exec db.DBTX,
	cc *combatContext,
	resp *RunStateResponse,
) (string, error) {
	cc.run.Status = Victory
	cc.run.Score = computeFinalScore(cc.run.Score, int(cc.char.HpCurrent), int(cc.run.TurnsTaken), true)

	newLevel, newHpMax := leveledUpStats(cc.char)
	updatedChar, err := s.charRepo.UpdateCharacterProgress(ctx, exec, cc.char.ID, newLevel, cc.char.Xp+uint64(cc.enemy.XPReward), newHpMax, cc.char.HpCurrent)
	if err != nil {
		return "", err
	}
	cc.char = updatedChar

	scoreEntry, err := s.recordDailyScore(ctx, exec, cc.run, cc.char, cc.run.TotalFloors, cc.char.HpCurrent, true)
	if err != nil {
		return "", err
	}
	resp.FinalScore = scoreEntry
	metrics.DailyRunsCompletedTotal.WithLabelValues(strconv.Itoa(int(cc.run.Tier)), string(Victory)).Inc()

	return fmt.Sprintf("VICTORY! Completed Tier %d Daily Run! Final Score: %d. Character is now Level %d (%d/%d HP).", cc.run.Tier, cc.run.Score, cc.char.Level, cc.char.HpCurrent, cc.char.HpMax), nil
}

func (s *Service) applyEnemyCounterAttack(
	ctx context.Context,
	exec db.DBTX,
	cc *combatContext,
	enemyHP int,
	turnRes *CombatTurnResult,
	resp *RunStateResponse,
) (string, bool, error) {
	roll, err := dice.RollD20()
	if err != nil {
		return "", false, err
	}

	atkBonus := int(cc.run.Tier) + 1
	total := roll + atkBonus
	turnRes.EnemyAttackRoll = roll
	turnRes.EnemyAttackTotal = total

	if roll == 1 || (roll != 20 && total < cc.playerAC) {
		return fmt.Sprintf("%s swung (%d vs AC %d) and missed!", cc.enemy.Name, total, cc.playerAC), false, nil
	}

	dmg := int(cc.enemy.Damage)
	if roll == 20 {
		dmg = (dmg * 3) / 2
	}
	turnRes.EnemyHit = true
	turnRes.EnemyDamage = dmg

	nextHP := int(cc.char.HpCurrent) - dmg
	if nextHP <= 0 {
		msg, err := s.handlePlayerDefeated(ctx, exec, cc, enemyHP, dmg, turnRes, resp)
		return msg, true, err
	}

	updatedChar, err := s.charRepo.UpdateCharacterHP(ctx, exec, cc.char.ID, uint8(nextHP))
	if err != nil {
		return "", false, err
	}
	cc.char = updatedChar
	return fmt.Sprintf("%s hit you (%d vs AC %d) for %d damage (%d/%d HP left).", cc.enemy.Name, total, cc.playerAC, dmg, updatedChar.HpCurrent, updatedChar.HpMax), false, nil
}

func (s *Service) handlePlayerDefeated(
	ctx context.Context,
	exec db.DBTX,
	cc *combatContext,
	enemyHP int,
	dmg int,
	turnRes *CombatTurnResult,
	resp *RunStateResponse,
) (string, error) {
	turnRes.PlayerDefeated = true

	updatedChar, err := s.charRepo.UpdateCharacterHP(ctx, exec, cc.char.ID, 1)
	if err != nil {
		return "", err
	}
	cc.char = updatedChar

	remainingEnemyHP := uint(enemyHP)
	cc.run.Status = Defeated
	cc.run.Score = computeFinalScore(cc.run.Score, 0, int(cc.run.TurnsTaken), false)
	cc.run.CurrentEnemyHP = &remainingEnemyHP

	scoreEntry, err := s.recordDailyScore(ctx, exec, cc.run, cc.char, cc.run.CurrentFloor-1, 0, false)
	if err != nil {
		return "", err
	}
	resp.FinalScore = scoreEntry
	metrics.DailyRunsCompletedTotal.WithLabelValues(strconv.Itoa(int(cc.run.Tier)), string(Defeated)).Inc()

	return fmt.Sprintf("%s struck for %d damage! You were defeated on Floor %d/%d (character left exhausted at 1 HP).", cc.enemy.Name, dmg, cc.run.CurrentFloor, cc.run.TotalFloors), nil
}

func (s *Service) recordDailyScore(
	ctx context.Context,
	exec db.DBTX,
	run *DailyRun,
	char *characters.Character,
	floorsCleared uint8,
	hpRemaining uint8,
	completed bool,
) (*DailyScore, error) {
	return s.repo.UpsertDailyScore(ctx, exec, &DailyScore{
		RunDate:           run.RunDate,
		Tier:              run.Tier,
		UserID:            char.OwnerID,
		CharacterID:       char.ID,
		CharacterName:     char.Name,
		CharacterClass:    char.Class,
		CharacterLevel:    char.Level,
		CharacterPrestige: char.Prestige,
		Score:             run.Score,
		FloorsCleared:     floorsCleared,
		TotalFloors:       run.TotalFloors,
		TurnsTaken:        run.TurnsTaken,
		HpRemaining:       hpRemaining,
		Completed:         completed,
	})
}

func shouldDropLoot(nodeType NodeType) bool {
	if nodeType == NodeElite || nodeType == NodeMiniBoss || nodeType == NodeBoss {
		return true
	}
	roll, err := dice.RollN(100)
	return err == nil && roll <= normalDropChancePercent
}

func leveledUpStats(char *characters.Character) (uint8, uint8) {
	if char.Level >= maxCharacterLevel {
		return char.Level, char.HpMax
	}
	newHpMax := min(maxCharacterHP, int(char.HpMax)+int(char.HPGainOnLevelUp()))
	return char.Level + 1, uint8(newHpMax)
}

func equippedBonuses(equipped []inventory.InventoryItem) (weaponDamage int, totalDefense int) {
	for _, eq := range equipped {
		if eq.Item == nil {
			continue
		}
		if eq.Item.Damage != nil {
			weaponDamage += int(*eq.Item.Damage)
		}
		if eq.Item.Defense != nil {
			totalDefense += int(*eq.Item.Defense)
		}
	}
	return weaponDamage, totalDefense
}

func computeFinalScore(accumulated uint, hpRemaining int, turnsTaken int, completed bool) uint {
	score := int(accumulated) + (hpRemaining * 5) - (turnsTaken * 2)
	if completed {
		score += bossCompletionScoreBonus
	}
	return uint(max(0, score))
}
