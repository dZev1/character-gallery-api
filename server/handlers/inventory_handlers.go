package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"dZev1/character-gallery/models/characters"
	"dZev1/character-gallery/models/inventory"
)

func (h *CharacterHandler) AddItemToCharacter(w http.ResponseWriter, r *http.Request) {
	characterIDStr := r.PathValue("character_id")
	itemIDStr := r.PathValue("item_id")
	quantityStr := r.URL.Query().Get("quantity")

	if quantityStr == "" {
		quantityStr = "1"
	}

	characterID, err := strconv.Atoi(characterIDStr)
	if err != nil {
		er := &Error{
			Error: "Invalid character id",
			Code:  "BAD_REQUEST",
			Details: struct {
				CharacterID string `json:"character_id"`
			}{
				CharacterID: characterIDStr,
			},
		}
		throwError(er, w, http.StatusBadRequest)
		return
	}

	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		er := &Error{
			Error: "Invalid item id",
			Code:  "BAD_REQUEST",
			Details: struct {
				ItemID string `json:"item_id"`
			}{
				ItemID: itemIDStr,
			},
		}
		throwError(er, w, http.StatusBadRequest)
		return
	}

	quantity, err := strconv.Atoi(quantityStr)
	if err != nil || quantity < 1 {
		er := &Error{
			Error: "Invalid item quantity",
			Code:  "BAD_REQUEST",
			Details: struct {
				Quantity string `json:"quantity"`
			}{
				Quantity: quantityStr,
			},
		}
		throwError(er, w, http.StatusBadRequest)
		return
	}

	item, err := h.Gallery.AddItemToCharacter(characters.CharacterID(characterID), inventory.ItemID(itemID), uint8(quantity))
	if err != nil {
		er := &Error{
			Error: "Could not add item to character",
			Code:  "INTERNAL_SERVER_ERROR",
		}
		throwError(er, w, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(item)
}

func (h *CharacterHandler) RemoveItemFromCharacter(w http.ResponseWriter, r *http.Request) {
	characterIDStr := r.PathValue("character_id")
	itemIDStr := r.PathValue("item_id")
	quantityStr := r.URL.Query().Get("quantity")

	if quantityStr == "" {
		quantityStr = "1"
	}

	characterID, err := strconv.Atoi(characterIDStr)
	if err != nil {
		er := &Error{
			Error: "Invalid character ID",
			Code:  "BAD_REQUEST",
			Details: struct {
				CharacterID string `json:"character_id"`
			}{
				CharacterID: characterIDStr,
			},
		}
		throwError(er, w, http.StatusBadRequest)
		return
	}

	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		er := &Error{
			Error: "Invalid item ID",
			Code:  "BAD_REQUEST",
			Details: struct {
				ItemID string `json:"item_id"`
			}{
				ItemID: itemIDStr,
			},
		}
		throwError(er, w, http.StatusBadRequest)
		return
	}

	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		er := &Error{
			Error: "Invalid quantity",
			Code:  "BAD_REQUEST",
			Details: struct {
				Quantity string `json:"quantity"`
			}{
				Quantity: quantityStr,
			},
		}
		throwError(er, w, http.StatusBadRequest)
		return
	}

	err = h.Gallery.RemoveItemFromCharacter(characters.CharacterID(characterID), inventory.ItemID(itemID), uint8(quantity))
	if err != nil {
		er := &Error{
			Error: "Could not remove item from character",
			Code:  "INTERNAL_SERVER_ERROR",
			Details: struct {
				CharacterID string `json:"character_id"`
				ItemID      string `json:"item_id"`
				Quantity    string `json:"quantity"`
			}{
				CharacterID: characterIDStr,
				ItemID:      itemIDStr,
				Quantity:    quantityStr,
			},
		}
		throwError(er, w, http.StatusInternalServerError)
		return
	}

	item, _ := h.Gallery.DisplayItem(inventory.ItemID(itemID))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(item)
}

func (h *CharacterHandler) GetCharacterInventory(w http.ResponseWriter, r *http.Request) {
	characterIDStr := r.PathValue("character_id")

	characterID, err := strconv.Atoi(characterIDStr)
	if err != nil {
		er := &Error{
			Error: "Invalid character ID",
			Code:  "BAD_REQUEST",
			Details: struct {
				CharacterID string `json:"character_id"`
			}{
				CharacterID: characterIDStr,
			},
		}
		throwError(er, w, http.StatusBadRequest)
		return
	}

	invItems, err := h.Gallery.GetCharacterInventory(characters.CharacterID(characterID))
	if err != nil {
		er := &Error{
			Error: "Could not retrieve inventory",
			Code:  "INTERNAL_SERVER_ERROR",
		}
		throwError(er, w, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(invItems)
}

func (h *CharacterHandler) ShowPoolItems(w http.ResponseWriter, r *http.Request) {
	items, err := h.Gallery.DisplayPoolItems()
	if err != nil {
		er := &Error{
			Error: "Could not retrieve pool items",
			Code:  "INTERNAL_SERVER_ERROR",
		}
		throwError(er, w, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(items)
}

func (h *CharacterHandler) ShowItem(w http.ResponseWriter, r *http.Request) {
	itemIDStr := r.PathValue("item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	item, err := h.Gallery.DisplayItem(inventory.ItemID(itemID) - 1)
	if err != nil {
		er := &Error{
			Error: "Could not retrieve item from item pool",
			Code:  "INTERNAL_SERVER_ERROR",
		}
		throwError(er, w, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(item)
}

func (h *CharacterHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	newItem := &inventory.Item{}

	err := json.NewDecoder(r.Body).Decode(newItem)
	if err != nil {
		er := &Error{
			Error: "Invalid request body",
			Code:  "BAD_REQUEST",
		}
		throwError(er, w, http.StatusBadRequest)
		return
	}
	if !validateItem(newItem, w) {
		er := &Error{
			Error: "Invalid item",
			Code:  "BAD_REQUEST",
		}
		throwError(er, w, http.StatusBadRequest)
		return
	}
	err = h.Gallery.CreateItem(newItem)
	if err != nil {
		er := &Error{
			Error: "Could not create item",
			Code:  "INTERNAL_SERVER_ERROR",
		}
		throwError(er, w, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(newItem)
}
