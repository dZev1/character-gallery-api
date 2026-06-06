-- name: AddItemToCharacter :one
INSERT INTO inventory (character_id, item_id, quantity, is_equipped)
VALUES ($1, $2, $3, false)
ON CONFLICT (character_id, item_id)
DO UPDATE SET quantity = inventory.quantity + EXCLUDED.quantity
RETURNING *;

-- name: RemoveItemFromCharacter :exec
WITH updated AS (
    UPDATE inventory
    SET quantity = inventory.quantity - $3
    WHERE inventory.character_id = $1 AND inventory.item_id = $2 AND inventory.quantity > $3
    RETURNING character_id
)
DELETE FROM inventory
WHERE inventory.character_id = $1 AND inventory.item_id = $2
  AND NOT EXISTS (SELECT 1 FROM updated);

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
       inv.quantity,
       inv.is_equipped
FROM inventory inv
JOIN items i ON i.id = inv.item_id
WHERE inv.character_id = $1;
