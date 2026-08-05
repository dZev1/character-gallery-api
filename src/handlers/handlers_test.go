package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"dZev1/character-gallery/internal/auth"
	"dZev1/character-gallery/internal/cache"
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/inventory"
	"dZev1/character-gallery/internal/items"
	"dZev1/character-gallery/internal/middleware"
	"dZev1/character-gallery/internal/postgres"
	redislib "dZev1/character-gallery/internal/redis"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const testJWTSecret = "test-secret-that-is-long-enough-for-hmac-123456789"

var (
	testServer    *httptest.Server
	testPool      *pgxpool.Pool
	testAuthSvc   *auth.Service
	testItemSvc   *items.Service
	testCharSvc   *characters.Service
	userCounter   int
)

func TestMain(m *testing.M) {
	dsn := os.Getenv("POSTGRES_TEST_DATABASE_URL")
	redisURL := os.Getenv("REDIS_TEST_URL")
	if dsn == "" || redisURL == "" {
		fmt.Println("SKIP: POSTGRES_TEST_DATABASE_URL / REDIS_TEST_URL not set")
		os.Exit(0)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect: %v\n", err)
		os.Exit(1)
	}

	if _, err := pool.Exec(ctx, `
		DROP TABLE IF EXISTS customizations CASCADE;
		DROP TABLE IF EXISTS stats CASCADE;
		DROP TABLE IF EXISTS inventory CASCADE;
		DROP TABLE IF EXISTS items CASCADE;
		DROP TABLE IF EXISTS characters CASCADE;
		DROP TABLE IF EXISTS users CASCADE;
	`); err != nil {
		fmt.Fprintf(os.Stderr, "failed to drop tables: %v\n", err)
		os.Exit(1)
	}

	schema, err := os.ReadFile("../database/schema/schema.sql")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read schema: %v\n", err)
		os.Exit(1)
	}
	if _, err := pool.Exec(ctx, string(schema)); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run schema: %v\n", err)
		os.Exit(1)
	}
	testPool = pool

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse redis url: %v\n", err)
		os.Exit(1)
	}
	opts.DB = 1
	rdb, err := redislib.NewRedisClient(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create redis client: %v\n", err)
		os.Exit(1)
	}
	if err := rdb.FlushDB(ctx).Err(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to flush redis: %v\n", err)
		os.Exit(1)
	}

	authRepo := postgres.NewAuthRepo(pool)
	testAuthSvc = auth.NewService(authRepo, pool, testJWTSecret)
	authHandler := NewAuthHandler(testAuthSvc)

	charCache := cache.NewRedisCache(rdb, 5*time.Minute, "character")
	itemCache := cache.NewRedisCache(rdb, 5*time.Minute, "item")
	invCache := cache.NewRedisCache(rdb, 5*time.Minute, "inventory")

	charRepo := postgres.NewCharacterRepo(pool)
	itemRepo := postgres.NewItemRepo(pool)
	invRepo := postgres.NewInventoryRepo(pool)

	testCharSvc = characters.NewService(charRepo, pool, charCache)
	testItemSvc = items.NewService(itemRepo, pool, itemCache)
	invService := inventory.NewService(invRepo, pool, invCache)

	if err := testItemSvc.SeedItems(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to seed items: %v\n", err)
		os.Exit(1)
	}

	gallery := NewGallery(testCharSvc, testItemSvc, invService)

	base := "/api/v2"
	mux := http.NewServeMux()
	mux.HandleFunc("POST "+base+"/auth/register", authHandler.HandleRegister)
	mux.HandleFunc("POST "+base+"/auth/login", authHandler.HandleLogin)
	mux.HandleFunc("GET /openapi.yaml", OpenAPIHandler)
	mux.HandleFunc("POST "+base+"/characters", gallery.HandleCreateCharacter)
	mux.HandleFunc("GET "+base+"/characters", gallery.HandleGetAllCharacters)
	mux.HandleFunc("GET "+base+"/characters/{characterId}", gallery.HandleGetCharacter)
	mux.HandleFunc("PUT "+base+"/characters/{characterId}", gallery.HandleUpdateCharacter)
	mux.HandleFunc("DELETE "+base+"/characters/{characterId}", gallery.HandleDeleteCharacter)
	mux.HandleFunc("POST "+base+"/items", gallery.HandleCreateItem)
	mux.HandleFunc("GET "+base+"/items", gallery.HandleGetAllItems)
	mux.HandleFunc("GET "+base+"/items/{itemId}", gallery.HandleGetItem)
	mux.HandleFunc("GET "+base+"/characters/{characterId}/inventory", gallery.HandleGetCharacterInventory)
	mux.HandleFunc("POST "+base+"/characters/{characterId}/inventory/{itemId}", gallery.HandleAddItemToCharacter)
	mux.HandleFunc("DELETE "+base+"/characters/{characterId}/inventory/{itemId}", gallery.HandleRemoveItemFromCharacter)

	_ = os.Setenv("CORS_ALLOW_ORIGIN", "http://192.168.1.32:3000")

	jwtMw := middleware.AuthMiddleware(testJWTSecret,
		"POST "+base+"/auth/register",
		"POST "+base+"/auth/login",
		"GET /openapi.yaml",
	)
	handler := middleware.EnableCors(jwtMw(mux))
	testServer = httptest.NewServer(handler)

	code := m.Run()

	testServer.Close()
	pool.Close()
	os.Exit(code)
}

