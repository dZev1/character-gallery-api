package uuidv7

import (
	"sort"
	"testing"

	"github.com/google/uuid"
)

func TestNew_IsUUIDV7(t *testing.T) {
	id, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if id.Version() != 7 {
		t.Fatalf("expected version 7, got %d", id.Version())
	}
	if id.Variant() != uuid.RFC4122 {
		t.Fatalf("expected RFC4122 variant, got %d", id.Variant())
	}
}

func TestNew_Unique(t *testing.T) {
	const n = 10000
	seen := make(map[uuid.UUID]struct{}, n)
	for i := 0; i < n; i++ {
		id, err := New()
		if err != nil {
			t.Fatal(err)
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("duplicate uuid generated: %s", id)
		}
		seen[id] = struct{}{}
	}
}

func TestNew_MonotonicOrdering(t *testing.T) {
	const n = 1000
	ids := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		id, _ := New()
		ids = append(ids, id)
	}

	sorted := make([]uuid.UUID, len(ids))
	copy(sorted, ids)
	sort.Slice(sorted, func(a, b int) bool {
		for i := 0; i < 16; i++ {
			if sorted[a][i] != sorted[b][i] {
				return sorted[a][i] < sorted[b][i]
			}
		}
		return false
	})

	for i := 0; i < len(ids); i++ {
		if ids[i] != sorted[i] {
			t.Fatalf("uuid %d out of order (not time-ordered)", i)
		}
	}
}
