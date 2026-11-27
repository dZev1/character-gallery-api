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

func TestInventoryOwnership_AddAndRemove(t *testing.T) {
	ownerToken := registerAndLogin(t, newUsername("invOwner"))
	charID := createCharacter(t, ownerToken)

	otherToken := registerAndLogin(t, newUsername("invOther"))

	// Seeded item 1 exists in the pool.
	invURL := fmt.Sprintf("%s/characters/%d/inventory/1?quantity=2", apiBase(), charID)

	// Other user cannot add items to owner's character.
	doReq(t, http.MethodPost, invURL, bearer(otherToken), "", http.StatusNotFound)

	// Owner can add items.
	doReq(t, http.MethodPost, invURL, bearer(ownerToken), "", http.StatusOK)

	// Other user cannot remove items from owner's character.
	remURL := fmt.Sprintf("%s/characters/%d/inventory/1?quantity=1", apiBase(), charID)
	doReq(t, http.MethodDelete, remURL, bearer(otherToken), "", http.StatusNotFound)

	// Owner can remove items.
	doReq(t, http.MethodDelete, remURL, bearer(ownerToken), "", http.StatusOK)

	// Admin can also modify inventory.
	adminUser := newUsername("invAdmin")
	registerAndLogin(t, adminUser)
	promoteUserToAdmin(t, adminUser)
	adminToken := loginOnly(t, adminUser)

	doReq(t, http.MethodDelete, remURL, bearer(adminToken), "", http.StatusOK)
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

func TestAuth_RegisterEmailValidationAndEmailLogin(t *testing.T) {
	u := newUsername("mail")
	email := u + "@Example.COM"

	// Invalid email rejected.
	doReq(t, http.MethodPost, apiBase()+"/auth/register", nil, fmt.Sprintf(`{"username":%q,"email":"not-an-email","password":"supersecret1"}`, u), http.StatusBadRequest)

	// Password over 72 bytes rejected (prevents bcrypt ErrPasswordTooLong 500).
	longPass := strings.Repeat("a", 73)
	doReq(t, http.MethodPost, apiBase()+"/auth/register", nil, fmt.Sprintf(`{"username":%q,"email":%q,"password":%q}`, u, email, longPass), http.StatusBadRequest)

	// Valid registration normalizes email to lowercase.
	raw := doReq(t, http.MethodPost, apiBase()+"/auth/register", nil, fmt.Sprintf(`{"username":%q,"email":%q,"password":"supersecret1"}`, u, email), http.StatusCreated)
	var reg struct {
		Email string `json:"email"`
	}
	decodeJSON(t, raw, &reg)
	if reg.Email != strings.ToLower(email) {
		t.Fatalf("got email %q, want normalized %q", reg.Email, strings.ToLower(email))
	}

	// Duplicate email (different casing, different username) returns 409.
	doReq(t, http.MethodPost, apiBase()+"/auth/register", nil, fmt.Sprintf(`{"username":%q,"email":%q,"password":"supersecret1"}`, newUsername("mail2"), strings.ToLower(email)), http.StatusConflict)

	// Login using email field works.
	doReq(t, http.MethodPost, apiBase()+"/auth/login", nil, fmt.Sprintf(`{"email":%q,"password":"supersecret1"}`, email), http.StatusOK)
}
