package characters

type Class string

const (
	BARBARIAN Class = "barbarian"
	BARD      Class = "bard"
	CLERIC    Class = "cleric"
	DRUID     Class = "druid"
	FIGHTER   Class = "fighter"
	MONK      Class = "monk"
	PALADIN   Class = "paladin"
	RANGER    Class = "ranger"
	ROGUE     Class = "rogue"
	SORCERER  Class = "sorcerer"
	WARLOCK   Class = "warlock"
	WIZARD    Class = "wizard"
)

func (c Class) String() string {
	return string(c)
}
