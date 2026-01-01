-- Investments support queries

-- name: FindAccountByType :one
SELECT id
FROM accounts
WHERE ledger_id = $1 AND type = $2 AND is_active = true
ORDER BY created_at ASC
LIMIT 1;

-- name: FindCategoryByName :one
SELECT id
FROM categories
WHERE ledger_id = $1 AND name = $2
LIMIT 1;

-- name: GetCategoryDirection :one
SELECT direction
FROM categories
WHERE ledger_id = $1 AND id = $2;

-- name: SumByCategory :one
SELECT COALESCE(SUM(e.amount_cents), 0)
FROM entries e
JOIN transactions t ON t.id = e.transaction_id
WHERE e.ledger_id = $1
  AND e.category_id = $2
  AND t.occurred_at >= $3 AND t.occurred_at <= $4;
