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

-- name: CountCasesByPerson :one
SELECT COUNT(*)
FROM picuinha_cases
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
SELECT (
  COALESCE(
    (
      SELECT SUM(CASE WHEN kind = 'PLUS' THEN amount ELSE -amount END)
      FROM picuinha_entries pe
      WHERE pe.person_id = $1
    ),
    0
  )
  +
  COALESCE(
    (
      SELECT SUM(i.amount + i.extra_amount)
      FROM picuinha_case_installments i
      JOIN picuinha_cases c ON c.picuinha_case_id = i.picuinha_case_id
      WHERE c.person_id = $1
        AND i.is_paid = false
        AND (
          c.case_type <> 'RECURRING'
          OR DATE_TRUNC('month', i.due_date) <= DATE_TRUNC('month', CURRENT_DATE)
        )
    ),
    0
  )
)::decimal;

-- name: CreatePicuinhaCase :one
INSERT INTO picuinha_cases (
  person_id,
  title,
  case_type,
  total_amount,
  installment_count,
  installment_amount,
  start_date,
  payment_method_id,
  installment_plan_id,
  category_id,
  interest_rate,
  interest_rate_unit,
  recurrence_interval_months
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING picuinha_case_id, person_id, title, case_type, total_amount, installment_count, installment_amount,
  start_date, payment_method_id, installment_plan_id, category_id, interest_rate, interest_rate_unit,
  recurrence_interval_months, created_at;

-- name: UpdatePicuinhaCase :one
UPDATE picuinha_cases
SET person_id = $2,
    title = $3,
    case_type = $4,
    total_amount = $5,
    installment_count = $6,
    installment_amount = $7,
    start_date = $8,
    payment_method_id = $9,
    installment_plan_id = $10,
    category_id = $11,
    interest_rate = $12,
    interest_rate_unit = $13,
    recurrence_interval_months = $14
WHERE picuinha_case_id = $1
RETURNING picuinha_case_id, person_id, title, case_type, total_amount, installment_count, installment_amount,
  start_date, payment_method_id, installment_plan_id, category_id, interest_rate, interest_rate_unit,
  recurrence_interval_months, created_at;

-- name: DeletePicuinhaCase :exec
DELETE FROM picuinha_cases
WHERE picuinha_case_id = $1;

-- name: GetPicuinhaCase :one
SELECT picuinha_case_id, person_id, title, case_type, total_amount, installment_count, installment_amount,
  start_date, payment_method_id, installment_plan_id, category_id, interest_rate, interest_rate_unit,
  recurrence_interval_months, created_at
FROM picuinha_cases
WHERE picuinha_case_id = $1;

-- name: ListPicuinhaCasesByPerson :many
SELECT
  c.picuinha_case_id,
  c.person_id,
  c.title,
  c.case_type,
  c.total_amount,
  c.installment_count,
  c.installment_amount,
  c.start_date,
  c.payment_method_id,
  c.installment_plan_id,
  c.category_id,
  c.interest_rate,
  c.interest_rate_unit,
  c.recurrence_interval_months,
  c.created_at,
  COUNT(i.picuinha_case_installment_id) AS installments_total,
  COUNT(i.picuinha_case_installment_id) FILTER (WHERE i.is_paid) AS installments_paid,
  COALESCE(SUM(i.amount + i.extra_amount) FILTER (WHERE i.is_paid), 0)::decimal AS amount_paid,
  COALESCE(SUM(i.amount + i.extra_amount) FILTER (WHERE NOT i.is_paid), 0)::decimal AS amount_remaining
FROM picuinha_cases c
LEFT JOIN picuinha_case_installments i ON i.picuinha_case_id = c.picuinha_case_id
WHERE c.person_id = $1
GROUP BY c.picuinha_case_id
ORDER BY c.created_at DESC;

-- name: CreatePicuinhaCaseInstallment :one
INSERT INTO picuinha_case_installments (
  picuinha_case_id,
  installment_number,
  due_date,
  amount,
  extra_amount,
  is_paid,
  paid_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING picuinha_case_installment_id, picuinha_case_id, installment_number, due_date,
  amount, extra_amount, is_paid, paid_at;

-- name: UpdatePicuinhaCaseInstallment :one
UPDATE picuinha_case_installments
SET amount = $2,
    extra_amount = $3,
    is_paid = $4,
    paid_at = $5
WHERE picuinha_case_installment_id = $1
RETURNING picuinha_case_installment_id, picuinha_case_id, installment_number, due_date,
  amount, extra_amount, is_paid, paid_at;

-- name: GetPicuinhaCaseInstallment :one
SELECT picuinha_case_installment_id, picuinha_case_id, installment_number, due_date,
  amount, extra_amount, is_paid, paid_at
FROM picuinha_case_installments
WHERE picuinha_case_installment_id = $1;

-- name: ListPicuinhaCaseInstallments :many
SELECT picuinha_case_installment_id, picuinha_case_id, installment_number, due_date,
  amount, extra_amount, is_paid, paid_at
FROM picuinha_case_installments
WHERE picuinha_case_id = $1
ORDER BY due_date ASC, installment_number ASC;
