-- name: CreatePaymentMethod :one
INSERT INTO payment_methods (name, kind, bank_name, credit_limit, closing_day, due_day, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING payment_method_id, name, kind, bank_name, credit_limit, closing_day, due_day, is_active;

-- name: ListPaymentMethods :many
SELECT payment_method_id, name, kind, bank_name, credit_limit, closing_day, due_day, is_active
FROM payment_methods
WHERE (sqlc.narg('is_active')::boolean IS NULL OR is_active = sqlc.narg('is_active'))
ORDER BY name;

-- name: GetPaymentMethod :one
SELECT payment_method_id, name, kind, bank_name, credit_limit, closing_day, due_day, is_active
FROM payment_methods
WHERE payment_method_id = $1;

-- name: UpdatePaymentMethod :one
UPDATE payment_methods
SET name = $2,
    kind = $3,
    bank_name = $4,
    credit_limit = $5,
    closing_day = $6,
    due_day = $7,
    is_active = $8
WHERE payment_method_id = $1
RETURNING payment_method_id, name, kind, bank_name, credit_limit, closing_day, due_day, is_active;

-- name: GetInvoiceEntries :many
SELECT 
    cf.cash_flow_id, 
    cf.date, 
    cf.title, 
    cf.amount, 
    cat.name as category_name
FROM cash_flows cf
JOIN flow_categories cat ON cf.category_id = cat.category_id
JOIN expense_details ed ON cf.cash_flow_id = ed.cash_flow_id
WHERE ed.payment_method_id = $1 
  AND DATE_TRUNC('month', cf.date) = DATE_TRUNC('month', $2::date)
ORDER BY cf.date ASC;

-- name: GetOutstandingAmount :one
SELECT COALESCE(SUM(cf.amount), 0)::float
FROM cash_flows cf
JOIN expense_details ed ON cf.cash_flow_id = ed.cash_flow_id
WHERE ed.payment_method_id = $1
  AND ed.affects_card_invoice = true
  AND DATE_TRUNC('month', cf.date) >= DATE_TRUNC('month', $2::date);
