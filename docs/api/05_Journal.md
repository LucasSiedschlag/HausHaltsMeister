# API — Journal (Transactions + Entries)

Este modulo cobre transacoes e entries com paginacao segura e regras de transferencia.

## Padroes globais aplicados

- JSON em `snake_case`.
- Datas e timestamps: ISO-8601 em UTC.
- IDs: UUID (string).
- Auth: `Authorization: Bearer <access_token>`.
- Ledger boundary obrigatorio em todos os endpoints.
- Paginacao via cursor (sem offset).

## Inventario de endpoints

| Metodo | Path | Descricao | Auth | Role |
|---|---|---|---|---|
| POST | /ledgers/{ledgerId}/transactions | Criar transacao | Sim | editor |
| GET | /ledgers/{ledgerId}/transactions | Listar transacoes (cursor) | Sim | viewer |
| GET | /ledgers/{ledgerId}/transactions/{transactionId} | Detalhe da transacao | Sim | viewer |
| PATCH | /ledgers/{ledgerId}/transactions/{transactionId} | Atualizar transacao | Sim | editor |
| DELETE | /ledgers/{ledgerId}/transactions/{transactionId} | Deletar transacao | Sim | editor |
| POST | /ledgers/{ledgerId}/transactions:bulk | Criacao em lote (opcional) | Sim | editor |

---

## Erros globais (5xx)

- 500 `INTERNAL_SERVER_ERROR`
- 503 `SERVICE_UNAVAILABLE`
- 504 `GATEWAY_TIMEOUT`

## Contratos detalhados

### POST /ledgers/{ledgerId}/transactions
1) Summary / Purpose
- Criar transacao com entries.

2) Auth & Authorization
- Role: editor+.

3) Request
- Headers: `Idempotency-Key` (opcional).
- Body:
```json
{
  "occurred_at": "2026-01-10",
  "description": "Mercado",
  "notes": null,
  "entries": [
    {
      "account_id": "uuid",
      "category_id": "uuid",
      "kind": "normal",
      "amount_cents": 12000,
      "memo": "Compra do mes"
    }
  ]
}
```

4) Response
- 201 (transaction + entries).

5) Errors
- 422 `VALIDATION_ERROR`
- 422 `TRANSFER_NOT_BALANCED`

6) Semantics / Notes
- `amount_cents` sempre positivo.
- `transfer` deve balancear IN/OUT.

7) Pagination
- n/a.

8) Idempotency
- `Idempotency-Key` evita duplicidade.

---

### GET /ledgers/{ledgerId}/transactions
1) Summary / Purpose
- Listar transacoes com paginacao por cursor.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Query params:
  - `from` (date-time)
  - `to` (date-time)
  - `account_id` (uuid)
  - `category_id` (uuid)
  - `q` (texto)
  - `limit` (int, default 50, max 200)
  - `cursor_occurred_at` (date-time)
  - `cursor_id` (uuid)

4) Response
- 200
```json
{
  "items": [
    { "id": "uuid", "occurred_at": "2026-01-10", "description": "Mercado", "entries": [ ... ] }
  ],
  "next_cursor": {
    "cursor_occurred_at": "2026-01-01T00:00:00Z",
    "cursor_id": "uuid"
  }
}
```

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Ordenacao fixa: `occurred_at DESC, id DESC`.
- Usa CTE de IDs para evitar fantasmas.

7) Pagination
- Cursor obrigatorio; offset nao suportado.

8) Idempotency
- n/a.

---

### GET /ledgers/{ledgerId}/transactions/{transactionId}
1) Summary / Purpose
- Detalhe da transacao.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Path: `transactionId`.

4) Response
- 200 (transaction + entries).

5) Errors
- 404 `TRANSACTION_NOT_FOUND`

6) Semantics / Notes
- Busca por `(ledger_id, transaction_id)`.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### PATCH /ledgers/{ledgerId}/transactions/{transactionId}
1) Summary / Purpose
- Atualizar transacao.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body parcial, pode substituir entries:
```json
{
  "occurred_at": "2026-01-10",
  "description": "Mercado ajustado",
  "notes": "Revisao",
  "entries": [
    {
      "account_id": "uuid",
      "category_id": "uuid",
      "kind": "normal",
      "amount_cents": 8000,
      "memo": "Ajuste"
    }
  ]
}
```

4) Response
- 200 (transaction atualizada).

5) Errors
- 409 `TRANSACTION_REFERENCED`
- 404 `TRANSACTION_NOT_FOUND`

6) Semantics / Notes
- Bloquear edicao se referenciada por parcelas/fatura.

7) Pagination
- n/a.

8) Idempotency
- Opcional.

---

### DELETE /ledgers/{ledgerId}/transactions/{transactionId}
1) Summary / Purpose
- Remover transacao.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body: n/a.

4) Response
- 204.

5) Errors
- 409 `TRANSACTION_REFERENCED`
- 404 `TRANSACTION_NOT_FOUND`

6) Semantics / Notes
- Preferir ajuste via `kind=adjust`.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### POST /ledgers/{ledgerId}/transactions:bulk (opcional)
1) Summary / Purpose
- Criacao em lote.

2) Auth & Authorization
- Role: editor+.

3) Request
- Headers: `Idempotency-Key` (obrigatorio).
- Body:
```json
{ "items": [ { "occurred_at": "...", "entries": [ ... ] } ] }
```

4) Response
- 202 (aceito).

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Processamento assinc.

7) Pagination
- n/a.

8) Idempotency
- Obrigatoria por lote.
