-- name: CreateLedgerAccount :one
INSERT INTO accounts (name, type, currency, is_active)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListLedgerAccounts :many
SELECT *
FROM accounts
ORDER BY name;

-- name: CreateLedgerTransaction :one
INSERT INTO transactions (occurred_at, description, reference, notes)
VALUES ($1, $2, $3, $4)
RETURNING *;

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

-- name: ListLedgerPostingsByTransaction :many
SELECT *
FROM postings
WHERE transaction_id = $1
ORDER BY posting_id;
