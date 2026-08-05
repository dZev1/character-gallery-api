package handlers

import (
	"dZev1/character-gallery/internal/auth"
	"dZev1/character-gallery/internal/items"
	"encoding/json"
	"net/http"
)

func (g *Gallery) HandleCreateItem(w http.ResponseWriter, r *http.Request) {
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

	item := &items.Item{}
	limitBody(w, r)
	err := json.NewDecoder(r.Body).Decode(item)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Error parsing request body")
		return
	}

	if !item.Validate() {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid item")
		return
	}

	item, err = g.itemsService.CreateItem(ctx, item)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error creating item")
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func (g *Gallery) HandleGetAllItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page, limit := paginationParams(r)

	itms, total, err := g.itemsService.GetAll(ctx, page, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error getting all items")
		return
	}

	writeJSON(w, http.StatusOK, paginatedResponse{
		Data: itms,
		Pagination: pagination{
			Page:    page,
			Limit:   limit,
			Total:   total,
			HasNext: total > uint64(page*limit),
		}})
}

func (g *Gallery) HandleGetItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseID(r, "itemId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid item id")
		return
	}

	item, err := g.itemsService.GetByID(ctx, items.ItemID(id))
	if err != nil {
		if isItemNotFound(err) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Item not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error getting item")
		return
	}

	writeJSON(w, http.StatusOK, item)
}
