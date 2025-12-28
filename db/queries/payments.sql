-- name: CreatePaymentMethod :one
INSERT INTO payment_methods (account_id, name, kind, bank_name, credit_limit, closing_day, due_day, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: ListPaymentMethods :many
SELECT *
FROM payment_methods
WHERE (sqlc.narg('is_active')::boolean IS NULL OR is_active = sqlc.narg('is_active'))
ORDER BY name;

-- name: GetPaymentMethod :one
SELECT *
FROM payment_methods
WHERE payment_method_id = $1;

-- name: UpdatePaymentMethod :one
UPDATE payment_methods
SET account_id = $2,
    name = $3,
    kind = $4,
    bank_name = $5,
    credit_limit = $6,
    closing_day = $7,
    due_day = $8,
    is_active = $9
WHERE payment_method_id = $1
RETURNING *;

-- name: GetInvoiceEntries :many
SELECT
    t.transaction_id,
    t.occurred_at,
    t.description AS title,
    CASE
        WHEN p.side = 'CREDIT' THEN p.amount
        WHEN p.side = 'DEBIT' THEN -p.amount
        ELSE 0
    END AS amount,
    COALESCE(c.name, '') AS category_name
FROM transactions t
JOIN postings p ON p.transaction_id = t.transaction_id
LEFT JOIN postings pc ON pc.transaction_id = t.transaction_id AND pc.category_id IS NOT NULL
LEFT JOIN categories c ON c.category_id = pc.category_id
WHERE p.account_id = (SELECT account_id FROM payment_methods WHERE payment_method_id = $1)
  AND DATE_TRUNC('month', t.occurred_at) = DATE_TRUNC('month', $2::date)
ORDER BY t.occurred_at ASC, t.transaction_id ASC;

-- name: GetOutstandingAmount :one
SELECT COALESCE(
  SUM(
    CASE
      WHEN p.side = 'CREDIT' THEN p.amount
      WHEN p.side = 'DEBIT' THEN -p.amount
      ELSE 0
    END
  ),
  0
)::float
FROM transactions t
JOIN postings p ON p.transaction_id = t.transaction_id
WHERE p.account_id = (SELECT account_id FROM payment_methods WHERE payment_method_id = $1)
  AND DATE_TRUNC('month', t.occurred_at) >= DATE_TRUNC('month', $2::date);
