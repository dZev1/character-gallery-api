package inventory

import (
	"dZev1/character-gallery/internal/items"
)

type InventoryItem struct {
	Item       *items.Item `db:"items" json:"items" redis:"items"`
	Quantity   uint8       `db:"quantity" json:"quantity" redis:"quantity"`
	IsEquipped bool        `db:"is_equipped" json:"is_equipped" redis:"is_equipped"`
}
