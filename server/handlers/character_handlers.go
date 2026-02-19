package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"dZev1/character-gallery/models"
	"dZev1/character-gallery/models/characters"
)

type CharacterHandler struct {
	Gallery models.CharacterGallery
}

type Pagination struct {
	Page    uint8  `json:"page"`
	Limit   uint8  `json:"limit"`
	Total   uint64 `json:"total"`
	HasNext bool   `json:"has_next"`
}

func (h *CharacterHandler) CreateCharacter(w http.ResponseWriter, r *http.Request) {
	newCharacter := &characters.Character{}

	err := json.NewDecoder(r.Body).Decode(newCharacter)
	if err != nil {
		er := &Error{
			Error: "Invalid request body",
			Code:  "BAD_REQUEST",
		}
		throwError(er, w, http.StatusBadRequest)
		return
	}

	if valid := validateCharacter(newCharacter, w); !valid {
		return
	}

	err = h.Gallery.Create(newCharacter)
	if err != nil {
		er := &Error{
			Error: "Could not create character",
			Code:  "INTERNAL_SERVER_ERROR",
		}
		throwError(er, w, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(*newCharacter)
}

func (h *CharacterHandler) GetAllCharacters(w http.ResponseWriter, r *http.Request) {
	page := 0

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil || p < 0 {
			er := &Error{
				Error: "Invalid page number",
				Code:  "BAD_REQUEST",
				Details: struct {
					Page string `json:"page"`
				}{
					Page: pageStr,
				},
			}
			throwError(er, w, http.StatusBadRequest)
			return
		}
		page = p * 20
	}

	chars, totalChars, err := h.Gallery.GetAll(page)

	response := struct {
		Data       []characters.Character `json:"data"`
		Pagination Pagination             `json:"pagination"`
	}{
		Data: chars,
		Pagination: Pagination{
			Page:    uint8(page / 20),
			Limit:   20,
			Total:   totalChars,
			HasNext: len(chars) == 20,
		},
	}
	if err != nil {
		er := &Error{
			Error: "Page not found",
			Code:  "NOT_FOUND",
		}
		throwError(er, w, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *CharacterHandler) GetCharacter(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		er := &Error{
			Error: "Invalid ID",
			Code:  "BAD_REQUEST",
			Details: struct {
				ID string `json:"id"`
			}{
				ID: idStr,
			},
		}
		throwError(er, w, http.StatusBadRequest)
		return
	}

	character, err := h.Gallery.Get(characters.CharacterID(id))
	if err != nil {
		er := &Error{
			Error: "Character not found",
			Code:  "NOT_FOUND",
			Details: struct {
				ID string `json:"id"`
			}{
				ID: idStr,
			},
		}
		throwError(er, w, http.StatusNotFound)
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
		er := &Error{
			Error: "Invalid ID",
			Code:  "BAD_REQUEST",
			Details: struct {
				ID string `json:"id"`
			}{
				ID: idStr,
			},
		}
		throwError(er, w, http.StatusBadRequest)
		return
	}

	characterToEdit := &characters.Character{}
	err = json.NewDecoder(r.Body).Decode(characterToEdit)
	if err != nil {
		er := &Error{
			Error: "Invalid Request Body",
			Code:  "BAD_REQUEST",
		}
		throwError(er, w, http.StatusBadRequest)
		return
	}

	if valid := validateCharacter(characterToEdit, w); !valid {
		er := &Error{
			Error: "Invalid character",
			Code:  "BAD_REQUEST",
		}
		throwError(er, w, http.StatusBadRequest)
		return
	}

	characterToEdit.ID = characters.CharacterID(id)
	characterToEdit.Stats.ID = characters.CharacterID(id)
	characterToEdit.Customization.ID = characters.CharacterID(id)

	err = h.Gallery.Edit(characterToEdit)
	if err != nil {
		er := &Error{
			Error: "Could not edit character",
			Code:  "INTERNAL_SERVER_ERROR",
		}
		throwError(er, w, http.StatusInternalServerError)
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
		er := &Error{
			Error: "Invalid ID",
			Code:  "BAD_REQUEST",
			Details: struct {
				ID string `json:"id"`
			}{
				ID: idStr,
			},
		}
		throwError(er, w, http.StatusBadRequest)
		return
	}

	err = h.Gallery.Remove(characters.CharacterID(id))
	if err != nil {
		er := &Error{
			Error: "Character not found",
			Code:  "NOT_FOUND",
			Details: struct {
				ID string `json:"id"`
			}{
				ID: idStr,
			},
		}
		throwError(er, w, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
