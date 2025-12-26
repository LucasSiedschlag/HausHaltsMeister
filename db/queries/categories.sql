-- name: CreateCategory :one
INSERT INTO flow_categories (
  name,
  direction,
  is_budget_relevant,
  is_active
) VALUES (
  $1, $2, $3, $4
)
RETURNING category_id, name, direction, is_budget_relevant, is_active, inactive_from_month;

-- name: ListCategories :many
SELECT category_id, name, direction, is_budget_relevant, is_active, inactive_from_month
FROM flow_categories
WHERE ($1::boolean = false OR is_active = true)
ORDER BY name;

-- name: ListCategoriesByMonth :many
SELECT category_id, name, direction, is_budget_relevant, is_active, inactive_from_month
FROM flow_categories fc
WHERE (
  $1::boolean = false
  OR fc.is_active = true
  OR EXISTS (
    SELECT 1
    FROM budget_items bi
    JOIN budget_periods bp ON bp.budget_period_id = bi.budget_period_id
    WHERE bi.category_id = fc.category_id
      AND date_trunc('month', bp.month) = date_trunc('month', $2::date)
  )
)
AND (
  fc.inactive_from_month IS NULL
  OR fc.inactive_from_month > $2::date
  OR EXISTS (
    SELECT 1
    FROM budget_items bi
    JOIN budget_periods bp ON bp.budget_period_id = bi.budget_period_id
    WHERE bi.category_id = fc.category_id
      AND date_trunc('month', bp.month) = date_trunc('month', $2::date)
  )
)
ORDER BY name;

-- name: GetCategoryByID :one
SELECT category_id, name, direction, is_budget_relevant, is_active, inactive_from_month
FROM flow_categories
WHERE category_id = $1;

-- name: UpdateCategory :one
UPDATE flow_categories
SET name = $2,
    direction = $3,
    is_budget_relevant = $4,
    is_active = $5
WHERE category_id = $1
RETURNING category_id, name, direction, is_budget_relevant, is_active, inactive_from_month;

-- name: DeactivateCategory :one
UPDATE flow_categories
SET is_active = false,
    inactive_from_month = $2
WHERE category_id = $1
RETURNING category_id, name, direction, is_budget_relevant, is_active, inactive_from_month;
