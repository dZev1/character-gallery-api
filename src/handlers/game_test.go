package handlers

import (
	"fmt"
	"net/http"
	"testing"
)

func TestRollDice_Endpoint(t *testing.T) {
	token := registerAndLogin(t, newUsername("roller"))

	raw := doReq(t, http.MethodPost, apiBase()+"/roll", bearer(token), `{"dice":"d20","count":2,"modifier":3}`, http.StatusOK)
	var res struct {
		Rolls    []int  `json:"rolls"`
		Total    int    `json:"total"`
		Modifier int    `json:"modifier"`
		Dice     string `json:"dice"`
		Count    int    `json:"count"`
	}
	decodeJSON(t, raw, &res)
	if len(res.Rolls) != 2 || res.Dice != "d20" || res.Modifier != 3 {
		t.Fatalf("unexpected roll response: %+v", res)
	}

	doReq(t, http.MethodPost, apiBase()+"/roll", bearer(token), `{"dice":"d99","count":1}`, http.StatusBadRequest)
}

func TestEquipAndUseItem_Flow(t *testing.T) {
	ownerToken := registerAndLogin(t, newUsername("equipOwner"))
	otherToken := registerAndLogin(t, newUsername("equipOther"))
	charID := createCharacter(t, ownerToken)

	// Add Light Armor (item id 1, equippable) to inventory.
	doReq(t, http.MethodPost, fmt.Sprintf("%s/characters/%d/inventory/1?quantity=1", apiBase(), charID), bearer(ownerToken), "", http.StatusOK)

	// Other user cannot equip on owner's character.
	equipURL := fmt.Sprintf("%s/characters/%d/equip/1", apiBase(), charID)
	doReq(t, http.MethodPut, equipURL, bearer(otherToken), `{"equip":true}`, http.StatusNotFound)

	// Owner equips Light Armor.
	raw := doReq(t, http.MethodPut, equipURL, bearer(ownerToken), `{"equip":true}`, http.StatusOK)
	var eqRes struct {
		AC            int      `json:"ac"`
		EquippedItems []uint64 `json:"equipped_items"`
	}
	decodeJSON(t, raw, &eqRes)
	if len(eqRes.EquippedItems) != 1 || eqRes.EquippedItems[0] != 1 {
		t.Fatalf("expected item 1 equipped, got %+v", eqRes)
	}
	if eqRes.AC < 10 {
		t.Fatalf("expected AC >= 10, got %d", eqRes.AC)
	}
}

func TestEnemies_AdminWriteAndPublicRead(t *testing.T) {
	user := newUsername("enemyUser")
	token := registerAndLogin(t, user)

	// Any authenticated user can list seeded enemies.
	doReq(t, http.MethodGet, apiBase()+"/enemies", bearer(token), "", http.StatusOK)
	doReq(t, http.MethodGet, apiBase()+"/enemies?tier=1", bearer(token), "", http.StatusOK)

	enemyBody := `{"name":"Test Mimic","type":"elite","tier":2,"description":"A hungry treasure chest","hp_max":50,"ac":14,"damage":11,"xp_reward":90,"score_reward":140}`

	// Regular user cannot create enemies.
	doReq(t, http.MethodPost, apiBase()+"/enemies", bearer(token), enemyBody, http.StatusForbidden)

	// Admin can create, update, and delete.
	promoteUserToAdmin(t, user)
	adminToken := loginOnly(t, user)

	raw := doReq(t, http.MethodPost, apiBase()+"/enemies", bearer(adminToken), enemyBody, http.StatusCreated)
	var created struct {
		ID uint64 `json:"id"`
	}
	decodeJSON(t, raw, &created)
	if created.ID == 0 {
		t.Fatal("expected created enemy id")
	}

	doReq(t, http.MethodDelete, fmt.Sprintf("%s/enemies/%d", apiBase(), created.ID), bearer(adminToken), "", http.StatusOK)
}

func TestDailyRun_StartResumeAndAction(t *testing.T) {
	ownerToken := registerAndLogin(t, newUsername("dailyRunner"))
	otherToken := registerAndLogin(t, newUsername("dailyOther"))
	charID := createCharacter(t, ownerToken)

	// View Tier 3 daily dungeon (validCharacterJSON creates a Level 5 character -> Tier 3, 20 floors).
	rawView := doReq(t, http.MethodGet, apiBase()+"/daily/3", bearer(ownerToken), "", http.StatusOK)
	var view struct {
		Dungeon struct {
			Tier        uint8 `json:"tier"`
			TotalFloors uint8 `json:"total_floors"`
		} `json:"dungeon"`
	}
	decodeJSON(t, rawView, &view)
	if view.Dungeon.Tier != 3 || view.Dungeon.TotalFloors != 20 {
		t.Fatalf("expected tier 3 with 20 floors, got %+v", view.Dungeon)
	}

	// Other user cannot start a run with someone else's character.
	startBody := fmt.Sprintf(`{"character_id":%d}`, charID)
	doReq(t, http.MethodPost, apiBase()+"/daily/start", bearer(otherToken), startBody, http.StatusNotFound)

	// Owner starts the run.
	rawStart := doReq(t, http.MethodPost, apiBase()+"/daily/start", bearer(ownerToken), startBody, http.StatusOK)
	var state struct {
		Run struct {
			ID     uint64 `json:"id"`
			Status string `json:"status"`
			Tier   uint8  `json:"tier"`
		} `json:"run"`
		CurrentEnemy *struct {
			Name string `json:"name"`
		} `json:"current_enemy"`
	}
	decodeJSON(t, rawStart, &state)
	if state.Run.Status != "in_combat" || state.Run.Tier != 3 || state.CurrentEnemy == nil {
		t.Fatalf("unexpected initial run state: %+v", state)
	}

	// Calling start again resumes the exact same run (pause & resume).
	rawResume := doReq(t, http.MethodPost, apiBase()+"/daily/start", bearer(ownerToken), startBody, http.StatusOK)
	var resumed struct {
		Run struct {
			ID uint64 `json:"id"`
		} `json:"run"`
	}
	decodeJSON(t, rawResume, &resumed)
	if resumed.Run.ID != state.Run.ID {
		t.Fatalf("expected resumed run ID %d, got %d", state.Run.ID, resumed.Run.ID)
	}

	// Perform one combat turn.
	actionBody := fmt.Sprintf(`{"character_id":%d,"action":"attack"}`, charID)
	doReq(t, http.MethodPost, apiBase()+"/daily/action", bearer(ownerToken), actionBody, http.StatusOK)

	// Check leaderboard endpoint.
	doReq(t, http.MethodGet, apiBase()+"/daily/3/leaderboard", bearer(ownerToken), "", http.StatusOK)
}

func TestCreateCharacter_MaxSixPerOwner(t *testing.T) {
	token := registerAndLogin(t, newUsername("limit6"))

	var firstID uint64
	for i := 0; i < 6; i++ {
		id := createCharacter(t, token)
		if i == 0 {
			firstID = id
		}
	}

	// 7th character is rejected with 409 Conflict.
	doReq(t, http.MethodPost, apiBase()+"/characters", bearer(token), validCharacterJSON(), http.StatusConflict)

	// Deleting one frees up a slot.
	doReq(t, http.MethodDelete, fmt.Sprintf("%s/characters/%d", apiBase(), firstID), bearer(token), "", http.StatusOK)
	createCharacter(t, token)
}
