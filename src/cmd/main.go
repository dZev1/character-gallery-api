package main

import (
	"context"
	"dZev1/character-gallery/database"
	"dZev1/character-gallery/internal/auth"
	"dZev1/character-gallery/internal/cache"
	"dZev1/character-gallery/internal/characters"
	"dZev1/character-gallery/internal/daily_run"
	"dZev1/character-gallery/internal/enemies"
	"dZev1/character-gallery/internal/inventory"
	"dZev1/character-gallery/internal/items"
	"dZev1/character-gallery/internal/metrics"
	"dZev1/character-gallery/internal/middleware"
	"dZev1/character-gallery/internal/postgres"
	redislib "dZev1/character-gallery/internal/redis"

	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
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

	if err = database.ApplySchema(context.Background(), pool); err != nil {
		log.Fatal(err)
	}
	metrics.RegisterDBPoolMetrics(pool)

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
	enemyCache := cache.NewRedisCache(rdb, 5*time.Minute, "enemy")

	charRepo := postgres.NewCharacterRepo(pool)
	itemRepo := postgres.NewItemRepo(pool)
	invRepo := postgres.NewInventoryRepo(pool)
	enemyRepo := postgres.NewEnemyRepo(pool)
	dailyRepo := postgres.NewDailyRunRepo(pool)

	charService := characters.NewService(charRepo, pool, charCache)
	itemService := items.NewService(itemRepo, pool, itemCache)
	invService := inventory.NewService(invRepo, pool, invCache).WithCharacterStore(charRepo, charCache)
	enemyService := enemies.NewService(enemyRepo, pool, enemyCache)
	dailyService := daily_run.NewService(dailyRepo, charRepo, enemyRepo, invRepo, itemRepo, pool, charCache, invCache)

	if err = itemService.SeedItems(context.Background()); err != nil {
		log.Print(err)
	}
	if err = enemyService.SeedEnemies(context.Background()); err != nil {
		log.Print(err)
	}

	gallery := handlers.NewGallery(charService, itemService, invService).WithGameServices(enemyService, dailyService)

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "/api/v1"
	}

	rl := middleware.NewRateLimiter(rdb, 1*time.Minute)
	rl.SetLimit("POST", baseURL+"/characters", 10)
	rl.SetLimit("DELETE", baseURL+"/characters/{characterId}", 30)
	rl.SetLimit("POST", baseURL+"/items", 10)
	rl.SetLimit("POST", baseURL+"/enemies", 10)
	rl.SetLimit("POST", baseURL+"/uploads", 20)
	rl.SetLimit("POST", baseURL+"/auth/register", 5)
	rl.SetLimit("POST", baseURL+"/auth/login", 10)

	enableDocs := strings.EqualFold(os.Getenv("ENABLE_DOCS"), "true")
	metricsAddr := strings.TrimSpace(os.Getenv("METRICS_ADDR"))
	uploadHandler := handlers.NewUploadHandler(os.Getenv("UPLOADS_DIR"))

	mux := http.NewServeMux()

	mux.HandleFunc("POST "+baseURL+"/auth/register", authHandler.HandleRegister)
	mux.HandleFunc("POST "+baseURL+"/auth/login", authHandler.HandleLogin)
	mux.HandleFunc("POST "+baseURL+"/uploads", uploadHandler.HandleUpload)
	mux.HandleFunc("GET /uploads/{category}/{filename}", uploadHandler.HandleServeUpload)
	if enableDocs {
		mux.HandleFunc("GET /openapi.yaml", handlers.OpenAPIHandler)
		mux.HandleFunc("GET /docs", handlers.DocsHandler)
	}

	mux.HandleFunc("POST "+baseURL+"/characters", gallery.HandleCreateCharacter)
	mux.HandleFunc("GET "+baseURL+"/characters", gallery.HandleGetAllCharacters)
	mux.HandleFunc("GET "+baseURL+"/characters/{characterId}", gallery.HandleGetCharacter)
	mux.HandleFunc("PUT "+baseURL+"/characters/{characterId}", gallery.HandleUpdateCharacter)
	mux.HandleFunc("DELETE "+baseURL+"/characters/{characterId}", gallery.HandleDeleteCharacter)
	mux.HandleFunc("POST "+baseURL+"/characters/{characterId}/prestige", gallery.HandlePrestigeCharacter)
	mux.HandleFunc("PUT "+baseURL+"/characters/{characterId}/equip/{itemId}", gallery.HandleEquipItem)
	mux.HandleFunc("POST "+baseURL+"/characters/{characterId}/use/{itemId}", gallery.HandleUseItem)

	mux.HandleFunc("POST "+baseURL+"/items", gallery.HandleCreateItem)
	mux.HandleFunc("GET "+baseURL+"/items", gallery.HandleGetAllItems)
	mux.HandleFunc("GET "+baseURL+"/items/{itemId}", gallery.HandleGetItem)
	mux.HandleFunc("PUT "+baseURL+"/items/{itemId}", gallery.HandleUpdateItem)
	mux.HandleFunc("DELETE "+baseURL+"/items/{itemId}", gallery.HandleDeleteItem)

	mux.HandleFunc("GET "+baseURL+"/characters/{characterId}/inventory", gallery.HandleGetCharacterInventory)
	mux.HandleFunc("POST "+baseURL+"/characters/{characterId}/inventory/{itemId}", gallery.HandleAddItemToCharacter)
	mux.HandleFunc("DELETE "+baseURL+"/characters/{characterId}/inventory/{itemId}", gallery.HandleRemoveItemFromCharacter)

	mux.HandleFunc("POST "+baseURL+"/roll", gallery.HandleRollDice)

	mux.HandleFunc("GET "+baseURL+"/enemies", gallery.HandleGetAllEnemies)
	mux.HandleFunc("GET "+baseURL+"/enemies/{enemyId}", gallery.HandleGetEnemy)
	mux.HandleFunc("POST "+baseURL+"/enemies", gallery.HandleCreateEnemy)
	mux.HandleFunc("PUT "+baseURL+"/enemies/{enemyId}", gallery.HandleUpdateEnemy)
	mux.HandleFunc("DELETE "+baseURL+"/enemies/{enemyId}", gallery.HandleDeleteEnemy)

	mux.HandleFunc("GET "+baseURL+"/daily/{tier}", gallery.HandleGetDaily)
	mux.HandleFunc("GET "+baseURL+"/daily/{tier}/leaderboard", gallery.HandleGetDailyLeaderboard)
	mux.HandleFunc("POST "+baseURL+"/daily/start", gallery.HandleStartDailyRun)
	mux.HandleFunc("POST "+baseURL+"/daily/move", gallery.HandleMoveDailyRun)
	mux.HandleFunc("POST "+baseURL+"/daily/action", gallery.HandleDailyCombatAction)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	publicRoutes := []string{
		"POST " + baseURL + "/auth/register",
		"POST " + baseURL + "/auth/login",
		"GET /uploads/*",
	}
	if enableDocs {
		publicRoutes = append(publicRoutes, "GET /openapi.yaml", "GET /docs")
	}

	if metricsAddr != "" {
		metricsMux := http.NewServeMux()
		metricsMux.Handle("GET /metrics", metrics.Handler())
		go func() {
			logger.Info("metrics server listening", "addr", metricsAddr)
			if err := http.ListenAndServe(metricsAddr, metricsMux); err != nil {
				logger.Error("metrics server stopped", "err", err)
			}
		}()
	} else {
		mux.Handle("GET /metrics", metrics.Handler())
		publicRoutes = append(publicRoutes, "GET /metrics")
	}

	jwtMw := middleware.AuthMiddleware(jwtSecret, publicRoutes...)
	handler := middleware.LoggingMiddleware(logger)(middleware.MetricsMiddleware(jwtMw(rl.Limit(mux))))
	handler = middleware.EnableCors(handler)

	addr := os.Getenv("PORT")
	if addr == "" {
		addr = ":8080"
	}

	logger.Info("server listening", "addr", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
