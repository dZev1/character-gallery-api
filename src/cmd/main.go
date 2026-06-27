package main

import (
	"context"
	"dZev1/character-gallery/internal/cache"
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/inventory"
	"dZev1/character-gallery/internal/items"
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
		log.Fatal(err)
	}

	gallery := handlers.NewGallery(charService, itemService, invService)

	rl := middleware.NewRateLimiter(rdb, 1*time.Minute)
	rl.SetLimit("POST", "/api/v1/characters", 10)
	rl.SetLimit("DELETE", "/api/v1/characters/{characterId}", 30)
	rl.SetLimit("POST", "/api/v1/items", 10)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /openapi.yaml", handlers.OpenAPIHandler)
	mux.HandleFunc("GET /docs", handlers.DocsHandler)

	mux.HandleFunc("POST /api/v1/characters", gallery.HandleCreateCharacter)
	mux.HandleFunc("GET /api/v1/characters", gallery.HandleGetAllCharacters)
	mux.HandleFunc("GET /api/v1/characters/{characterId}", gallery.HandleGetCharacter)
	mux.HandleFunc("PUT /api/v1/characters/{characterId}", gallery.HandleUpdateCharacter)
	mux.HandleFunc("DELETE /api/v1/characters/{characterId}", gallery.HandleDeleteCharacter)

	mux.HandleFunc("POST /api/v1/items", gallery.HandleCreateItem)
	mux.HandleFunc("GET /api/v1/items", gallery.HandleGetAllItems)
	mux.HandleFunc("GET /api/v1/items/{itemId}", gallery.HandleGetItem)

	mux.HandleFunc("GET /api/v1/characters/{characterId}/inventory", gallery.HandleGetCharacterInventory)
	mux.HandleFunc("POST /api/v1/characters/{characterId}/inventory/{itemId}", gallery.HandleAddItemToCharacter)
	mux.HandleFunc("DELETE /api/v1/characters/{characterId}/inventory/{itemId}", gallery.HandleRemoveItemFromCharacter)

	handler := rl.Limit(mux)
	handler = middleware.EnableCors(handler)

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