// --- helpers ---

func apiBase() string { return testServer.URL + "/api/v2" }

func newUsername(prefix string) string {
	userCounter++
	return fmt.Sprintf("%s_%d_%d", prefix, userCounter, time.Now().UnixNano())
}

// registerAndLogin creates a user through the real HTTP endpoints and returns
// the JWT for that user.
func registerAndLogin(t *testing.T, username string) string {
	t.Helper()

	regBody := fmt.Sprintf(`{"username":%q,"password":"supersecret1"}`, username)
	raw := doReq(t, http.MethodPost, apiBase()+"/auth/register", nil, regBody, http.StatusCreated)

	var created map[string]any
	decodeJSON(t, raw, &created)
	if _, ok := created["id"]; !ok {
		t.Fatalf("register response missing id: %v", created)
	}
	return loginOnly(t, username)
}

// loginOnly assumes the user already exists and returns a fresh JWT (as used
// after changing a user's role, since the role is baked into the token).
func loginOnly(t *testing.T, username string) string {
	t.Helper()

	loginBody := fmt.Sprintf(`{"username":%q,"password":"supersecret1"}`, username)
	raw := doReq(t, http.MethodPost, apiBase()+"/auth/login", nil, loginBody, http.StatusOK)

	var login struct {
		Token string `json:"token"`
	}
	decodeJSON(t, raw, &login)
	if login.Token == "" {
		t.Fatal("login returned empty token")
	}
	return login.Token
}

// promoteUserToAdmin flips the role in the DB directly (there is no endpoint).
func promoteUserToAdmin(t *testing.T, username string) {
	t.Helper()
	ctx := context.Background()
	tag, err := testPool.Exec(ctx, `UPDATE users SET role = 'admin' WHERE username = $1`, username)
	if err != nil {
		t.Fatal(err)
	}
	if tag.RowsAffected() != 1 {
		t.Fatalf("expected 1 user promoted, got %d", tag.RowsAffected())
	}
}

func bearer(token string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + token}
}

// doReq performs an HTTP request, asserts the status code, and returns the
// response body. The body is read eagerly so callers don't manage closures.
func doReq(t *testing.T, method, url string, headers map[string]string, body string, wantStatus int) []byte {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = bytes.NewReader([]byte(body))
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatal(err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := testServer.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != wantStatus {
		t.Fatalf("%s %s: got status %d, want %d (body: %s)", method, url, resp.StatusCode, wantStatus, raw)
	}
	return raw
}

func decodeJSON(t *testing.T, raw []byte, dest any) {
	t.Helper()
	if err := json.Unmarshal(raw, dest); err != nil {
		t.Fatalf("failed to decode response %q: %v", raw, err)
	}
}

func validCharacterJSON() string {
	return `{
		"name": "Aragorn",
		"body_type": "type_a",
		"species": "human",
		"class": "ranger",
		"level": 5,
		"xp": 120,
		"hp_max": 20,
		"hp_current": 18,
		"stats": {"strength":15,"dexterity":14,"constitution":13,"intelligence":12,"wisdom":11,"charisma":10},
		"customization": {"hair":5,"face":3,"shirt":12,"pants":8,"shoes":2}
	}`
}

func createCharacter(t *testing.T, token string) uint64 {
	t.Helper()
	raw := doReq(t, http.MethodPost, apiBase()+"/characters", bearer(token), validCharacterJSON(), http.StatusCreated)

	var created struct {
		ID uint64 `json:"id"`
	}
	decodeJSON(t, raw, &created)
	if created.ID == 0 {
		t.Fatal("created character has no id")
	}
	return created.ID
}
