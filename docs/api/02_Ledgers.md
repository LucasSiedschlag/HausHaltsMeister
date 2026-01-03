# API — Ledgers e Members

Este modulo cobre ledgers e controle de membros (RBAC) por ledger.

## Padroes globais aplicados

- JSON em `snake_case`.
- Datas e timestamps: ISO-8601 em UTC.
- IDs: UUID (string).
- Auth: `Authorization: Bearer <access_token>`.
- Ledger boundary obrigatorio em todos os endpoints.
- Roles por ledger: `owner`, `editor`, `viewer`.

## Inventario de endpoints

| Metodo | Path | Descricao | Auth | Role |
|---|---|---|---|---|
| GET | /ledgers | Listar ledgers do usuario | Sim | viewer |
| GET | /ledgers/{ledgerId}/me | Role do usuario no ledger | Sim | viewer |
| POST | /ledgers | Criar ledger | Sim | viewer |
| GET | /ledgers/{ledgerId} | Detalhe do ledger | Sim | viewer |
| PATCH | /ledgers/{ledgerId} | Atualizar ledger | Sim | editor |
| DELETE | /ledgers/{ledgerId} | Soft delete (planejado) | Sim | owner |
| GET | /ledgers/{ledgerId}/members | Listar membros | Sim | owner |
| POST | /ledgers/{ledgerId}/members | Adicionar membro | Sim | owner |
| PATCH | /ledgers/{ledgerId}/members/{userId} | Alterar role | Sim | owner |
| DELETE | /ledgers/{ledgerId}/members/{userId} | Remover membro | Sim | owner |

---

## Erros globais (5xx)

- 500 `INTERNAL_SERVER_ERROR`
- 503 `SERVICE_UNAVAILABLE`
- 504 `GATEWAY_TIMEOUT`

## Contratos detalhados

### GET /ledgers
1) Summary / Purpose
- Lista ledgers acessiveis ao usuario.

2) Auth & Authorization
- Token: sim.
- Role: viewer+.

3) Request
- Path params: n/a.
- Query params: n/a.
- Headers: `Authorization`.

4) Response
- 200
```json
[
  {
    "id": "uuid",
    "owner_user_id": "uuid",
    "name": "Pessoal",
    "currency_code": "BRL",
    "created_at": "...",
    "updated_at": "...",
    "role": "owner"
  }
]
```

5) Errors
- 401 `AUTH_INVALID_CREDENTIALS`

6) Semantics / Notes
- Filtra apenas ledgers do usuario.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### GET /ledgers/{ledgerId}/me
1) Summary / Purpose
- Retorna a role do usuario no ledger.

2) Auth & Authorization
- Token: sim.
- Role: viewer+.

3) Request
- Path params:
  - `ledgerId` (uuid, obrigatorio).

4) Response
- 200
```json
{
  "ledger_id": "uuid",
  "role": "viewer"
}
```

5) Errors
- 404 `LEDGER_NOT_FOUND`
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Usa o contexto do usuario autenticado.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### POST /ledgers
1) Summary / Purpose
- Criar ledger.

2) Auth & Authorization
- Token: sim.
- Role: viewer+.

3) Request
- Body:
```json
{ "name": "Pessoal", "currency_code": "BRL" }
```

4) Response
- 201
```json
{
  "id": "uuid",
  "owner_user_id": "uuid",
  "name": "Pessoal",
  "currency_code": "BRL",
  "created_at": "...",
  "updated_at": "...",
  "role": "owner"
}
```

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- O criador vira `owner`.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### GET /ledgers/{ledgerId}
1) Summary / Purpose
- Detalhe do ledger.

2) Auth & Authorization
- Token: sim.
- Role: viewer+.

3) Request
- Path params:
  - `ledgerId` (uuid, obrigatorio).

4) Response
- 200
```json
{
  "id": "uuid",
  "owner_user_id": "uuid",
  "name": "Pessoal",
  "currency_code": "BRL",
  "created_at": "...",
  "updated_at": "...",
  "role": "viewer"
}
```

5) Errors
- 404 `LEDGER_NOT_FOUND`
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Busca por `(user_id, ledger_id)`.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### PATCH /ledgers/{ledgerId}
1) Summary / Purpose
- Atualizar ledger.

2) Auth & Authorization
- Role: editor+.

3) Request
- Path: `ledgerId`.
- Body:
```json
{ "name": "Novo Nome" }
```

4) Response
- 200 (ledger atualizado).

5) Errors
- 404 `LEDGER_NOT_FOUND`
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Somente campos mutaveis.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### DELETE /ledgers/{ledgerId} (planejado)
1) Summary / Purpose
- Soft delete de ledger.

2) Auth & Authorization
- Role: owner.

3) Request
- Path: `ledgerId`.

4) Response
- 204.

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Nao remove historico.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### GET /ledgers/{ledgerId}/members
1) Summary / Purpose
- Listar membros e roles.

2) Auth & Authorization
- Role: owner.

3) Request
- Path: `ledgerId`.

4) Response
- 200
```json
[
  { "ledger_id": "uuid", "user_id": "uuid", "role": "viewer", "created_at": "...", "updated_at": "..." }
]
```

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Nao expor dados sensiveis de usuario.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### POST /ledgers/{ledgerId}/members
1) Summary / Purpose
- Adicionar membro ao ledger.

2) Auth & Authorization
- Role: owner.

3) Request
- Body:
```json
{ "user_id": "uuid", "role": "viewer" }
```

4) Response
- 201
```json
{ "ledger_id": "uuid", "user_id": "uuid", "role": "viewer", "created_at": "...", "updated_at": "..." }
```

5) Errors
- 409 `MEMBER_ALREADY_EXISTS`
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Validar role.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### PATCH /ledgers/{ledgerId}/members/{userId}
1) Summary / Purpose
- Atualizar role de um membro.

2) Auth & Authorization
- Role: owner.

3) Request
- Path: `userId`.
- Body:
```json
{ "role": "editor" }
```

4) Response
- 200 (membro atualizado).

5) Errors
- 403 `LEDGER_ACCESS_DENIED`
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Nao permitir remover o ultimo owner.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### DELETE /ledgers/{ledgerId}/members/{userId}
1) Summary / Purpose
- Remover membro.

2) Auth & Authorization
- Role: owner.

3) Request
- Path: `userId`.

4) Response
- 204.

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Nao permitir remover o ultimo owner.

7) Pagination
- n/a.

8) Idempotency
- n/a.
