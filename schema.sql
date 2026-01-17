CREATE TABLE
IF NOT EXISTS "item"
(
  "id" BIGSERIAL PRIMARY KEY,
  "name" text,
  "type" text CHECK
(type IN
('weapon', 'armor', 'potion', 'spell', 'explosive', 'misc')),
  "description" text,
  "equippable" boolean,
  "rarity" int,
  "damage" int,
  "defense" int,
  "heal_amount" int,
  "mana_cost" int,
  "duration" text
);

CREATE TABLE
IF NOT EXISTS "inventory"
(
  "character_id" bigserial,
  "item_id" integer,
  "quantity" int NOT NULL,
  "is_equipped" boolean DEFAULT FALSE
);

CREATE TABLE
IF NOT EXISTS "characters"
(
  "id" BIGSERIAL PRIMARY KEY,
  "name" TEXT NOT NULL,
  "body_type" TEXT NOT NULL CHECK
(body_type IN
('type_a', 'type_b')),
  "species" TEXT NOT NULL CHECK
(species IN
(
                    'aasimar', 'dragonborn', 'dwarf', 'elf', 'gnome', 
                    'goliath', 'halfling', 'human', 'orc', 'tiefling'
                )),
  "class" TEXT NOT NULL CHECK
(class IN
(
                    'barbarian', 'bard', 'cleric', 'druid', 'fighter', 
                    'monk', 'paladin', 'ranger', 'rogue', 'sorcerer', 
                    'warlock', 'wizard'
                ))
);

CREATE TABLE
IF NOT EXISTS "stats"
(
  "id" BIGSERIAL PRIMARY KEY,
  "strength" SMALLINT NOT NULL CHECK
(strength BETWEEN 1 AND 99),
  "dexterity" SMALLINT NOT NULL CHECK
(dexterity BETWEEN 1 AND 99),
  "constitution" SMALLINT NOT NULL CHECK
(constitution BETWEEN 1 AND 99),
  "intelligence" SMALLINT NOT NULL CHECK
(intelligence BETWEEN 1 AND 99),
  "wisdom" SMALLINT NOT NULL CHECK
(wisdom BETWEEN 1 AND 99),
  "charisma" SMALLINT NOT NULL CHECK
(charisma BETWEEN 1 AND 99)
);

CREATE TABLE
IF NOT EXISTS "customizations"
(
  "id" BIGSERIAL PRIMARY KEY,
  "hair" SMALLINT NOT NULL CHECK
(hair BETWEEN 0 AND 30),
  "face" SMALLINT NOT NULL CHECK
(face BETWEEN 0 AND 30),
  "shirt" SMALLINT NOT NULL CHECK
(shirt BETWEEN 0 AND 30),
  "pants" SMALLINT NOT NULL CHECK
(pants BETWEEN 0 AND 30),
  "shoes" SMALLINT NOT NULL CHECK
(shoes BETWEEN 0 AND 30)
);

ALTER TABLE "inventory" ADD FOREIGN KEY ("item_id") REFERENCES "item" ("id") ON DELETE CASCADE;

ALTER TABLE "stats" ADD FOREIGN KEY ("id") REFERENCES "characters" ("id") ON DELETE CASCADE;

ALTER TABLE "customizations" ADD FOREIGN KEY ("id") REFERENCES "characters" ("id") ON DELETE CASCADE;

ALTER TABLE "inventory" ADD FOREIGN KEY ("character_id") REFERENCES "characters" ("id") ON DELETE CASCADE;
