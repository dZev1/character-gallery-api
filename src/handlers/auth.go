package handlers

import (
	"dZev1/character-gallery/internal/auth"
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
	var req registerRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if err := validateRegistration(req.Username, req.Email, req.Password); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	resp, err := h.svc.Register(r.Context(), req.Username, req.Email, req.Password)
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
	var req loginRequest
	if !decodeBody(w, r, &req) {
		return
	}

	identifier, err := validateLogin(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid credentials")
		return
	}

	resp, err := h.svc.Login(r.Context(), identifier, req.Password)
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
