package handlers

import (
	"dZev1/character-gallery/internal/auth"
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/daily_run"
	"dZev1/character-gallery/internal/inventory"
	"dZev1/character-gallery/internal/items"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type dailyStartRequest struct {
	CharacterID uint64 `json:"character_id"`
}

type dailyMoveRequest struct {
	CharacterID uint64 `json:"character_id"`
	NodeID      string `json:"node_id"`
}

type dailyActionRequest struct {
	CharacterID uint64               `json:"character_id"`
	Action      daily_run.ActionType `json:"action"`
	ItemID      *uint64              `json:"item_id,omitempty"`
}

func parseTier(r *http.Request) (uint8, error) {
	raw := r.PathValue("tier")
	val, err := strconv.Atoi(raw)
	if err != nil || val < 1 || val > 5 {
		return 0, daily_run.ErrInvalidTier
	}
	return uint8(val), nil
}

func (g *Gallery) HandleGetDaily(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	tier, err := parseTier(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Tier must be between 1 and 5")
		return
	}

	view, err := g.dailyService.GetDailyView(ctx, claims.UserID, tier)
	if err != nil {
		writeDailyError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, view)
}

func (g *Gallery) HandleGetDailyLeaderboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tier, err := parseTier(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Tier must be between 1 and 5")
		return
	}

	page, limit := paginationParams(r)
	scores, total, err := g.dailyService.GetLeaderboard(ctx, tier, page, limit)
	if err != nil {
		writeDailyError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, paginatedResponse{
		Data: scores,
		Pagination: pagination{
			Page:    page,
			Limit:   limit,
			Total:   total,
			HasNext: total > uint64(page*limit),
		},
	})
}

func (g *Gallery) HandleStartDailyRun(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req dailyStartRequest
	limitBody(w, r)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CharacterID == 0 {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Valid character_id is required")
		return
	}

	resp, err := g.dailyService.StartOrResumeRun(ctx, claims.UserID, claims.Role == "admin", characters.ID(req.CharacterID))
	if err != nil {
		writeDailyError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (g *Gallery) HandleMoveDailyRun(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req dailyMoveRequest
	limitBody(w, r)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CharacterID == 0 || req.NodeID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Valid character_id and node_id are required")
		return
	}

	resp, err := g.dailyService.Move(ctx, claims.UserID, claims.Role == "admin", characters.ID(req.CharacterID), req.NodeID)
	if err != nil {
		writeDailyError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (g *Gallery) HandleDailyCombatAction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req dailyActionRequest
	limitBody(w, r)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CharacterID == 0 {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Valid character_id and action are required")
		return
	}

	var itemID *items.ID
	if req.ItemID != nil {
		id := items.ID(*req.ItemID)
		itemID = &id
	}

	resp, err := g.dailyService.CombatAction(ctx, claims.UserID, claims.Role == "admin", characters.ID(req.CharacterID), req.Action, itemID)
	if err != nil {
		writeDailyError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeDailyError(w http.ResponseWriter, err error) {
	switch {
	case isCharacterNotFound(err):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Character not found")
	case errors.Is(err, daily_run.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Active daily run not found")
	case errors.Is(err, daily_run.ErrRunAlreadyFinished):
		writeError(w, http.StatusConflict, "CONFLICT", err.Error())
	case errors.Is(err, daily_run.ErrInvalidTier),
		errors.Is(err, daily_run.ErrCharacterDead),
		errors.Is(err, daily_run.ErrRunNotInCombat),
		errors.Is(err, daily_run.ErrRunInCombat),
		errors.Is(err, daily_run.ErrInvalidMove),
		errors.Is(err, daily_run.ErrInvalidAction),
		errors.Is(err, inventory.ErrItemNotConsumable),
		errors.Is(err, inventory.ErrInsufficientQuantity):
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error processing daily run")
	}
}
