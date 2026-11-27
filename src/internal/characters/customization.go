package characters

import "fmt"

type Customization struct {
	ID          ID    `db:"id" json:"-"`
	Hair        uint8 `db:"hair" json:"hair" redis:"hair"`
	Beard       uint8 `db:"beard" json:"beard" redis:"beard"`
	SkinTone    uint8 `db:"skin_tone" json:"skin_tone" redis:"skin_tone"`
	HairColor   uint8 `db:"hair_color" json:"hair_color" redis:"hair_color"`
	Cloak       uint8 `db:"cloak" json:"cloak" redis:"cloak"`
	AccentColor uint8 `db:"accent_color" json:"accent_color" redis:"accent_color"`
}

func (c *Customization) String() string {
	return fmt.Sprintf(
		"{ Hair: %v, Beard: %v, SkinTone: %v, HairColor: %v, Cloak: %v, AccentColor: %v }",
		c.Hair,
		c.Beard,
		c.SkinTone,
		c.HairColor,
		c.Cloak,
		c.AccentColor,
	)
}

func (c *Customization) IsValid() bool {
	return c.Hair <= 20 &&
		c.Beard <= 15 &&
		c.SkinTone <= 15 &&
		c.HairColor <= 15 &&
		c.Cloak <= 15 &&
		c.AccentColor <= 15
}
