package characters

type Species string

const (
	Aasimar    Species = "aasimar"
	Dragonborn Species = "dragonborn"
	Dwarf      Species = "dwarf"
	Elf        Species = "elf"
	Gnome      Species = "gnome"
	Goliath    Species = "goliath"
	Halfling   Species = "halfling"
	Human      Species = "human"
	Orc        Species = "orc"
	Tiefling   Species = "tiefling"
)

func (s Species) String() string {
	return string(s)
}

func (s Species) Validate() bool {
	switch s {
	case Aasimar, Dragonborn, Dwarf, Elf, Gnome, Goliath, Halfling, Human, Orc, Tiefling:
		return true
	}
	return false
}