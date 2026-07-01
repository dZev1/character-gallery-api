package handlers

import (
	"dZev1/character-gallery/internal/characters"
	"encoding/json"
	"net/http"
)

func (g *Gallery) HandleCreateCharacter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	character := &characters.Character{}

	err := json.NewDecoder(r.Body).Decode(character)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Error parsing request body")
		return
	}

	if !character.Validate() {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid character")
		return
	}

	character, err = g.characterService.Create(ctx, character)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error creating character")
		return
	}

	writeJSON(w, http.StatusCreated, character)
}

func (g *Gallery) HandleGetAllCharacters(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page := max(parseQueryInt(r, "page", 1), 1)
	limit := max(parseQueryInt(r, "limit", 20), 20)

	chars, total, err := g.characterService.GetAll(ctx, page, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error getting all characters")
		return
	}

	data := paginatedResponse{
		Data: chars,
		Pagination: pagination{
			Page:    page,
			Limit:   limit,
			Total:   total,
			HasNext: total > uint64(page*limit),
		},
	}

	writeJSON(w, http.StatusOK, data)
}

func (g *Gallery) HandleGetCharacter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseID(r, "characterId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid character id")
		return
	}

	char, err := g.characterService.GetByID(ctx, characters.CharacterID(id))
	if err != nil {
		if isCharacterNotFound(err) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Character not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error getting character")
		return
	}

	writeJSON(w, http.StatusOK, char)
}

func (g *Gallery) HandleUpdateCharacter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseID(r, "characterId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid character id")
		return
	}

	character := &characters.Character{}
	err = json.NewDecoder(r.Body).Decode(character)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Error parsing request body")
		return
	}

	character.ID = characters.CharacterID(id)
	if !character.Validate() {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid character")
		return
	}

	character, err = g.characterService.Update(ctx, character)
	if err != nil {
		if isCharacterNotFound(err) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Character not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error updating character")
		return
	}

	writeJSON(w, http.StatusOK, character)
}

func (g *Gallery) HandleDeleteCharacter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseID(r, "characterId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid character id")
		return
	}

	err = g.characterService.Delete(ctx, characters.CharacterID(id))
	if err != nil {
		if isCharacterNotFound(err) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Character not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error deleting character")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"result": "success"})
}
