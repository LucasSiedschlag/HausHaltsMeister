-- name: CreatePaymentMethod :one
INSERT INTO payment_methods (name, kind, bank_name, closing_day, due_day, is_active)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING payment_method_id, name, kind, bank_name, closing_day, due_day, is_active;

-- name: ListPaymentMethods :many
SELECT payment_method_id, name, kind, bank_name, closing_day, due_day, is_active
FROM payment_methods
WHERE (sqlc.narg('is_active')::boolean IS NULL OR is_active = sqlc.narg('is_active'))
ORDER BY name;

-- name: GetPaymentMethod :one
SELECT payment_method_id, name, kind, bank_name, closing_day, due_day, is_active
FROM payment_methods
WHERE payment_method_id = $1;
