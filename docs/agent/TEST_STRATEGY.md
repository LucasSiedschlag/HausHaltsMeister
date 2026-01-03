# Test Strategy

This document defines the testing policy and required coverage for critical flows. No feature should be merged without satisfying these rules.

---

## Test types

**Unit tests**
- Target pure business rules and small helpers.
- No database, no HTTP.

**Service tests**
- Target application rules that orchestrate repositories and validations.
- Mocks/stubs allowed, but must verify domain invariants.

**Handler tests**
- Target HTTP layer behavior: status codes, error payloads, auth checks.
- Use minimal fixtures and mock services when possible.

**Integration tests**
- Use real Postgres.
- Validate schema constraints, SQL queries, and transaction boundaries.

---

## Mandatory flows (must have tests)

1. **Transfer balance**
   - Transfer must contain at least one IN and one OUT.
   - SUM(OUT) == SUM(IN).

2. **Budget monthly calculation**
   - Income base uses only IN with `is_budget_base=true`.
   - Budget consumption uses OUT with `is_budget_relevant=true`.
   - Version selection by `effective_from_month`.

3. **Installment posting idempotency**
   - Posting the same month twice does not duplicate entries.
   - Only `scheduled` installments can be posted.

4. **Refresh token rotation**
   - Refresh invalidates the old session and creates a new session.
   - Using a revoked refresh token must fail.

5. **Ledger access denial**
   - Requests to another ledger return 403/404.
6. **Ledger scope scan**
   - `TestQueriesScopedByLedgerID` must pass (no `WHERE id =` without `ledger_id`).
7. **Role matrix (owner)**
   - Owner-only service methods must have explicit tests for non-owner denial.

---

## Baseline rules

- Every new endpoint must include:
  - 1 handler test (status + errors)
  - 1 service test (business rule)
  - 1 integration test **if** it touches the DB

- If an endpoint affects the journal, add a transaction integrity test.
- If an endpoint affects the card lifecycle, add an idempotency test.

---

## Templates

### Handler test template

```txt
Test: <endpoint> returns <status> when <condition>
- Arrange: mock service + request payload
- Act: call handler
- Assert: status code, error code, response shape
```

### Service test template

```txt
Test: <service> enforces <rule>
- Arrange: inputs + mock repos
- Act: call service
- Assert: domain rule enforced, error code returned
```

### Integration test template

```txt
Test: <feature> persists correctly in DB
- Arrange: setup DB state
- Act: run service method
- Assert: DB rows, constraints, and joins are correct
```
