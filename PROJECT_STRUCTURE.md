# Project Structure: character-gallery-api

A RESTful Go API for an RPG character gallery. Manages characters (stats, customization, class, species), inventory (items per character), and an item pool.

- **Language:** Go 1.26
- **Database:** PostgreSQL (pgx/v5 driver, sqlc v2 codegen)
- **Cache:** Redis (go-redis/v9)
- **HTTP:** Go 1.22+ standard library mux
- **Port:** `:8080`

---

## 1. Top-Level Layout

```
character-gallery-api/
├── Dockerfile              # Multi-stage build (golang:1.26.4-alpine → alpine)
├── README.md               # Full project docs, setup, API reference
├── PROJECT_STRUCTURE.md    # This file
├── db-diagram.svg          # Entity-relationship diagram
├── openapi.yaml            # OpenAPI 3.0 spec (served at /openapi.yaml)
├── .github/
│   └── workflows/
│       └── test.yml        # CI: runs tests with PostgreSQL service container
├── .gitignore              # ⚠ Has unresolved merge conflict markers
└── src/                    # All source code (see below)
```

---

## 2. Source Directory (`src/`)

```
src/
├── .gitignore                # Ignores .exe, .test, .out, coverage, .env, .vscode/
├── config.env                # DATABASE_TYPE, REDIS_HOST/PORT/PASSWORD
├── go.mod                    # Module: dZev1/character-gallery (Go 1.26.0)
│                             # Deps: pgx/v5, pgtype, redis/go-redis/v9
├── go.sum                    #
├── item_pool.json            # Seed data: 12 RPG items
├── sqlc.yaml                 # sqlc v2 config (engine: postgresql, gen: pgx/v5)
│                             # Output: internal/postgres/db
│
├── cmd/
│   └── main.go               # Entry point — HTTP server; serves /docs, /openapi.yaml
│                              # NO database connection, NO API routes wired — STUB
│
├── database/
│   ├── schema/
│   │   └── schema.sql        # Full DDL: 6 tables with CHECK constraints, FKs, cascades
│   └── queries/
│       ├── characters.sql    # 11 queries: full character + stats + customization CRUD
│       ├── inventory.sql     # 3 queries: AddItemToCharacter (upsert), RemoveItem, GetInventory
│       └── items.sql         # 4 queries: FindAllItems, FindItem, CreateItem, SeedItems
│
├── handlers/
│   ├── character_handlers.go # EMPTY stub (just `package handlers`)
│   ├── inventory_handlers.go # EMPTY stub (just `package handlers`)
│   ├── validation.go         # EMPTY stub (just `package handlers`)
│   ├── errors.go             # EMPTY stub (just `package handlers`)
│   └── docs.go               # Swagger UI docs page (serves /openapi.yaml via SwaggerUIBundle)
│
└── internal/
    ├── cache/
    │   ├── cache.go           # Cache interface (Get, Set, Delete, Invalidate, Ping, Close)
    │   ├── redis_cache.go     # RedisCache — fully implemented (JSON serialization, prefix isolation, TTL)
    │   └── redis_cache_test.go # 10 test functions: Get/Set, ErrMiss, TTL, Delete, Invalidate, PrefixIsolation, DefaultTTL, Ping, JSON types
    │
    ├── characters/
    │   ├── character.go       # Character domain model (ID, Name, BodyType, Species, Class, Stats, Customization)
    │   ├── id.go              # CharacterID (uint64)
    │   ├── body_type.go       # Enum: type_a, type_b
    │   ├── class.go           # Enum: 12 RPG classes
    │   ├── species.go         # Enum: 10 RPG species
    │   ├── stats.go           # Stats struct (6 attributes, 1-99)
    │   ├── customization.go   # Appearance (hair, face, shirt, pants, shoes, 0-30)
    │   ├── repository.go      # Repository interface (CRUD methods, accepts db.DBTX)
    │   └── service.go         # Service struct — HAS full CRUD methods with cache integration
    │
    ├── inventory/
    │   ├── inventory_item.go  # InventoryItem model (Item + Quantity + IsEquipped)
    │   ├── repository.go      # Repository interface (Add/Remove/Get with RepositoryParam)
    │   └── service.go         # Service struct (repo, pool, cache fields) — NO methods yet
    │
    ├── items/
    │   ├── item.go            # Item domain model (name, type, description, optional stats, validation)
    │   ├── id.go              # ItemID (uint64)
    │   ├── type.go            # Enum: 14 item types (armor, weapon, scroll, etc.)
    │   ├── repository.go      # Repository interface (FindAll, Find, Save, Seed)
    │   └── service.go         # Service struct (repo, pool, cache fields) — NO methods yet
    │
    ├── middleware/
    │   ├── enable_cors.go     # CORS middleware (Allow-Origin: *, handles OPTIONS preflight)
    │   ├── ratelimit.go       # Redis-backed rate limiter (per-route, sliding window, X-RateLimit-* headers)
    │   └── ratelimit_test.go  # 7 test functions: under/over limit, headers, routes, Retry-After, POST/GET separation
    │
    ├── postgres/
    │   ├── postgres_test.go       # Test infrastructure: TestMain, newTxQueries, newTestTx, test helpers
    │   ├── character_repo.go      # PostgreSQL character repo (Save, Find, FindAll, Update, Delete with stats+customization)
    │   ├── character_repo_test.go # 7 test functions for character CRUD + stats/customization/cascade
    │   ├── items_repo.go          # PostgreSQL item repo (FindAll, Find, Save, Seed with helper funcs)
    │   ├── items_repo_test.go     # 7 test functions for item CRUD + seed
    │   ├── inventory_repo.go      # PostgreSQL inventory repo (Add, Remove, Get with item validation)
    │   ├── inventory_repo_test.go # 10 test functions for inventory add/remove/list/cascade
    │   └── db/                    # sqlc GENERATED CODE (package: postgres)
    │       ├── db.go             # DBTX interface, New(), WithTx()
    │       ├── models.go         # DB structs (Character, Customization, Inventory, Item, Stat)
    │       ├── characters.sql.go # Generated: full character CRUD + stats + customizations
    │       ├── items.sql.go      # Generated: FindAllItems, FindItem, CreateItem
    │       ├── inventory.sql.go  # Generated: AddItemToCharacter, RemoveItemFromCharacter, GetCharacterInventory
    │       └── copyfrom.go       # Generated: bulk SeedItems via COPY protocol
    │
    └── redis/
        └── client.go           # Redis client factory (NewRedisClient with Config, connection test)
```

