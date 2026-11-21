package handlers

import (
	"dZev1/character-gallery/internal/auth"
	"dZev1/character-gallery/internal/enemies"
	"encoding/json"
	"errors"
	"net/http"
)

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
	ctx := r.Context()

	id, err := parseID(r, "enemyId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid enemy id")
		return
	}

	enemy, err := g.enemyService.GetByID(ctx, enemies.ID(id))
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
	ctx := r.Context()

	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}
	if claims.Role != "admin" {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Admin role required")
		return
	}

	enemy := &enemies.Enemy{}
	limitBody(w, r)
	if err := json.NewDecoder(r.Body).Decode(enemy); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Error parsing request body")
		return
	}
	if !enemy.IsValid() {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid enemy")
		return
	}

	created, err := g.enemyService.Create(ctx, enemy)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error creating enemy")
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (g *Gallery) HandleUpdateEnemy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}
	if claims.Role != "admin" {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Admin role required")
		return
	}

	id, err := parseID(r, "enemyId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid enemy id")
		return
	}

	enemy := &enemies.Enemy{}
	limitBody(w, r)
	if err := json.NewDecoder(r.Body).Decode(enemy); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Error parsing request body")
		return
	}
	enemy.ID = enemies.ID(id)
	if !enemy.IsValid() {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid enemy")
		return
	}

	updated, err := g.enemyService.Update(ctx, enemy)
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
	ctx := r.Context()

	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}
	if claims.Role != "admin" {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Admin role required")
		return
	}

	id, err := parseID(r, "enemyId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid enemy id")
		return
	}

	if err := g.enemyService.Delete(ctx, enemies.ID(id)); err != nil {
		if errors.Is(err, enemies.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Enemy not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error deleting enemy")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"result": "success"})
}
