package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/dZev1/character-gallery/models"
	"github.com/dZev1/character-gallery/models/characters"
	"github.com/dZev1/character-gallery/models/inventory"
)

type CharacterHandler struct {
	Gallery models.CharacterGallery
}

func (h *CharacterHandler) CreateCharacter(w http.ResponseWriter, r *http.Request) {
	newCharacter := &characters.Character{}

	err := json.NewDecoder(r.Body).Decode(newCharacter)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.Gallery.Create(newCharacter)
	if err != nil {
		http.Error(w, "Could not create character", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(*newCharacter)
}

func (h *CharacterHandler) GetAllCharacters(w http.ResponseWriter, r *http.Request) {
	page := 0
	limit := 20

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err == nil && p >= 0 {
			page = p
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	chars, err := h.Gallery.GetAll(page, limit)

	if err != nil {
		http.Error(w, err.Error(), http.StatusFailedDependency)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chars)
}

func (h *CharacterHandler) GetCharacter(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	character, err := h.Gallery.Get(characters.CharacterID(id))
	if err != nil {
		http.Error(w, "Character not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(character)
}

func (h *CharacterHandler) EditCharacter(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	characterToEdit := &characters.Character{}
	err = json.NewDecoder(r.Body).Decode(characterToEdit)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	characterToEdit.ID = characters.CharacterID(id)
	characterToEdit.Stats.ID = characters.CharacterID(id)
	characterToEdit.Customization.ID = characters.CharacterID(id)

	err = h.Gallery.Edit(characterToEdit)
	if err != nil {
		http.Error(w, "Could not edit character", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(characterToEdit)
}

func (h *CharacterHandler) DeleteCharacter(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = h.Gallery.Remove(characters.CharacterID(id))
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *CharacterHandler) AddItemToCharacter(w http.ResponseWriter, r *http.Request) {
	characterIDStr := r.PathValue("character_id")
	itemIDStr := r.PathValue("item_id")
	quantityStr := r.URL.Query().Get("quantity")

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

	err = h.Gallery.AddItemToCharacter(characters.CharacterID(characterID), inventory.ItemID(itemID), uint8(quantity))
	if err != nil {
		http.Error(w, "Could not add item to character", http.StatusInternalServerError)
		return
	}

	item := inventory.Items[itemID-1]

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(item)
}

func (h *CharacterHandler) RemoveItemFromCharacter(w http.ResponseWriter, r *http.Request) {
	characterIDStr := r.PathValue("character_id")
	itemIDStr := r.PathValue("item_id")
	quantityStr := r.URL.Query().Get("quantity")

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

	item := inventory.Items[itemID-1]

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