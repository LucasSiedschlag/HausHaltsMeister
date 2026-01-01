# API — Accounts e Categories

## Accounts

- `GET /ledgers/{ledgerId}/accounts`
- `GET /ledgers/{ledgerId}/accounts/{accountId}`
- `POST /ledgers/{ledgerId}/accounts`
- `PATCH /ledgers/{ledgerId}/accounts/{accountId}`
- `DELETE /ledgers/{ledgerId}/accounts/{accountId}` (soft delete recomendado)

### POST /ledgers/{ledgerId}/accounts

Body:
```json
{
  "name": "Pessoal",
  "type": "cash",
  "is_active": true
}
```

Tipos validos: `cash`, `investment`, `credit_card`.

## Categories

- `GET /ledgers/{ledgerId}/categories?direction=in|out&active=true`
- `GET /ledgers/{ledgerId}/categories/{categoryId}`
- `POST /ledgers/{ledgerId}/categories`
- `PATCH /ledgers/{ledgerId}/categories/{categoryId}`
- `DELETE /ledgers/{ledgerId}/categories/{categoryId}` (soft delete recomendado)

### POST /ledgers/{ledgerId}/categories

Body:
```json
{
  "name": "Mercado",
  "direction": "out",
  "is_budget_base": false,
  "is_budget_relevant": true,
  "parent_id": null
}
```

Regras:
- `direction=in` -> `is_budget_relevant=false`
- `direction=out` -> `is_budget_base=false`
