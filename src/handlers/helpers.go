package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"unicode"

	"dZev1/character-gallery/internal/auth"
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
	ErrInvalidEmail    = errors.New("invalid email address")
	ErrInvalidPassword = errors.New("password must be 8-72 bytes")
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

func requireClaims(w http.ResponseWriter, r *http.Request) (*auth.Claims, bool) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return nil, false
	}
	return claims, true
}

func requireAdmin(w http.ResponseWriter, r *http.Request) (*auth.Claims, bool) {
	claims, ok := requireClaims(w, r)
	if !ok {
		return nil, false
	}
	if claims.Role != "admin" {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Admin role required")
		return nil, false
	}
	return claims, true
}

func decodeBody(w http.ResponseWriter, r *http.Request, dst any, errMsg ...string) bool {
	limitBody(w, r)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		msg := "Error parsing request body"
		if len(errMsg) > 0 && errMsg[0] != "" {
			msg = errMsg[0]
		}
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", msg)
		return false
	}
	return true
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
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func validateUsername(username string) error {
	if len(username) < 3 || len(username) > 30 {
		return ErrInvalidUsername
	}
	for _, c := range username {
		if !(unicode.IsLetter(c) || unicode.IsDigit(c) || c == '_') {
			return ErrInvalidUsername
		}
	}
	return nil
}

func validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if len(email) < 5 || len(email) > 254 {
		return ErrInvalidEmail
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return ErrInvalidEmail
	}
	at := strings.LastIndex(email, "@")
	if at <= 0 || at >= len(email)-1 {
		return ErrInvalidEmail
	}
	domain := email[at+1:]
	if !strings.Contains(domain, ".") || strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return ErrInvalidEmail
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 || len(password) > 72 {
		return ErrInvalidPassword
	}
	return nil
}

func validateRegistration(username, email, password string) error {
	if err := validateUsername(username); err != nil {
		return err
	}
	if err := validateEmail(email); err != nil {
		return err
	}
	return validatePassword(password)
}

func validateLogin(req loginRequest) (string, error) {
	if err := validatePassword(req.Password); err != nil {
		return "", err
	}
	if strings.TrimSpace(req.Email) != "" {
		if err := validateEmail(req.Email); err != nil {
			return "", err
		}
		return strings.TrimSpace(req.Email), nil
	}
	if strings.Contains(req.Username, "@") {
		if err := validateEmail(req.Username); err != nil {
			return "", err
		}
		return strings.TrimSpace(req.Username), nil
	}
	if err := validateUsername(req.Username); err != nil {
		return "", err
	}
	return req.Username, nil
}
