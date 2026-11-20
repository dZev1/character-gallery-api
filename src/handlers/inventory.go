package handlers

import (
	"dZev1/character-gallery/internal/auth"
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/inventory"
	"dZev1/character-gallery/internal/items"
	"errors"
	"net/http"
)

func (g *Gallery) HandleGetCharacterInventory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseID(r, "characterId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid character id")
		return
	}

	inv, err := g.inventoryService.GetByCharacterID(ctx, characters.ID(id))
	if err != nil {
		if isCharacterNotFound(err) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Character not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error getting inventory")
		return
	}

	writeJSON(w, http.StatusOK, inv)
}

func (g *Gallery) HandleAddItemToCharacter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	charID, err := parseID(r, "characterId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid character id")
		return
	}

	itemID, err := parseID(r, "itemId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid item id")
		return
	}

	quantity := parseQueryInt(r, "quantity", 1)
	if quantity < 1 {
		quantity = 1
	}
	if quantity > 255 {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid quantity")
		return
	}

	item, err := g.inventoryService.AddItem(ctx, characters.ID(charID), ownerIDForRequest(claims), items.ID(itemID), uint8(quantity))
	if err != nil {
		switch {
		case isCharacterNotFound(err):
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Character not found")
		case isItemNotFound(err):
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Item not found")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error adding item to character")
		}
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (g *Gallery) HandleRemoveItemFromCharacter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	charID, err := parseID(r, "characterId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid character id")
		return
	}

	itemID, err := parseID(r, "itemId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid item id")
		return
	}

	quantity := parseQueryInt(r, "quantity", 1)
	if quantity < 1 {
		quantity = 1
	}
	if quantity > 255 {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid quantity")
		return
	}

	_, err = g.inventoryService.DeleteItem(ctx, characters.ID(charID), ownerIDForRequest(claims), items.ID(itemID), uint8(quantity))
	if err != nil {
		switch {
		case isCharacterNotFound(err):
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Character not found")
		case isItemNotFound(err):
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Item not found")
		case errors.Is(err, inventory.ErrInsufficientQuantity):
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Cannot remove more items than character owns")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error removing item from character")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"result": "success"})
}
