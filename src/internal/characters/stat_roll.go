package characters

import (
	"crypto/rand"
	"errors"
	"math/big"
	"sort"
	"time"
)

const (
	WeeklyCooldown         = 7 * 24 * time.Hour
	MaxRerollsPerCharacter = 3
)

var (
	ErrNoRerollsLeft            = errors.New("no stat rerolls remaining for this character")
	ErrDeleteRequiresFullRoster = errors.New("you can only delete a character once all 6 character slots are full")
	ErrDeleteCooldown           = errors.New("you can only delete one character per week")
	ErrNoStatRoll               = errors.New("no active stat roll found; fetch stat roll first")
	ErrInvalidStatAllocation    = errors.New("stats must match the active server stat roll")
)

type StatModifiers struct {
	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Constitution int `json:"constitution"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Charisma     int `json:"charisma"`
}

var ClassModifiers = map[Class]StatModifiers{
	Fighter:   {Strength: 2, Constitution: 1, Intelligence: -1},
	Barbarian: {Strength: 3, Constitution: 1, Intelligence: -2},
	Paladin:   {Strength: 2, Charisma: 1, Dexterity: -1},
	Ranger:    {Dexterity: 2, Wisdom: 1, Charisma: -1},
	Rogue:     {Dexterity: 3, Strength: -1},
	Monk:      {Dexterity: 2, Wisdom: 1, Charisma: -1},
	Cleric:    {Wisdom: 2, Constitution: 1, Dexterity: -1},
	Druid:     {Wisdom: 2, Intelligence: 1, Strength: -1},
	Bard:      {Charisma: 2, Dexterity: 1, Strength: -1},
	Wizard:    {Intelligence: 3, Strength: -1, Constitution: -1},
	Sorcerer:  {Charisma: 2, Constitution: 1, Strength: -1},
	Warlock:   {Charisma: 2, Intelligence: 1, Wisdom: -1},
}

type UserStatRoll struct {
	Rolls                  []int
	RerollsUsed            int
	LastCharacterDeletedAt *time.Time
}

type StatRollView struct {
	Rolls            []int                   `json:"rolls"`
	RerollsRemaining int                     `json:"rerolls_remaining"`
	MaxRerolls       int                     `json:"max_rerolls"`
	CanReroll        bool                    `json:"can_reroll"`
	ClassModifiers   map[Class]StatModifiers `json:"class_modifiers"`
}

// RollStatPool rolls 6 ability scores using 4d6 drop lowest (cryptographically random).
func RollStatPool() ([]int, error) {
	pool := make([]int, 6)
	six := big.NewInt(6)

	for i := range pool {
		dice := make([]int, 4)
		for j := range dice {
			n, err := rand.Int(rand.Reader, six)
			if err != nil {
				return nil, err
			}
			dice[j] = int(n.Int64()) + 1
		}
		sort.Ints(dice)
		pool[i] = dice[1] + dice[2] + dice[3]
	}

	sort.Sort(sort.Reverse(sort.IntSlice(pool)))
	return pool, nil
}

// MatchesRollPool checks that the 6 assigned stats are an exact permutation of rolls.
func MatchesRollPool(s *Stats, rolls []int) bool {
	if s == nil || len(rolls) != 6 {
		return false
	}
	assigned := []int{
		int(s.Strength),
		int(s.Dexterity),
		int(s.Constitution),
		int(s.Intelligence),
		int(s.Wisdom),
		int(s.Charisma),
	}
	expected := make([]int, 6)
	copy(expected, rolls)

	sort.Ints(assigned)
	sort.Ints(expected)

	for i := 0; i < 6; i++ {
		if assigned[i] != expected[i] {
			return false
		}
	}
	return true
}

func clampStat(val int) uint8 {
	if val < 1 {
		return 1
	}
	if val > 99 {
		return 99
	}
	return uint8(val)
}

// ApplyClassModifiers returns a new Stats with the class buffs/debuffs applied.
func ApplyClassModifiers(class Class, base *Stats) *Stats {
	mod := ClassModifiers[class]
	return &Stats{
		ID:           base.ID,
		Strength:     clampStat(int(base.Strength) + mod.Strength),
		Dexterity:    clampStat(int(base.Dexterity) + mod.Dexterity),
		Constitution: clampStat(int(base.Constitution) + mod.Constitution),
		Intelligence: clampStat(int(base.Intelligence) + mod.Intelligence),
		Wisdom:       clampStat(int(base.Wisdom) + mod.Wisdom),
		Charisma:     clampStat(int(base.Charisma) + mod.Charisma),
	}
}
