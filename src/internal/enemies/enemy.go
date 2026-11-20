package enemies

type ID uint64

type Enemy struct {
	ID          ID     `db:"id" json:"id" redis:"id"`
	Name        string `db:"name" json:"name" redis:"name"`
	Tier        uint8  `db:"tier" json:"tier" redis:"tier"`
	Description string `db:"description" json:"description" redis:"description"`
	MaxHP       uint8  `db:"max_hp" json:"max_hp" redis:"max_hp"`
}
