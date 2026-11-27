package handlers

import (
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/inventory"
	"dZev1/character-gallery/internal/items"
	"errors"
	"net/http"
)

func parseCharAndItemIDs(w http.ResponseWriter, r *http.Request) (characters.ID, items.ID, bool) {
	charID, err := parseID(r, "characterId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid character id")
		return 0, 0, false
	}
	itemID, err := parseID(r, "itemId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid item id")
		return 0, 0, false
	}
	return characters.ID(charID), items.ID(itemID), true
}

func parseQuantityParam(w http.ResponseWriter, r *http.Request) (uint8, bool) {
	qty := parseQueryInt(r, "quantity", 1)
	if qty < 1 {
		qty = 1
	}
	if qty > 255 {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid quantity")
		return 0, false
	}
	return uint8(qty), true
}

func (g *Gallery) HandleGetCharacterInventory(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "characterId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid character id")
		return
	}

	inv, err := g.inventoryService.GetByCharacterID(r.Context(), characters.ID(id))
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
	claims, ok := requireClaims(w, r)
	if !ok {
		return
	}
	charID, itemID, ok := parseCharAndItemIDs(w, r)
	if !ok {
		return
	}
	qty, ok := parseQuantityParam(w, r)
	if !ok {
		return
	}

	item, err := g.inventoryService.AddItem(r.Context(), charID, ownerIDForRequest(claims), itemID, qty)
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
	claims, ok := requireClaims(w, r)
	if !ok {
		return
	}
	charID, itemID, ok := parseCharAndItemIDs(w, r)
	if !ok {
		return
	}
	qty, ok := parseQuantityParam(w, r)
	if !ok {
		return
	}

	_, err := g.inventoryService.DeleteItem(r.Context(), charID, ownerIDForRequest(claims), itemID, qty)
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

type equipRequest struct {
	Equip bool `json:"equip"`
}

func (g *Gallery) HandleEquipItem(w http.ResponseWriter, r *http.Request) {
	claims, ok := requireClaims(w, r)
	if !ok {
		return
	}
	charID, itemID, ok := parseCharAndItemIDs(w, r)
	if !ok {
		return
	}

	var req equipRequest
	if !decodeBody(w, r, &req) {
		return
	}

	result, err := g.inventoryService.EquipItem(r.Context(), charID, ownerIDForRequest(claims), itemID, req.Equip)
	if err != nil {
		switch {
		case isCharacterNotFound(err):
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Character not found")
		case isItemNotFound(err):
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Item not found")
		case errors.Is(err, inventory.ErrItemNotInInventory):
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Item not in character inventory")
		case errors.Is(err, inventory.ErrItemNotEquippable):
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Item is not equippable")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error equipping item")
		}
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (g *Gallery) HandleUseItem(w http.ResponseWriter, r *http.Request) {
	claims, ok := requireClaims(w, r)
	if !ok {
		return
	}
	charID, itemID, ok := parseCharAndItemIDs(w, r)
	if !ok {
		return
	}

	char, err := g.inventoryService.UseItem(r.Context(), charID, ownerIDForRequest(claims), itemID)
	if err != nil {
		switch {
		case isCharacterNotFound(err):
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Character not found")
		case isItemNotFound(err):
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Item not found")
		case errors.Is(err, inventory.ErrInsufficientQuantity):
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Item not in character inventory")
		case errors.Is(err, inventory.ErrItemNotConsumable):
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Item cannot be consumed for healing")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error using item")
		}
		return
	}
	writeJSON(w, http.StatusOK, char)
}
