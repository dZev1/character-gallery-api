package handlers

import (
	"dZev1/character-gallery/internal/auth"
	"dZev1/character-gallery/internal/characters"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type queryCharactersRequest struct {
	Mine  bool `json:"mine"`
	Page  int  `json:"page"`
	Limit int  `json:"limit"`
}

func ownerIDForRequest(claims *auth.Claims) uuid.UUID {
	if claims.Role == "admin" {
		return uuid.Nil
	}
	return claims.UserID
}

func (g *Gallery) HandleGetStatRoll(w http.ResponseWriter, r *http.Request) {
	claims, ok := requireClaims(w, r)
	if !ok {
		return
	}

	view, err := g.characterService.GetOrCreateStatRoll(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error getting stat roll")
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (g *Gallery) HandleRerollStats(w http.ResponseWriter, r *http.Request) {
	claims, ok := requireClaims(w, r)
	if !ok {
		return
	}

	view, err := g.characterService.RerollStats(r.Context(), claims.UserID)
	if err != nil {
		switch {
		case errors.Is(err, characters.ErrCreateCooldown):
			writeError(w, http.StatusConflict, "CHARACTER_CREATE_COOLDOWN", "You can only create one character per week")
		case errors.Is(err, characters.ErrRerollCooldown):
			writeError(w, http.StatusConflict, "REROLL_COOLDOWN", "You can only reroll stats once per week")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error rerolling stats")
		}
		return
	}
	writeJSON(w, http.StatusOK, view)
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

	useRoll := claims.Role != "admin" || strings.EqualFold(r.URL.Query().Get("from_roll"), "true")

	var (
		created *characters.Character
		err     error
	)
	if useRoll {
		created, err = g.characterService.CreateWithStatRoll(r.Context(), character, claims.UserID)
	} else {
		created, err = g.characterService.Create(r.Context(), character, claims.UserID)
	}
	if err != nil {
		switch {
		case errors.Is(err, characters.ErrCharacterLimitReached):
			writeError(w, http.StatusConflict, "CHARACTER_LIMIT_REACHED", "Maximum of 6 characters per user reached")
		case errors.Is(err, characters.ErrCreateCooldown):
			writeError(w, http.StatusConflict, "CHARACTER_CREATE_COOLDOWN", "You can only create one character per week")
		case errors.Is(err, characters.ErrNoStatRoll):
			writeError(w, http.StatusBadRequest, "NO_STAT_ROLL", "You must roll stats before creating a character")
		case errors.Is(err, characters.ErrInvalidStatAllocation):
			writeError(w, http.StatusBadRequest, "INVALID_STAT_ALLOCATION", "Stats must match your active dice roll")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error creating character")
		}
		return
	}
	created.OwnerUsername = claims.Username
	writeJSON(w, http.StatusCreated, created)
}

func (g *Gallery) HandleGetAllCharacters(w http.ResponseWriter, r *http.Request) {
	page, limit := paginationParams(r)
	mine := strings.EqualFold(r.URL.Query().Get("mine"), "true")
	g.listCharactersResponse(w, r, mine, page, limit)
}

func (g *Gallery) HandleQueryCharacters(w http.ResponseWriter, r *http.Request) {
	var req queryCharactersRequest
	limitBody(w, r)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Error parsing request body")
		return
	}

	page := req.Page
	if page < 1 {
		page = defaultPage
	} else if page > maxPage {
		page = maxPage
	}

	limit := req.Limit
	if limit < minLimit {
		limit = defaultLimit
	} else if limit > maxLimit {
		limit = maxLimit
	}

	g.listCharactersResponse(w, r, req.Mine, page, limit)
}

func (g *Gallery) listCharactersResponse(w http.ResponseWriter, r *http.Request, mine bool, page, limit int) {
	var (
		chars []characters.Character
		total uint64
		err   error
	)

	if mine {
		claims, ok := requireClaims(w, r)
		if !ok {
			return
		}
		chars, total, err = g.characterService.GetByOwner(r.Context(), claims.UserID, page, limit)
	} else {
		chars, total, err = g.characterService.GetAll(r.Context(), page, limit)
	}

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