---

## 3. Database Schema (6 tables)

| Table            | Key Columns | Notes |
|-----------------|-------------|-------|
| `items`         | name, rarity, type, damage, defense, ... | CHECK on type, UNIQUE(name, rarity) |
| `characters`    | name, body_type, species, class | CHECK on enum fields |
| `inventory`     | character_id, item_id, quantity, is_equipped | Composite PK, CASCADE deletes |
| `stats`         | character_id + 6 int16 stats (1-99) | 1:1 with characters |
| `customizations`| character_id + 5 uint16 fields (0-30) | 1:1 with characters |
| `api_keys`      | **(removed from code, schema still defines it)** | No code references remain |

---

## 4. API Endpoints (Planned)

All handler files are currently empty stubs — no routes are wired yet.

| Method | Path | Handler |
|--------|------|---------|
| POST | `/characters` | CreateCharacter |
| GET | `/characters` | GetAllCharacters (pagination) |
| GET | `/characters/{id}` | GetCharacter |
| PUT | `/characters/{id}` | EditCharacter |
| DELETE | `/characters/{id}` | DeleteCharacter |
| POST | `/characters/{id}/inventory` | AddItemToCharacter |
| DELETE | `/characters/{id}/inventory` | RemoveItemFromCharacter |
| GET | `/characters/{id}/inventory` | GetCharacterInventory |
| POST | `/items` | CreateItem |
| GET | `/items` | ShowPoolItems |
| GET | `/items/{id}` | ShowItem |

