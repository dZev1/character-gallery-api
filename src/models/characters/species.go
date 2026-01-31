package characters

type Species string

const (
	AASIMAR    Species = "aasimar"
	DRAGONBORN Species = "dragonborn"
	DWARF      Species = "dwarf"
	ELF        Species = "elf"
	GNOME      Species = "gnome"
	GOLIATH    Species = "goliath"
	HALFLING   Species = "halfling"
	HUMAN      Species = "human"
	ORC        Species = "orc"
	TIEFLING   Species = "tiefling"
)

func (s Species) String() string {
	return string(s)
}
