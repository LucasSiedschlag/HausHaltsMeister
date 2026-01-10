# API Contracts (Frontend Reference)

Este documento e a fonte de verdade de contratos entre frontend e backend. Ele cobre todos os modulos do projeto e inclui rotas existentes e planejadas. Sempre que um endpoint novo for criado, ele deve ser adicionado aqui seguindo o template fixo.

## 1) Escopo fechado (modulos cobertos)

- Auth
- Ledgers (e Members)
- Accounts
- Categories
- Journal (transactions + entries)
- Budget
- Investments
- Credit Card (Cartao)
- Reports (Relatorios)

---

## 2) Padroes globais (obrigatorio)

### 2.1 Datas, UUID, dinheiro
- Datas e timestamps: ISO-8601 em UTC.
- Meses: `YYYY-MM-01` (date).
- IDs: UUID (string).
- Dinheiro: `*_cents` (int64).

### 2.2 JSON e nomes
- JSON sempre em `snake_case`.
- DB em `snake_case`.

### 2.3 Auth e sessao
- Access token em `Authorization: Bearer <token>`.
- Refresh token em cookie HttpOnly (web).
- Roles por ledger: `owner`, `editor`, `viewer`.
- Ledger boundary e obrigatorio em todos os endpoints de dominio.

### 2.4 Paginacao (cursor)
- Apenas Journal usa cursor.
- Ordenacao fixa: `occurred_at DESC, id DESC`.
- Cursor shape:
```json
{
  "cursor_occurred_at": "2026-01-10T00:00:00Z",
  "cursor_id": "uuid"
}
```

