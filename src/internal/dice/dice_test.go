package dice

import (
	"errors"
	"testing"
)

func TestRoll_ValidDice(t *testing.T) {
	res, err := Roll("d20", 3, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Rolls) != 3 {
		t.Fatalf("got %d rolls, want 3", len(res.Rolls))
	}
	sum := 0
	for _, r := range res.Rolls {
		if r < 1 || r > 20 {
			t.Fatalf("roll %d out of range 1..20", r)
		}
		sum += r
	}
	if res.Total != sum+5 {
		t.Fatalf("got total %d, want %d", res.Total, sum+5)
	}
}

func TestRoll_InvalidArgs(t *testing.T) {
	if _, err := Roll("d7", 1, 0); !errors.Is(err, ErrInvalidDice) {
		t.Fatalf("expected ErrInvalidDice, got %v", err)
	}
	if _, err := Roll("d20", 0, 0); !errors.Is(err, ErrInvalidCount) {
		t.Fatalf("expected ErrInvalidCount, got %v", err)
	}
	if _, err := Roll("d20", 1, 200); !errors.Is(err, ErrInvalidMod) {
		t.Fatalf("expected ErrInvalidMod, got %v", err)
	}
}
