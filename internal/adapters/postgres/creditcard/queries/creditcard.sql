-- Card networks

-- name: ListCardNetworks :many
SELECT code, display_name, created_at, updated_at
FROM card_networks
ORDER BY code;

-- name: GetCardNetwork :one
SELECT code, display_name, created_at, updated_at
FROM card_networks
WHERE code = $1;

-- name: CreateCardNetwork :one
INSERT INTO card_networks (code, display_name)
VALUES ($1, $2)
RETURNING code, display_name, created_at, updated_at;

-- name: UpdateCardNetwork :one
UPDATE card_networks
SET display_name = $2, updated_at = $3
WHERE code = $1
RETURNING code, display_name, created_at, updated_at;

-- name: DeleteCardNetwork :exec
DELETE FROM card_networks
WHERE code = $1;

-- Credit cards

-- name: ListCreditCards :many
SELECT c.id, c.ledger_id, c.parent_account_id, c.liability_account_id, c.label, c.brand, c.last4, c.cvv, c.holder_name,
       c.active, c.color, c.style, c.closing_day, c.due_day, c.created_at, c.updated_at
FROM credit_cards c
WHERE c.ledger_id = $1 AND c.parent_account_id = $2
ORDER BY c.created_at;

-- name: GetCreditCard :one
SELECT c.id, c.ledger_id, c.parent_account_id, c.liability_account_id, c.label, c.brand, c.last4, c.cvv, c.holder_name,
       c.active, c.color, c.style, c.closing_day, c.due_day, c.created_at, c.updated_at
FROM credit_cards c
WHERE c.id = $1;

