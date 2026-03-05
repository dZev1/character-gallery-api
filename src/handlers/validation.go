package handlers

import (
	"dZev1/character-gallery/models/characters"
	"dZev1/character-gallery/models/inventory"
)

func validateCharacter(character *characters.Character) (bool, string) {
	if len(character.Name) < 2 {
		return false, "Character's name is too short" 
	}
	if !character.BodyType.Validate() {
		return false, "Character's body type not valid"
	}
	if !character.Class.Validate() {
		return false, "Character's class not valid"
	}
	if !character.Species.Validate() {
		return false, "Character's species not valid"
	}
	return true, ""
}

func validateItem(item *inventory.Item) (bool, string) {
	if !item.Validate() || !item.Type.Validate() {
		return false, "Item not valid"
	}
	return true, ""
}
