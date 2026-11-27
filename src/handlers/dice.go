package handlers

import (
	"dZev1/character-gallery/internal/dice"
	"errors"
	"net/http"
)

type rollRequest struct {
	Dice     string `json:"dice"`
	Count    int    `json:"count"`
	Modifier int    `json:"modifier"`
}

func (g *Gallery) HandleRollDice(w http.ResponseWriter, r *http.Request) {
	var req rollRequest
	if !decodeBody(w, r, &req) {
		return
	}

	res, err := dice.Roll(req.Dice, req.Count, req.Modifier)
	if err != nil {
		switch {
		case errors.Is(err, dice.ErrInvalidDice),
			errors.Is(err, dice.ErrInvalidCount),
			errors.Is(err, dice.ErrInvalidMod):
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error rolling dice")
		}
		return
	}
	writeJSON(w, http.StatusOK, res)
}
