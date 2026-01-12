# API — Accounts

Este modulo cobre todas as contas internas. Cartoes de credito sao tratados no modulo de cartao.

## Padroes globais aplicados

- JSON em `snake_case`.
- Datas e timestamps: ISO-8601 em UTC.
- IDs: UUID (string).
- Auth: `Authorization: Bearer <access_token>`.
- Ledger boundary obrigatorio em todos os endpoints.

## Inventario de endpoints

| Metodo | Path | Descricao | Auth | Role |
|---|---|---|---|---|
| GET | /ledgers/{ledgerId}/accounts | Listar contas | Sim | viewer |
| GET | /ledgers/{ledgerId}/accounts/{accountId} | Detalhe da conta | Sim | viewer |
| POST | /ledgers/{ledgerId}/accounts | Criar conta | Sim | editor |
| PATCH | /ledgers/{ledgerId}/accounts/{accountId} | Atualizar conta | Sim | editor |
| DELETE | /ledgers/{ledgerId}/accounts/{accountId} | Soft delete (planejado) | Sim | editor |

---

## Erros globais (5xx)

- 500 `INTERNAL_SERVER_ERROR`
- 503 `SERVICE_UNAVAILABLE`
- 504 `GATEWAY_TIMEOUT`

## Contratos detalhados

### GET /ledgers/{ledgerId}/accounts
1) Summary / Purpose
- Listar contas do ledger.

2) Auth & Authorization
- Token: sim.
- Role: viewer+.

3) Request
- Path params: `ledgerId` (uuid).
- Query params:
  - `active` (bool, default true).
- Headers: `Authorization`.

4) Response
- 200
```json
[
  {
    "id": "uuid",
    "ledger_id": "uuid",
    "name": "Pessoal",
    "type": "wallet",
    "nature": "asset",
    "is_active": true,
    "created_at": "...",
    "updated_at": "..."
  }
]
```

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- `type` tem validacao via CHECK no DB.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### GET /ledgers/{ledgerId}/accounts/{accountId}
1) Summary / Purpose
- Detalhe da conta.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Path params: `ledgerId`, `accountId`.

4) Response
- 200
```json
{
  "id": "uuid",
  "ledger_id": "uuid",
  "name": "Pessoal",
  "type": "wallet",
  "nature": "asset",
  "is_active": true,
  "created_at": "...",
  "updated_at": "..."
}
```

5) Errors
- 404 `ACCOUNT_NOT_FOUND`
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Busca por `(ledger_id, account_id)`.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### POST /ledgers/{ledgerId}/accounts
1) Summary / Purpose
- Criar conta.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body:
```json
{
  "name": "Pessoal",
  "type": "wallet",
  "nature": "asset",
  "is_active": true
}
```

4) Response
- 201
```json
{
  "id": "uuid",
  "ledger_id": "uuid",
  "name": "Pessoal",
  "type": "wallet",
  "nature": "asset",
  "is_active": true,
  "created_at": "...",
  "updated_at": "..."
}
```

5) Errors
- 409 `CONFLICT_DUPLICATE_NAME`
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- `type` aceito: `current`, `business`, `investment`, `exchange`, `wallet`.
- `nature` aceito: `asset` (padrao) ou `liability`.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### PATCH /ledgers/{ledgerId}/accounts/{accountId}
1) Summary / Purpose
- Atualizar conta.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body parcial:
```json
{ "name": "Conta Nova", "is_active": false }
```

4) Response
- 200 (conta atualizada).

5) Errors
- 404 `ACCOUNT_NOT_FOUND`
- 409 `CONFLICT_DUPLICATE_NAME`

6) Semantics / Notes
- Nao apagar historico.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### DELETE /ledgers/{ledgerId}/accounts/{accountId}
1) Summary / Purpose
- Soft delete (planejado).

2) Auth & Authorization
- Role: editor+.

3) Request
- Body: n/a.

4) Response
- 204.

5) Errors
- 404 `ACCOUNT_NOT_FOUND`

6) Semantics / Notes
- Pode virar inativa.

7) Pagination
- n/a.

8) Idempotency
- n/a.
