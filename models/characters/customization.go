package characters

import "fmt"

type Customization struct {
	ID    CharacterID    `db:"id" json:"-"`
	Hair  uint8 `db:"hair"  json:"hair"`
	Face  uint8 `db:"face"  json:"face"`
	Shirt uint8 `db:"shirt" json:"shirt"`
	Pants uint8 `db:"pants" json:"pants"`
	Shoes uint8 `db:"shoes" json:"shoes"`
}

func (c *Customization) String() string {
	return fmt.Sprintf("{ Hair: %v, Face: %v, Shirt: %v, Pants: %v, Shoes: %v }",
		c.Hair,
		c.Face,
		c.Shirt,
		c.Pants,
		c.Shoes,
	)
}
