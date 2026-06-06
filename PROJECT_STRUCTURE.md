# Project Structure: character-gallery-api

A RESTful Go API for an RPG character gallery. Manages characters (stats, customization, class, species), inventory (items per character), and an item pool.

- **Language:** Go 1.26
- **Database:** PostgreSQL (pgx driver, sqlc codegen)
- **HTTP:** Go 1.22+ standard library mux
- **Port:** `:8080`

---

## 1. Top-Level Layout

```
character-gallery-api/
├── Dockerfile              # Multi-stage build (golang:1.26-alpine → alpine)
├── README.md               # Full project docs, setup, API reference
├── db-diagram.svg          # Entity-relationship diagram
├── .gitignore              # Ignores .sh files
├── .idea/                  # GoLand IDE config
└── src/                    # All source code (see below)
```

---

## 2. Source Directory (`src/`)

```
src/
├── cmd/
│   ├── main.go                # Entry point — HTTP server, routing, graceful shutdown
│   └── apikey_gen/
│       └── main.go            # CLI tool: generates API keys, stores hash in DB
├── config.env                 # DATABASE_TYPE="postgres"
├── go.mod / go.sum            # Module: dZev1/character-gallery
│                              #   pgx/v5, sqlx, godotenv, sqlmock
├── item_pool.json             # Seed data: 11 RPG items (armor, weapons, etc.)
├── sqlc.yaml                  # sqlc config (engine: postgresql)
│
├── database/
│   ├── schema/
│   │   └── schema.sql         # Full DDL: items, characters, inventory, stats,
│   │                          #   customizations, api_keys
│   └── queries/
│       ├── auth.sql           # Empty
│       ├── characters.sql     # Named queries for character CRUD
│       ├── inventory.sql      # Empty
│       └── items.sql          # Named queries for item CRUD + seed
│
├── handlers/
│   ├── character_handlers.go  # Character CRUD: POST/GET/PUT/DELETE
│   ├── inventory_handlers.go  # Inventory add/remove/list + item pool CRUD
│   ├── errors.go              # Standard JSON error response helper
│   └── validation.go          # Input validation (character fields, item fields)
│
└── internal/
    ├── auth/
    │   ├── api_key.go         # APIKey struct, generate/hash SHA-256
    │   ├── repository.go      # Repository interface (ValidateAPIKey, etc.)
    │   └── service.go         # Empty service stub
    │
    ├── characters/
    │   ├── character.go       # Character domain model
    │   ├── id.go              # CharacterID (uint64)
    │   ├── body_type.go       # Enum: type_a, type_b
    │   ├── class.go           # Enum: 12 RPG classes
    │   ├── species.go         # Enum: 10 RPG species
    │   ├── stats.go           # Stats struct (6 attributes, 1-99)
    │   ├── customization.go   # Appearance (hair, face, shirt, pants, shoes, 0-30)
    │   ├── repository.go      # Repository interface (CRUD methods)
    │   └── service.go         # Empty service stub
    │
    ├── inventory/
    │   ├── inventory_item.go  # InventoryItem model (item + quantity + equipped)
    │   ├── repository.go      # Repository interface
    │   └── service.go         # Empty service stub
    │
    ├── items/
    │   ├── item.go            # Item domain model (name, type, stats, validation)
    │   ├── id.go              # ItemID (uint64)
    │   ├── type.go            # Enum: 14 item types (armor, weapon, scroll, etc.)
    │   ├── repository.go      # Repository interface
    │   └── service.go         # Empty service stub
    │
    ├── middleware/
    │   ├── api_key_auth.go       # X-API-Key header validation middleware
    │   ├── api_key_auth_test.go  # 5 unit tests with MockAuthStore
    │   └── enable_cors.go        # CORS middleware (Allow-Origin: *)
    │
    └── postgres/
        ├── character_repo.go     # PostgreSQL character repo — all panic("implement me")
        ├── character_repo_test.go# Empty test placeholder
        ├── items_repo.go         # PostgreSQL item repo — partial, has recursive-call bug
        └── db/
            ├── db.go             # sqlc-generated: DBTX interface, New(), WithTx()
            ├── models.go         # sqlc-generated: DB structs (Character, Item, etc.)
            ├── items.sql.go      # sqlc-generated: FindAllItems, FindItem, SaveItem
            └── copyfrom.go       # sqlc-generated: bulk SeedItems iterator
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

## 5. Test Files

| File | Description |
|------|-------------|
| `internal/middleware/api_key_auth_test.go` | 5 tests for auth middleware (missing header, invalid key, valid key, DB error, key hashing) |
| `internal/postgres/character_repo_test.go` | Placeholder (empty test function) |

---

## 6. Known Issues

1. **Broken imports** — Handlers reference old `dZev1/character-gallery/models` and `dZev1/character-gallery/internal/database` packages that no longer exist. Domain types were moved to `internal/` but imports not updated.
2. **Character repo not implemented** — `internal/postgres/character_repo.go` is all `panic("implement me")`.
3. **Recursive call bug** — `items_repo.go`'s `FindAllItems` and `FindItem` call themselves recursively (infinite loop at runtime).
4. **Dockerfile path mismatch** — References `internal/database/postgres_gallery/schema.sql`; actual path is `database/schema/schema.sql`.
5. **sqlc output mismatch** — `sqlc.yaml` specifies output `internal/db/db`; actual generated code is in `internal/postgres/db`.
6. **Empty service layer** — All 4 service packages (auth, characters, inventory, items) have repo fields but zero methods.
7. **Empty SQL query files** — `auth.sql` and `inventory.sql` are empty; no sqlc code generated for them.
