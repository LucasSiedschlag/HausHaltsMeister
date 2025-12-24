-- name: CreatePerson :one
INSERT INTO picuinha_persons (name, notes)
VALUES ($1, $2)
RETURNING person_id, name, notes;

-- name: ListPersons :many
SELECT person_id, name, notes
FROM picuinha_persons
ORDER BY name;

-- name: GetPerson :one
SELECT person_id, name, notes
FROM picuinha_persons
WHERE person_id = $1;

-- name: CreatePicuinhaEntry :one
INSERT INTO picuinha_entries (person_id, date, kind, amount, cash_flow_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING picuinha_entry_id, person_id, date, kind, amount, cash_flow_id;

-- name: ListEntriesByPerson :many
SELECT picuinha_entry_id, person_id, date, kind, amount, cash_flow_id
FROM picuinha_entries
WHERE person_id = $1
ORDER BY date DESC;

-- name: GetPersonBalance :one
SELECT COALESCE(SUM(CASE WHEN kind = 'PLUS' THEN amount ELSE -amount END), 0)::decimal
FROM picuinha_entries
WHERE person_id = $1;
