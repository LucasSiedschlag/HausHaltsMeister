# API — Categories

Este modulo cobre categorias de entrada e saida, incluindo regras de budget.

## Padroes globais aplicados

- JSON em `snake_case`.
- Datas e timestamps: ISO-8601 em UTC.
- IDs: UUID (string).
- Auth: `Authorization: Bearer <access_token>`.
- Ledger boundary obrigatorio em todos os endpoints.

## Inventario de endpoints

| Metodo | Path | Descricao | Auth | Role |
|---|---|---|---|---|
| GET | /ledgers/{ledgerId}/categories | Listar categorias | Sim | viewer |
| GET | /ledgers/{ledgerId}/categories/{categoryId} | Detalhe da categoria | Sim | viewer |
| POST | /ledgers/{ledgerId}/categories | Criar categoria | Sim | editor |
| PATCH | /ledgers/{ledgerId}/categories/{categoryId} | Atualizar categoria | Sim | editor |
| DELETE | /ledgers/{ledgerId}/categories/{categoryId} | Soft delete (planejado) | Sim | editor |

---

## Erros globais (5xx)

- 500 `INTERNAL_SERVER_ERROR`
- 503 `SERVICE_UNAVAILABLE`
- 504 `GATEWAY_TIMEOUT`

## Contratos detalhados

### GET /ledgers/{ledgerId}/categories
1) Summary / Purpose
- Listar categorias do ledger.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Path params: `ledgerId` (uuid).
- Query params:
  - `direction` (in|out, opcional)
  - `active` (bool, default true)

4) Response
- 200
```json
[
  {
    "id": "uuid",
    "ledger_id": "uuid",
    "name": "Mercado",
    "direction": "out",
    "is_budget_base": false,
    "is_budget_relevant": true,
    "parent_id": null,
    "created_at": "...",
    "updated_at": "..."
  }
]
```

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- `direction=in` -> `is_budget_relevant=false`.
- `direction=out` -> `is_budget_base=false`.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### GET /ledgers/{ledgerId}/categories/{categoryId}
1) Summary / Purpose
- Detalhe da categoria.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Path params: `ledgerId`, `categoryId`.

4) Response
- 200
```json
{
  "id": "uuid",
  "ledger_id": "uuid",
  "name": "Mercado",
  "direction": "out",
  "is_budget_base": false,
  "is_budget_relevant": true,
  "parent_id": null,
  "created_at": "...",
  "updated_at": "..."
}
```

5) Errors
- 404 `CATEGORY_NOT_FOUND`
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Busca por `(ledger_id, category_id)`.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### POST /ledgers/{ledgerId}/categories
1) Summary / Purpose
- Criar categoria.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body:
```json
{
  "name": "Mercado",
  "direction": "out",
  "is_budget_base": false,
  "is_budget_relevant": true,
  "parent_id": null
}
```

4) Response
- 201
```json
{
  "id": "uuid",
  "ledger_id": "uuid",
  "name": "Mercado",
  "direction": "out",
  "is_budget_base": false,
  "is_budget_relevant": true,
  "parent_id": null,
  "created_at": "...",
  "updated_at": "..."
}
```

5) Errors
- 409 `CONFLICT_DUPLICATE_NAME`
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- `parent_id` opcional.
- Validacoes por `direction`.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### PATCH /ledgers/{ledgerId}/categories/{categoryId}
1) Summary / Purpose
- Atualizar categoria.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body parcial:
```json
{ "name": "Mercado 2", "is_budget_relevant": false }
```

4) Response
- 200 (categoria atualizada).

5) Errors
- 404 `CATEGORY_NOT_FOUND`
- 409 `CONFLICT_DUPLICATE_NAME`
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Regras de budget aplicadas.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### DELETE /ledgers/{ledgerId}/categories/{categoryId}
1) Summary / Purpose
- Soft delete (planejado).

2) Auth & Authorization
- Role: editor+.

3) Request
- Body: n/a.

4) Response
- 204.

5) Errors
- 404 `CATEGORY_NOT_FOUND`

6) Semantics / Notes
- Pode virar inativa.

7) Pagination
- n/a.

8) Idempotency
- n/a.
