package handlers

import (
	"dZev1/character-gallery/internal/enemies"
	"errors"
	"fmt"
	"net/http"
)

const maxBulkEnemies = 500

func (g *Gallery) HandleGetAllEnemies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if tier := parseQueryInt(r, "tier", 0); tier >= 1 && tier <= 5 {
		list, err := g.enemyService.GetByTier(ctx, uint8(tier))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error getting enemies by tier")
			return
		}
		writeJSON(w, http.StatusOK, list)
		return
	}

	page, limit := paginationParams(r)
	list, total, err := g.enemyService.GetAll(ctx, page, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error getting enemies")
		return
	}

	writeJSON(w, http.StatusOK, paginatedResponse{
		Data: list,
		Pagination: pagination{
			Page:    page,
			Limit:   limit,
			Total:   total,
			HasNext: total > uint64(page*limit),
		},
	})
}

func (g *Gallery) HandleGetEnemy(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "enemyId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid enemy id")
		return
	}

	enemy, err := g.enemyService.GetByID(r.Context(), enemies.ID(id))
	if err != nil {
		if errors.Is(err, enemies.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Enemy not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error getting enemy")
		return
	}
	writeJSON(w, http.StatusOK, enemy)
}

func (g *Gallery) HandleCreateEnemy(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	enemy := &enemies.Enemy{}
	if !decodeBody(w, r, enemy) {
		return
	}
	if !enemy.IsValid() {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid enemy")
		return
	}

	created, err := g.enemyService.Create(r.Context(), enemy)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error creating enemy")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (g *Gallery) HandleBulkCreateEnemies(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	var list []enemies.Enemy
	if !decodeBody(w, r, &list, "Error parsing request body (expected JSON array of enemies)") {
		return
	}
	if len(list) == 0 {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Enemy list cannot be empty")
		return
	}
	if len(list) > maxBulkEnemies {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", fmt.Sprintf("Maximum %d enemies per import", maxBulkEnemies))
		return
	}

	for i := range list {
		if !list[i].IsValid() {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", fmt.Sprintf("Invalid enemy at index %d (%q)", i, list[i].Name))
			return
		}
	}

	result, err := g.enemyService.BulkCreate(r.Context(), list)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error importing enemies")
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (g *Gallery) HandleUpdateEnemy(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	id, err := parseID(r, "enemyId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid enemy id")
		return
	}

	enemy := &enemies.Enemy{}
	if !decodeBody(w, r, enemy) {
		return
	}
	enemy.ID = enemies.ID(id)
	if !enemy.IsValid() {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid enemy")
		return
	}

	updated, err := g.enemyService.Update(r.Context(), enemy)
	if err != nil {
		if errors.Is(err, enemies.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Enemy not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error updating enemy")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (g *Gallery) HandleDeleteEnemy(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	id, err := parseID(r, "enemyId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid enemy id")
		return
	}

	if err := g.enemyService.Delete(r.Context(), enemies.ID(id)); err != nil {
		if errors.Is(err, enemies.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Enemy not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error deleting enemy")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"result": "success"})
}
