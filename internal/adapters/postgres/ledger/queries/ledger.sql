-- Ledgers

-- name: ListLedgersForUser :many
SELECT l.id, l.owner_user_id, l.name, l.currency_code, l.created_at, l.updated_at,
  CASE WHEN l.owner_user_id = $1 THEN 'owner' ELSE lm.role END AS role
FROM ledgers l
LEFT JOIN ledger_members lm ON lm.ledger_id = l.id AND lm.user_id = $1 AND lm.removed_at IS NULL
WHERE l.owner_user_id = $1 OR lm.user_id = $1
ORDER BY l.created_at DESC;

-- name: CreateLedger :one
INSERT INTO ledgers (owner_user_id, name, currency_code)
VALUES ($1, $2, $3)
RETURNING id, owner_user_id, name, currency_code, created_at, updated_at;

-- name: GetLedgerForUser :one
SELECT l.id, l.owner_user_id, l.name, l.currency_code, l.created_at, l.updated_at,
  CASE WHEN l.owner_user_id = $1 THEN 'owner' ELSE lm.role END AS role
FROM ledgers l
LEFT JOIN ledger_members lm ON lm.ledger_id = l.id AND lm.user_id = $1 AND lm.removed_at IS NULL
WHERE l.id = $2 AND (l.owner_user_id = $1 OR lm.user_id = $1);

-- name: GetLedgerByID :one
SELECT id, owner_user_id, name, currency_code, created_at, updated_at
FROM ledgers
WHERE id = $1;

-- name: UpdateLedger :one
UPDATE ledgers
SET name = $2, updated_at = $3
WHERE id = $1
RETURNING id, owner_user_id, name, currency_code, created_at, updated_at;

-- name: DeleteLedger :exec
DELETE FROM ledgers
WHERE id = $1;

-- name: LedgerExists :one
SELECT EXISTS (SELECT 1 FROM ledgers WHERE id = $1) AS exists;

-- name: GetLedgerRole :one
SELECT CASE WHEN l.owner_user_id = $1 THEN 'owner' ELSE lm.role END AS role
FROM ledgers l
LEFT JOIN ledger_members lm ON lm.ledger_id = l.id AND lm.user_id = $1 AND lm.removed_at IS NULL
WHERE l.id = $2 AND (l.owner_user_id = $1 OR lm.user_id = $1);

-- Members

-- name: ListMembers :many
SELECT lm.ledger_id, lm.user_id, lm.role, lm.created_at, lm.updated_at,
  COALESCE(u.display_name, '') AS display_name,
  u.email,
  u.avatar_url
FROM ledger_members lm
JOIN users u ON u.id = lm.user_id
WHERE lm.ledger_id = $1 AND lm.removed_at IS NULL
ORDER BY lm.created_at;

-- name: AddMember :one
WITH upserted AS (
  INSERT INTO ledger_members (ledger_id, user_id, role, updated_at)
  VALUES ($1, $2, $3, $4)
  ON CONFLICT (ledger_id, user_id) DO UPDATE
  SET role = EXCLUDED.role,
    removed_at = NULL,
    updated_at = EXCLUDED.updated_at
  WHERE ledger_members.removed_at IS NOT NULL
  RETURNING ledger_id, user_id, role, created_at, updated_at
)
SELECT upserted.ledger_id, upserted.user_id, upserted.role, upserted.created_at, upserted.updated_at,
  COALESCE(u.display_name, '') AS display_name,
  u.email,
  u.avatar_url
FROM upserted
JOIN users u ON u.id = upserted.user_id;

-- name: UpdateMemberRole :one
WITH updated AS (
  UPDATE ledger_members
  SET role = $3, updated_at = $4
  WHERE ledger_id = $1 AND user_id = $2 AND removed_at IS NULL
  RETURNING ledger_id, user_id, role, created_at, updated_at
)
SELECT updated.ledger_id, updated.user_id, updated.role, updated.created_at, updated.updated_at,
  COALESCE(u.display_name, '') AS display_name,
  u.email,
  u.avatar_url
FROM updated
JOIN users u ON u.id = updated.user_id;

-- name: RemoveMember :exec
UPDATE ledger_members
SET removed_at = now(), updated_at = now()
WHERE ledger_id = $1 AND user_id = $2 AND removed_at IS NULL;
