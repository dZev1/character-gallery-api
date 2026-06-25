package characters

import "fmt"

type Customization struct {
	ID    CharacterID `db:"id" json:"-"`
	Hair  uint8       `db:"hair"  json:"hair" redis:"hair"`
	Face  uint8       `db:"face"  json:"face" redis:"face"`
	Shirt uint8       `db:"shirt" json:"shirt" redis:"shirt"`
	Pants uint8       `db:"pants" json:"pants" redis:"pants"`
	Shoes uint8       `db:"shoes" json:"shoes" redis:"shoes"`
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

func (c *Customization) Validate() bool {
	for _, customization := range []uint8{c.Hair, c.Face, c.Shirt, c.Pants, c.Shoes} {
		if customization > 30 {
			return false
		}
	}
	return true
}
