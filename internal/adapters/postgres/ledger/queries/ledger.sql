-- Ledgers

-- name: ListLedgersForUser :many
SELECT l.id, l.owner_user_id, l.name, l.currency_code, l.created_at, l.updated_at,
  CASE WHEN l.owner_user_id = $1 THEN 'owner' ELSE lm.role END AS role
FROM ledgers l
LEFT JOIN ledger_members lm ON lm.ledger_id = l.id AND lm.user_id = $1
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
LEFT JOIN ledger_members lm ON lm.ledger_id = l.id AND lm.user_id = $1
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
LEFT JOIN ledger_members lm ON lm.ledger_id = l.id AND lm.user_id = $1
WHERE l.id = $2 AND (l.owner_user_id = $1 OR lm.user_id = $1);

-- Members

-- name: ListMembers :many
SELECT ledger_id, user_id, role, created_at, updated_at
FROM ledger_members
WHERE ledger_id = $1
ORDER BY created_at;

-- name: AddMember :one
INSERT INTO ledger_members (ledger_id, user_id, role, updated_at)
VALUES ($1, $2, $3, $4)
RETURNING ledger_id, user_id, role, created_at, updated_at;

-- name: UpdateMemberRole :one
UPDATE ledger_members
SET role = $3, updated_at = $4
WHERE ledger_id = $1 AND user_id = $2
RETURNING ledger_id, user_id, role, created_at, updated_at;

-- name: RemoveMember :exec
DELETE FROM ledger_members
WHERE ledger_id = $1 AND user_id = $2;
