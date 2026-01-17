package inventory

type Item struct {
	ID          ItemID `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	Type        Type   `db:"type" json:"type"`
	Description string `db:"description" json:"description"`
	Equippable  bool   `db:"equippable" json:"equippable"`
	Rarity      uint8  `db:"rarity" json:"rarity"`

	Damage     *uint64 `db:"damage" json:"damage,omitempty"`
	Defense    *uint64 `db:"defense" json:"defense,omitempty"`
	HealAmount *uint64 `db:"heal_amount" json:"heal_amount,omitempty"`
	ManaCost   *uint64 `db:"mana_cost" json:"mana_cost,omitempty"`
	Duration   *string `db:"duration" json:"duration,omitempty"`
}

var Items = []Item{
	{
		ID:          1,
		Name:        "Master Sword",
		Type:        Weapon,
		Description: "A legendary sword with immense power.",
		Equippable:  true,
		Rarity:      5,
		Damage:      uint64Ptr(100),
	},
	{
		ID:          2,
		Name:        "Carl's Doomsday Scenario",
		Type:        Explosive,
		Description: "Created  by  a  man  who  murders  babies  and  steals  rare collectibles from his elders, this device is powerful enough to level an entire city and all the suburbs around it. It is created by combining  a  massively  overloaded  soul crystal and  a  Sheol Glass Reaper Case.",
		Equippable:  false,
		Rarity:      5,
		Damage:      uint64Ptr(1000),
	},
	{
		ID:          3,
		Name:        "Healing Potion",
		Type:        Potion,
		Description: "A potion that restores health.",
		Equippable:  false,
		Rarity:      1,
		HealAmount:  uint64Ptr(60),
	},
	{
		ID:          4,
		Name:        "Steel Armor",
		Type:        Armor,
		Description: "Sturdy armor made of steel.",
		Equippable:  true,
		Rarity:      3,
		Defense:     uint64Ptr(40),
	},
	{
		ID:          5,
		Name:        "Magic Missile",
		Type:        Spell,
		Description: "A spell that launches a magic missile at the target.",
		Equippable:  false,
		Rarity:      2,
		ManaCost:    uint64Ptr(30),
	},
	{
		ID:          6,
		Name:        "Invisibility Cloak",
		Type:        Misc,
		Description: "A cloak that grants invisibility to the wearer.",
		Equippable:  true,
		Rarity:      4,
		Duration:    strPtr("5 minutes"),
	},
	{
		ID:          7,
		Name:        "Paris's Bow",
		Type:        Weapon,
		Description: "A finely crafted bow used by the legendary archer Paris.",
		Equippable:  true,
		Rarity:      4,
		Damage:      uint64Ptr(80),
	},
	{
		ID:          8,
		Name:        "Favor and Protection Ring",
		Type:        Misc,
		Description: "A ring symbolizing the favor and protection of the goddess Fina, known in legend to possess 'fateful beauty'. This ring boosts its wearer's HP, stamina, and max equipment load, but breaks if ever removed.",
		Equippable:  true,
		Rarity:      5,
		Defense:     uint64Ptr(10),
		ManaCost:    uint64Ptr(20),
	},
}

func uint64Ptr(i uint64) *uint64 {
	return &i
}

func strPtr(s string) *string {
	return &s
}
