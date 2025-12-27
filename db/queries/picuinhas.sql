-- name: CreatePerson :one
INSERT INTO picuinha_persons (name, notes)
VALUES ($1, $2)
RETURNING person_id, name, notes;

-- name: UpdatePerson :one
UPDATE picuinha_persons
SET name = $2,
    notes = $3
WHERE person_id = $1
RETURNING person_id, name, notes;

-- name: DeletePerson :exec
DELETE FROM picuinha_persons
WHERE person_id = $1;

-- name: CountCasesByPerson :one
SELECT COUNT(*)
FROM installment_plans
WHERE person_id = $1;

-- name: ListPersons :many
SELECT person_id, name, notes
FROM picuinha_persons
ORDER BY name;

-- name: GetPerson :one
SELECT person_id, name, notes
FROM picuinha_persons
WHERE person_id = $1;

-- name: GetPersonBalance :one
SELECT (
  COALESCE(
    (
      SELECT SUM(i.amount + i.extra_amount)
      FROM installment_plan_items i
      JOIN installment_plans p ON p.installment_plan_id = i.installment_plan_id
      WHERE p.person_id = $1
        AND i.is_paid = false
        AND (
          p.plan_type <> 'RECURRING'
          OR DATE_TRUNC('month', i.due_date) <= DATE_TRUNC('month', CURRENT_DATE)
        )
    ),
    0
  )
)::decimal;

-- name: CreatePicuinhaCase :one
INSERT INTO installment_plans (
  person_id,
  description,
  plan_type,
  total_amount,
  installment_count,
  installment_amount,
  start_date,
  payment_method_id,
  category_id,
  interest_rate,
  interest_rate_unit,
  recurrence_interval_months
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING installment_plan_id, description, total_amount, installment_count, installment_amount,
  start_date, payment_method_id, starts_on_current_invoice, plan_type, person_id,
  category_id, interest_rate, interest_rate_unit, recurrence_interval_months, created_at;

-- name: UpdatePicuinhaCase :one
UPDATE installment_plans
SET person_id = $2,
    description = $3,
    plan_type = $4,
    total_amount = $5,
    installment_count = $6,
    installment_amount = $7,
    start_date = $8,
    payment_method_id = $9,
    category_id = $10,
    interest_rate = $11,
    interest_rate_unit = $12,
    recurrence_interval_months = $13
WHERE installment_plan_id = $1
RETURNING installment_plan_id, description, total_amount, installment_count, installment_amount,
  start_date, payment_method_id, starts_on_current_invoice, plan_type, person_id,
  category_id, interest_rate, interest_rate_unit, recurrence_interval_months, created_at;

-- name: DeletePicuinhaCase :exec
DELETE FROM installment_plans
WHERE installment_plan_id = $1;

-- name: GetPicuinhaCase :one
SELECT installment_plan_id, description, total_amount, installment_count, installment_amount,
  start_date, payment_method_id, starts_on_current_invoice, plan_type, person_id,
  category_id, interest_rate, interest_rate_unit, recurrence_interval_months, created_at
FROM installment_plans
WHERE installment_plan_id = $1
  AND person_id IS NOT NULL;

-- name: ListPicuinhaCasesByPerson :many
SELECT
  p.installment_plan_id,
  p.person_id,
  p.description,
  p.plan_type,
  p.total_amount,
  p.installment_count,
  p.installment_amount,
  p.start_date,
  p.payment_method_id,
  p.category_id,
  p.interest_rate,
  p.interest_rate_unit,
  p.recurrence_interval_months,
  p.created_at,
  COUNT(i.installment_plan_item_id) AS installments_total,
  COUNT(i.installment_plan_item_id) FILTER (WHERE i.is_paid) AS installments_paid,
  COALESCE(SUM(i.amount + i.extra_amount) FILTER (WHERE i.is_paid), 0)::decimal AS amount_paid,
  COALESCE(SUM(i.amount + i.extra_amount) FILTER (WHERE NOT i.is_paid), 0)::decimal AS amount_remaining
FROM installment_plans p
LEFT JOIN installment_plan_items i ON i.installment_plan_id = p.installment_plan_id
WHERE p.person_id = $1
GROUP BY p.installment_plan_id
ORDER BY p.created_at DESC;

-- name: CreatePicuinhaCaseInstallment :one
INSERT INTO installment_plan_items (
  installment_plan_id,
  installment_number,
  due_date,
  amount,
  extra_amount,
  is_paid,
  paid_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING installment_plan_item_id, installment_plan_id, installment_number, due_date,
  amount, extra_amount, is_paid, paid_at, cash_flow_id;

-- name: UpdatePicuinhaCaseInstallment :one
UPDATE installment_plan_items
SET amount = $2,
    extra_amount = $3,
    is_paid = $4,
    paid_at = $5
WHERE installment_plan_item_id = $1
RETURNING installment_plan_item_id, installment_plan_id, installment_number, due_date,
  amount, extra_amount, is_paid, paid_at, cash_flow_id;

-- name: GetPicuinhaCaseInstallment :one
SELECT installment_plan_item_id, installment_plan_id, installment_number, due_date,
  amount, extra_amount, is_paid, paid_at, cash_flow_id
FROM installment_plan_items
WHERE installment_plan_item_id = $1;

-- name: ListPicuinhaCaseInstallments :many
SELECT installment_plan_item_id, installment_plan_id, installment_number, due_date,
  amount, extra_amount, is_paid, paid_at, cash_flow_id
FROM installment_plan_items
WHERE installment_plan_id = $1
ORDER BY due_date ASC, installment_number ASC;
