package inventory

type Type string

const (
	// Equipment
	Weapon    Type = "weapon"
	Sorcery   Type = "sorcery"
	Armor     Type = "armor"
	Accessory Type = "accessory"

	// Consumables
	Potion Type = "potion"
	Spell  Type = "spell"
	Food   Type = "food"

	// Other
	Explosive Type = "explosive"
	Ammo      Type = "ammo"
	Treasure  Type = "treasure"
	Misc      Type = "misc"
)
