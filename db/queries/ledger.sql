-- name: CreateLedgerAccount :one
INSERT INTO accounts (name, type, currency, is_active)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateLedgerAccount :one
UPDATE accounts
SET name = $2,
    type = $3,
    is_active = $4,
    updated_at = now()
WHERE account_id = $1
RETURNING *;

-- name: ListLedgerAccounts :many
SELECT *
FROM accounts
ORDER BY name;

-- name: CreateLedgerTransaction :one
INSERT INTO transactions (occurred_at, description, reference, notes)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateLedgerTransaction :one
UPDATE transactions
SET occurred_at = $2,
    description = $3,
    reference = $4,
    notes = $5,
    updated_at = now()
WHERE transaction_id = $1
RETURNING *;

-- name: DeleteLedgerTransaction :exec
DELETE FROM transactions
WHERE transaction_id = $1;

-- name: ListLedgerTransactionsByMonth :many
SELECT *
FROM transactions
WHERE occurred_at >= $1
  AND occurred_at < ($1 + interval '1 month')
ORDER BY occurred_at DESC, transaction_id DESC;

-- name: CreateLedgerPosting :one
INSERT INTO postings (transaction_id, account_id, category_id, party_id, side, amount, memo)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateLedgerPosting :one
UPDATE postings
SET account_id = $2,
    category_id = $3,
    party_id = $4,
    side = $5,
    amount = $6,
    memo = $7
WHERE posting_id = $1
RETURNING *;

-- name: ListLedgerPostingsByTransaction :many
SELECT *
FROM postings
WHERE transaction_id = $1
ORDER BY posting_id;
