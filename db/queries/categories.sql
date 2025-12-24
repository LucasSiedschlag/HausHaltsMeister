-- name: CreateCategory :one
INSERT INTO flow_categories (
  name,
  direction,
  is_budget_relevant,
  is_active
) VALUES (
  $1, $2, $3, $4
)
RETURNING category_id, name, direction, is_budget_relevant, is_active;

-- name: ListCategories :many
SELECT category_id, name, direction, is_budget_relevant, is_active
FROM flow_categories
WHERE ($1::boolean = false OR is_active = true)
ORDER BY name;

-- name: GetCategoryByID :one
SELECT category_id, name, direction, is_budget_relevant, is_active
FROM flow_categories
WHERE category_id = $1;
