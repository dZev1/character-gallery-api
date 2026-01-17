package characters

import "fmt"

const formatString = "\nName: %v\nSpecies: %v\nBody Type: %v\nClass: %v\n\n-STATS-\n%v\n\nCustomization: %v\n\n"

type Character struct {
	ID            CharacterID             `db:"id" json:"id"`
	Name          string         `db:"name" json:"name"`
	BodyType      BodyType       `db:"body_type" json:"body_type"`
	Species       Species        `db:"species" json:"species"`
	Class         Class          `db:"class" json:"class"`
	Stats         *Stats         `json:"stats"`
	Customization *Customization `json:"customization"`
}

func (char *Character) String() string {
	return fmt.Sprintf(formatString,
		char.Name,
		char.Species,
		char.BodyType,
		char.Class,
		char.Stats,
		char.Customization,
	)
}
