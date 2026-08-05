package characters

import (
	"fmt"

	"github.com/google/uuid"
)

const formatString = "\nName: %v\nSpecies: %v\nBody Type: %v\nClass: %v\nLevel: %v\nXP: %v\nHP: %v/%v\n\n-STATS-\n%v\n\nCustomization: %v\n\n"

type Character struct {
	ID            CharacterID    `db:"id" json:"id" redis:"id"`
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

func (char *Character) Validate() bool {
	if char.Stats == nil || char.Customization == nil {
		return false
	}
	return char.BodyType.Validate() &&
		char.Species.Validate() &&
		char.Class.Validate() &&
		char.Stats.Validate() &&
		char.Customization.Validate()
}
