package inventory

type InventoryItem struct {
	Item       *Item `db:"item" json:"item"`
	Quantity   uint8 `db:"quantity" json:"quantity"`
	IsEquipped bool  `db:"is_equipped" json:"is_equipped"`
}
