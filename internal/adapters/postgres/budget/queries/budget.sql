-- Budget plans

-- name: GetBudgetPlanByLedger :one
SELECT id, ledger_id, name, created_at, updated_at
FROM budget_plans
WHERE ledger_id = $1;

-- name: CreateBudgetPlan :one
INSERT INTO budget_plans (ledger_id, name, updated_at)
VALUES ($1, $2, $3)
RETURNING id, ledger_id, name, created_at, updated_at;

-- name: UpdateBudgetPlan :one
UPDATE budget_plans
SET name = $2, updated_at = $3
WHERE ledger_id = $1
RETURNING id, ledger_id, name, created_at, updated_at;

-- name: DeleteBudgetPlan :exec
DELETE FROM budget_plans
WHERE ledger_id = $1;

-- Budget versions

-- name: CreateBudgetVersion :one
INSERT INTO budget_plan_versions (plan_id, effective_from_month, created_by_user_id, updated_at)
VALUES ($1, $2, $3, $4)
RETURNING id, plan_id, effective_from_month, created_by_user_id, created_at, updated_at;

-- name: GetBudgetVersionByID :one
SELECT id, plan_id, effective_from_month, created_by_user_id, created_at, updated_at
FROM budget_plan_versions
WHERE id = $1;

-- name: ListBudgetVersions :many
SELECT id, plan_id, effective_from_month, created_by_user_id, created_at, updated_at
FROM budget_plan_versions
WHERE plan_id = $1
  AND ($2::date IS NULL OR effective_from_month >= $2)
  AND ($3::date IS NULL OR effective_from_month <= $3)
ORDER BY effective_from_month DESC;

-- name: DeleteBudgetVersion :exec
DELETE FROM budget_plan_versions
WHERE id = $1;

-- name: GetApplicableBudgetVersion :one
SELECT id, plan_id, effective_from_month, created_by_user_id, created_at, updated_at
FROM budget_plan_versions
WHERE plan_id = $1 AND effective_from_month <= $2
ORDER BY effective_from_month DESC
LIMIT 1;

-- Budget lines

-- name: CreateBudgetLine :one
INSERT INTO budget_plan_lines (version_id, category_id, percent, include_children, updated_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, version_id, category_id, percent, include_children, created_at, updated_at;

-- name: UpdateBudgetLine :one
UPDATE budget_plan_lines
SET percent = $3, include_children = $4, updated_at = $5
WHERE id = $1 AND version_id = $2
RETURNING id, version_id, category_id, percent, include_children, created_at, updated_at;

-- name: DeleteBudgetLine :exec
DELETE FROM budget_plan_lines
WHERE id = $1 AND version_id = $2;

-- name: ListBudgetLinesByVersion :many
SELECT id, version_id, category_id, percent, include_children, created_at, updated_at
FROM budget_plan_lines
WHERE version_id = $1
ORDER BY created_at;

-- name: ListBudgetLinesWithCategory :many
SELECT l.id, l.version_id, l.category_id, l.percent, l.include_children, l.created_at, l.updated_at,
  c.name, c.direction, c.is_budget_base, c.is_budget_relevant
FROM budget_plan_lines l
JOIN categories c ON c.id = l.category_id
WHERE l.version_id = $1
ORDER BY c.name;

-- Monthly calculations

-- name: SumIncomeBase :one
SELECT COALESCE(SUM(e.amount_cents), 0) AS income_base_cents
FROM entries e
JOIN transactions t ON t.id = e.transaction_id
JOIN categories c ON c.id = e.category_id
WHERE e.ledger_id = $1
  AND t.occurred_at >= $2 AND t.occurred_at < $3
  AND c.direction = 'in'
  AND c.is_budget_base = true;

-- name: SumSpentForCategory :one
SELECT COALESCE(SUM(e.amount_cents), 0) AS spent_cents
FROM entries e
JOIN transactions t ON t.id = e.transaction_id
JOIN categories c ON c.id = e.category_id
WHERE e.ledger_id = $1
  AND t.occurred_at >= $2 AND t.occurred_at < $3
  AND c.direction = 'out'
  AND c.is_budget_relevant = true
  AND e.category_id = $4;

-- name: SumSpentForCategoryWithChildren :one
WITH RECURSIVE tree AS (
  SELECT categories.id, categories.parent_id
  FROM categories
  WHERE categories.ledger_id = $1 AND categories.id = $4
  UNION ALL
  SELECT c.id, c.parent_id
  FROM categories c
  JOIN tree t ON t.id = c.parent_id
  WHERE c.ledger_id = $1
)
SELECT COALESCE(SUM(e.amount_cents), 0) AS spent_cents
FROM entries e
JOIN transactions t ON t.id = e.transaction_id
JOIN categories c ON c.id = e.category_id
JOIN tree ON tree.id = c.id
WHERE e.ledger_id = $1
  AND t.occurred_at >= $2 AND t.occurred_at < $3
  AND c.direction = 'out'
  AND c.is_budget_relevant = true;

-- name: SumOutOfBudget :one
SELECT COALESCE(SUM(e.amount_cents), 0) AS spent_cents
FROM entries e
JOIN transactions t ON t.id = e.transaction_id
JOIN categories c ON c.id = e.category_id
WHERE e.ledger_id = $1
  AND t.occurred_at >= $2 AND t.occurred_at < $3
  AND c.direction = 'out'
  AND c.is_budget_relevant = false;
