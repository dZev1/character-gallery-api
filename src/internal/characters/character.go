package characters

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const MaxCharactersPerOwner = 6

var ErrNotFound = errors.New("character not found")
var ErrPrestigeRequirement = errors.New("character must be at least level 10 to prestige")
var ErrCharacterLimitReached = errors.New("maximum number of characters per user reached (6)")

type ID uint64

func (id ID) String() string {
	return fmt.Sprintf("Nº%d", id)
}

const formatString = "\nName: %v\nSpecies: %v\nBody Type: %v\nClass: %v\nLevel: %v\nXP: %v\nHP: %v/%v\nPrestige: %v\n\n-STATS-\n%v\n\nCustomization: %v\n\n"

type Character struct {
	ID            ID             `db:"id" json:"id" redis:"id"`
	OwnerID       uuid.UUID      `db:"owner_id" json:"owner_id,omitempty" redis:"owner_id"`
	OwnerUsername string         `db:"owner_username" json:"owner_username,omitempty" redis:"owner_username"`
	Name          string         `db:"name" json:"name" redis:"name"`
	BodyType      BodyType       `db:"body_type" json:"body_type" redis:"body_type"`
	Species       Species        `db:"species" json:"species" redis:"species"`
	Class         Class          `db:"class" json:"class" redis:"class"`
	Level         uint8          `db:"level" json:"level" redis:"level"`
	Xp            uint64         `db:"xp" json:"xp" redis:"xp"`
	HpMax         uint8          `db:"hp_max" json:"hp_max" redis:"hp_max"`
	HpCurrent     uint8          `db:"hp_current" json:"hp_current" redis:"hp_current"`
	Prestige      uint           `db:"prestige" json:"prestige" redis:"prestige"`
	Stats         *Stats         `json:"stats" redis:"stats"`
	Customization *Customization `json:"customization" redis:"customization"`
}

func (char *Character) String() string {
	return fmt.Sprintf(formatString,
		char.Name,
		char.Species,
		char.BodyType,
		char.Class,
		char.Level,
		char.Xp,
		char.HpMax,
		char.HpCurrent,
		char.Prestige,
		char.Stats,
		char.Customization,
	)
}

func (char *Character) IsValid() bool {
	if strings.TrimSpace(char.Name) == "" {
		return false
	}
	if char.Stats == nil || char.Customization == nil {
		return false
	}
	if char.Level == 0 {
		char.Level = 1
	}
	if char.HpMax == 0 {
		char.HpMax = 1
		if char.HpCurrent == 0 {
			char.HpCurrent = 1
		}
	}
	return char.BodyType.IsValid() &&
		char.Species.IsValid() &&
		char.Class.IsValid() &&
		char.Stats.IsValid() &&
		char.Customization.IsValid()
}
