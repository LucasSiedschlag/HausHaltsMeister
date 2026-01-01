-- Categories

-- name: ListCategoriesByLedger :many
SELECT id, ledger_id, parent_id, name, direction, is_budget_base, is_budget_relevant, is_active, created_at, updated_at
FROM categories
WHERE ledger_id = $1
  AND ($2::text IS NULL OR direction = $2)
  AND ($3::boolean IS NULL OR is_active = $3)
ORDER BY name;

-- name: GetCategoryByID :one
SELECT id, ledger_id, parent_id, name, direction, is_budget_base, is_budget_relevant, is_active, created_at, updated_at
FROM categories
WHERE ledger_id = $1 AND id = $2;

-- name: CreateCategory :one
INSERT INTO categories (ledger_id, parent_id, name, direction, is_budget_base, is_budget_relevant, is_active, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, ledger_id, parent_id, name, direction, is_budget_base, is_budget_relevant, is_active, created_at, updated_at;

-- name: UpdateCategory :one
UPDATE categories
SET parent_id = $3, name = $4, direction = $5, is_budget_base = $6, is_budget_relevant = $7, is_active = $8, updated_at = $9
WHERE ledger_id = $1 AND id = $2
RETURNING id, ledger_id, parent_id, name, direction, is_budget_base, is_budget_relevant, is_active, created_at, updated_at;

-- name: DeactivateCategory :one
UPDATE categories
SET is_active = false, updated_at = $3
WHERE ledger_id = $1 AND id = $2
RETURNING id, ledger_id, parent_id, name, direction, is_budget_base, is_budget_relevant, is_active, created_at, updated_at;

-- name: FindCategoryByName :one
SELECT id
FROM categories
WHERE ledger_id = $1 AND name = $2
LIMIT 1;

-- name: GetCategoryDirection :one
SELECT direction
FROM categories
WHERE ledger_id = $1 AND id = $2;

-- name: GetCategoryBudgetInfo :one
SELECT direction, is_budget_relevant
FROM categories
WHERE ledger_id = $1 AND id = $2;

-- name: ListCategoryDescendants :many
WITH RECURSIVE tree AS (
  SELECT categories.id, categories.parent_id
  FROM categories
  WHERE categories.ledger_id = $1 AND categories.id = $2
  UNION ALL
  SELECT c.id, c.parent_id
  FROM categories c
  JOIN tree t ON t.id = c.parent_id
  WHERE c.ledger_id = $1
)
SELECT tree.id
FROM tree;
