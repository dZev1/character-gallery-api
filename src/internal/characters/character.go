package characters

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("character not found")

type ID uint64

func (id ID) String() string {
	return fmt.Sprintf("Nº%d", id)
}

const formatString = "\nName: %v\nSpecies: %v\nBody Type: %v\nClass: %v\nLevel: %v\nXP: %v\nHP: %v/%v\n\n-STATS-\n%v\n\nCustomization: %v\n\n"

type Character struct {
	ID            ID             `db:"id" json:"id" redis:"id"`
	OwnerID       uuid.UUID      `db:"owner_id" json:"-"`
	Name          string         `db:"name" json:"name" redis:"name"`
	BodyType      BodyType       `db:"body_type" json:"body_type" redis:"body_type"`
	Species       Species        `db:"species" json:"species" redis:"species"`
	Class         Class          `db:"class" json:"class" redis:"class"`
	Level         uint8          `json:"level"`
	Xp            uint64         `json:"xp"`
	HpMax         uint8          `json:"hp_max"`
	HpCurrent     uint8          `json:"hp_current"`
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
