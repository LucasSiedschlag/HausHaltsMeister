-- Journal: transactions + entries

-- name: CreateTransaction :one
INSERT INTO transactions (ledger_id, occurred_at, description, notes, created_by_user_id, credit_card_id, investment_action, external_source, external_id, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, ledger_id, occurred_at, description, notes, created_by_user_id, credit_card_id, investment_action, created_at, updated_at;

-- name: UpdateTransaction :one
UPDATE transactions
SET occurred_at = $3, description = $4, notes = $5, updated_at = $6
WHERE ledger_id = $1 AND id = $2
RETURNING id, ledger_id, occurred_at, description, notes, created_by_user_id, credit_card_id, investment_action, created_at, updated_at;

-- name: DeleteTransaction :exec
DELETE FROM transactions
WHERE ledger_id = $1 AND id = $2;

-- name: CreateEntry :one
INSERT INTO entries (ledger_id, transaction_id, account_id, category_id, kind, amount_cents, memo, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, ledger_id, transaction_id, account_id, category_id, kind, amount_cents, memo, created_at, updated_at;

-- name: DeleteEntriesByTransaction :exec
DELETE FROM entries
WHERE ledger_id = $1 AND transaction_id = $2;

-- name: GetTransactionWithEntries :many
SELECT t.id, t.ledger_id, t.occurred_at, t.description, t.notes, t.created_by_user_id, t.credit_card_id, t.investment_action, t.created_at, t.updated_at,
  e.id, e.ledger_id, e.transaction_id, e.account_id, e.category_id, e.kind, e.amount_cents, e.memo, e.created_at, e.updated_at
FROM transactions t
JOIN entries e ON e.transaction_id = t.id
WHERE t.ledger_id = $1 AND t.id = $2
ORDER BY e.created_at ASC;

-- name: GetTransactionByExternal :one
SELECT id, ledger_id, occurred_at, description, notes, created_by_user_id, credit_card_id, investment_action, external_source, external_id, created_at, updated_at
FROM transactions
WHERE ledger_id = $1 AND external_source = $2 AND external_id = $3;

-- name: ListEntriesByTransaction :many
SELECT id, ledger_id, transaction_id, account_id, category_id, kind, amount_cents, memo, created_at, updated_at
FROM entries
WHERE ledger_id = $1 AND transaction_id = $2
ORDER BY created_at ASC;

-- name: ListTransactionIDs :many
SELECT t.id, t.occurred_at
FROM transactions t
WHERE t.ledger_id = $1
  AND ($2::timestamptz IS NULL OR t.occurred_at >= $2)
  AND ($3::timestamptz IS NULL OR t.occurred_at <= $3)
  AND ($4::uuid IS NULL OR EXISTS (
    SELECT 1 FROM entries e
    WHERE e.transaction_id = t.id AND e.account_id = $4
  ))
  AND ($5::uuid IS NULL OR EXISTS (
    SELECT 1 FROM entries e
    WHERE e.transaction_id = t.id AND e.category_id = $5
  ))
  AND ($6::text IS NULL OR (t.description ILIKE $6 OR t.notes ILIKE $6))
  AND ($7::timestamptz IS NULL OR (t.occurred_at < $7 OR (t.occurred_at = $7 AND t.id < $8)))
ORDER BY t.occurred_at DESC, t.id DESC
LIMIT $9;

-- name: ListTransactionsWithEntries :many
SELECT t.id, t.ledger_id, t.occurred_at, t.description, t.notes, t.created_by_user_id, t.credit_card_id, t.investment_action, t.created_at, t.updated_at,
  e.id, e.ledger_id, e.transaction_id, e.account_id, e.category_id, e.kind, e.amount_cents, e.memo, e.created_at, e.updated_at
FROM transactions t
JOIN entries e ON e.transaction_id = t.id
WHERE t.id = ANY($1::uuid[])
ORDER BY t.occurred_at DESC, t.id DESC, e.created_at ASC;

-- name: CheckTransactionReferences :one
SELECT
  EXISTS (SELECT 1 FROM installments WHERE posted_transaction_id = $1) AS has_installment_post,
  EXISTS (SELECT 1 FROM credit_card_statements WHERE payment_transaction_id = $1) AS has_statement_payment;