### 2.5 Formato de erro (padrao)
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Human readable message",
    "details": {
      "field": "reason"
    }
  }
}
```

### 2.6 Enum values oficiais
- `category_direction`: `in`, `out`
- `entry_kind`: `normal`, `transfer`, `adjust`
- `installment_plan_status`: `active`, `cancelled`, `finished`
- `installment_status`: `scheduled`, `posted`, `paid`, `skipped`
- `statement_status`: `open`, `closed`, `paid`

### 2.7 Erros 5xx (globais)
- `INTERNAL_SERVER_ERROR` (500) — erro inesperado.
- `SERVICE_UNAVAILABLE` (503) — dependencia/servico indisponivel.
- `GATEWAY_TIMEOUT` (504) — timeout em dependencia.

---

## 3) Inventario de endpoints por modulo

### Auth
| Metodo | Path | Descricao | Auth | Role |
|---|---|---|---|---|
| POST | /auth/signup | Criar usuario e ledger padrao | Nao | - |
| POST | /auth/login | Login | Nao | - |
| POST | /auth/refresh | Rotacionar sessao | Nao (cookie) | - |
| POST | /auth/logout | Encerrar sessao | Sim | viewer |
| GET | /auth/me | Usuario autenticado | Sim | viewer |
| GET | /auth/oauth/{provider}/start | OAuth start (futuro) | Nao | - |
| GET | /auth/oauth/{provider}/callback | OAuth callback (futuro) | Nao | - |
| POST | /auth/forgot-password | Reset (fase 2) | Nao | - |
| POST | /auth/reset-password | Reset (fase 2) | Nao | - |
| POST | /auth/verify-email | Verificacao (fase 2) | Nao | - |
| POST | /auth/resend-verification | Reenvio (fase 2) | Nao | - |
| GET | /auth/sessions | Listar sessoes (opcional) | Sim | viewer |
| DELETE | /auth/sessions/{sessionId} | Encerrar sessao (opcional) | Sim | viewer |
| POST | /auth/logout-all | Encerrar todas as sessoes (opcional) | Sim | viewer |
| GET | /auth/providers | Providers ativos (opcional) | Sim | viewer |
| POST | /auth/link/{provider}/start | Linkar provider (opcional) | Sim | viewer |
| POST | /auth/unlink/{provider} | Desvincular provider (opcional) | Sim | viewer |

### Ledgers
| Metodo | Path | Descricao | Auth | Role |
|---|---|---|---|---|
| GET | /ledgers | Listar ledgers do usuario | Sim | viewer |
| POST | /ledgers | Criar ledger | Sim | viewer |
| GET | /ledgers/{ledgerId} | Detalhe do ledger | Sim | viewer |
| PATCH | /ledgers/{ledgerId} | Atualizar ledger | Sim | editor |
| DELETE | /ledgers/{ledgerId} | Soft delete (planejado) | Sim | owner |
| GET | /ledgers/{ledgerId}/members | Listar membros | Sim | owner |
| POST | /ledgers/{ledgerId}/members | Adicionar membro | Sim | owner |
| PATCH | /ledgers/{ledgerId}/members/{userId} | Alterar role | Sim | owner |
| DELETE | /ledgers/{ledgerId}/members/{userId} | Remover membro | Sim | owner |

### Accounts
| Metodo | Path | Descricao | Auth | Role |
|---|---|---|---|---|
| GET | /ledgers/{ledgerId}/accounts | Listar contas | Sim | viewer |
| GET | /ledgers/{ledgerId}/accounts/{accountId} | Detalhe da conta | Sim | viewer |
| POST | /ledgers/{ledgerId}/accounts | Criar conta | Sim | editor |
| PATCH | /ledgers/{ledgerId}/accounts/{accountId} | Atualizar conta | Sim | editor |
| DELETE | /ledgers/{ledgerId}/accounts/{accountId} | Soft delete (planejado) | Sim | editor |

### Categories
| Metodo | Path | Descricao | Auth | Role |
|---|---|---|---|---|
| GET | /ledgers/{ledgerId}/categories | Listar categorias | Sim | viewer |
| GET | /ledgers/{ledgerId}/categories/{categoryId} | Detalhe da categoria | Sim | viewer |
| POST | /ledgers/{ledgerId}/categories | Criar categoria | Sim | editor |
| PATCH | /ledgers/{ledgerId}/categories/{categoryId} | Atualizar categoria | Sim | editor |
| DELETE | /ledgers/{ledgerId}/categories/{categoryId} | Soft delete (planejado) | Sim | editor |

### Journal
| Metodo | Path | Descricao | Auth | Role |
|---|---|---|---|---|
| POST | /ledgers/{ledgerId}/transactions | Criar transacao | Sim | editor |
| GET | /ledgers/{ledgerId}/transactions | Listar transacoes (cursor) | Sim | viewer |
| GET | /ledgers/{ledgerId}/transactions/{transactionId} | Detalhe da transacao | Sim | viewer |
| PATCH | /ledgers/{ledgerId}/transactions/{transactionId} | Atualizar transacao | Sim | editor |
| DELETE | /ledgers/{ledgerId}/transactions/{transactionId} | Deletar transacao | Sim | editor |
| POST | /ledgers/{ledgerId}/transactions:bulk | Criacao em lote (opcional) | Sim | editor |

### Budget
| Metodo | Path | Descricao | Auth | Role |
|---|---|---|---|---|
| GET | /ledgers/{ledgerId}/budget/plan | Obter plano | Sim | viewer |
| POST | /ledgers/{ledgerId}/budget/plan | Criar plano | Sim | editor |
| PATCH | /ledgers/{ledgerId}/budget/plan | Atualizar plano | Sim | editor |
| DELETE | /ledgers/{ledgerId}/budget/plan | Reset administrativo | Sim | owner |
| POST | /ledgers/{ledgerId}/budget/versions | Criar versao | Sim | editor |
| GET | /ledgers/{ledgerId}/budget/versions | Listar versoes | Sim | viewer |
| GET | /ledgers/{ledgerId}/budget/versions/{versionId} | Detalhe da versao | Sim | viewer |
| PATCH | /ledgers/{ledgerId}/budget/versions/{versionId} | Atualizar (restrito) | Sim | owner |
| DELETE | /ledgers/{ledgerId}/budget/versions/{versionId} | Remover versao | Sim | owner |
| POST | /ledgers/{ledgerId}/budget/versions/{versionId}/lines | Criar linha | Sim | editor |
| PATCH | /ledgers/{ledgerId}/budget/versions/{versionId}/lines/{lineId} | Atualizar linha | Sim | editor |
| DELETE | /ledgers/{ledgerId}/budget/versions/{versionId}/lines/{lineId} | Remover linha | Sim | editor |
| GET | /ledgers/{ledgerId}/budget/monthly | Painel mensal | Sim | viewer |
| GET | /ledgers/{ledgerId}/budget/period | Painel por periodo | Sim | viewer |

### Investments
| Metodo | Path | Descricao | Auth | Role |
|---|---|---|---|---|
| POST | /ledgers/{ledgerId}/investments/contributions | Aporte | Sim | editor |
| POST | /ledgers/{ledgerId}/investments/redemptions | Resgate | Sim | editor |
| POST | /ledgers/{ledgerId}/investments/earnings | Rendimento | Sim | editor |
| GET | /ledgers/{ledgerId}/investments/summary | Resumo | Sim | viewer |

### Credit Card (Cartao)
| Metodo | Path | Descricao | Auth | Role |
|---|---|---|---|---|
| GET | /card-networks | Listar bandeiras | Sim | viewer |
| POST | /card-networks | Criar bandeira | Sim | owner |
| PATCH | /card-networks/{code} | Atualizar bandeira | Sim | owner |
| DELETE | /card-networks/{code} | Remover bandeira | Sim | owner |
| GET | /ledgers/{ledgerId}/credit-cards | Listar cartoes | Sim | viewer |
| GET | /ledgers/{ledgerId}/credit-cards/{cardAccountId} | Detalhe do cartao | Sim | viewer |
| POST | /ledgers/{ledgerId}/credit-cards | Criar cartao | Sim | editor |
| PATCH | /ledgers/{ledgerId}/credit-cards/{cardAccountId} | Atualizar cartao | Sim | editor |
| DELETE | /ledgers/{ledgerId}/credit-cards/{cardAccountId} | Remover cartao | Sim | editor |
| POST | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans | Criar plano | Sim | editor |
| GET | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans | Listar planos | Sim | viewer |
| GET | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans/{planId} | Detalhe do plano | Sim | viewer |
| PATCH | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans/{planId} | Atualizar plano | Sim | editor |
| DELETE | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans/{planId} | Remover plano | Sim | editor |
| GET | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/installments | Listar parcelas | Sim | viewer |
| PATCH | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/installments/{installmentId} | Atualizar parcela | Sim | editor |
| POST | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/post | Postar parcelas do mes | Sim | editor |
| GET | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements | Listar faturas | Sim | viewer |
| GET | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/{statementId} | Detalhe da fatura | Sim | viewer |
| POST | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/close | Fechar fatura | Sim | editor |
| POST | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/pay | Pagar fatura | Sim | editor |
| PATCH | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/{statementId} | Atualizar (restrito) | Sim | owner |

### Reports (Relatorios)
| Metodo | Path | Descricao | Auth | Role |
|---|---|---|---|---|
| GET | /ledgers/{ledgerId}/reports/balances | Saldos por conta | Sim | viewer |
| GET | /ledgers/{ledgerId}/reports/categories | Categoria x gasto | Sim | viewer |
| GET | /ledgers/{ledgerId}/reports/cashflow | Fluxo de caixa | Sim | viewer |

---

## 4) Contratos detalhados (template fixo)

### Auth

#### POST /auth/signup
1) Summary / Purpose
- Criar usuario e ledger padrao.

2) Auth & Authorization
- Token: nao.
- Role: n/a.

3) Request
- Path params: n/a.
- Query params: n/a.
- Headers: `Content-Type: application/json`.
- Body:
```json
{
  "email": "user@example.com",
  "password": "min8chars",
  "display_name": "Nome"
}
```

4) Response
- 201
```json
{
  "access_token": "...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "display_name": "Nome"
  }
}
```

5) Errors
- 409 `CONFLICT_DUPLICATE_EMAIL`
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Cria ledger padrao e associa role owner.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### POST /auth/login
1) Summary / Purpose
- Autenticar usuario.

2) Auth & Authorization
- Token: nao.

3) Request
- Body:
```json
{ "email": "user@example.com", "password": "..." }
```

4) Response
- 200
```json
{
  "access_token": "...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "display_name": "Nome"
  }
}
```

5) Errors
- 401 `AUTH_INVALID_CREDENTIALS`
- 403 `AUTH_USER_INACTIVE`

6) Semantics / Notes
- Refresh token rotacionavel e criado.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### POST /auth/refresh
1) Summary / Purpose
- Rotacionar refresh token e emitir novo access.

2) Auth & Authorization
- Token: refresh em cookie HttpOnly.

3) Request
- Headers: Cookie com refresh.
- Body: vazio.

4) Response
- 200 (mesmo payload do login).

5) Errors
- 401 `AUTH_REFRESH_REVOKED`

6) Semantics / Notes
- Rotacao invalida o refresh anterior.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### POST /auth/logout
1) Summary / Purpose
- Revogar sessao atual.

2) Auth & Authorization
- Token: sim.

3) Request
- Body: vazio.

4) Response
- 204

5) Errors
- 401 `AUTH_INVALID_CREDENTIALS`

6) Semantics / Notes
- Revoga refresh atual.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### GET /auth/me
1) Summary / Purpose
- Retornar perfil basico.

2) Auth & Authorization
- Token: sim.

3) Request
- Path/query: n/a.

4) Response
- 200
```json
{ "id": "uuid", "email": "user@example.com", "display_name": "Nome" }
```

5) Errors
- 401 `AUTH_INVALID_CREDENTIALS`

6) Semantics / Notes
- Pode incluir ledgers no futuro.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### GET /auth/oauth/{provider}/start (futuro)
1) Summary / Purpose
- Inicia OAuth (PKCE).

2) Auth & Authorization
- Token: nao.

3) Request
- Path params: `provider` (google|github).

4) Response
- 302 redirect.

5) Errors
- 502 `AUTH_OAUTH_PROVIDER_ERROR`

6) Semantics / Notes
- Gera state + code_verifier.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### GET /auth/oauth/{provider}/callback (futuro)
1) Summary / Purpose
- Finaliza OAuth e cria sessao.

2) Auth & Authorization
- Token: nao.

3) Request
- Query: `code`, `state`.

4) Response
- 302 redirect + cookie refresh.

5) Errors
- 401 `AUTH_OAUTH_STATE_INVALID`
- 422 `AUTH_OAUTH_EMAIL_REQUIRED`

6) Semantics / Notes
- Cria/vincula `auth_identities`.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### Ledgers

#### GET /ledgers
1) Summary / Purpose
- Lista ledgers acessiveis.

2) Auth & Authorization
- Token: sim.
- Role: viewer+.

3) Request
- Query: n/a.

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
- Sempre filtra por user_id.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### GET /ledgers/{ledgerId}/me
1) Summary / Purpose
- Role do usuario no ledger.

2) Auth & Authorization
- Token: sim.
- Role: viewer+.

3) Request
- Path: `ledgerId` (uuid).

4) Response
- 200
```json
{ "ledger_id": "uuid", "role": "viewer" }
```

5) Errors
- 404 `LEDGER_NOT_FOUND`
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Usa o usuario autenticado.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### POST /ledgers
1) Summary / Purpose
- Cria novo ledger.

2) Auth & Authorization
- Token: sim.
- Role: viewer+.

3) Request
- Body:
```json
{ "name": "Pessoal", "currency_code": "BRL" }
```

4) Response
- 201 (ledger + role).

5) Errors
- 422 `VALIDATION_ERROR`
- 429 `RATE_LIMITED`

6) Semantics / Notes
- User vira owner.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### GET /ledgers/{ledgerId}
1) Summary / Purpose
- Detalhe do ledger.

2) Auth & Authorization
- Token: sim.
- Role: viewer+.

3) Request
- Path: `ledgerId` (uuid).

4) Response
- 200 (ledger + role, flat).

5) Errors
- 404 `LEDGER_NOT_FOUND`
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- ledger boundary obrigatorio.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### PATCH /ledgers/{ledgerId}
1) Summary / Purpose
- Atualizar ledger.

2) Auth & Authorization
- Role: editor+.

3) Request
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

#### DELETE /ledgers/{ledgerId} (planejado)
1) Summary / Purpose
- Soft delete de ledger.

2) Auth & Authorization
- Role: owner.

3) Request
- Body: n/a.

4) Response
- 204.

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Mantem historico.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### GET /ledgers/{ledgerId}/members
1) Summary / Purpose
- Listar membros.

2) Auth & Authorization
- Role: owner.

3) Request
- Path: `ledgerId`.

4) Response
- 200
```json
[
  {
    "ledger_id": "uuid",
    "user_id": "uuid",
    "role": "viewer",
    "display_name": "Nome",
    "email": "user@example.com",
    "avatar_url": "https://...",
    "created_at": "...",
    "updated_at": "..."
  }
]
```

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Expor apenas nome, email e avatar (sem dados sensiveis adicionais).

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### POST /ledgers/{ledgerId}/members
1) Summary / Purpose
- Adicionar membro.

2) Auth & Authorization
- Role: owner.

3) Request
- Body:
```json
{ "user_id": "uuid", "email": "user@example.com", "role": "viewer" }
```

4) Response
- 201
```json
{
  "ledger_id": "uuid",
  "user_id": "uuid",
  "role": "viewer",
  "display_name": "Nome",
  "email": "user@example.com",
  "avatar_url": "https://...",
  "created_at": "...",
  "updated_at": "..."
}
```

5) Errors
- 409 `MEMBER_ALREADY_EXISTS`
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Enviar `user_id` **ou** `email` (um dos dois).
- Role valida: owner/editor/viewer.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### PATCH /ledgers/{ledgerId}/members/{userId}
1) Summary / Purpose
- Atualizar role.

2) Auth & Authorization
- Role: owner.

3) Request
- Body:
```json
{ "role": "editor" }
```

4) Response
- 200 (membro atualizado).

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Nao remover ultimo owner.
- Nao remove historico (soft remove via `removed_at`).

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### DELETE /ledgers/{ledgerId}/members/{userId}
1) Summary / Purpose
- Remover membro.

2) Auth & Authorization
- Role: owner.

3) Request
- Body: n/a.

4) Response
- 204.

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Nao remover ultimo owner.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### Accounts

#### GET /ledgers/{ledgerId}/accounts
1) Summary / Purpose
- Listar contas.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Query: `active` (bool, default true).

4) Response
- 200
```json
[
  { "id": "uuid", "ledger_id": "uuid", "name": "Pessoal", "type": "cash", "is_active": true, "created_at": "...", "updated_at": "..." }
]
```

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Ledger boundary obrigatorio.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### GET /ledgers/{ledgerId}/accounts/{accountId}
1) Summary / Purpose
- Detalhe de conta.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Path: `ledgerId`, `accountId`.

4) Response
- 200 (conta).

5) Errors
- 404 `ACCOUNT_NOT_FOUND`
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- n/a.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### POST /ledgers/{ledgerId}/accounts
1) Summary / Purpose
- Criar conta.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body:
```json
{ "name": "Pessoal", "type": "cash", "is_active": true }
```

4) Response
- 201 (conta criada).

5) Errors
- 409 `CONFLICT_DUPLICATE_NAME`
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- `type` atual usa regras via CHECK.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### PATCH /ledgers/{ledgerId}/accounts/{accountId}
1) Summary / Purpose
- Atualizar conta.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body (parcial):
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

#### DELETE /ledgers/{ledgerId}/accounts/{accountId}
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
- Pode virar inativo.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### Categories

#### GET /ledgers/{ledgerId}/categories
1) Summary / Purpose
- Listar categorias.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Query: `direction` (in|out), `active` (bool).

4) Response
- 200
```json
[
  { "id": "uuid", "ledger_id": "uuid", "name": "Mercado", "direction": "out", "is_budget_base": false, "is_budget_relevant": true, "parent_id": null, "created_at": "...", "updated_at": "..." }
]
```

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Regras: `direction=in` -> `is_budget_relevant=false`; `direction=out` -> `is_budget_base=false`.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### GET /ledgers/{ledgerId}/categories/{categoryId}
1) Summary / Purpose
- Detalhe de categoria.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Path: `categoryId`.

4) Response
- 200 (categoria).

5) Errors
- 404 `CATEGORY_NOT_FOUND`

6) Semantics / Notes
- n/a.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### POST /ledgers/{ledgerId}/categories
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
- 201 (categoria criada).

5) Errors
- 409 `CONFLICT_DUPLICATE_NAME`
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- `parent_id` opcional.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### PATCH /ledgers/{ledgerId}/categories/{categoryId}
1) Summary / Purpose
- Atualizar categoria.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body parcial.

4) Response
- 200 (categoria atualizada).

5) Errors
- 404 `CATEGORY_NOT_FOUND`
- 409 `CONFLICT_DUPLICATE_NAME`

6) Semantics / Notes
- Validacoes por direction.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### DELETE /ledgers/{ledgerId}/categories/{categoryId}
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

---

### Journal

#### POST /ledgers/{ledgerId}/transactions
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
- 429 `RATE_LIMITED`
- 422 `TRANSFER_NOT_BALANCED`
- 409 `IDEMPOTENCY_KEY_CONFLICT`
- 409 `IDEMPOTENCY_KEY_EXPIRED`

6) Semantics / Notes
- `amount_cents` sempre positivo.
- `transfer` deve balancear IN/OUT.

7) Pagination
- n/a.

8) Idempotency
- `Idempotency-Key` evita duplicidade.
- Replays validos retornam a transacao original.

#### GET /ledgers/{ledgerId}/transactions
1) Summary / Purpose
- Listar transacoes com paginacao por cursor.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Query: `from`, `to`, `account_id`, `category_id`, `q`, `limit` (default 50, max 200), `cursor_occurred_at`, `cursor_id`.

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
- Usa CTE de IDs para evitar fantasmas.

7) Pagination
- Cursor obrigatorio (offset nao suportado).

8) Idempotency
- n/a.

#### GET /ledgers/{ledgerId}/transactions/{transactionId}
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
- Busca por (ledger_id, transaction_id).

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### PATCH /ledgers/{ledgerId}/transactions/{transactionId}
1) Summary / Purpose
- Atualizar transacao.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body parcial, pode substituir entries.

4) Response
- 200 (transaction atualizada).

5) Errors
- 409 `TRANSACTION_REFERENCED`
- 404 `TRANSACTION_NOT_FOUND`

6) Semantics / Notes
- Bloquear edicao se referenciada.

7) Pagination
- n/a.

8) Idempotency
- Opcional (Idempotency-Key).

#### DELETE /ledgers/{ledgerId}/transactions/{transactionId}
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

#### POST /ledgers/{ledgerId}/transactions:bulk (opcional)
1) Summary / Purpose
- Criacao em lote.

2) Auth & Authorization
- Role: editor+.

3) Request
- Headers: `Idempotency-Key`.
- Body:
```json
{ "items": [ { "occurred_at": "...", "entries": [ ... ] } ] }
```

4) Response
- 202 (aceito).

5) Errors
- 422 `VALIDATION_ERROR`
- 429 `RATE_LIMITED`

6) Semantics / Notes
- Processamento assinc.

7) Pagination
- n/a.

8) Idempotency
- Obrigatoria para lote.

---

### Budget

#### GET /ledgers/{ledgerId}/budget/plan
1) Summary / Purpose
- Retornar plano atual.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Query: n/a.

4) Response
- 200
```json
{ "id": "uuid", "ledger_id": "uuid", "name": "Padrao", "is_active": true, "created_at": "...", "updated_at": "..." }
```

5) Errors
- 404 `LEDGER_NOT_FOUND`

6) Semantics / Notes
- Pode ser criado automaticamente no signup.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### POST /ledgers/{ledgerId}/budget/plan
1) Summary / Purpose
- Criar plano.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body:
```json
{ "name": "Padrao", "is_active": true }
```

4) Response
- 201 (plano).

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Apenas um plano ativo por ledger.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### PATCH /ledgers/{ledgerId}/budget/plan
1) Summary / Purpose
- Atualizar plano.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body parcial.

4) Response
- 200.

5) Errors
- 404 `LEDGER_NOT_FOUND`

6) Semantics / Notes
- Pode desativar.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### DELETE /ledgers/{ledgerId}/budget/plan
1) Summary / Purpose
- Reset administrativo.

2) Auth & Authorization
- Role: owner.

3) Request
- Body: n/a.

4) Response
- 204.

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Uso restrito.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### POST /ledgers/{ledgerId}/budget/versions
1) Summary / Purpose
- Criar versao de budget.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body:
```json
{
  "effective_from_month": "2026-01-01",
  "lines": [ { "category_id": "uuid", "percent": 10, "include_children": false } ]
}
```

4) Response
- 201 (versao).

5) Errors
- 409 `BUDGET_VERSION_CONFLICT`
- 422 `BUDGET_CATEGORY_NOT_ELIGIBLE`

6) Semantics / Notes
- Percentual baseado em renda base.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### GET /ledgers/{ledgerId}/budget/versions
1) Summary / Purpose
- Listar versoes.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Query: `from`, `to` (month).

4) Response
- 200 (lista de versoes).

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Pode filtrar intervalo.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### GET /ledgers/{ledgerId}/budget/versions/{versionId}
1) Summary / Purpose
- Detalhe de versao.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Path: `versionId`.

4) Response
- 200 (versao + linhas).

5) Errors
- 404 `BUDGET_VERSION_CONFLICT` (usar not found quando aplicavel).

6) Semantics / Notes
- Inclui linhas.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### PATCH /ledgers/{ledgerId}/budget/versions/{versionId}
1) Summary / Purpose
- Atualizar versao (restrito).

2) Auth & Authorization
- Role: owner.

3) Request
- Body: parcial.

4) Response
- 200.

5) Errors
- 501 `NOT_IMPLEMENTED`

6) Semantics / Notes
- Normalmente evitar em meses fechados.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### DELETE /ledgers/{ledgerId}/budget/versions/{versionId}
1) Summary / Purpose
- Remover versao.

2) Auth & Authorization
- Role: owner.

3) Request
- Body: n/a.

4) Response
- 204.

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Evitar em meses passados.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### POST /ledgers/{ledgerId}/budget/versions/{versionId}/lines
1) Summary / Purpose
- Criar linha.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body:
```json
{ "category_id": "uuid", "percent": 10, "include_children": false }
```

4) Response
- 201 (linha).

5) Errors
- 422 `BUDGET_CATEGORY_NOT_ELIGIBLE`

6) Semantics / Notes
- Percent > 0.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### PATCH /ledgers/{ledgerId}/budget/versions/{versionId}/lines/{lineId}
1) Summary / Purpose
- Atualizar linha.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body parcial.

4) Response
- 200.

5) Errors
- 404 `VALIDATION_ERROR` (quando linha nao existir).

6) Semantics / Notes
- Ajusta percentuais.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### DELETE /ledgers/{ledgerId}/budget/versions/{versionId}/lines/{lineId}
1) Summary / Purpose
- Remover linha.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body: n/a.

4) Response
- 204.

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Nao afeta historico mensal.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### GET /ledgers/{ledgerId}/budget/monthly
1) Summary / Purpose
- Resumo mensal.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Query: `month` (YYYY-MM-01, required).

4) Response
- 200 (resumo + linhas).

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Mostra orcado vs realizado.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### GET /ledgers/{ledgerId}/budget/period
1) Summary / Purpose
- Resumo por periodo.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Query: `from`, `to` (YYYY-MM-01, required).

4) Response
- 200 (agregacao + totais).

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Retorna totals por categoria.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### Investments

#### POST /ledgers/{ledgerId}/investments/contributions
1) Summary / Purpose
- Registrar aporte.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body:
```json
{ "amount_cents": 50000, "occurred_at": "2026-02-01", "memo": "Aporte" }
```

4) Response
- 201 (transaction criada).

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Gera entries de investimento.

7) Pagination
- n/a.

8) Idempotency
- opcional.

#### POST /ledgers/{ledgerId}/investments/redemptions
1) Summary / Purpose
- Registrar resgate.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body:
```json
{ "amount_cents": 50000, "occurred_at": "2026-03-01", "memo": "Resgate" }
```

4) Response
- 201 (transaction criada).

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Movimenta saldo de investimento.

7) Pagination
- n/a.

8) Idempotency
- opcional.

#### POST /ledgers/{ledgerId}/investments/earnings
1) Summary / Purpose
- Registrar rendimento.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body:
```json
{ "amount_cents": 500, "occurred_at": "2026-03-31", "memo": "Rendimento" }
```

4) Response
- 201 (transaction criada).

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Lanca ganho no investimento.

7) Pagination
- n/a.

8) Idempotency
- opcional.

#### GET /ledgers/{ledgerId}/investments/summary
1) Summary / Purpose
- Resumo de investimentos.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Query: `from`, `to` (date-time).

4) Response
- 200
```json
{
  "ledger_id": "uuid",
  "from": "2026-02-01T00:00:00Z",
  "to": "2026-03-31T23:59:59Z",
  "total_contributions": 0,
  "total_redemptions": 0,
  "total_earnings": 0,
  "total_losses": 0,
  "net_variation": 0
}
```

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Intervalo fechado.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### Credit Card (Cartao)

#### GET /card-networks
1) Summary / Purpose
- Listar bandeiras.

2) Auth & Authorization
- Role: viewer+.

3) Request
- n/a.

4) Response
- 200
```json
[ { "code": "visa", "name": "Visa", "is_active": true } ]
```

5) Errors
- 401 `AUTH_INVALID_CREDENTIALS`

6) Semantics / Notes
- Catalogo global.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### POST /card-networks
1) Summary / Purpose
- Criar bandeira.

2) Auth & Authorization
- Role: owner.

3) Request
- Body:
```json
{ "code": "visa", "name": "Visa", "is_active": true }
```

4) Response
- 201 (bandeira).

5) Errors
- 409 `CONFLICT_DUPLICATE_NAME`

6) Semantics / Notes
- `code` unico.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### PATCH /card-networks/{code}
1) Summary / Purpose
- Atualizar bandeira.

2) Auth & Authorization
- Role: owner.

3) Request
- Body parcial.

4) Response
- 200.

5) Errors
- 404 `VALIDATION_ERROR`.

6) Semantics / Notes
- Ativar/desativar.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### DELETE /card-networks/{code}
1) Summary / Purpose
- Remover bandeira.

2) Auth & Authorization
- Role: owner.

3) Request
- n/a.

4) Response
- 204.

5) Errors
- 404 `VALIDATION_ERROR`.

6) Semantics / Notes
- Restrito.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### POST /ledgers/{ledgerId}/credit-cards
1) Summary / Purpose
- Criar cartao.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body:
```json
{
  "account_id": "uuid",
  "issuer_name": "Banco X",
  "network": "visa",
  "nickname": "Cartao Principal",
  "last4": "1234",
  "credit_limit_cents": 500000,
  "closing_day": 25,
  "due_day": 10
}
```

4) Response
- 201 (cartao).

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- `account_id` deve ser tipo credit_card.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### GET /ledgers/{ledgerId}/credit-cards
1) Summary / Purpose
- Listar cartoes.

2) Auth & Authorization
- Role: viewer+.

3) Request
- n/a.

4) Response
- 200 (lista).

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Ledger boundary.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### GET /ledgers/{ledgerId}/credit-cards/{cardAccountId}
1) Summary / Purpose
- Detalhe do cartao.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Path: `cardAccountId`.

4) Response
- 200 (cartao).

5) Errors
- 404 `CREDITCARD_CARD_NOT_FOUND`

6) Semantics / Notes
- n/a.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### PATCH /ledgers/{ledgerId}/credit-cards/{cardAccountId}
1) Summary / Purpose
- Atualizar cartao.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body parcial.

4) Response
- 200.

5) Errors
- 404 `CREDITCARD_CARD_NOT_FOUND`

6) Semantics / Notes
- Permite alterar limites e datas.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### DELETE /ledgers/{ledgerId}/credit-cards/{cardAccountId}
1) Summary / Purpose
- Remover cartao.

2) Auth & Authorization
- Role: editor+.

3) Request
- n/a.

4) Response
- 204.

5) Errors
- 404 `CREDITCARD_CARD_NOT_FOUND`

6) Semantics / Notes
- Bloquear se houver planos ativos.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### POST /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans
1) Summary / Purpose
- Criar plano de parcelas.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body:
```json
{
  "purchase_occurred_at": "2026-02-15",
  "merchant": "Loja X",
  "description": "Notebook",
  "category_id": "uuid",
  "total_amount_cents": 600000,
  "installments_count": 6,
  "installment_amount_cents": 100000,
  "first_due_month": "2026-03-01"
}
```

4) Response
- 201 (plano).

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Compra vira plano + parcelas.

7) Pagination
- n/a.

8) Idempotency
- opcional.

#### GET /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans
1) Summary / Purpose
- Listar planos.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Query: `status` (active|cancelled|finished).

4) Response
- 200 (lista).

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Filtra status.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### GET /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans/{planId}
1) Summary / Purpose
- Detalhe do plano.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Path: `planId`.

4) Response
- 200 (plano + parcelas).

5) Errors
- 404 `CREDITCARD_CARD_NOT_FOUND`

6) Semantics / Notes
- Inclui installments.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### PATCH /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans/{planId}
1) Summary / Purpose
- Atualizar plano.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body parcial.

4) Response
- 200.

5) Errors
- 409 `CREDITCARD_INSTALLMENT_ALREADY_POSTED`

6) Semantics / Notes
- Bloquear se parcelas postadas.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### DELETE /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans/{planId}
1) Summary / Purpose
- Remover plano.

2) Auth & Authorization
- Role: editor+.

3) Request
- n/a.

4) Response
- 204.

5) Errors
- 409 `CREDITCARD_INSTALLMENT_ALREADY_POSTED`

6) Semantics / Notes
- Apenas se nenhuma parcela postada.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### GET /ledgers/{ledgerId}/credit-cards/{cardAccountId}/installments
1) Summary / Purpose
- Listar parcelas.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Query: `month` (YYYY-MM-01), `status`.

4) Response
- 200 (lista).

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Usa status derivado.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### PATCH /ledgers/{ledgerId}/credit-cards/{cardAccountId}/installments/{installmentId}
1) Summary / Purpose
- Atualizar parcela (ex.: skipped).

2) Auth & Authorization
- Role: editor+.

3) Request
- Body:
```json
{ "status": "skipped" }
```

4) Response
- 200.

5) Errors
- 409 `CREDITCARD_INSTALLMENT_ALREADY_POSTED`

6) Semantics / Notes
- Somente enquanto nao postada.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### POST /ledgers/{ledgerId}/credit-cards/{cardAccountId}/post
1) Summary / Purpose
- Postar parcelas do mes.

2) Auth & Authorization
- Role: editor+.

3) Request
- Query: `month` (YYYY-MM-01).
- Headers: `Idempotency-Key` (recomendado).

4) Response
- 200
```json
{ "posted_count": 3, "transaction_id": "uuid" }
```

5) Errors
- 409 `CREDITCARD_INSTALLMENT_ALREADY_POSTED`

6) Semantics / Notes
- Idempotente por mes.

7) Pagination
- n/a.

8) Idempotency
- Obrigatoria por mes.

#### GET /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements
1) Summary / Purpose
- Listar faturas.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Query: `month` (YYYY-MM-01, opcional).

4) Response
- 200 (lista).

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Inclui status.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### GET /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/{statementId}
1) Summary / Purpose
- Detalhe da fatura.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Path: `statementId`.

4) Response
- 200 (statement).

5) Errors
- 404 `CREDITCARD_STATEMENT_NOT_FOUND`

6) Semantics / Notes
- n/a.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### POST /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/close
1) Summary / Purpose
- Fechar fatura do mes.

2) Auth & Authorization
- Role: editor+.

3) Request
- Query: `month` (YYYY-MM-01).

4) Response
- 200 (statement fechado).

5) Errors
- 409 `VALIDATION_ERROR`

6) Semantics / Notes
- Fecha periodo e calcula total.

7) Pagination
- n/a.

8) Idempotency
- Recomendada por mes.

#### POST /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/pay
1) Summary / Purpose
- Pagar fatura.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body:
```json
{
  "statement_id": "uuid",
  "payment_date": "2026-03-10",
  "pay_amount_cents": 300000,
  "cash_account_id": "uuid"
}
```

4) Response
- 200 (statement pago).

5) Errors
- 404 `CREDITCARD_STATEMENT_NOT_FOUND`
- 409 `CREDITCARD_STATEMENT_ALREADY_PAID`
- 422 `CREDITCARD_PAYMENT_EXCEEDS_TOTAL`

6) Semantics / Notes
- Gera transaction de pagamento.

7) Pagination
- n/a.

8) Idempotency
- Recomendada por pagamento.

#### PATCH /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/{statementId}
1) Summary / Purpose
- Ajuste manual (restrito).

2) Auth & Authorization
- Role: owner.

3) Request
- Body parcial.

4) Response
- 200.

5) Errors
- 501 `NOT_IMPLEMENTED`

6) Semantics / Notes
- Uso administrativo.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### Reports (Relatorios)

#### GET /ledgers/{ledgerId}/reports/balances
1) Summary / Purpose
- Saldos por conta no mes.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Query: `month` (YYYY-MM-01).

4) Response
- 200
```json
{
  "ledger_id": "uuid",
  "month": "2026-01-01",
  "items": [ { "account_id": "uuid", "account_name": "Conta", "account_type": "cash", "balance_cents": 0 } ]
}
```

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Regras por tipo de conta.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### GET /ledgers/{ledgerId}/reports/categories
1) Summary / Purpose
- Gasto por categoria.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Query: `from`, `to` (date-time).

4) Response
- 200
```json
{
  "ledger_id": "uuid",
  "from": "...",
  "to": "...",
  "items": [ { "category_id": "uuid", "name": "Mercado", "direction": "out", "total_cents": 0 } ]
}
```

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Sempre filtra ledger.

7) Pagination
- n/a.

8) Idempotency
- n/a.

#### GET /ledgers/{ledgerId}/reports/cashflow
1) Summary / Purpose
- Fluxo de caixa por periodo.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Query: `from`, `to` (date-time).

4) Response
- 200
```json
{
  "ledger_id": "uuid",
  "from": "...",
  "to": "...",
  "items": [ { "month": "2026-01-01", "total_in_cents": 0, "total_out_cents": 0, "net_cents": 0 } ]
}
```

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Baseado em entries do journal.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

## 5) Matriz de autorizacao (resumo)

- Viewer: GET em recursos do ledger.
- Editor: cria/atualiza/deleta recursos do ledger.
- Owner: gerencia membros, reset administrativo, ajustes sensiveis.

---

## 6) Coverage 1:1 (verificacao)

- `docs/api/01_Autenticacao.md` -> Auth
- `docs/api/02_Ledgers.md` -> Ledgers + Members
- `docs/api/03_Accounts.md` -> Accounts
- `docs/api/04_Categories.md` -> Categories
- `docs/api/05_Journal.md` -> Journal
- `docs/api/06_Budget.md` -> Budget
- `docs/api/07_Investimentos.md` -> Investments
- `docs/api/08_Cartao.md` -> Credit Card
- `docs/api/09_Relatorios.md` -> Reports
