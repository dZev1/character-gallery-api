package dice

import (
	"crypto/rand"
	"errors"
	"math/big"
)

var (
	ErrInvalidDice  = errors.New("dice must be one of d4, d6, d8, d10, d12, d20, d100")
	ErrInvalidCount = errors.New("count must be between 1 and 100")
	ErrInvalidMod   = errors.New("modifier must be between -100 and 100")
)

var validSides = map[string]int{
	"d4":   4,
	"d6":   6,
	"d8":   8,
	"d10":  10,
	"d12":  12,
	"d20":  20,
	"d100": 100,
}

type Result struct {
	Rolls    []int  `json:"rolls"`
	Total    int    `json:"total"`
	Modifier int    `json:"modifier"`
	Dice     string `json:"dice"`
	Count    int    `json:"count"`
}

// Roll rolls `count` dice of the given type (e.g. "d20") using crypto/rand
// and adds `modifier` to the sum.
func Roll(diceType string, count, modifier int) (*Result, error) {
	sides, ok := validSides[diceType]
	if !ok {
		return nil, ErrInvalidDice
	}
	if count < 1 || count > 100 {
		return nil, ErrInvalidCount
	}
	if modifier < -100 || modifier > 100 {
		return nil, ErrInvalidMod
	}

	rolls := make([]int, count)
	sum := 0
	for i := 0; i < count; i++ {
		v, err := rollOne(sides)
		if err != nil {
			return nil, err
		}
		rolls[i] = v
		sum += v
	}

	return &Result{
		Rolls:    rolls,
		Total:    sum + modifier,
		Modifier: modifier,
		Dice:     diceType,
		Count:    count,
	}, nil
}

// RollD20 is a helper for standard d20 combat/skill checks.
func RollD20() (int, error) {
	return rollOne(20)
}

// RollN rolls a single die with the specified number of sides (1..sides).
func RollN(sides int) (int, error) {
	return rollOne(sides)
}

func rollOne(sides int) (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(sides)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()) + 1, nil
}
