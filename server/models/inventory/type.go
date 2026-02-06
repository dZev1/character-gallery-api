package inventory

type Type string

const (
	// Equipment
	Armor           Type = "armor"
	Ring            Type = "ring"
	Weapon          Type = "weapon"
	Shield          Type = "shield"
	Tool            Type = "tool"
	AdventuringGear Type = "adventuring_gear"

	// Magic Equipment
	Rod    Type = "rod"
	Staff  Type = "staff"
	Wand   Type = "wand"
	Scroll Type = "scroll"

	// Consumables
	Potion       Type = "potion"
	Ammo         Type = "ammo"
	Consumable   Type = "consumable"
	WondrousItem Type = "wondrous_item"
)

func (t Type) Validate() bool {
	switch t {
	case Armor, Ring, Weapon, Shield, Tool, AdventuringGear, Rod, Staff, Wand, Scroll, Potion, Ammo, Consumable, WondrousItem:
		return true
	}
	return false
}