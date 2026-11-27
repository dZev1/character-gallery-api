package handlers

import (
	"dZev1/character-gallery/internal/auth"
	"dZev1/character-gallery/internal/characters"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

func ownerIDForRequest(claims *auth.Claims) uuid.UUID {
	if claims.Role == "admin" {
		return uuid.Nil
	}
	return claims.UserID
}

func (g *Gallery) HandleCreateCharacter(w http.ResponseWriter, r *http.Request) {
	claims, ok := requireClaims(w, r)
	if !ok {
		return
	}

	character := &characters.Character{}
	if !decodeBody(w, r, character) {
		return
	}
	if !character.IsValid() {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid character")
		return
	}

	created, err := g.characterService.Create(r.Context(), character, claims.UserID)
	if err != nil {
		if errors.Is(err, characters.ErrCharacterLimitReached) {
			writeError(w, http.StatusConflict, "CHARACTER_LIMIT_REACHED", "Maximum of 6 characters per user reached")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error creating character")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (g *Gallery) HandleGetAllCharacters(w http.ResponseWriter, r *http.Request) {
	page, limit := paginationParams(r)

	chars, total, err := g.characterService.GetAll(r.Context(), page, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error getting all characters")
		return
	}

	writeJSON(w, http.StatusOK, paginatedResponse{
		Data: chars,
		Pagination: pagination{
			Page:    page,
			Limit:   limit,
			Total:   total,
			HasNext: total > uint64(page*limit),
		},
	})
}

func (g *Gallery) HandleGetCharacter(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "characterId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid character id")
		return
	}

	char, err := g.characterService.GetByID(r.Context(), characters.ID(id))
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
	claims, ok := requireClaims(w, r)
	if !ok {
		return
	}

	id, err := parseID(r, "characterId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid character id")
		return
	}

	character := &characters.Character{}
	if !decodeBody(w, r, character) {
		return
	}
	character.ID = characters.ID(id)
	if !character.IsValid() {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid character")
		return
	}

	updated, err := g.characterService.Update(r.Context(), character, ownerIDForRequest(claims))
	if err != nil {
		if isCharacterNotFound(err) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Character not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error updating character")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (g *Gallery) HandleDeleteCharacter(w http.ResponseWriter, r *http.Request) {
	claims, ok := requireClaims(w, r)
	if !ok {
		return
	}

	id, err := parseID(r, "characterId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid character id")
		return
	}

	if err = g.characterService.Delete(r.Context(), characters.ID(id), ownerIDForRequest(claims)); err != nil {
		if isCharacterNotFound(err) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Character not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error deleting character")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (g *Gallery) HandlePrestigeCharacter(w http.ResponseWriter, r *http.Request) {
	claims, ok := requireClaims(w, r)
	if !ok {
		return
	}

	id, err := parseID(r, "characterId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid character id")
		return
	}

	char, err := g.characterService.Prestige(r.Context(), characters.ID(id), ownerIDForRequest(claims))
	if err != nil {
		switch {
		case isCharacterNotFound(err):
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Character not found")
		case errors.Is(err, characters.ErrPrestigeRequirement):
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Character must be at least level 10 to prestige")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error prestiging character")
		}
		return
	}
	writeJSON(w, http.StatusOK, char)
}
