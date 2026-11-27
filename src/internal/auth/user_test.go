package auth

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUserResponse_MarshalsIDAsUUIDString(t *testing.T) {
	rawUUID := uuid.MustParse("01a20628-a185-7026-b7f8-d0871156becb")
	resp := UserResponse{
		ID:        UserID(rawUUID),
		Username:  "dezeta1",
		Role:      "user",
		CreatedAt: time.Date(2026, 11, 23, 2, 7, 27, 0, time.UTC),
	}

	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}

	gotID, ok := decoded["id"].(string)
	if !ok {
		t.Fatalf("expected id to serialize as JSON string, got %T (%v)", decoded["id"], decoded["id"])
	}
	if gotID != rawUUID.String() {
		t.Fatalf("got id %q, want %q", gotID, rawUUID.String())
	}

	var roundTrip UserResponse
	if err := json.Unmarshal(b, &roundTrip); err != nil {
		t.Fatal(err)
	}
	if uuid.UUID(roundTrip.ID) != rawUUID {
		t.Fatalf("round-trip id %s != %s", roundTrip.ID, rawUUID)
	}
}
