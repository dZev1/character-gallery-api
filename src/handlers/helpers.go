package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/items"
)

func isCharacterNotFound(err error) bool {
	return errors.Is(err, characters.ErrNotFound)
}

func isItemNotFound(err error) bool {
	return errors.Is(err, items.ErrNotFound)
}

type apiError struct {
	Error   string      `json:"error"`
	Code    string      `json:"code"`
	Details interface{} `json:"details,omitempty"`
}

type pagination struct {
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
	Total   uint64 `json:"total"`
	HasNext bool   `json:"has_next"`
}

type paginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination pagination  `json:"pagination"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, apiError{
		Error: msg,
		Code:  code,
	})
}

func parseID(r *http.Request, name string) (uint64, error) {
	s := r.PathValue(name)
	id, err := strconv.ParseUint(s, 10, 64)
	if err != nil || id == 0 {
		return 0, fmt.Errorf("invalid %s: %q", name, s)
	}
	return id, nil
}

func parseQueryInt(r *http.Request, name string, defaultVal int) int {
	s := r.URL.Query().Get(name)
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil || val < 0 {
		return defaultVal
	}
	return val
}
