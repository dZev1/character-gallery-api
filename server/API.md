# Character Gallery

## Description

A RESTful API built in Go to create, manage and see a gallery of Role Playing Game characters. The API allows CRUD complete operations, including management of base stats and character appearance customization.

## Getting Started

1. Clone repository

    ```(bash)
    git clone https://github.com/dZev1/fantasy-character-gallery.git
    cd character-gallery
    ```

2. Install dependencies

    ```(bash)
    go mod tidy
    ```

3. Configure the database

    - Create a database for the project.
    - Select one of the supported database engines (PostgreSQL, MariaDB, SQLite), de-comment it in `config.env` file.
    - The database schema will be created automatically when the application starts.

4. Configure environment variables

    - Create a `.env` file in the root of the project.
    - Add the following variable with the connection URL to your database:

        ```(.env)
        DATABASE_URL="postgres://user:password@localhost:XXXX/database_name?params"
        ```

5. Run the application:

    - Build the application:

        ```(bash)
        go build ./cmd/
        ```

    - Run `cmd.exe`.

    - Server will be listening in `http://localhost:8080`.

## API References

---

### Character Management

#### Create a character

- **Endpoint**: `POST /characters`
- **Description**: Creates a new character with their stats and customization.
- **Request Body**:

```JSON
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

- **Succesful Response (`201 Created`)**: Returns the object of the created character, including their new `id`.

#### Get all characters

- **Endpoint**: `GET /characters`
- **Description**: Returns an array of all characters in the database.

- **Succesful Response (`200 OK`)**: Returns an array of character objects, each including stats and customization fields:

```JSON
[
    {
        "id": 1,
        "name": "Shallan",
        "body_type": "type_b",
        "species": "human",
        "class": "monk",
        "stats": {
            ...
        },
        "customization": {
            ...
        }
    },
    {
        "id": 2,
        "name": "Dalinar",
        "body_type": "type_a",
        "species": "human",
        "class": "fighter",
        "stats": {
            ...
        },
        "customization": {
            ...
        }
    }
]
```

#### Get a character

- **Endpoint**: `GET /characters/{id}`
- **Description**: Returns a single character by their `id`.
- **Succesful Response (`200 OK`)**: Returns the object of the character with the specified `id`, including stats and customization fields:

```JSON
{
    "id": 1,
    "name": "Shallan",
    "body_type": "type_b",
    "species": "human",
    "class": "monk",
    "stats": {
        ...
    },
    "customization": {
        ...
    }
}
```

#### Edit a character

- **Endpoint**: `PUT /characters/{id}`
- **Description**: Updates an existing character by their `id`.
- **Request Body**: A character object with the updated fields.

```JSON
{
    "name": "Shallan",
    "body_type": "type_b",
    "species": "human",
    "class": "monk",
    "stats": {
        "strength": 8,
        "dexterity": 7,
        "constitution": 9,
        "intelligence": 6,
        "wisdom": 8,
        "charisma": 5
    },
    "customization": {
        "hair": 1,
        "face": 2,
        "shirt": 3,
        "pants": 4,
        "shoes": 0
    }
}
```

- **Succesful Response (`200 OK`)**: Returns the object of the updated character, including their `id`.

#### Delete a character

- **Endpoint**: `DELETE /characters/{id}`
- **Description**: Deletes an existing character by their `id`.
- **Succesful Response (`200 OK`)**.

### Character Inventory Management

#### Add item to character inventory

- **Endpoint**: `POST /characters/{character_id}/inventory/{item_id}`
- **Description**: Adds a specific item to the character's inventory.
- **Path Variables**:
  - `character_id`: The ID of the character.
  - `item_id`: The ID of the item to be added to the character's inventory.
- **Query Parameters**:
  - `quantity`: *(OPTIONAL)* The amount of items to add. If no value is specified, defaults to 1.
- **Successful Response (`200 OK`)**: returns the object of the item added:

```JSON
{
  "id": 3,
  "name": "Healing Potion",
  "type": "potion",
  "description": "A potion that restores health.",
  "equippable": false,
  "rarity": 1,
  "heal_amount": 60
}
```

#### Delete item from character's inventory

- **Endpoint**: `DELETE /characters/{character_id}/inventory/{item_id}`
- **Description**: Deletes a specific item.
- **Path Variables**:
  - `character_id`: The ID of the character.
  - `item_id`: The ID of the item to be added to the character's inventory.
- **Query Parameters**:
  - `quantity`: *(OPTIONAL)* The amount of items to add. If no value is specified, defaults to 1.
- **Successful Response (`200 OK`)**: returns the object of the item added:

```JSON
{
  "id": 3,
  "name": "Healing Potion",
  "type": "potion",
  "description": "A potion that restores health.",
  "equippable": false,
  "rarity": 1,
  "heal_amount": 60
}
```

#### Get a character's inventory

- **Endpoint**: `GET /characters/{character_id}/inventory`
- **Description**: Gets a character's inventory.
- **Path Variables**:
  - `character_id`: The ID of the character.
- **Successful Response(`200 ok`)**: returns an array of the items belonging to the character.

```JSON
[
  {
    "item": {
      "id": 2,
      "name": "Carl's Doomsday Scenario",
      "type": "explosive",
      "description": "Created  by  a  man  who  murders  babies  and  steals  rare collectibles from his elders, this device is powerful enough to level an entire city and all the suburbs around it. It is created by combining  a  massively  overloaded  soul crystal and  a  Sheol Glass Reaper Case.",
      "equippable": false,
      "rarity": 5,
      "damage": 1000
    },
    "quantity": 4,
    "is_equipped": false
  },
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

### Item Pool Management

#### Get the current Item Pool

- **Endpoint**: `GET /items`
- **Description**: Gets the whole item pool.
- **Successful Response(`200 ok`)**: returns an array that represents the current item pool.

```JSON
[
  {
    "id": 1,
    "name": "Master Sword",
    "type": "weapon",
    "description": "A legendary sword with immense power.",
    "equippable": true,
    "rarity": 5,
    "damage": 100
  },
  {
    "id": 2,
    "name": "Carl's Doomsday Scenario",
    "type": "explosive",
    "description": "Created  by  a  man  who  murders  babies  and  steals  rare collectibles from his elders, this device is powerful enough to level an entire city and all the suburbs around it. It is created by combining  a  massively  overloaded  soul crystal and  a  Sheol Glass Reaper Case.",
    "equippable": false,
    "rarity": 5,
    "damage": 1000
  },
  {
    "id": 3,
    "name": "Healing Potion",
    "type": "potion",
    "description": "A potion that restores health.",
    "equippable": false,
    "rarity": 1,
    "heal_amount": 60
  },
  {
    "id": 4,
    "name": "Steel Armor",
    "type": "armor",
    "description": "Sturdy armor made of steel.",
    "equippable": true,
    "rarity": 3,
    "defense": 40
  },
  {
    "id": 5,
    "name": "Magic Missile",
    "type": "spell",
    "description": "A spell that launches a magic missile at the target.",
    "equippable": false,
    "rarity": 2,
    "mana_cost": 30
  },
  {
    "id": 6,
    "name": "Invisibility Cloak",
    "type": "misc",
    "description": "A cloak that grants invisibility to the wearer.",
    "equippable": true,
    "rarity": 4,
    "duration": "5 minutes"
  },
  {
    "id": 7,
    "name": "Paris's Bow",
    "type": "weapon",
    "description": "A finely crafted bow used by the legendary archer Paris.",
    "equippable": true,
    "rarity": 4,
    "damage": 80
  },
  {
    "id": 8,
    "name": "Favor and Protection Ring",
    "type": "accessory",
    "description": "A ring symbolizing the favor and protection of the goddess Fina, known in legend to possess 'fateful beauty'. This ring boosts its wearer's HP, stamina, and max equipment load, but breaks if ever removed.",
    "equippable": true,
    "rarity": 5,
    "defense": 10,
    "mana_cost": 20
  }
]
```

#### Get item from current Item Pool

- **Endpoint**: `GET /items`
- **Description**: Gets an item from the current item pool.
- **Path Parameters**:
- **Successful Response(`200 ok`)**: returns an object of the item from the item pool.

```JSON
  {
    "id": 8,
    "name": "Favor and Protection Ring",
    "type": "accessory",
    "description": "A ring symbolizing the favor and protection of the goddess Fina, known in legend to possess 'fateful beauty'. This ring boosts its wearer's HP, stamina, and max equipment load, but breaks if ever removed.",
    "equippable": true,
    "rarity": 5,
    "defense": 10,
    "mana_cost": 20
  }
```
