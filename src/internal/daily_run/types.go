package daily_run

import (
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/enemies"
	"dZev1/character-gallery/internal/items"
	"errors"
)

var (
	ErrInvalidTier         = errors.New("tier must be between 1 and 5")
	ErrCharacterDead       = errors.New("character has 0 HP and must heal before entering a run")
	ErrRunAlreadyFinished  = errors.New("daily run for this character and tier is already completed today")
	ErrRunNotInCombat      = errors.New("daily run is not currently in combat")
	ErrRunInCombat         = errors.New("cannot move to another room while in combat")
	ErrInvalidMove         = errors.New("target node is not reachable from current room")
	ErrInvalidAction       = errors.New("invalid combat action: must be attack, heavy_attack, or use_item")
	ErrNoEnemiesConfigured = errors.New("no enemies configured for this tier")
)

type ActionType string

const (
	ActionAttack      ActionType = "attack"
	ActionHeavyAttack ActionType = "heavy_attack"
	ActionUseItem     ActionType = "use_item"
)

func (a ActionType) IsValid() bool {
	switch a {
	case ActionAttack, ActionHeavyAttack, ActionUseItem:
		return true
	}
	return false
}

type DailyView struct {
	Dungeon    *DungeonGraph `json:"dungeon"`
	ActiveRuns []DailyRun    `json:"active_runs"`
	UserScores []DailyScore  `json:"user_scores"`
}

type RunStateResponse struct {
	Run          *DailyRun             `json:"run"`
	Character    *characters.Character `json:"character"`
	CurrentNode  GraphNode             `json:"current_node"`
	CurrentEnemy *enemies.Enemy        `json:"current_enemy,omitempty"`
	NextNodeIDs  []string              `json:"next_node_ids"`
	EventLog     []string              `json:"event_log,omitempty"`
	LootItem     *items.Item           `json:"loot_item,omitempty"`
	CheckRoll    *SkillCheckResult     `json:"check_roll,omitempty"`
	TurnResult   *CombatTurnResult     `json:"turn_result,omitempty"`
	FinalScore   *DailyScore           `json:"final_score,omitempty"`
}

type SkillCheckResult struct {
	Stat     string `json:"stat"`
	DC       int    `json:"dc"`
	Roll     int    `json:"roll"`
	Modifier int    `json:"modifier"`
	Total    int    `json:"total"`
	Success  bool   `json:"success"`
}

type CombatTurnResult struct {
	Action            ActionType `json:"action"`
	PlayerAttackRoll  int        `json:"player_attack_roll,omitempty"`
	PlayerAttackTotal int        `json:"player_attack_total,omitempty"`
	PlayerHit         bool       `json:"player_hit,omitempty"`
	PlayerCritical    bool       `json:"player_critical,omitempty"`
	PlayerDamage      int        `json:"player_damage,omitempty"`
	HealedHP          int        `json:"healed_hp,omitempty"`
	EnemyDefeated     bool       `json:"enemy_defeated"`
	EnemyAttackRoll   int        `json:"enemy_attack_roll,omitempty"`
	EnemyAttackTotal  int        `json:"enemy_attack_total,omitempty"`
	EnemyHit          bool       `json:"enemy_hit,omitempty"`
	EnemyDamage       int        `json:"enemy_damage,omitempty"`
	PlayerDefeated    bool       `json:"player_defeated"`
}
