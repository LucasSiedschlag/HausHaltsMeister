-- name: CreatePerson :one
INSERT INTO parties (name, notes, kind, is_active)
VALUES ($1, $2, 'PERSON', true)
RETURNING party_id, name, notes;

-- name: UpdatePerson :one
UPDATE parties
SET name = $2,
    notes = $3
WHERE party_id = $1
  AND kind = 'PERSON'
RETURNING party_id, name, notes;

-- name: DeletePerson :exec
DELETE FROM parties
WHERE party_id = $1
  AND kind = 'PERSON';

-- name: CountCasesByPerson :one
SELECT COUNT(*)
FROM installment_plans
WHERE party_id = $1;

-- name: ListPersons :many
SELECT party_id, name, notes
FROM parties
WHERE kind = 'PERSON'
ORDER BY name;

-- name: GetPerson :one
SELECT party_id, name, notes
FROM parties
WHERE party_id = $1
  AND kind = 'PERSON';

-- name: GetPersonBalance :one
SELECT (
  COALESCE(
    (
      SELECT SUM(i.amount + i.extra_amount)
      FROM installment_plan_items i
      JOIN installment_plans p ON p.installment_plan_id = i.installment_plan_id
      WHERE p.party_id = $1
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
  party_id,
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
  recurrence_interval_months,
  starts_on_current_invoice
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, true)
RETURNING *;

-- name: UpdatePicuinhaCase :one
UPDATE installment_plans
SET party_id = $2,
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
RETURNING *;

-- name: DeletePicuinhaCase :exec
DELETE FROM installment_plans
WHERE installment_plan_id = $1;

-- name: GetPicuinhaCase :one
SELECT *
FROM installment_plans
WHERE installment_plan_id = $1
  AND party_id IS NOT NULL;

-- name: ListPicuinhaCasesByPerson :many
SELECT
  p.*,
  COUNT(i.installment_plan_item_id) AS installments_total,
  COUNT(i.installment_plan_item_id) FILTER (WHERE i.is_paid) AS installments_paid,
  COALESCE(SUM(i.amount + i.extra_amount) FILTER (WHERE i.is_paid), 0)::decimal AS amount_paid,
  COALESCE(SUM(i.amount + i.extra_amount) FILTER (WHERE NOT i.is_paid), 0)::decimal AS amount_remaining
FROM installment_plans p
LEFT JOIN installment_plan_items i ON i.installment_plan_id = p.installment_plan_id
WHERE p.party_id = $1
GROUP BY p.installment_plan_id
ORDER BY p.created_at DESC;

-- name: CreatePicuinhaCaseInstallment :one
INSERT INTO installment_plan_items (
  installment_plan_id,
  sequence,
  due_date,
  amount,
  extra_amount,
  is_paid,
  paid_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdatePicuinhaCaseInstallment :one
UPDATE installment_plan_items
SET amount = $2,
    extra_amount = $3,
    is_paid = $4,
    paid_at = $5
WHERE installment_plan_item_id = $1
RETURNING *;

-- name: GetPicuinhaCaseInstallment :one
SELECT *
FROM installment_plan_items
WHERE installment_plan_item_id = $1;

-- name: ListPicuinhaCaseInstallments :many
SELECT *
FROM installment_plan_items
WHERE installment_plan_id = $1
ORDER BY due_date ASC, sequence ASC;
