-- name: CreateCashFlowEntry :one
INSERT INTO cashflow_entries (
  transaction_id,
  category_id,
  payment_method_id,
  installment_plan_id,
  installment_plan_item_id,
  direction,
  title,
  amount,
  is_fixed,
  occurred_at,
  reversal_of_entry_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: UpdateCashFlowEntry :one
UPDATE cashflow_entries
SET transaction_id = $2,
    category_id = $3,
    payment_method_id = $4,
    installment_plan_id = $5,
    installment_plan_item_id = $6,
    direction = $7,
    title = $8,
    amount = $9,
    is_fixed = $10,
    occurred_at = $11,
    reversal_of_entry_id = $12,
    updated_at = now()
WHERE cashflow_entry_id = $1
RETURNING *;

-- name: DeleteCashFlowEntry :exec
DELETE FROM cashflow_entries
WHERE cashflow_entry_id = $1;

-- name: GetCashFlowEntry :one
SELECT
  ce.*,
  c.name AS category_name,
  pm.name AS payment_method_name
FROM cashflow_entries ce
JOIN categories c ON c.category_id = ce.category_id
JOIN payment_methods pm ON pm.payment_method_id = ce.payment_method_id
WHERE ce.cashflow_entry_id = $1;

-- name: ListCashFlowEntriesByMonth :many
WITH picuinha_category AS (
  SELECT category_id
  FROM categories
  WHERE name = 'Picuinhas'
  LIMIT 1
),
combined AS (
  SELECT
    ce.cashflow_entry_id,
    ce.transaction_id,
    ce.category_id,
    ce.payment_method_id,
    ce.installment_plan_id,
    ce.installment_plan_item_id,
    ce.direction,
    ce.title,
    ce.amount,
    ce.is_fixed,
    ce.occurred_at,
    ce.reversal_of_entry_id,
    ce.created_at,
    ce.updated_at
  FROM cashflow_entries ce
  WHERE date_trunc('month', ce.occurred_at) = date_trunc('month', $1::date)

  UNION ALL

  SELECT
    -i.installment_plan_item_id AS cashflow_entry_id,
    COALESCE(i.transaction_id, 0) AS transaction_id,
    COALESCE(p.category_id, (SELECT category_id FROM picuinha_category)) AS category_id,
    COALESCE(p.payment_method_id, 0) AS payment_method_id,
    p.installment_plan_id,
    i.installment_plan_item_id,
    c.direction,
    CASE
      WHEN p.installment_count IS NOT NULL THEN
        CONCAT(p.description, ' (', i.sequence, '/', p.installment_count, ')')
      ELSE
        p.description
    END AS title,
    (i.amount + i.extra_amount) AS amount,
    false AS is_fixed,
    i.due_date AS occurred_at,
    NULL::int AS reversal_of_entry_id,
    i.created_at,
    i.updated_at
  FROM installment_plan_items i
  JOIN installment_plans p ON p.installment_plan_id = i.installment_plan_id
  LEFT JOIN cashflow_entries ce ON ce.installment_plan_item_id = i.installment_plan_item_id
  JOIN categories c ON c.category_id = COALESCE(p.category_id, (SELECT category_id FROM picuinha_category))
  WHERE ce.cashflow_entry_id IS NULL
    AND date_trunc('month', i.due_date) = date_trunc('month', $1::date)
)
SELECT
  combined.*,
  c.name AS category_name,
  COALESCE(pm.name, 'Sem meio') AS payment_method_name
FROM combined
JOIN categories c ON c.category_id = combined.category_id
LEFT JOIN payment_methods pm ON pm.payment_method_id = combined.payment_method_id
WHERE (sqlc.narg('direction')::varchar IS NULL OR combined.direction = sqlc.narg('direction'))
  AND (sqlc.narg('is_fixed')::boolean IS NULL OR combined.is_fixed = sqlc.narg('is_fixed'))
ORDER BY combined.occurred_at, combined.cashflow_entry_id;

-- name: GetMonthlySummary :one
SELECT
  COALESCE(
    SUM(
      CASE
        WHEN c.direction = 'IN' AND p.side = 'CREDIT' THEN p.amount
        WHEN c.direction = 'IN' AND p.side = 'DEBIT' THEN -p.amount
        ELSE 0
      END
    ),
    0
  )::float AS total_income,
  COALESCE(
    SUM(
      CASE
        WHEN c.direction = 'OUT' AND p.side = 'DEBIT' THEN p.amount
        WHEN c.direction = 'OUT' AND p.side = 'CREDIT' THEN -p.amount
        ELSE 0
      END
    ),
    0
  )::float AS total_expense
FROM postings p
JOIN transactions t ON t.transaction_id = p.transaction_id
JOIN categories c ON c.category_id = p.category_id
WHERE date_trunc('month', t.occurred_at) = date_trunc('month', $1::date)
  AND c.is_budget_relevant = true;

-- name: GetCategorySummary :many
SELECT
  c.name,
  c.direction,
  COALESCE(
    SUM(
      CASE
        WHEN c.direction = 'IN' AND p.side = 'CREDIT' THEN p.amount
        WHEN c.direction = 'IN' AND p.side = 'DEBIT' THEN -p.amount
        WHEN c.direction = 'OUT' AND p.side = 'DEBIT' THEN p.amount
        WHEN c.direction = 'OUT' AND p.side = 'CREDIT' THEN -p.amount
        ELSE 0
      END
    ),
    0
  )::float AS total_amount
FROM postings p
JOIN transactions t ON t.transaction_id = p.transaction_id
JOIN categories c ON c.category_id = p.category_id
WHERE date_trunc('month', t.occurred_at) = date_trunc('month', $1::date)
  AND c.is_budget_relevant = true
GROUP BY c.name, c.direction
ORDER BY total_amount DESC;
