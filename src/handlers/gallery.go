package handlers

import (
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/daily_run"
	"dZev1/character-gallery/internal/enemies"
	"dZev1/character-gallery/internal/inventory"
	"dZev1/character-gallery/internal/items"
)

type Gallery struct {
	characterService *characters.Service
	itemsService     *items.Service
	inventoryService *inventory.Service
	enemyService     *enemies.Service
	dailyService     *daily_run.Service
}

func NewGallery(chS *characters.Service, itS *items.Service, ivS *inventory.Service) *Gallery {
	return &Gallery{
		characterService: chS,
		itemsService:     itS,
		inventoryService: ivS,
	}
}

func (g *Gallery) WithGameServices(enemySvc *enemies.Service, dailySvc *daily_run.Service) *Gallery {
	g.enemyService = enemySvc
	g.dailyService = dailySvc
	return g
}
