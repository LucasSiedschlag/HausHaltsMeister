# Conventions (Mechanical Standards)

This document defines the mechanical standards the agent must follow without asking each time.

---

## Routing and URL naming

- Base pattern: `/ledgers/{ledgerId}/...` for all ledger-scoped resources.
- Never expose endpoints that accept only `:id` without `ledgerId`.
- State transitions are explicit (e.g., `/statements/close`, `/statements/pay`).

---

## Naming conventions

- **Database**: `snake_case` for tables and columns.
- **JSON**: `snake_case` for request/response fields.
- **IDs**: `uuid` fields are named `*_id`.
- **Money**: use `*_cents` as integer.

---

## Code layout (mandatory)

- HTTP handlers live in `internal/adapters/http/handlers`.
- Middleware lives in `internal/adapters/http/middleware`.
- HTTP DTOs live in `internal/adapters/http/dto`.
- Shared HTTP helpers live in `internal/adapters/http/httpx`.
- Postgres repositories live in `internal/adapters/postgres/<module>`.
- SQL files (sqlc) live in `internal/adapters/postgres/<module>/queries`.
- sqlc generated code lives in `internal/adapters/postgres/sqlc`.
- Domain interfaces live in `internal/domain/<module>`.
- Shared infrastructure lives in `internal/infra`.

---

## Error response format

Standard error payload:
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Human readable message",
  "details": {
    "field": "reason"
  }
}
```

- `code` is stable and machine-readable (see `docs/agent/ERRORS.md`).
- `message` is user-facing.
- `details` is optional and structured.

---

## Timezones and timestamps

- DB uses `timestamptz` and stores **UTC**.
- API accepts and returns ISO-8601; convert to UTC in storage.
- `created_at` is set by DB; `updated_at` is set by the application.

---

## Pagination (cursor)

- Cursor pagination is mandatory for transaction lists.
- Default ordering: `occurred_at DESC, id DESC`.
- Cursor shape:
```json
{
  "cursor_occurred_at": "2026-01-10T00:00:00Z",
  "cursor_id": "uuid"
}
```
- Response should include `next_cursor` with the same shape.

---

## Default filters

Standard query parameters:
- `from` / `to` (date range)
- `q` (text search)
- `account_id`
- `category_id`
- `limit`

For card-specific endpoints:
- `month` (YYYY-MM-01)
- `status`
