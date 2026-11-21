package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"unicode"

	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/items"
)

// maxBodyBytes caps the size of request bodies to avoid memory exhaustion via
// oversized payloads. Character customizations make bodies small; 1 MiB is a
// generous ceiling.
const maxBodyBytes = 1 << 20

func limitBody(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
}

const (
	defaultPage  = 1
	defaultLimit = 20
	minLimit     = 20
	maxLimit     = 100
	maxPage      = 10000
)

// paginationParams resolves and clamps page/limit query values so clients
// cannot trigger unbounded scans (e.g. limit=10000000).
func paginationParams(r *http.Request) (page, limit int) {
	page = parseQueryInt(r, "page", defaultPage)
	if page < 1 {
		page = defaultPage
	}
	if page > maxPage {
		page = maxPage
	}

	limit = parseQueryInt(r, "limit", defaultLimit)
	if limit < minLimit {
		limit = minLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return page, limit
}

var (
	ErrInvalidUsername = errors.New("username must be 3-30 alphanumeric characters or underscores")
	ErrInvalidPassword = errors.New("password must be 8-128 characters")
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

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func validateCredentials(username, password string) error {
	if len(username) < 3 || len(username) > 30 {
		return ErrInvalidUsername
	}
	for _, c := range username {
		if !(unicode.IsLetter(c) || unicode.IsDigit(c) || c == '_') {
			return ErrInvalidUsername
		}
	}
	if len(password) < 8 || len(password) > 128 {
		return ErrInvalidPassword
	}
	return nil
}
