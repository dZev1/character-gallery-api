package handlers

import (
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/inventory"
	"dZev1/character-gallery/internal/items"
)

type Gallery struct {
	characterService *characters.Service
	itemsService     *items.Service
	inventoryService *inventory.Service
}

func NewGallery(chS *characters.Service, itS *items.Service, ivS *inventory.Service) *Gallery {
	return &Gallery{
		characterService: chS,
		itemsService:     itS,
		inventoryService: ivS,
	}
}
