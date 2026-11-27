package main

import (
	"context"
	"dZev1/character-gallery/database"
	"dZev1/character-gallery/handlers"
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
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	pgConnStr := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_URL")
	jwtSecret := os.Getenv("JWT_SECRET")
	if pgConnStr == "" || redisURL == "" || jwtSecret == "" {
		logger.Error("DATABASE_URL, REDIS_URL and JWT_SECRET must be set")
		os.Exit(1)
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, pgConnStr)
	if err != nil {
		logger.Error("failed to connect to postgres", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err = database.ApplySchema(ctx, pool); err != nil {
		logger.Error("failed to apply schema", "err", err)
		os.Exit(1)
	}
	metrics.RegisterDBPoolMetrics(pool)

	redisOpts, err := redis.ParseURL(redisURL)
	if err != nil {
		logger.Error("invalid REDIS_URL", "err", err)
		os.Exit(1)
	}
	rdb, err := redislib.NewRedisClient(redisOpts)
	if err != nil {
		logger.Error("failed to connect to redis", "err", err)
		os.Exit(1)
	}
	defer rdb.Close()

	// Repositories
	authRepo := postgres.NewAuthRepo(pool)
	charRepo := postgres.NewCharacterRepo(pool)
	itemRepo := postgres.NewItemRepo(pool)
	invRepo := postgres.NewInventoryRepo(pool)
	enemyRepo := postgres.NewEnemyRepo(pool)
	dailyRepo := postgres.NewDailyRunRepo(pool)

	// Caches
	charCache := cache.NewRedisCache(rdb, 5*time.Minute, "character")
	itemCache := cache.NewRedisCache(rdb, 5*time.Minute, "item")
	invCache := cache.NewRedisCache(rdb, 5*time.Minute, "inventory")
	enemyCache := cache.NewRedisCache(rdb, 5*time.Minute, "enemy")

	// Services
	authSvc := auth.NewService(authRepo, pool, jwtSecret)
	charSvc := characters.NewService(charRepo, pool, charCache)
	itemSvc := items.NewService(itemRepo, pool, itemCache)
	invSvc := inventory.NewService(invRepo, pool, invCache).WithCharacterStore(charRepo, charCache)
	enemySvc := enemies.NewService(enemyRepo, pool, enemyCache)
	dailySvc := daily_run.NewService(dailyRepo, charRepo, enemyRepo, invRepo, itemRepo, pool, charCache, invCache)

	if strings.EqualFold(os.Getenv("ENABLE_SEED"), "true") {
		if err = itemSvc.SeedItems(ctx); err != nil {
			logger.Warn("seed items failed", "err", err)
		}
		if err = enemySvc.SeedEnemies(ctx); err != nil {
			logger.Warn("seed enemies failed", "err", err)
		}
	}

	baseURL := strings.Trim(strings.TrimSpace(os.Getenv("BASE_URL")), `"'`)
	if baseURL == "" {
		baseURL = "/api"
	}
	enableDocs := strings.EqualFold(os.Getenv("ENABLE_DOCS"), "true")
	metricsAddr := strings.TrimSpace(os.Getenv("METRICS_ADDR"))

	authHandler := handlers.NewAuthHandler(authSvc)
	uploadHandler := handlers.NewUploadHandler(os.Getenv("UPLOADS_DIR"))
	gallery := handlers.NewGallery(charSvc, itemSvc, invSvc).WithGameServices(enemySvc, dailySvc)

	mux := http.NewServeMux()
	publicRoutes := registerRoutes(mux, baseURL, enableDocs, authHandler, uploadHandler, gallery)
	rl := configureRateLimits(rdb, baseURL)

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
	handler := middleware.EnableCors(
		middleware.LoggingMiddleware(logger)(
			middleware.MetricsMiddleware(jwtMw(rl.Limit(mux))),
		),
	)

	addr := os.Getenv("PORT")
	if addr == "" {
		addr = ":8080"
	}

	logger.Info("server listening", "addr", addr, "base_url", baseURL)
	if err := http.ListenAndServe(addr, handler); err != nil {
		logger.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func configureRateLimits(rdb *redis.Client, baseURL string) *middleware.RateLimiter {
	rl := middleware.NewRateLimiter(rdb, 1*time.Minute)
	rl.SetLimit("POST", baseURL+"/characters", 10)
	rl.SetLimit("DELETE", baseURL+"/characters/{characterId}", 30)
	rl.SetLimit("POST", baseURL+"/items", 10)
	rl.SetLimit("POST", baseURL+"/items/bulk", 10)
	rl.SetLimit("POST", baseURL+"/enemies", 10)
	rl.SetLimit("POST", baseURL+"/enemies/bulk", 10)
	rl.SetLimit("POST", baseURL+"/uploads", 20)
	rl.SetLimit("POST", baseURL+"/auth/register", 5)
	rl.SetLimit("POST", baseURL+"/auth/login", 10)
	return rl
}

func registerRoutes(
	mux *http.ServeMux,
	baseURL string,
	enableDocs bool,
	authHandler *handlers.AuthHandler,
	uploadHandler *handlers.UploadHandler,
	gallery *handlers.Gallery,
) []string {
	mux.HandleFunc("POST "+baseURL+"/auth/register", authHandler.HandleRegister)
	mux.HandleFunc("POST "+baseURL+"/auth/login", authHandler.HandleLogin)
	mux.HandleFunc("POST "+baseURL+"/uploads", uploadHandler.HandleUpload)
	mux.HandleFunc("GET /uploads/{category}/{filename}", uploadHandler.HandleServeUpload)

	publicRoutes := []string{
		"POST " + baseURL + "/auth/register",
		"POST " + baseURL + "/auth/login",
		"GET /uploads/*",
	}
	if enableDocs {
		mux.HandleFunc("GET /openapi.yaml", handlers.OpenAPIHandler)
		mux.HandleFunc("GET /docs", handlers.DocsHandler)
		publicRoutes = append(publicRoutes, "GET /openapi.yaml", "GET /docs")
	}

	// Characters
	mux.HandleFunc("GET "+baseURL+"/characters/stat-roll", gallery.HandleGetStatRoll)
	mux.HandleFunc("POST "+baseURL+"/characters/stat-roll/reroll", gallery.HandleRerollStats)
	mux.HandleFunc("POST "+baseURL+"/characters", gallery.HandleCreateCharacter)
	mux.HandleFunc("GET "+baseURL+"/characters", gallery.HandleGetAllCharacters)
	mux.HandleFunc("QUERY "+baseURL+"/characters", gallery.HandleQueryCharacters)
	mux.HandleFunc("GET "+baseURL+"/characters/{characterId}", gallery.HandleGetCharacter)
	mux.HandleFunc("PUT "+baseURL+"/characters/{characterId}", gallery.HandleUpdateCharacter)
	mux.HandleFunc("DELETE "+baseURL+"/characters/{characterId}", gallery.HandleDeleteCharacter)
	mux.HandleFunc("POST "+baseURL+"/characters/{characterId}/prestige", gallery.HandlePrestigeCharacter)
	mux.HandleFunc("PUT "+baseURL+"/characters/{characterId}/equip/{itemId}", gallery.HandleEquipItem)
	mux.HandleFunc("POST "+baseURL+"/characters/{characterId}/use/{itemId}", gallery.HandleUseItem)

	// Inventory
	mux.HandleFunc("GET "+baseURL+"/characters/{characterId}/inventory", gallery.HandleGetCharacterInventory)
	mux.HandleFunc("POST "+baseURL+"/characters/{characterId}/inventory/{itemId}", gallery.HandleAddItemToCharacter)
	mux.HandleFunc("DELETE "+baseURL+"/characters/{characterId}/inventory/{itemId}", gallery.HandleRemoveItemFromCharacter)

	// Items
	mux.HandleFunc("POST "+baseURL+"/items", gallery.HandleCreateItem)
	mux.HandleFunc("POST "+baseURL+"/items/bulk", gallery.HandleBulkCreateItems)
	mux.HandleFunc("GET "+baseURL+"/items", gallery.HandleGetAllItems)
	mux.HandleFunc("GET "+baseURL+"/items/{itemId}", gallery.HandleGetItem)
	mux.HandleFunc("PUT "+baseURL+"/items/{itemId}", gallery.HandleUpdateItem)
	mux.HandleFunc("DELETE "+baseURL+"/items/{itemId}", gallery.HandleDeleteItem)

	// Enemies
	mux.HandleFunc("GET "+baseURL+"/enemies", gallery.HandleGetAllEnemies)
	mux.HandleFunc("GET "+baseURL+"/enemies/{enemyId}", gallery.HandleGetEnemy)
	mux.HandleFunc("POST "+baseURL+"/enemies", gallery.HandleCreateEnemy)
	mux.HandleFunc("POST "+baseURL+"/enemies/bulk", gallery.HandleBulkCreateEnemies)
	mux.HandleFunc("PUT "+baseURL+"/enemies/{enemyId}", gallery.HandleUpdateEnemy)
	mux.HandleFunc("DELETE "+baseURL+"/enemies/{enemyId}", gallery.HandleDeleteEnemy)

	// Dice & Daily Run
	mux.HandleFunc("POST "+baseURL+"/roll", gallery.HandleRollDice)
	mux.HandleFunc("GET "+baseURL+"/daily/{tier}", gallery.HandleGetDaily)
	mux.HandleFunc("GET "+baseURL+"/daily/{tier}/leaderboard", gallery.HandleGetDailyLeaderboard)
	mux.HandleFunc("POST "+baseURL+"/daily/start", gallery.HandleStartDailyRun)
	mux.HandleFunc("POST "+baseURL+"/daily/move", gallery.HandleMoveDailyRun)
	mux.HandleFunc("POST "+baseURL+"/daily/action", gallery.HandleDailyCombatAction)

	return publicRoutes
}
