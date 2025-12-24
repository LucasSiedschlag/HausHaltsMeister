-- name: CreateCashFlow :one
INSERT INTO cash_flows (
  date,
  category_id,
  direction,
  title,
  amount,
  is_fixed
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING cash_flow_id, date, category_id, direction, title, amount, is_fixed;

-- name: ListCashFlowsByMonth :many
SELECT
  cf.cash_flow_id,
  cf.date,
  cf.category_id,
  cf.direction,
  cf.title,
  cf.amount,
  cf.is_fixed,
  fc.name AS category_name
FROM cash_flows cf
JOIN flow_categories fc ON fc.category_id = cf.category_id
WHERE date_trunc('month', cf.date) = date_trunc('month', $1::date)
ORDER BY cf.date, cf.cash_flow_id;

-- name: GetMonthlySummary :one
SELECT
  SUM(CASE WHEN direction = 'IN' THEN amount ELSE 0 END)::float AS total_income,
  SUM(CASE WHEN direction = 'OUT' THEN amount ELSE 0 END)::float AS total_expense
FROM cash_flows
WHERE date_trunc('month', date) = date_trunc('month', $1::date);

-- name: GetCategorySummary :many
SELECT
  fc.name,
  fc.direction,
  SUM(cf.amount)::float AS total_amount
FROM cash_flows cf
JOIN flow_categories fc ON fc.category_id = cf.category_id
WHERE date_trunc('month', cf.date) = date_trunc('month', $1::date)
GROUP BY fc.name, fc.direction
ORDER BY total_amount DESC;
