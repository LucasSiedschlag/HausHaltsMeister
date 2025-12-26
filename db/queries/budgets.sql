-- name: GetBudgetPeriodByMonth :one
SELECT budget_period_id, month, analysis_mode, is_closed
FROM budget_periods
WHERE month = $1;

-- name: GetLatestBudgetPeriodWithItemsBefore :one
SELECT budget_period_id, month, analysis_mode, is_closed
FROM budget_periods
WHERE month < $1
  AND EXISTS (
    SELECT 1
    FROM budget_items
    WHERE budget_period_id = budget_periods.budget_period_id
  )
ORDER BY month DESC
LIMIT 1;

-- name: CreateBudgetPeriod :one
INSERT INTO budget_periods (month, analysis_mode, is_closed)
VALUES ($1, $2, $3)
RETURNING budget_period_id, month, analysis_mode, is_closed;

-- name: UpsertBudgetItem :one
INSERT INTO budget_items (budget_period_id, category_id, mode, planned_amount, target_percent, notes)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (budget_period_id, category_id) DO UPDATE -- Need unique constraint first!
SET mode = EXCLUDED.mode,
    planned_amount = EXCLUDED.planned_amount,
    target_percent = EXCLUDED.target_percent,
    notes = EXCLUDED.notes
RETURNING budget_item_id, budget_period_id, category_id, mode, planned_amount, target_percent, notes;

-- name: GetBudgetItemsByPeriod :many
SELECT 
    bi.budget_item_id, 
    bi.budget_period_id, 
    bi.category_id, 
    bi.mode, 
    bi.planned_amount, 
    bi.target_percent, 
    bi.notes,
    fc.name as category_name
FROM budget_items bi
JOIN flow_categories fc ON fc.category_id = bi.category_id
WHERE bi.budget_period_id = $1;

-- name: GetBudgetItemByID :one
SELECT 
    bi.budget_item_id, 
    bi.budget_period_id, 
    bi.category_id, 
    bi.mode, 
    bi.planned_amount, 
    bi.target_percent, 
    bi.notes,
    fc.name as category_name
FROM budget_items bi
JOIN flow_categories fc ON fc.category_id = bi.category_id
WHERE bi.budget_item_id = $1;

-- name: UpdateBudgetItem :one
UPDATE budget_items
SET mode = $2,
    planned_amount = $3,
    target_percent = $4,
    notes = $5
WHERE budget_item_id = $1
RETURNING budget_item_id, budget_period_id, category_id, mode, planned_amount, target_percent, notes,
  (SELECT name FROM flow_categories WHERE category_id = budget_items.category_id) AS category_name;
