# API — Budget (% flexivel)

## Planos e versoes

- `GET /ledgers/{ledgerId}/budget/plan`
- `POST /ledgers/{ledgerId}/budget/plan` (opcional, se nao criar automaticamente)
- `PATCH /ledgers/{ledgerId}/budget/plan`
- `DELETE /ledgers/{ledgerId}/budget/plan` (reset administrativo)

- `POST /ledgers/{ledgerId}/budget/versions`
- `GET /ledgers/{ledgerId}/budget/versions?from=&to=`
- `GET /ledgers/{ledgerId}/budget/versions/{versionId}`
- `PATCH /ledgers/{ledgerId}/budget/versions/{versionId}` (uso restrito)
- `DELETE /ledgers/{ledgerId}/budget/versions/{versionId}` (evitar em meses passados)

### POST /ledgers/{ledgerId}/budget/versions

Body:
```json
{
  "effective_from_month": "2026-01-01",
  "lines": [
    {"category_id": "uuid", "percent": 10, "include_children": false}
  ]
}
```

## Linhas do orçamento

- `POST /ledgers/{ledgerId}/budget/versions/{versionId}/lines`
- `PATCH /ledgers/{ledgerId}/budget/versions/{versionId}/lines/{lineId}`
- `DELETE /ledgers/{ledgerId}/budget/versions/{versionId}/lines/{lineId}`

## Painel mensal

- `GET /ledgers/{ledgerId}/budget/monthly?month=YYYY-MM-01`