---

## 5. Test Database Connection

All PostgreSQL integration tests live in `src/internal/postgres/`. They run against a **real PostgreSQL database** using per-test transactions for isolation.

### Connection Setup (Env Var)

Tests read the connection string from the environment variable:

```
POSTGRES_TEST_DATABASE_URL
```

Set it to a PostgreSQL DSN such as:

```sh
export POSTGRES_TEST_DATABASE_URL="postgres://user:password@host:5432/test_db?sslmode=disable"
```

### How Tests Use the Connection

1. `TestMain` (`postgres_test.go:29`) checks `POSTGRES_TEST_DATABASE_URL` — if unset, all tests **skip silently** with exit code 0.
2. `pgxpool.New(ctx, dsn)` creates a connection pool for the test run.
3. The full DDL from `database/schema/schema.sql` is applied on every test run.
4. Each test function acquires a **fresh transaction** via `newTxQueries(t)` or `newTestTx(t)`.
5. Transactions auto-rollback on test teardown via `t.Cleanup(func() { _ = tx.Rollback(ctx) })` — no cleanup fixtures needed.
6. Tests use sqlc-generated `*db.Queries` directly (no service/repository layer abstraction).

### Running Tests

```sh
cd src
POSTGRES_TEST_DATABASE_URL="postgres://..." go test ./internal/postgres/ -v
```

### IDE Reference

The `.idea/dataSources.xml` configuration references a local test database at `jdbc:postgresql://192.168.1.51:5432/test`.

---

## 6. Test Files

| File | Tests | Description |
|------|-------|-------------|
| `internal/postgres/postgres_test.go` | Infrastructure | `TestMain`, `newTxQueries`, `newTestTx`, `createTestCharacter`, `createTestCharacterFull`, `createTestItem` helpers |
| `internal/postgres/character_repo_test.go` | 7 funcs | Character CRUD, pagination, stats & customization, cascading delete |
| `internal/postgres/items_repo_test.go` | 7 funcs | Item CRUD, full fields, seed items, empty pool |
| `internal/postgres/inventory_repo_test.go` | 10 funcs | Add/remove/list inventory, upsert, over-removal, cascading delete |
| `internal/cache/redis_cache_test.go` | 10 funcs | Redis cache Get/Set, miss, TTL, delete, invalidate, prefix isolation, default TTL, ping, JSON types |
| `internal/middleware/ratelimit_test.go` | 7 funcs | Rate limiter under/over limit, headers, route isolation, Retry-After, POST/GET separation |

---

## 7. Known Issues

1. **Handler files are empty stubs** — `character_handlers.go`, `inventory_handlers.go`, `validation.go`, and `errors.go` contain only `package handlers`. No actual handler logic exists yet.

2. **Main entry point is a stub** — `cmd/main.go` starts an HTTP server and serves `/docs` + `/openapi.yaml`, but has **no database connection**, **no repository wiring**, and **no API routes** wired up.

3. **Empty service layer (inventory + items)** — `internal/inventory/service.go` and `internal/items/service.go` have repo/pool/cache fields and `NewService` constructors but **zero methods**. The character service (`internal/characters/service.go`) is fully implemented.

4. **No `DATABASE_URL` reader** — No code reads a `DATABASE_URL` env var or connects to a database at startup.

5. **No `REDIS_URL` reader** — Dockerfile sets `REDIS_URL` and `REDIS_PASSWORD` env vars but no Go code reads them.

6. **Auth system removed** — Entire `internal/auth/` package was deleted (API key auth no longer exists). `internal/middleware/api_key_auth.go` was also deleted. The `api_keys` table remains in the DDL schema.

7. **`.gitignore` merge conflicts** — Root `.gitignore` contains unresolved `<<<<<<< HEAD` / `=======` / `>>>>>>>` markers.

8. **Password in config.env** — `src/config.env` contains a plaintext Redis password (`hashingIsShit,123`) — should be gitignored or loaded from env vars.
