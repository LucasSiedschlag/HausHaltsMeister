# API — Ledgers e Members

## Ledgers

- `GET /ledgers`
- `POST /ledgers`
- `GET /ledgers/{ledgerId}`
- `PATCH /ledgers/{ledgerId}`
- `DELETE /ledgers/{ledgerId}` (soft delete recomendado)

### POST /ledgers

Body:
```json
{
  "name": "Pessoal",
  "currency_code": "BRL"
}
```

### PATCH /ledgers/{ledgerId}

Body:
```json
{
  "name": "Novo Nome"
}
```

## Members (RBAC)

- `GET /ledgers/{ledgerId}/members`
- `POST /ledgers/{ledgerId}/members`
- `PATCH /ledgers/{ledgerId}/members/{userId}`
- `DELETE /ledgers/{ledgerId}/members/{userId}`

### POST /ledgers/{ledgerId}/members

Body:
```json
{
  "user_id": "uuid",
  "role": "viewer"
}
```

Roles validos: `owner`, `editor`, `viewer`.
