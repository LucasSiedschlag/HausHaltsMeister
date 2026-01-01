# API — Relatorios

Este modulo cobre relatorios agregados por ledger.

## Padroes globais aplicados

- JSON em `snake_case`.
- Datas e timestamps: ISO-8601 em UTC.
- Meses: `YYYY-MM-01`.
- IDs: UUID (string).
- Auth: `Authorization: Bearer <access_token>`.
- Ledger boundary obrigatorio em todos os endpoints.

## Inventario de endpoints

| Metodo | Path | Descricao | Auth | Role |
|---|---|---|---|---|
| GET | /ledgers/{ledgerId}/reports/balances | Saldos por conta | Sim | viewer |
| GET | /ledgers/{ledgerId}/reports/categories | Categoria x gasto | Sim | viewer |
| GET | /ledgers/{ledgerId}/reports/cashflow | Fluxo de caixa | Sim | viewer |

---

## Erros globais (5xx)

- 500 `INTERNAL_SERVER_ERROR`
- 503 `SERVICE_UNAVAILABLE`
- 504 `GATEWAY_TIMEOUT`

## Contratos detalhados

### GET /ledgers/{ledgerId}/reports/balances
1) Summary / Purpose
- Saldos por conta no mes.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Path: `ledgerId`.
- Query: `month` (YYYY-MM-01).

4) Response
- 200
```json
{ "month": "2026-01-01", "accounts": [ { "account_id": "uuid", "balance_cents": 0 } ] }
```

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Regras por tipo de conta (cash/investment vs credit_card).

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### GET /ledgers/{ledgerId}/reports/categories
1) Summary / Purpose
- Gasto por categoria.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Query: `from`, `to` (date-time).

4) Response
- 200
```json
{ "from": "...", "to": "...", "items": [ { "category_id": "uuid", "spent_cents": 0 } ] }
```

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Sempre filtra por ledger.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### GET /ledgers/{ledgerId}/reports/cashflow
1) Summary / Purpose
- Fluxo de caixa por periodo.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Query: `from`, `to` (date-time).

4) Response
- 200
```json
{ "from": "...", "to": "...", "in_cents": 0, "out_cents": 0, "net_cents": 0 }
```

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Baseado em entries do journal.

7) Pagination
- n/a.

8) Idempotency
- n/a.
