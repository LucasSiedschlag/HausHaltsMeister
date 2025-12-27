-- name: CreateInstallmentPlan :one
INSERT INTO installment_plans (description, total_amount, installment_count, installment_amount, start_date, payment_method_id, starts_on_current_invoice)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING installment_plan_id, description, total_amount, installment_count, installment_amount, start_date, payment_method_id, starts_on_current_invoice, plan_type, person_id, category_id, interest_rate, interest_rate_unit, recurrence_interval_months, created_at;

-- name: CreateExpenseDetail :exec
INSERT INTO expense_details (cash_flow_id, payment_method_id, is_fixed, is_future, installment_plan_id, affects_card_invoice)
VALUES ($1, $2, false, true, $3, $4);
