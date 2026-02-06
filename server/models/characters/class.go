package characters

type Class string

const (
	Barbarian Class = "barbarian"
	Bard      Class = "bard"
	Cleric    Class = "cleric"
	Druid     Class = "druid"
	Fighter   Class = "fighter"
	Monk      Class = "monk"
	Paladin   Class = "paladin"
	Ranger    Class = "ranger"
	Rogue     Class = "rogue"
	Sorcerer  Class = "sorcerer"
	Warlock   Class = "warlock"
	Wizard    Class = "wizard"
)

func (c Class) String() string {
	return string(c)
}

func (c Class) Validate() bool {
	switch c {
	case Barbarian, Bard, Cleric, Druid, Fighter, Monk, Paladin, Ranger, Rogue, Sorcerer, Warlock, Wizard:
		return true
	}
	return false
}