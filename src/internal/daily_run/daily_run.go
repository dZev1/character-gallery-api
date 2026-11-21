package daily_run

import (
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/enemies"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("daily run not found")

const formatString = "RUN {\nID: %d\nRunDate: %s\nTier: %d\nCharacterID: %d\nFloor: %d/%d\nNode: %s\nTurns: %d\nScore: %d\nStatus: %s\n}\n"

var floorCountPerTier = []uint8{10, 15, 20, 25, 30}

type Status string

const (
	InProgress Status = "in_progress"
	InCombat   Status = "in_combat"
	Victory    Status = "victory"
	Defeated   Status = "defeat"
)

type ID uint64

func (id ID) String() string {
	return fmt.Sprintf("Nº%d", id)
}

type DailyRun struct {
	ID             ID            `db:"id" json:"id" redis:"id"`
	RunDate        time.Time     `db:"run_date" json:"run_date" redis:"run_date"`
	Tier           uint8         `db:"tier" json:"tier" redis:"tier"`
	UserID         uuid.UUID     `db:"user_id" json:"user_id" redis:"user_id"`
	CharacterID    characters.ID `db:"character_id" json:"character_id" redis:"character_id"`
	TotalFloors    uint8         `db:"total_floors" json:"total_floors" redis:"total_floors"`
	CurrentFloor   uint8         `db:"current_floor" json:"current_floor" redis:"current_floor"`
	CurrentNodeID  string        `db:"current_node_id" json:"current_node_id" redis:"current_node_id"`
	VisitedNodes   []string      `db:"visited_nodes" json:"visited_nodes" redis:"visited_nodes"`
	TurnsTaken     uint          `db:"turns_taken" json:"turns_taken" redis:"turns_taken"`
	Score          uint          `db:"score" json:"score" redis:"score"`
	Status         Status        `db:"status" json:"status" redis:"status"`
	CurrentEnemyID *enemies.ID   `db:"current_enemy_id" json:"current_enemy_id,omitempty" redis:"current_enemy_id"`
	CurrentEnemyHP *uint         `db:"current_enemy_hp" json:"current_enemy_hp,omitempty" redis:"current_enemy_hp"`
	CreatedAt      time.Time     `db:"created_at" json:"created_at" redis:"created_at"`
	UpdatedAt      time.Time     `db:"updated_at" json:"updated_at" redis:"updated_at"`
}

func (d *DailyRun) String() string {
	return fmt.Sprintf(
		formatString,
		d.ID,
		d.RunDate.Format("2006-01-02"),
		d.Tier,
		d.CharacterID,
		d.CurrentFloor,
		d.TotalFloors,
		d.CurrentNodeID,
		d.TurnsTaken,
		d.Score,
		d.Status,
	)
}

func (d *DailyRun) IsValid() bool {
	if d.Tier < 1 || d.Tier > 5 {
		return false
	}
	validTotalFloors := floorCountPerTier[d.Tier-1] == d.TotalFloors
	validCurrentFloor := d.CurrentFloor > 0 && d.CurrentFloor <= d.TotalFloors
	validCurrentNode := len(strings.TrimSpace(d.CurrentNodeID)) > 0
	validVisitedNodes := len(d.VisitedNodes) > 0
	validStatus := d.Status.IsValid()

	if d.Status == InCombat && (d.CurrentEnemyID == nil || *d.CurrentEnemyID == 0 || d.CurrentEnemyHP == nil) {
		return false
	}

	return validTotalFloors &&
		validCurrentFloor &&
		validCurrentNode &&
		validVisitedNodes &&
		validStatus
}

func (s Status) IsValid() bool {
	switch s {
	case InProgress, InCombat, Victory, Defeated:
		return true
	}
	return false
}

type DailyScore struct {
	ID                uint64           `db:"id" json:"id"`
	RunDate           time.Time        `db:"run_date" json:"run_date"`
	Tier              uint8            `db:"tier" json:"tier"`
	UserID            uuid.UUID        `db:"user_id" json:"user_id"`
	Username          string           `db:"username" json:"username,omitempty"`
	CharacterID       characters.ID    `db:"character_id" json:"character_id"`
	CharacterName     string           `db:"character_name" json:"character_name"`
	CharacterClass    characters.Class `db:"character_class" json:"character_class"`
	CharacterLevel    uint8            `db:"character_level" json:"character_level"`
	CharacterPrestige uint             `db:"character_prestige" json:"character_prestige"`
	Score             uint             `db:"score" json:"score"`
	FloorsCleared     uint8            `db:"floors_cleared" json:"floors_cleared"`
	TotalFloors       uint8            `db:"total_floors" json:"total_floors"`
	TurnsTaken        uint             `db:"turns_taken" json:"turns_taken"`
	HpRemaining       uint8            `db:"hp_remaining" json:"hp_remaining"`
	Completed         bool             `db:"completed" json:"completed"`
	CreatedAt         time.Time        `db:"created_at" json:"created_at"`
}
