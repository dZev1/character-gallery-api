package inventory

type Item struct {
	ID          ItemID `db:"id" json:"id,omitempty"`
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

func uint64Ptr(i uint64) *uint64 {
	return &i
}

func strPtr(s string) *string {
	return &s
}
