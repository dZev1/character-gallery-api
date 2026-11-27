package characters

import "testing"

func TestRollStatPool(t *testing.T) {
	rolls, err := RollStatPool()
	if err != nil {
		t.Fatalf("RollStatPool failed: %v", err)
	}
	if len(rolls) != 6 {
		t.Fatalf("expected 6 rolls, got %d", len(rolls))
	}
	for i, v := range rolls {
		if v < 3 || v > 18 {
			t.Fatalf("roll[%d]=%d out of 4d6 drop lowest range 3..18", i, v)
		}
		if i > 0 && rolls[i-1] < v {
			t.Fatalf("expected descending order, got %v", rolls)
		}
	}
}

func TestMatchesRollPool(t *testing.T) {
	pool := []int{16, 14, 13, 12, 10, 8}
	valid := &Stats{Strength: 8, Dexterity: 16, Constitution: 12, Intelligence: 14, Wisdom: 10, Charisma: 13}
	if !MatchesRollPool(valid, pool) {
		t.Fatalf("expected valid permutation to match")
	}

	tampered := &Stats{Strength: 18, Dexterity: 16, Constitution: 12, Intelligence: 14, Wisdom: 10, Charisma: 13}
	if MatchesRollPool(tampered, pool) {
		t.Fatalf("expected tampered stats not to match")
	}
}

func TestApplyClassModifiers(t *testing.T) {
	base := &Stats{Strength: 10, Dexterity: 10, Constitution: 10, Intelligence: 10, Wisdom: 10, Charisma: 10}
	out := ApplyClassModifiers(Wizard, base)
	if out.Intelligence != 13 || out.Strength != 9 || out.Constitution != 9 {
		t.Fatalf("unexpected wizard stats: %+v", out)
	}
}