-- name: CreateCreditCard :one
INSERT INTO credit_cards (
  ledger_id,
  parent_account_id,
  liability_account_id,
  label,
  brand,
  last4,
  cvv,
  holder_name,
  active,
  color,
  style,
  closing_day,
  due_day
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING id, ledger_id, parent_account_id, liability_account_id, label, brand, last4, cvv, holder_name,
  active, color, style, closing_day, due_day, created_at, updated_at;

-- name: UpdateCreditCard :one
UPDATE credit_cards c
SET label = $3,
    brand = $4,
    last4 = $5,
    cvv = $6,
    holder_name = $7,
    active = $8,
    color = $9,
    style = $10,
    closing_day = $11,
    due_day = $12,
    updated_at = $13
WHERE c.ledger_id = $1 AND c.id = $2
RETURNING c.id, c.ledger_id, c.parent_account_id, c.liability_account_id, c.label, c.brand, c.last4, c.cvv, c.holder_name,
  c.active, c.color, c.style, c.closing_day, c.due_day, c.created_at, c.updated_at;

-- name: DeleteCreditCard :exec
DELETE FROM credit_cards
WHERE ledger_id = $1 AND id = $2;

-- Accounts lookup

-- name: GetAccount :one
SELECT a.id, a.ledger_id, a.type, a.nature, a.is_active
FROM accounts a
JOIN ledger_members lm ON lm.ledger_id = a.ledger_id AND lm.user_id = $2 AND lm.removed_at IS NULL
WHERE a.id = $1;

-- name: CreateAccount :one
INSERT INTO accounts (ledger_id, name, type, nature, is_active)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, ledger_id, type, nature, is_active;

-- name: GetAccountNature :one
SELECT nature
FROM accounts
WHERE ledger_id = $1 AND id = $2;

-- Categories lookup

-- name: GetCategoryBudgetInfo :one
SELECT direction, is_budget_relevant
FROM categories
WHERE ledger_id = $1 AND id = $2;

-- name: FindCategoryByName :one
SELECT id
FROM categories
WHERE ledger_id = $1 AND name = $2
LIMIT 1;

-- Installment plans

-- name: CreateInstallmentPlan :one
INSERT INTO installment_plans (ledger_id, credit_card_id, purchase_occurred_at, merchant, description, category_id, total_amount_cents, installments_count, installment_amount_cents, first_due_month, status, created_by_user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING id, ledger_id, credit_card_id, purchase_occurred_at, merchant, description, category_id, total_amount_cents, installments_count, installment_amount_cents, first_due_month, status, created_by_user_id, created_at, updated_at;

-- name: CreateInstallment :exec
INSERT INTO installments (ledger_id, plan_id, installment_no, due_month, amount_cents, status)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: ListPlans :many
SELECT id, ledger_id, credit_card_id, purchase_occurred_at, merchant, description, category_id, total_amount_cents, installments_count, installment_amount_cents, first_due_month, status, created_by_user_id, created_at, updated_at
FROM installment_plans
WHERE ledger_id = $1 AND credit_card_id = $2
  AND ($3::text IS NULL OR status = $3)
ORDER BY created_at DESC;

-- name: GetPlan :one
SELECT id, ledger_id, credit_card_id, purchase_occurred_at, merchant, description, category_id, total_amount_cents, installments_count, installment_amount_cents, first_due_month, status, created_by_user_id, created_at, updated_at
FROM installment_plans
WHERE ledger_id = $1 AND credit_card_id = $2 AND id = $3;

-- name: UpdatePlanStatus :exec
UPDATE installment_plans
SET status = $2, updated_at = $3
WHERE ledger_id = $1 AND id = $4;

-- name: DeletePlan :exec
DELETE FROM installment_plans
WHERE ledger_id = $1 AND id = $2;

-- name: HasPostedInstallments :one
SELECT EXISTS (
  SELECT 1
  FROM installments
  WHERE ledger_id = $1 AND plan_id = $2 AND status IN ('posted', 'paid')
) AS exists;

-- Installments

-- name: ListInstallments :many
SELECT i.id, i.ledger_id, i.plan_id, i.installment_no, i.due_month, i.amount_cents, i.status, i.posted_transaction_id, i.paid_statement_id, i.created_at, i.updated_at
FROM installments i
JOIN installment_plans p ON p.id = i.plan_id
WHERE i.ledger_id = $1 AND p.credit_card_id = $2
  AND ($3::date IS NULL OR i.due_month = $3)
  AND ($4::text IS NULL OR i.status = $4)
ORDER BY i.due_month, i.installment_no;

-- name: UpdateInstallmentStatus :one
UPDATE installments
SET status = $2, updated_at = $3
WHERE ledger_id = $1 AND id = $4
RETURNING id, ledger_id, plan_id, installment_no, due_month, amount_cents, status, posted_transaction_id, paid_statement_id, created_at, updated_at;

-- name: ListInstallmentsForPosting :many
SELECT i.id, i.ledger_id, i.plan_id, i.installment_no, i.due_month, i.amount_cents, i.status, i.posted_transaction_id, i.paid_statement_id, i.created_at, i.updated_at
FROM installments i
JOIN installment_plans p ON p.id = i.plan_id
WHERE i.ledger_id = $1
  AND p.credit_card_id = $2
  AND i.due_month = $3
  AND i.status = 'scheduled'
ORDER BY i.installment_no;

-- name: GetPlanCategory :one
SELECT category_id
FROM installment_plans
WHERE ledger_id = $1 AND id = $2;

-- name: MarkInstallmentPosted :exec
UPDATE installments
SET status = 'posted', posted_transaction_id = $2, updated_at = $3
WHERE ledger_id = $1 AND id = $4 AND status = 'scheduled';

-- Statements

-- name: GetStatement :one
SELECT id, ledger_id, credit_card_id, statement_month, closing_date, due_date, total_charges_cents, total_payments_cents, status, payment_transaction_id, created_at, updated_at
FROM credit_card_statements
WHERE ledger_id = $1 AND credit_card_id = $2 AND id = $3;

-- name: GetStatementByMonth :one
SELECT id, ledger_id, credit_card_id, statement_month, closing_date, due_date, total_charges_cents, total_payments_cents, status, payment_transaction_id, created_at, updated_at
FROM credit_card_statements
WHERE ledger_id = $1 AND credit_card_id = $2 AND statement_month = $3;

-- name: CreateStatement :one
INSERT INTO credit_card_statements (ledger_id, credit_card_id, statement_month, closing_date, due_date, total_charges_cents, total_payments_cents, status, payment_transaction_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, ledger_id, credit_card_id, statement_month, closing_date, due_date, total_charges_cents, total_payments_cents, status, payment_transaction_id, created_at, updated_at;

-- name: UpdateStatementTotals :one
UPDATE credit_card_statements
SET total_charges_cents = $4, total_payments_cents = $5, status = $6, updated_at = $7
WHERE ledger_id = $1 AND credit_card_id = $2 AND id = $3
RETURNING id, ledger_id, credit_card_id, statement_month, closing_date, due_date, total_charges_cents, total_payments_cents, status, payment_transaction_id, created_at, updated_at;

-- name: ListStatements :many
SELECT id, ledger_id, credit_card_id, statement_month, closing_date, due_date, total_charges_cents, total_payments_cents, status, payment_transaction_id, created_at, updated_at
FROM credit_card_statements
WHERE ledger_id = $1 AND credit_card_id = $2
  AND ($3::date IS NULL OR statement_month = $3)
ORDER BY statement_month DESC;

-- name: SetStatementPayment :one
UPDATE credit_card_statements
SET payment_transaction_id = $4, total_payments_cents = $5, status = $6, updated_at = $7
WHERE ledger_id = $1 AND credit_card_id = $2 AND id = $3
RETURNING id, ledger_id, credit_card_id, statement_month, closing_date, due_date, total_charges_cents, total_payments_cents, status, payment_transaction_id, created_at, updated_at;

-- name: MarkInstallmentsPaid :exec
UPDATE installments i
SET status = 'paid', paid_statement_id = $4, updated_at = $5
FROM installment_plans p
WHERE i.plan_id = p.id
  AND i.ledger_id = $1
  AND p.credit_card_id = $2
  AND i.due_month = $3
  AND i.status = 'posted';

-- name: SumStatementCharges :one
SELECT COALESCE(SUM(i.amount_cents), 0)
FROM installments i
JOIN installment_plans p ON p.id = i.plan_id
WHERE i.ledger_id = $1
  AND p.credit_card_id = $2
  AND i.due_month = $3
  AND i.status IN ('posted', 'paid');
