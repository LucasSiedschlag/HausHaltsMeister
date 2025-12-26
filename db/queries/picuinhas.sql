-- name: CreatePerson :one
INSERT INTO picuinha_persons (name, notes)
VALUES ($1, $2)
RETURNING person_id, name, notes;

-- name: UpdatePerson :one
UPDATE picuinha_persons
SET name = $2,
    notes = $3
WHERE person_id = $1
RETURNING person_id, name, notes;

-- name: DeletePerson :exec
DELETE FROM picuinha_persons
WHERE person_id = $1;

-- name: CountEntriesByPerson :one
SELECT COUNT(*)
FROM picuinha_entries
WHERE person_id = $1;

-- name: ListPersons :many
SELECT person_id, name, notes
FROM picuinha_persons
ORDER BY name;

-- name: GetPerson :one
SELECT person_id, name, notes
FROM picuinha_persons
WHERE person_id = $1;

-- name: CreatePicuinhaEntry :one
INSERT INTO picuinha_entries (person_id, date, kind, amount, cash_flow_id, payment_method_id, card_owner)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING picuinha_entry_id, person_id, date, kind, amount, cash_flow_id, payment_method_id, card_owner;

-- name: UpdatePicuinhaEntry :one
UPDATE picuinha_entries
SET person_id = $2,
    kind = $3,
    amount = $4,
    cash_flow_id = $5,
    payment_method_id = $6,
    card_owner = $7
WHERE picuinha_entry_id = $1
RETURNING picuinha_entry_id, person_id, date, kind, amount, cash_flow_id, payment_method_id, card_owner;

-- name: DeletePicuinhaEntry :exec
DELETE FROM picuinha_entries
WHERE picuinha_entry_id = $1;

-- name: ListEntries :many
SELECT picuinha_entry_id, person_id, date, kind, amount, cash_flow_id, payment_method_id, card_owner
FROM picuinha_entries
ORDER BY date DESC;

-- name: GetEntry :one
SELECT picuinha_entry_id, person_id, date, kind, amount, cash_flow_id, payment_method_id, card_owner
FROM picuinha_entries
WHERE picuinha_entry_id = $1;

-- name: ListEntriesByPerson :many
SELECT picuinha_entry_id, person_id, date, kind, amount, cash_flow_id, payment_method_id, card_owner
FROM picuinha_entries
WHERE person_id = $1
ORDER BY date DESC;

-- name: GetPersonBalance :one
SELECT COALESCE(SUM(CASE WHEN kind = 'PLUS' THEN amount ELSE -amount END), 0)::decimal
FROM picuinha_entries
WHERE person_id = $1;
