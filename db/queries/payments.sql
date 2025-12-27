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
    i.installment_plan_item_id,
    i.due_date,
    CASE
        WHEN p.installment_count IS NOT NULL AND p.installment_count > 1 THEN
            p.description || ' (' || i.sequence || '/' || p.installment_count || ')'
        ELSE
            p.description
    END AS title,
    (i.amount + i.extra_amount) AS amount,
    COALESCE(c.name, '') AS category_name
FROM installment_plan_items i
JOIN installment_plans p ON p.installment_plan_id = i.installment_plan_id
LEFT JOIN categories c ON c.category_id = p.category_id
WHERE p.payment_method_id = $1
  AND DATE_TRUNC('month', i.due_date) = DATE_TRUNC('month', $2::date)
  AND p.is_active = true
  AND i.status <> 'CANCELED'
ORDER BY i.due_date ASC, i.sequence ASC;

-- name: GetOutstandingAmount :one
SELECT COALESCE(SUM(i.amount + i.extra_amount), 0)::float
FROM installment_plan_items i
JOIN installment_plans p ON p.installment_plan_id = i.installment_plan_id
WHERE p.payment_method_id = $1
  AND DATE_TRUNC('month', i.due_date) >= DATE_TRUNC('month', $2::date)
  AND p.is_active = true
  AND i.status <> 'CANCELED'
  AND i.is_paid = false;
