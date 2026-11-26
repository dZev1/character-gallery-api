package items

import (
	"errors"
	"fmt"
	"strings"
)

var ErrNotFound = errors.New("item not found")

type ID uint64

func (id ID) String() string {
	return fmt.Sprintf("Nº%d", id)
}

type Item struct {
	ID          ID     `db:"id" json:"id,omitempty" redis:"id"`
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
	ImageURL   string  `db:"image_url" json:"image_url,omitempty" redis:"image_url"`
}

func (i *Item) IsValid() bool {
	if len(i.Name) < 3 || len(i.Name) > 50 {
		return false
	}
	if !i.Type.IsValid() {
		return false
	}
	if len(i.Description) < 3 || len(i.Description) > 300 {
		return false
	}
	if i.Rarity < 1 || i.Rarity > 5 {
		return false
	}
	if !HasValidStats(i) && i.Equippable {
		return false
	}
	if i.ImageURL != "" {
		if len(i.ImageURL) > 500 {
			return false
		}
		if !strings.HasPrefix(i.ImageURL, "/uploads/") &&
			!strings.HasPrefix(i.ImageURL, "http://") &&
			!strings.HasPrefix(i.ImageURL, "https://") {
			return false
		}
	}
	return true
}

func HasValidStats(i *Item) bool {
	return (i.Damage != nil && *i.Damage > 0) ||
		(i.Defense != nil && *i.Defense > 0) ||
		(i.HealAmount != nil && *i.HealAmount > 0) ||
		(i.ManaCost != nil && *i.ManaCost > 0) ||
		(i.Duration != nil && *i.Duration > 0) ||
		(i.Cooldown != nil && *i.Cooldown > 0) ||
		(i.Capacity != nil && *i.Capacity > 0)
}
