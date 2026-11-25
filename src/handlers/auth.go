package handlers

import (
	"dZev1/character-gallery/internal/auth"
	"encoding/json"
	"errors"
	"net/http"
)

type AuthHandler struct {
	svc *auth.Service
}

func NewAuthHandler(svc *auth.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req registerRequest
	limitBody(w, r)
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Error parsing request body")
		return
	}

	if err := validateRegistration(req.Username, req.Email, req.Password); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	resp, err := h.svc.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrDuplicateUser) {
			writeError(w, http.StatusConflict, "CONFLICT", "Username or email already in use")
		} else {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error registering user")
		}
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req loginRequest
	limitBody(w, r)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Error parsing request body")
		return
	}

	identifier, err := validateLogin(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid credentials")
		return
	}

	resp, err := h.svc.Login(ctx, identifier, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid credentials")
		} else {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error logging in")
		}
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
