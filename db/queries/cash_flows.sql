-- name: CreateCashFlow :one
INSERT INTO cash_flows (
  date,
  category_id,
  direction,
  title,
  amount
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING cash_flow_id, date, category_id, direction, title, amount;

-- name: ListCashFlowsByMonth :many
SELECT
  cf.cash_flow_id,
  cf.date,
  cf.category_id,
  cf.direction,
  cf.title,
  cf.amount,
  fc.name AS category_name
FROM cash_flows cf
JOIN flow_categories fc ON fc.category_id = cf.category_id
WHERE date_trunc('month', cf.date) = date_trunc('month', $1::date)
ORDER BY cf.date, cf.cash_flow_id;

