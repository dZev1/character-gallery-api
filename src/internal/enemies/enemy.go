package enemies

import (
	"errors"
	"fmt"
	"strings"
)

var ErrNotFound = errors.New("enemy not found")

type ID uint64

func (id ID) String() string {
	return fmt.Sprintf("Nº%d", id)
}

const formatString = "ENEMY {\nID: %d\nName: %s\nType: %s\nTier: %d\nDescription: %s\nMax HP: %d\nAC: %d\nDamage: %d\nXP: %d\nSCORE: %d\n}\n"

type Enemy struct {
	ID          ID     `db:"id" json:"id" redis:"id"`
	Name        string `db:"name" json:"name" redis:"name"`
	Type        Type   `db:"type" json:"type" redis:"type"`
	Tier        uint8  `db:"tier" json:"tier" redis:"tier"`
	Description string `db:"description" json:"description" redis:"description"`
	HpMax       uint   `db:"hp_max" json:"hp_max" redis:"hp_max"`
	ArmorClass  uint8  `db:"ac" json:"ac" redis:"ac"`
	Damage      uint   `db:"damage" json:"damage" redis:"damage"`
	XPReward    uint   `db:"xp_reward" json:"xp_reward" redis:"xp_reward"`
	ScoreReward uint   `db:"score_reward" json:"score_reward" redis:"score_reward"`
	ImageURL    string `db:"image_url" json:"image_url,omitempty" redis:"image_url"`
}

func (e *Enemy) String() string {
	return fmt.Sprintf(
		formatString,
		e.ID,
		e.Name,
		e.Type,
		e.Tier,
		e.Description,
		e.HpMax,
		e.ArmorClass,
		e.Damage,
		e.XPReward,
		e.ScoreReward,
	)
}

func (e *Enemy) IsValid() bool {
	validName := len(strings.TrimSpace(e.Name)) > 0
	validType := e.Type.IsValid()
	validTier := e.Tier >= 1 && e.Tier <= 5
	validDescription := len(strings.TrimSpace(e.Description)) > 0 && len(e.Description) <= 255
	validHpMax := e.HpMax > 0 && e.HpMax <= 1000
	validArmorClass := e.ArmorClass >= 5 && e.ArmorClass <= 30
	validDamage := e.Damage > 0
	validXPReward := e.XPReward > 0
	validScoreReward := e.ScoreReward > 0
	validImageURL := e.ImageURL == "" || (len(e.ImageURL) <= 500 &&
		(strings.HasPrefix(e.ImageURL, "/uploads/") ||
			strings.HasPrefix(e.ImageURL, "http://") ||
			strings.HasPrefix(e.ImageURL, "https://")))

	return validName &&
		validType &&
		validTier &&
		validDescription &&
		validHpMax &&
		validArmorClass &&
		validDamage &&
		validXPReward &&
		validScoreReward &&
		validImageURL
}
