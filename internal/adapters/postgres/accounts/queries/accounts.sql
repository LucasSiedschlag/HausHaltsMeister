-- Accounts

-- name: ListAccountsByLedger :many
SELECT id, ledger_id, name, type, is_active, created_at, updated_at
FROM accounts
WHERE ledger_id = $1
ORDER BY created_at;

-- name: GetAccountByID :one
SELECT id, ledger_id, name, type, is_active, created_at, updated_at
FROM accounts
WHERE ledger_id = $1 AND id = $2;

-- name: CreateAccount :one
INSERT INTO accounts (ledger_id, name, type, is_active, updated_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, ledger_id, name, type, is_active, created_at, updated_at;

-- name: UpdateAccount :one
UPDATE accounts
SET name = $3, type = $4, is_active = $5, updated_at = $6
WHERE ledger_id = $1 AND id = $2
RETURNING id, ledger_id, name, type, is_active, created_at, updated_at;

-- name: DeactivateAccount :one
UPDATE accounts
SET is_active = false, updated_at = $3
WHERE ledger_id = $1 AND id = $2
RETURNING id, ledger_id, name, type, is_active, created_at, updated_at;

-- name: FindAccountByType :one
SELECT id
FROM accounts
WHERE ledger_id = $1 AND type = $2 AND is_active = true
ORDER BY created_at ASC
LIMIT 1;
