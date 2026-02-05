package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/dZev1/character-gallery/models/characters"
	"github.com/dZev1/character-gallery/models/inventory"
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
		http.Error(w, "Invalid character ID", http.StatusBadRequest)
		return
	}

	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	quantity, err := strconv.Atoi(quantityStr)

	if err != nil || quantity < 1 {
		http.Error(w, "Invalid quantity", http.StatusBadRequest)
		return
	}

	err = h.Gallery.AddItemToCharacter(characters.CharacterID(characterID), inventory.ItemID(itemID), uint8(quantity))
	if err != nil {
		http.Error(w, "Could not add item to character", http.StatusInternalServerError)
		return
	}

	item, _ := h.Gallery.DisplayItem(inventory.ItemID(itemID))
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
		http.Error(w, "Invalid character ID", http.StatusBadRequest)
		return
	}

	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		http.Error(w, "Invalid quantity", http.StatusBadRequest)
		return
	}

	err = h.Gallery.RemoveItemFromCharacter(characters.CharacterID(characterID), inventory.ItemID(itemID), uint8(quantity))
	if err != nil {
		http.Error(w, "Could not remove item from character", http.StatusInternalServerError)
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
		http.Error(w, "Invalid character ID", http.StatusBadRequest)
		return
	}

	invItems, err := h.Gallery.GetCharacterInventory(characters.CharacterID(characterID))
	if err != nil {
		http.Error(w, "Could not retrieve inventory", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(invItems)
}

func (h *CharacterHandler) ShowPoolItems(w http.ResponseWriter, r *http.Request) {
	items, err := h.Gallery.DisplayPoolItems()
	if err != nil {
		http.Error(w, "Could not retrieve pool items", http.StatusInternalServerError)
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
		http.Error(w, "Could not retrieve item from item pool", http.StatusInternalServerError)
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
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.Gallery.CreateItem(newItem)
	if err != nil {
		http.Error(w, "Could not create item: "+err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(newItem)
}