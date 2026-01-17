package inventory

type InventoryItem struct {
	Item       *Item  `db:"item" json:"item"`
	Quantity   uint64 `db:"qty" json:"qty"`
	IsEquipped bool   `db:"is_equipped" json:"is_equipped"`
}
