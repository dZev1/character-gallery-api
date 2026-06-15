# Project Structure: character-gallery-api

A RESTful Go API for an RPG character gallery. Manages characters (stats, customization, class, species), inventory (items per character), and an item pool.

- **Language:** Go 1.26
- **Database:** PostgreSQL (pgx/v5 driver, sqlc v1.31.1 codegen)
- **HTTP:** Go 1.22+ standard library mux
- **Port:** `:8080`

---

## 1. Top-Level Layout

```
character-gallery-api/
├── Dockerfile              # Multi-stage build (golang:1.26-alpine → alpine)
├── README.md               # Full project docs, setup, API reference
├── PROJECT_STRUCTURE.md    # This file
├── db-diagram.svg          # Entity-relationship diagram
├── .gitignore              # Ignores *.sh files
├── .idea/                  # GoLand IDE config (includes test DB data source ref)
└── src/                    # All source code (see below)
```

---

## 2. Source Directory (`src/`)

```
src/
├── .gitignore                # Ignores .exe, .test, .out, coverage, .env, .vscode/
├── config.env                # DATABASE_TYPE="postgres"
├── go.mod                    # Module: dZev1/character-gallery (Go 1.26.0)
├── go.sum                    # Deps: pgx/v5, sqlx, godotenv, go-sqlmock
├── item_pool.json            # Seed data: 12 RPG items
├── sqlc.yaml                 # sqlc v2 config (engine: postgresql, gen: pgx/v5)
│
├── cmd/
│   ├── main.go               # Entry point — STUB (only prints "Hello, world!")
│   └── apikey_gen/
│       └── main.go           # CLI tool — STUB (prints hardcoded info, not functional)
│
├── database/
│   ├── schema/
│   │   └── schema.sql        # Full DDL: 6 tables with CHECK constraints, FKs, cascades
│   └── queries/
│       ├── auth.sql          # 3 queries: ValidateAPIKey, UpdateLastUsed, CreateAPIKey
│       ├── characters.sql    # 11 queries: full character + stats + customization CRUD
│       ├── inventory.sql     # 3 queries: AddItemToCharacter (upsert), RemoveItem, GetInventory
│       └── items.sql         # 4 queries: FindAllItems, FindItem, CreateItem, SeedItems
│
├── handlers/
│   ├── character_handlers.go # Character CRUD: POST/GET/PUT/DELETE with pagination
│   ├── inventory_handlers.go # Inventory add/remove/list + item pool CRUD
│   ├── errors.go             # Standard JSON error response helper
│   └── validation.go         # Input validation (character fields, item fields)
│
└── internal/
    ├── auth/
    │   ├── api_key.go        # APIKey struct, SHA-256 key generation/hashing
    │   ├── repository.go     # Repository interface (ValidateAPIKey, UpdateLastUsed, CreateAPIKey)
    │   └── service.go        # Service struct — EMPTY (no methods)
    │
    ├── characters/
    │   ├── character.go      # Character domain model (ID, Name, BodyType, Species, Class, Stats, Customization)
    │   ├── id.go             # CharacterID (uint64)
    │   ├── body_type.go      # Enum: type_a, type_b
    │   ├── class.go          # Enum: 12 RPG classes
    │   ├── species.go        # Enum: 10 RPG species
    │   ├── stats.go          # Stats struct (6 attributes, 1-99)
    │   ├── customization.go  # Appearance (hair, face, shirt, pants, shoes, 0-30)
    │   ├── repository.go     # Repository interface (CRUD methods)
    │   └── service.go        # Service struct — EMPTY (no methods)
    │
    ├── inventory/
    │   ├── inventory_item.go # InventoryItem model (Item + Quantity + IsEquipped)
    │   ├── repository.go     # Repository interface (Add/Remove/Get)
    │   └── service.go        # Service struct — EMPTY (no methods)
    │
    ├── items/
    │   ├── item.go           # Item domain model (name, type, description, optional stats, validation)
    │   ├── id.go             # ItemID (uint64)
    │   ├── type.go           # Enum: 14 item types (armor, weapon, scroll, etc.)
    │   ├── repository.go     # Repository interface (FindAll, Find, Save, Seed)
    │   └── service.go        # Service struct — EMPTY (no methods)
    │
    ├── middleware/
    │   ├── api_key_auth.go    # X-API-Key auth middleware — ENTIRELY COMMENTED OUT (disabled)
    │   └── enable_cors.go     # CORS middleware (Allow-Origin: *, handles OPTIONS preflight)
    │
    └── postgres/
        ├── postgres_test.go       # Test infrastructure: TestMain, newTxQueries, newTestTx, test helpers
        ├── auth_repo_test.go      # 6 tests for API key CRUD + validation
        ├── character_repo.go      # PostgreSQL character repo — ALL methods panic("implement me")
        ├── character_repo_test.go # 13 tests for character CRUD + stats/customization/cascade
        ├── items_repo.go          # PostgreSQL item repo — partial impl, BUG: recursive calls
        ├── items_repo_test.go     # 7 tests for item CRUD + seed
        ├── inventory_repo_test.go # 10 tests for inventory add/remove/list/cascade
        └── db/                    # sqlc v1.31.1 GENERATED CODE (package: postgres)
            ├── db.go             # DBTX interface, New(), WithTx()
            ├── models.go         # DB structs (ApiKey, Character, Customization, Inventory, Item, Stat)
            ├── auth.sql.go       # Generated: CreateAPIKey, UpdateLastUsed, ValidateAPIKey
            ├── characters.sql.go # Generated: full character CRUD + stats + customizations
            ├── items.sql.go      # Generated: FindAllItems, FindItem, CreateItem
            ├── inventory.sql.go  # Generated: AddItemToCharacter, RemoveItemFromCharacter, GetCharacterInventory
            └── copyfrom.go       # Generated: bulk SeedItems via COPY protocol
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
| `api_keys`      | key_hash (SHA-256), name, is_active | SHA-256 hex hash stored |

---

## 4. API Endpoints

All under `/api/{version}/`:

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

1. `TestMain` (`postgres_test.go:20`) checks `POSTGRES_TEST_DATABASE_URL` — if unset, all tests **skip silently** with exit code 0.
2. `pgx.Connect(ctx, dsn)` creates a single connection for the test run.
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
| `internal/postgres/auth_repo_test.go` | 6 | API key CRUD, duplicate hash, validation, inactive key, update last used |
| `internal/postgres/character_repo_test.go` | 13 | Character CRUD, pagination, stats & customization, cascading delete |
| `internal/postgres/items_repo_test.go` | 7 | Item CRUD, full fields, seed items, empty pool |
| `internal/postgres/inventory_repo_test.go` | 10 | Add/remove/list inventory, upsert, over-removal, cascading delete |
| `internal/middleware/api_key_auth_test.go` | 5 | Unit tests with `MockAuthStore` (no DB needed) |

---

## 7. Known Issues

1. **Broken imports** — Handlers reference old `dZev1/character-gallery/models` and `dZev1/character-gallery/internal/database` packages that no longer exist.
2. **Main entry point is a stub** — `cmd/main.go` only prints "Hello, world!". No HTTP server, no DB connection, no routing.
3. **Character repo not implemented** — `internal/postgres/character_repo.go` is all `panic("implement me")`.
4. **Recursive call bug** — `items_repo.go`'s `FindAllItems` and `FindItem` call themselves instead of the embedded `Queries` methods (infinite loop).
5. **Dockerfile path mismatch** — References `internal/database/postgres_gallery/schema.sql`; actual path is `database/schema/schema.sql`.
6. **sqlc output mismatch** — `sqlc.yaml` specifies output `internal/db/db`; actual generated code is in `internal/postgres/db`.
7. **Empty service layer** — All 4 service packages (auth, characters, inventory, items) have repo fields but zero methods.
8. **Auth middleware disabled** — Entire `api_key_auth.go` file is commented out.
9. **No `DATABASE_URL` reader** — No code reads a `DATABASE_URL` env var or connects to a database at startup.
