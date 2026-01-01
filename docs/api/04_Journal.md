# API — Journal (Transactions + Entries)

## Endpoints

- `POST /ledgers/{ledgerId}/transactions`
- `GET /ledgers/{ledgerId}/transactions?from=&to=&account_id=&category_id=&q=&limit=&cursor_occurred_at=&cursor_id=`
- `GET /ledgers/{ledgerId}/transactions/{transactionId}`
- `PATCH /ledgers/{ledgerId}/transactions/{transactionId}`
- `DELETE /ledgers/{ledgerId}/transactions/{transactionId}`
- `POST /ledgers/{ledgerId}/transactions:bulk` (opcional)

### POST /ledgers/{ledgerId}/transactions

Body:
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

`kind` valido: `normal`, `transfer`, `adjust`.

Regras:
- `transfer` exige pelo menos uma IN e uma OUT, balanceadas.
- `amount_cents` sempre positivo.

## Listagem (GET /transactions) — paginacao segura

**Ordenacao fixa:** `occurred_at DESC, id DESC`.

**Cursor estavel:** usar `cursor_occurred_at` + `cursor_id`.

**Evitar offset:** se for mantido, documentar como instavel sob insercoes concorrentes.

Fluxo recomendado (2 etapas, evita “fantasmas”):
1. **CTE de IDs** com filtros e ordenacao:
   - aplica `from`, `to`, `account_id`, `category_id`, `q`.
   - limita por `limit`.
   - se cursor presente, filtra `occurred_at < cursor_occurred_at OR (occurred_at = cursor_occurred_at AND id < cursor_id)`.
2. **Busca detalhada** de transactions + `json_agg(entries)` apenas para esses IDs.

Resposta deve incluir `next_cursor` com `occurred_at` e `id` do ultimo item.

## GET por id

- Buscar por `(ledger_id, transaction_id)`.
- Retornar transaction + entries (sem paginacao).

## PATCH / DELETE (restricoes minimas)

- Bloquear DELETE/edicao quebradora quando a transaction estiver referenciada por:
  - `installments.posted_transaction_id`
  - `credit_card_statements.payment_transaction_id`
- Em caso de conflito, exigir ajuste via `kind=adjust`.

### PATCH /ledgers/{ledgerId}/transactions/{transactionId}

Body (campos opcionais, substitui entries quando enviado):
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
