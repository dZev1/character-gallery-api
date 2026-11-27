-- name: AddItemToCharacter :one
INSERT INTO inventory (character_id, item_id, quantity, is_equipped)
VALUES ($1, $2, $3, false)
ON CONFLICT (character_id, item_id)
DO UPDATE SET quantity = inventory.quantity + EXCLUDED.quantity
RETURNING *;

-- name: RemoveItemFromCharacter :one
WITH updated AS (
    UPDATE inventory
    SET quantity = inventory.quantity - $3
    WHERE inventory.character_id = $1 AND inventory.item_id = $2 AND inventory.quantity > $3
    RETURNING *
), deleted AS (
    DELETE FROM inventory
    WHERE inventory.character_id = $1 AND inventory.item_id = $2
      AND inventory.quantity = $3
    RETURNING *
)
SELECT * FROM updated
UNION ALL
SELECT * FROM deleted;

-- name: GetCharacterInventory :many
SELECT i.id             AS item_id,
       i.name,
       i.type,
       i.description,
       i.equippable,
       i.rarity,
       i.damage,
       i.defense,
       i.heal_amount,
       i.mana_cost,
       i.duration,
       i.cooldown,
       i.capacity,
       i.image_url,
       inv.quantity,
       inv.is_equipped
FROM inventory inv
JOIN items i ON i.id = inv.item_id
WHERE inv.character_id = $1;

-- name: GetCharacterEquippedItems :many
SELECT i.id             AS item_id,
       i.name,
       i.type,
       i.description,
       i.equippable,
       i.rarity,
       i.damage,
       i.defense,
       i.heal_amount,
       i.mana_cost,
       i.duration,
       i.cooldown,
       i.capacity,
       i.image_url,
       inv.quantity,
       inv.is_equipped
FROM inventory inv
JOIN items i ON i.id = inv.item_id
WHERE inv.character_id = $1 AND inv.is_equipped = TRUE;

-- name: SetInventoryItemEquipped :one
UPDATE inventory
SET is_equipped = $3
WHERE character_id = $1 AND item_id = $2
RETURNING *;

-- name: UnequipItemsByType :exec
UPDATE inventory
SET is_equipped = FALSE
FROM items i
WHERE inventory.item_id = i.id
  AND inventory.character_id = $1
  AND i.type = $2
  AND inventory.item_id <> $3;
