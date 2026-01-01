# API — Convencoes

Este documento consolida os padroes globais e o template obrigatorio de contratos para todos os modulos de API.

## 1) Padroes globais

### 1.1 JSON e nomes
- JSON em `snake_case`.
- Campos de ID: `*_id` (UUID).

### 1.2 Datas, horarios e timezone
- Datas e timestamps: ISO-8601 em UTC.
- Meses: `YYYY-MM-01`.
- DB usa `timestamptz` e armazena UTC.

### 1.3 Dinheiro
- Usar `*_cents` (int64).
- Valores sempre inteiros e positivos; direcao e sinal sao definidos pelo contexto do entry.

### 1.4 Auth e sessoes
- Access token via `Authorization: Bearer <token>`.
- Refresh token via cookie HttpOnly.
- JWT curto (ex.: 15 min).
- Refresh longo (ex.: 30-90 dias).

### 1.5 RBAC por ledger
- Roles: `owner`, `editor`, `viewer`.
- Toda rota de dominio exige `ledgerId`.
- Sempre filtrar por `ledger_id` para evitar vazamentos.

### 1.6 Paginacao
- Cursor para journal (transactions).
- Ordenacao fixa: `occurred_at DESC, id DESC`.
- Cursor shape:
```json
{ "cursor_occurred_at": "2026-01-10T00:00:00Z", "cursor_id": "uuid" }
```

### 1.7 Erros
- Payload padrao:
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Human readable message",
    "details": { "field": "reason" }
  }
}
```

### 1.8 Enum values oficiais
- `category_direction`: `in`, `out`
- `entry_kind`: `normal`, `transfer`, `adjust`
- `installment_plan_status`: `active`, `cancelled`, `finished`
- `installment_status`: `scheduled`, `posted`, `paid`, `skipped`
- `statement_status`: `open`, `closed`, `paid`

### 1.9 Erros 5xx (globais)
- `INTERNAL_SERVER_ERROR` (500) — erro inesperado.
- `SERVICE_UNAVAILABLE` (503) — dependencia/servico indisponivel.
- `GATEWAY_TIMEOUT` (504) — timeout em dependencia.

---

## 2) Template obrigatorio por endpoint

Todo endpoint deve seguir este template:

1. Summary / Purpose
2. Auth & Authorization
   - requer token?
   - roles minimos
   - ledger boundary (sempre)
3. Request
   - Path params (tipos + validacao)
   - Query params (tipos + defaults + limites)
   - Headers (Authorization, Idempotency-Key se usar)
   - Body schema + exemplos
4. Response
   - Success status code
   - Body schema + exemplos
5. Errors
   - status codes
   - error codes do docs/agent/ERRORS.md
   - exemplos de erro
6. Semantics / Notes
   - invariantes importantes
   - side effects
7. Pagination (quando aplicavel)
   - cursor shape
8. Idempotency (quando aplicavel)

---

## 3) Padrões de exemplo

### Exemplo de request com Idempotency-Key
```
POST /ledgers/{ledgerId}/transactions
Idempotency-Key: 6e1a1f1c-7a3d-4cf6-8cbe-6d6f0f81b6c2
```

### Exemplo de erro
```json
{
  "error": {
    "code": "LEDGER_ACCESS_DENIED",
    "message": "Sem acesso ao ledger",
    "details": { "ledger_id": "uuid" }
  }
}
```
