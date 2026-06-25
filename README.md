# Character Gallery API

A RESTful API built in Go to create, manage and see a gallery of Role Playing Game characters. Full CRUD operations with base stats, character customization, an item pool, and per-character inventory management.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Architecture](#architecture)
3. [Characters](#characters)
   - [Classes](#classes)
   - [Species](#species)
   - [Stats & Customization](#stats--customization)
4. [Items](#items)
   - [Item Types](#item-types)
5. [API Reference](#api-reference)
   - [Characters](#characters-endpoints)
   - [Items](#items-endpoints)
   - [Inventory](#inventory-endpoints)
6. [Entity Diagram](#entity-diagram)

---

## Getting Started

1. Clone the repo and enter the source directory:
   ```bash
   git clone https://github.com/dZev1/character-gallery-api.git
   cd character-gallery-api/src
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Set up a PostgreSQL database and a Redis instance, then configure:
   ```bash
   export DATABASE_URL="postgres://user:password@host:5432/dbname"
   export REDIS_URL="redis://:password@host:6379/0"
   ```

4. Run the application:
   ```bash
   go run ./cmd
   ```
   Server listens on `http://localhost:8080`.

5. Open the API docs:
   ```
   http://localhost:8080/docs
   ```

### Docker

```bash
docker build -t character-gallery-api .
docker run -p 8080:8080 \
  -e DATABASE_URL="postgres://..." \
  -e REDIS_URL="redis://:..." \
  character-gallery-api
```

---

## Architecture

```
Handler (net/http) → Service (business logic + cache) → Repository (PostgreSQL via sqlc)
                                ↕
                           Redis Cache
```

- **Go 1.26** with `net/http` (Go 1.22+ routing with path params)
- **PostgreSQL** via `pgx/v5` + sqlc-generated queries
- **Redis** via `go-redis/v9` — cache-aside pattern + fixed-window rate limiting
- **CORS** middleware enabled for all origins
- Concurrent by design — every request runs in its own goroutine, all layers are goroutine-safe

---

## Characters

### Classes

| Class      | JSON Tag     |
|------------|--------------|
| Barbarian  | `barbarian`  |
| Bard       | `bard`       |
| Cleric     | `cleric`     |
| Druid      | `druid`      |
| Fighter    | `fighter`    |
| Monk       | `monk`       |
| Paladin    | `paladin`    |
| Ranger     | `ranger`     |
| Rogue      | `rogue`      |
| Sorcerer   | `sorcerer`   |
| Warlock    | `warlock`    |
| Wizard     | `wizard`     |

### Species

| Species    | JSON Tag     |
|------------|--------------|
| Aasimar    | `aasimar`    |
| Dragonborn | `dragonborn` |
| Dwarf      | `dwarf`      |
| Elf        | `elf`        |
| Gnome      | `gnome`      |
| Goliath    | `goliath`    |
| Halfling   | `halfling`   |
| Human      | `human`      |
| Orc        | `orc`        |
| Tiefling   | `tiefling`   |

### Body Types

| Type    | JSON Tag   |
|---------|------------|
| Type A  | `type_a`   |
| Type B  | `type_b`   |

### Stats & Customization

Each character has six stats (1–99) and five customization fields (0–30).

| Stat         | JSON Tag        |
|--------------|-----------------|
| Strength     | `strength`      |
| Dexterity    | `dexterity`     |
| Constitution | `constitution`  |
| Intelligence | `intelligence`  |
| Wisdom       | `wisdom`        |
| Charisma     | `charisma`      |

| Customization | JSON Tag |
|---------------|----------|
| Hair          | `hair`   |
| Face          | `face`   |
| Shirt         | `shirt`  |
| Pants         | `pants`  |
| Shoes         | `shoes`  |

---

## Items

### Item Types

| Type             | JSON Tag            |
|------------------|---------------------|
| Armor            | `armor`             |
| Ring             | `ring`              |
| Weapon           | `weapon`            |
| Shield           | `shield`            |
| Tool             | `tool`              |
| Adventuring Gear | `adventuring_gear`  |
| Rod              | `rod`               |
| Staff            | `staff`             |
| Wand             | `wand`              |
| Scroll           | `scroll`            |
| Potion           | `potion`            |
| Ammo             | `ammo`              |
| Consumable       | `consumable`        |
| Wondrous Item    | `wondrous_item`     |

---

## API Reference

Base path: `/api/v1`

---

### Characters Endpoints

#### Create a character

```
POST /api/v1/characters
```

```json
{
    "name": "Arwen",
    "body_type": "type_b",
    "species": "elf",
    "class": "wizard",
    "stats": {
        "strength": 10,
        "dexterity": 5,
        "constitution": 10,
        "intelligence": 5,
        "wisdom": 7,
        "charisma": 3
    },
    "customization": {
        "hair": 0,
        "face": 3,
        "shirt": 4,
        "pants": 2,
        "shoes": 1
    }
}
```

`201 Created` — Returns the created character with its new `id`.

---

#### Get all characters

```
GET /api/v1/characters?page=1
```

| Parameter | Type   | Default | Description                  |
|-----------|--------|---------|------------------------------|
| `page`    | int    | 1       | Page number (1-indexed)      |

`200 OK`

```json
{
    "data": [
        {
            "id": 1,
            "name": "Shallan",
            "body_type": "type_b",
            "species": "human",
            "class": "monk",
            "stats": { ... },
            "customization": { ... }
        }
    ],
    "pagination": {
        "page": 1,
        "limit": 20,
        "total": 150,
        "has_next": true
    }
}
```

---

#### Get a character

```
GET /api/v1/characters/{id}
```

`200 OK` — Returns the character object.

```
404 Not Found` — Character doesn't exist.

---

#### Update a character

```
PUT /api/v1/characters/{id}
```

Body is the same as create (excluding `id`).

`200 OK` — Returns the updated character.

---

#### Delete a character

```
DELETE /api/v1/characters/{id}
```

`200 OK`

```json
{ "result": "success" }
```

---

### Items Endpoints

#### Create an item

```
POST /api/v1/items
```

```json
{
    "name": "Master Sword",
    "type": "weapon",
    "description": "A legendary sword with immense power.",
    "equippable": true,
    "rarity": 5,
    "damage": 34,
    "defense": 23
}
```

`201 Created` — Returns the item with its new `id`.

---

#### Get all items

```
GET /api/v1/items
```

`200 OK` — Returns the full item pool array.

---

#### Get an item

```
GET /api/v1/items/{id}
```

`200 OK` — Returns the item object.

---

### Inventory Endpoints

#### Get a character's inventory

```
GET /api/v1/characters/{characterId}/inventory
```

`200 OK`

```json
[
    {
        "item": {
            "id": 3,
            "name": "Healing Potion",
            "type": "potion",
            "description": "A potion that restores health.",
            "equippable": false,
            "rarity": 1,
            "heal_amount": 60
        },
        "quantity": 3,
        "is_equipped": false
    }
]
```

---

#### Add item to character

```
POST /api/v1/characters/{characterId}/inventory/{itemId}?quantity=1
```

| Parameter  | Type | Default | Description            |
|------------|------|---------|------------------------|
| `quantity` | int  | 1       | Amount to add (min 1)  |

`200 OK` — Returns the inventory entry for the added item.

---

#### Remove item from character

```
DELETE /api/v1/characters/{characterId}/inventory/{itemId}?quantity=1
```

| Parameter  | Type | Default | Description               |
|------------|------|---------|---------------------------|
| `quantity` | int  | 1       | Amount to remove (min 1)  |

`200 OK`

```json
{ "result": "success" }
```

---

## Entity Diagram

![Entity Diagram](./db-diagram.svg)
