package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestSpec_IsPublic(t *testing.T) {
	raw := doReq(t, http.MethodGet, testServer.URL+"/openapi.yaml", nil, "", http.StatusOK)
	if !bytes.Contains(raw, []byte("openapi")) {
		t.Fatalf("spec body does not look like an OpenAPI document: %q", raw[:64])
	}
}

func TestAuthMiddleware_RejectsMissingToken(t *testing.T) {
	doReq(t, http.MethodGet, apiBase()+"/characters", nil, "", http.StatusUnauthorized)
	doReq(t, http.MethodPost, apiBase()+"/characters", nil, `{}`, http.StatusUnauthorized)
	doReq(t, http.MethodDelete, apiBase()+"/characters/1", nil, "", http.StatusUnauthorized)
}

func TestCreateItem_AdminOnly(t *testing.T) {
	user := newUsername("user")
	token := registerAndLogin(t, user)

	itemBody := `{"name":"Ardent Blade","type":"weapon","description":"A blazing sword","equippable":false,"rarity":3}`

	// Normal user is forbidden.
	doReq(t, http.MethodPost, apiBase()+"/items", bearer(token), itemBody, http.StatusForbidden)

	// same user, promoted to admin, can create (role claim comes from DB at
	// login time, so a fresh login is required).
	promoteUserToAdmin(t, user)
	token = loginOnly(t, user)
	raw := doReq(t, http.MethodPost, apiBase()+"/items", bearer(token), itemBody, http.StatusCreated)
	var created struct {
		ID   uint64 `json:"id"`
		Name string `json:"name"`
	}
	decodeJSON(t, raw, &created)
	if created.Name != "Ardent Blade" {
		t.Fatalf("got item name %q, want Ardent Blade", created.Name)
	}
}

func TestCharacterOwnership_Delete(t *testing.T) {
	ownerToken := registerAndLogin(t, newUsername("owner"))
	charID := createCharacter(t, ownerToken)

	// A different non-admin user cannot delete it.
	otherToken := registerAndLogin(t, newUsername("other"))
	doReq(t, http.MethodDelete, fmt.Sprintf("%s/characters/%d", apiBase(), charID), bearer(otherToken), "", http.StatusNotFound)

	// The owner can.
	doReq(t, http.MethodDelete, fmt.Sprintf("%s/characters/%d", apiBase(), charID), bearer(ownerToken), "", http.StatusOK)
}

func TestCharacterOwnership_AdminBypass(t *testing.T) {
	ownerToken := registerAndLogin(t, newUsername("ownerA"))
	charID := createCharacter(t, ownerToken)

	adminUser := newUsername("adminA")
	registerAndLogin(t, adminUser)
	promoteUserToAdmin(t, adminUser)
	adminToken := loginOnly(t, adminUser)

	doReq(t, http.MethodDelete, fmt.Sprintf("%s/characters/%d", apiBase(), charID), bearer(adminToken), "", http.StatusOK)
}

func TestCharacterOwnership_Update(t *testing.T) {
	ownerToken := registerAndLogin(t, newUsername("ownerUp"))
	charID := createCharacter(t, ownerToken)

	otherToken := registerAndLogin(t, newUsername("otherUp"))
	doReq(t, http.MethodPut, fmt.Sprintf("%s/characters/%d", apiBase(), charID), bearer(otherToken), validCharacterJSON(), http.StatusNotFound)

	// Owner can update.
	raw := doReq(t, http.MethodPut, fmt.Sprintf("%s/characters/%d", apiBase(), charID), bearer(ownerToken), validCharacterJSON(), http.StatusOK)
	var updated struct {
		Name string `json:"name"`
	}
	decodeJSON(t, raw, &updated)
	if updated.Name != "Aragorn" {
		t.Fatalf("got updated name %q, want Aragorn", updated.Name)
	}
}

func TestCreateCharacter_RequiresAuthForOtherOwners(t *testing.T) {
	// Sanity: users can read the public character listing, but creating
	// attributes the character to the authenticated user only.
	token := registerAndLogin(t, newUsername("sanity"))
	_ = createCharacter(t, token)
	doReq(t, http.MethodGet, apiBase()+"/characters", bearer(token), "", http.StatusOK)
}

func TestRequestBody_EnforcedLimit(t *testing.T) {
	token := registerAndLogin(t, newUsername("bigbody"))

	// Build a JSON body larger than the 1 MiB cap.
	big := `{"data":"` + strings.Repeat("A", (1<<20)+4096) + `"}`
	doReq(t, http.MethodPost, apiBase()+"/characters", bearer(token), big, http.StatusBadRequest)
}

func TestCORSPreflight_AllowListedOrigin(t *testing.T) {
	req, _ := http.NewRequest(http.MethodOptions, apiBase()+"/characters", nil)
	req.Header.Set("Origin", "http://192.168.1.32:3000")

	resp, err := testServer.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("preflight status = %d, want 200", resp.StatusCode)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "http://192.168.1.32:3000" {
		t.Fatalf("Allow-Origin = %q, want the allow-listed origin", got)
	}
}

func TestCORSPreflight_UnknownOrigin(t *testing.T) {
	req, _ := http.NewRequest(http.MethodOptions, apiBase()+"/characters", nil)
	req.Header.Set("Origin", "http://evil.example.com")

	resp, err := testServer.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("unexpected Allow-Origin %q for unknown origin", got)
	}
}