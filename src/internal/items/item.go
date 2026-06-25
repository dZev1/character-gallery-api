package items

type Item struct {
	ID          ItemID `db:"id" json:"id,omitempty" redis:"id"`
	Name        string `db:"name" json:"name" redis:"name"`
	Type        Type   `db:"type" json:"type" redis:"type"`
	Description string `db:"description" json:"description" redis:"description"`
	Equippable  bool   `db:"equippable" json:"equippable" redis:"equippable"`
	Rarity      uint8  `db:"rarity" json:"rarity" redis:"rarity"`

	Damage     *uint64 `db:"damage" json:"damage,omitempty" redis:"damage"`
	Defense    *uint64 `db:"defense" json:"defense,omitempty" redis:"defense"`
	HealAmount *uint64 `db:"heal_amount" json:"heal_amount,omitempty" redis:"heal_amount"`
	ManaCost   *uint64 `db:"mana_cost" json:"mana_cost,omitempty" redis:"mana_cost"`
	Duration   *uint64 `db:"duration" json:"duration,omitempty" redis:"duration"`
	Cooldown   *uint64 `db:"cooldown" json:"cooldown,omitempty" redis:"cooldown"`
	Capacity   *uint64 `db:"capacity" json:"capacity,omitempty" redis:"capacity"`
}

func (i *Item) Validate() bool {
	if len(i.Name) < 3 || len(i.Name) > 50 {
		return false
	}
	if len(i.Description) < 3 || len(i.Description) > 300 {
		return false
	}
	if i.Rarity < 1 || i.Rarity > 5 {
		return false
	}
	if !ValidateStats(i) && i.Equippable {
		return false
	}
	return true
}

func ValidateStats(i *Item) bool {
	return (i.Damage != nil && *i.Damage > 0) ||
		(i.Defense != nil && *i.Defense > 0) ||
		(i.HealAmount != nil && *i.HealAmount > 0) ||
		(i.ManaCost != nil && *i.ManaCost > 0) ||
		(i.Duration != nil && *i.Duration > 0) ||
		(i.Cooldown != nil && *i.Cooldown > 0) ||
		(i.Capacity != nil && *i.Capacity > 0)
}
