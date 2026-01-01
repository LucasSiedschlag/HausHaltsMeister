# Implementation Playbook

This document defines how the agent must work on every feature. It is the default operating procedure.

---

## Feature checklist (mandatory)

1. **Docs/API**
   - Update `docs/api/` or confirm the endpoint already exists.
   - Update `docs/ledger/` when rules or flows change.

2. **Migrations**
   - Add/adjust migrations if schema changes are required.
   - Keep `created_at` default and set `updated_at` in the app.

3. **SQLC queries**
   - Add/modify queries in `db/queries/`.
   - Ensure queries always filter by `ledger_id`.

4. **Service layer**
   - Implement domain rules and validations first.
   - Enforce ledger boundary and transfer balance.

5. **Handlers**
   - Implement HTTP layer with correct status codes and errors.
   - Use the standard error payload from `docs/agent/ERRORS.md`.

6. **Tests**
   - Follow `docs/agent/TEST_STRATEGY.md`.
   - Add handler + service tests; integration test if DB touched.

7. **Decisions (ADR)**
   - If a new irreversible decision is made, update `docs/agent/DECISIONS.md`.

8. **Seeds**
   - Update seeds if a new technical category or catalog entry is required.

---

## PR template

```
## O que foi feito
- ...

## Decisões
- ...

## Migrations
- ...

## Queries sqlc
- ...

## Testes
- ...

## Como validar
- ...
```
