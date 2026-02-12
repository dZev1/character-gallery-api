package handlers

import (
	"dZev1/character-gallery/models/characters"
	"dZev1/character-gallery/models/inventory"
	"net/http"
)

func validateCharacter(character *characters.Character, w http.ResponseWriter) bool {
	if len(character.Name) < 2 {
		http.Error(w, "Character's name is too short", http.StatusBadRequest)
		return false
	}
	if !character.BodyType.Validate() {
		http.Error(w, "Character's body type not valid", http.StatusBadRequest)
		return false
	}
	if !character.Class.Validate() {
		http.Error(w, "Character's class not valid", http.StatusBadRequest)
		return false
	}
	if !character.Species.Validate() {
		http.Error(w, "Character's species not valid", http.StatusBadRequest)
		return false
	}
	return true
}

func validateItem(item *inventory.Item, w http.ResponseWriter) bool {
	if !item.Validate() || !item.Type.Validate() {
		http.Error(w, "Item not valid", http.StatusBadRequest)
		return false
	}
	return true
}