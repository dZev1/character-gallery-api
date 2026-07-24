package main

import (
	"context"
	"dZev1/character-gallery/internal/auth"
	"dZev1/character-gallery/internal/cache"
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/inventory"
	"dZev1/character-gallery/internal/items"
	"dZev1/character-gallery/internal/metrics"
	"dZev1/character-gallery/internal/middleware"
	"dZev1/character-gallery/internal/postgres"
	redislib "dZev1/character-gallery/internal/redis"

	"log"
	"net/http"
	"os"
	"time"

	"dZev1/character-gallery/handlers"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	pgConnStr := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_URL")

	if pgConnStr == "" || redisURL == "" {
		log.Fatal("DATABASE_URL or REDIS_URL environment variables not set")
	}

	pool, err := pgxpool.New(context.Background(), pgConnStr)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	redisOpts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal(err)
	}
	rdb, err := redislib.NewRedisClient(redisOpts)
	if err != nil {
		log.Fatal(err)
	}
	defer rdb.Close()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable not set")
	}

	authRepo := postgres.NewAuthRepo(pool)
	authSvc := auth.NewService(authRepo, pool, jwtSecret)
	authHandler := handlers.NewAuthHandler(authSvc)

	charCache := cache.NewRedisCache(rdb, 5*time.Minute, "character")
	itemCache := cache.NewRedisCache(rdb, 5*time.Minute, "item")
	invCache := cache.NewRedisCache(rdb, 5*time.Minute, "inventory")

	charRepo := postgres.NewCharacterRepo(pool)
	itemRepo := postgres.NewItemRepo(pool)
	invRepo := postgres.NewInventoryRepo(pool)

	charService := characters.NewService(charRepo, pool, charCache)
	itemService := items.NewService(itemRepo, pool, itemCache)
	invService := inventory.NewService(invRepo, pool, invCache)

	if err = itemService.SeedItems(context.Background()); err != nil {
		log.Print(err)
	}

	gallery := handlers.NewGallery(charService, itemService, invService)

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "/api/v1"
	}

	rl := middleware.NewRateLimiter(rdb, 1*time.Minute)
	rl.SetLimit("POST", baseURL+"/characters", 10)
	rl.SetLimit("DELETE", baseURL+"/characters/{characterId}", 30)
	rl.SetLimit("POST", baseURL+"/items", 10)
	rl.SetLimit("POST", baseURL+"/auth/register", 5)
	rl.SetLimit("POST", baseURL+"/auth/login", 10)

	mux := http.NewServeMux()

	mux.HandleFunc("POST "+baseURL+"/auth/register", authHandler.HandleRegister)
	mux.HandleFunc("POST "+baseURL+"/auth/login", authHandler.HandleLogin)
	mux.HandleFunc("GET /openapi.yaml", handlers.OpenAPIHandler)
	mux.HandleFunc("GET /docs", handlers.DocsHandler)

	mux.HandleFunc("POST "+baseURL+"/characters", gallery.HandleCreateCharacter)
	mux.HandleFunc("GET "+baseURL+"/characters", gallery.HandleGetAllCharacters)
	mux.HandleFunc("GET "+baseURL+"/characters/{characterId}", gallery.HandleGetCharacter)
	mux.HandleFunc("PUT "+baseURL+"/characters/{characterId}", gallery.HandleUpdateCharacter)
	mux.HandleFunc("DELETE "+baseURL+"/characters/{characterId}", gallery.HandleDeleteCharacter)

	mux.HandleFunc("POST "+baseURL+"/items", gallery.HandleCreateItem)
	mux.HandleFunc("GET "+baseURL+"/items", gallery.HandleGetAllItems)
	mux.HandleFunc("GET "+baseURL+"/items/{itemId}", gallery.HandleGetItem)

	mux.HandleFunc("GET "+baseURL+"/characters/{characterId}/inventory", gallery.HandleGetCharacterInventory)
	mux.HandleFunc("POST "+baseURL+"/characters/{characterId}/inventory/{itemId}", gallery.HandleAddItemToCharacter)
	mux.HandleFunc("DELETE "+baseURL+"/characters/{characterId}/inventory/{itemId}", gallery.HandleRemoveItemFromCharacter)

	mux.Handle("GET /metrics", metrics.Handler())

	jwtMw := middleware.AuthMiddleware(jwtSecret,
		"POST "+baseURL+"/auth/register",
		"POST "+baseURL+"/auth/login",
		"GET /openapi.yaml",
		"GET /docs",
		"GET /metrics",
	)
	handler := middleware.MetricsMiddleware(jwtMw(rl.Limit(mux)))
	handler = middleware.EnableCors(handler)

	addr := os.Getenv("PORT")
	if addr == "" {
		addr = ":8080"
	}

	log.Println("Server listening on " + addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
