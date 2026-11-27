package handlers

import (
	"dZev1/character-gallery/internal/items"
	"fmt"
	"net/http"
)

const maxBulkItems = 500

func (g *Gallery) HandleCreateItem(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	item := &items.Item{}
	if !decodeBody(w, r, item) {
		return
	}
	if !item.IsValid() {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid item")
		return
	}

	created, err := g.itemsService.CreateItem(r.Context(), item)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error creating item")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (g *Gallery) HandleBulkCreateItems(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	var list []items.Item
	if !decodeBody(w, r, &list, "Error parsing request body (expected JSON array of items)") {
		return
	}
	if len(list) == 0 {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Item list cannot be empty")
		return
	}
	if len(list) > maxBulkItems {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", fmt.Sprintf("Maximum %d items per import", maxBulkItems))
		return
	}

	for i := range list {
		if !list[i].IsValid() {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", fmt.Sprintf("Invalid item at index %d (%q)", i, list[i].Name))
			return
		}
	}

	result, err := g.itemsService.BulkCreate(r.Context(), list)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error importing items")
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (g *Gallery) HandleGetAllItems(w http.ResponseWriter, r *http.Request) {
	page, limit := paginationParams(r)

	itms, total, err := g.itemsService.GetAll(r.Context(), page, limit)
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
		},
	})
}

func (g *Gallery) HandleGetItem(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "itemId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid item id")
		return
	}

	item, err := g.itemsService.GetByID(r.Context(), items.ID(id))
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

func (g *Gallery) HandleUpdateItem(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	id, err := parseID(r, "itemId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid item id")
		return
	}

	item := &items.Item{}
	if !decodeBody(w, r, item) {
		return
	}
	item.ID = items.ID(id)
	if !item.IsValid() {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid item")
		return
	}

	updated, err := g.itemsService.UpdateItem(r.Context(), item)
	if err != nil {
		if isItemNotFound(err) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Item not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error updating item")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (g *Gallery) HandleDeleteItem(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	id, err := parseID(r, "itemId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid item id")
		return
	}

	if err := g.itemsService.DeleteItem(r.Context(), items.ID(id)); err != nil {
		if isItemNotFound(err) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Item not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error deleting item")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"result": "success"})
}
