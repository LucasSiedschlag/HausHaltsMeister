-- name: CreateInstallmentPlan :one
INSERT INTO installment_plans (
  description,
  plan_type,
  total_amount,
  installment_count,
  installment_amount,
  start_date,
  payment_method_id,
  account_id,
  category_id,
  starts_on_current_invoice
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING installment_plan_id, description, plan_type, total_amount, installment_count, installment_amount, start_date, payment_method_id, account_id, category_id, party_id, interest_rate, recurrence_interval_months, is_active, created_at, updated_at, interest_rate_unit, starts_on_current_invoice;

-- name: CreateInstallmentPlanItem :one
INSERT INTO installment_plan_items (
  installment_plan_id,
  sequence,
  due_date,
  amount,
  status,
  transaction_id,
  extra_amount,
  is_paid,
  paid_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING installment_plan_item_id, installment_plan_id, sequence, due_date, amount, status, transaction_id, created_at, updated_at, extra_amount, is_paid, paid_at;

-- name: ListInstallmentPlanItemsByPlan :many
SELECT
  installment_plan_item_id,
  installment_plan_id,
  sequence,
  due_date,
  amount,
  status,
  transaction_id,
  created_at,
  updated_at,
  extra_amount,
  is_paid,
  paid_at
FROM installment_plan_items
WHERE installment_plan_id = $1
ORDER BY sequence ASC;
