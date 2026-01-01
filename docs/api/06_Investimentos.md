# API — Investimentos

## Endpoints

- `POST /ledgers/{ledgerId}/investments/contributions` — aporte.
- `POST /ledgers/{ledgerId}/investments/redemptions` — resgate.
- `POST /ledgers/{ledgerId}/investments/earnings` — rendimento.
- `GET /ledgers/{ledgerId}/investments/summary?from=&to=`

### POST /ledgers/{ledgerId}/investments/contributions

Body:
```json
{
  "amount_cents": 50000,
  "occurred_at": "2026-02-01",
  "memo": "Aporte mensal"
}
```

### POST /ledgers/{ledgerId}/investments/redemptions

Body:
```json
{
  "amount_cents": 50000,
  "occurred_at": "2026-03-01",
  "memo": "Resgate"
}
```

### POST /ledgers/{ledgerId}/investments/earnings

Body:
```json
{
  "amount_cents": 500,
  "occurred_at": "2026-03-31",
  "memo": "Rendimento mensal"
}
```
