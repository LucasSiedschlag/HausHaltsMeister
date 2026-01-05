# Checklist zero brecha (Ledger Security)

## Status atual (resumo rapido)
- Implementado:
  - LedgerId do path como fonte de verdade.
  - Repos sempre filtram `ledger_id`.
  - Role/membership avaliados no service.
  - `/ledgers/{ledgerId}/me` para role.
  - Auth + refresh com rate limit.
  - LedgerGuard middleware (membership + role minima).
  - Validacao de IDs de path (UUID).
  - Idempotency-key para journal (create).
  - Auditoria basica (`audit_log`).
  - Testes cross-ledger + matriz de roles (LedgerGuard).
  - Scan automatizado para `WHERE id =` sem `ledger_id` (rg com allowlist de auth/ledger/budget/access.go).
- Decisao: politica transparente (403 para nao-membro, 404 para nao existe).

## Plano de implementacao (ordem recomendada)
1. Scan de repos sem `ledger_id`
   - Concluido: teste com `rg` em go test (exclui auth/ledger/budget/access.go).

---

## A) Regras absolutas (nao negocie)
1. Fonte de verdade do ledger = ledgerId do PATH
   - Em qualquer POST/PATCH, ignore ledgerId do body, mesmo que exista no DTO.
2. Toda query de recurso deve filtrar por ledger_id
   - `WHERE ledger_id = $1 AND id = $2` (sempre).
3. Nunca autorize por “existe no banco”
   - Autorize por: (user_id, ledger_id) membership + role.
4. Role/perms so do servidor
   - Nunca aceite role no request (exceto endpoints de owner que alteram membership, e ainda assim validando owner).

---

## B) Middleware/Policy obrigatorio (padrao por endpoint)
### 1) Middleware de autenticacao
- Resolve userID da sessao/JWT.
- Injeta no context.

### 2) Middleware de LedgerGuard
Aplica em qualquer rota que tenha `:ledgerId`.

Passos do LedgerGuard:
1. Ler ledgerId do path.
2. Buscar membership:
   - `SELECT role FROM ledger_members WHERE ledger_id=$1 AND user_id=$2 AND removed_at IS NULL`
3. Se nao existir: 403 (ou 404 se voce quiser “nao revelar” que o ledger existe).
4. Verificar role minima exigida pela rota:
   - `viewer < editor < owner`
5. Anexar no context:
   - `ctx.ledgerId`
   - `ctx.ledgerRole`
   - `ctx.permissions` (opcional derivado do role)

Boas praticas:
- Use uma funcao unica: `RequireLedgerRole(minRole)`.
- Evita “esquecer” check em handler.

---

## C) Padrao de Service/Repo (para nao errar)
### 1) Handler so extrai input e chama service
- Nada de SQL no handler.
- Nada de regra de autorizacao no handler (fica no guard/policy).

### 2) Service recebe sempre (ctx, ledgerId, ...)
Mesmo que o ledger ja esteja no ctx, passe explicitamente pra ficar claro.

### 3) Repo: funcoes sempre scoped
Exemplos de assinaturas (ideia):
- `GetCategory(ctx, ledgerId, categoryId)`
- `ListTransactions(ctx, ledgerId, cursor, limit)`
- `UpdateAccount(ctx, ledgerId, accountId, patch)`

Proibido:
- `GetCategoryByID(categoryId)` sem ledgerId.

---

## D) Padroes SQL que evitam vazamento
### 1) Read de recurso
```sql
SELECT ...
FROM categories
WHERE ledger_id = $1
  AND id = $2
  AND deleted_at IS NULL;
```

### 2) Update de recurso (garantindo scoping)
```sql
UPDATE categories
SET name = $3, updated_at = now()
WHERE ledger_id = $1
  AND id = $2
  AND deleted_at IS NULL
RETURNING ...;
```

Se nao retornou linha: responde 404 (ou 403/404 conforme sua politica).

### 3) Delete (soft delete planejado)
```sql
UPDATE categories
SET deleted_at = now()
WHERE ledger_id = $1
  AND id = $2
  AND deleted_at IS NULL;
```

---

## E) Padrao de erros (seguranca + UX)
Escolha 1: “Nao revelar existencia” (mais seguro)
- Se o usuario nao e membro do ledger: 404 em tudo (ledger e recursos).
- Se e membro mas sem role: 403.

Escolha 2: “Transparente”
- Nao membro: 403.
- Sem permissao: 403.
- Nao existe: 404.

Escolha 1 costuma ser melhor pra evitar enumeracao.

---

## F) Protecoes anti-brecha comuns
1. Validacao de IDs
   - ledgerId, accountId, etc. com formato unico (UUID/ULID).
   - Rejeitar strings invalidas com 400.

2. Rate limit e lockout
   - Rate limit por IP + user em rotas sensiveis (login, bulk, reports).
   - Evita brute force e enumeracao.

3. Auditoria (muito recomendado em sistema financeiro)
   - Registrar eventos:
     - ledger.member.add/remove/role_change
     - transaction.create/update/delete
     - budget.plan.update
   - Guardar: user_id, ledger_id, entity_id, action, created_at, ip/user-agent (se tiver).

4. Idempotencia em POST criticos
   - Para transactions e bulk: Idempotency-Key (header) por ledger.
   - Evita duplicar transacoes por retry.

---

## G) Testes que garantem que nao tem furo
1. Testes de autorizacao por role
   - viewer tenta acao editor -> deve falhar
   - editor tenta acao owner -> deve falhar
   - nao-membro -> deve falhar (404 ou 403)

2. Testes de cross-ledger
   - ledger A com transaction X
   - ledger B com usuario atacante
   - Acessar `/ledgers/B/transactions/X` -> deve falhar, mesmo com ID valido.

3. Teste de query sem scope
   - Suite que procura repos com `WHERE id =` sem `ledger_id`.
   - Pega regressao.

---

## H) Como encaixa no seu contrato atual
Mantem o que voce implementou:
- Flat por padrao
- ledgerId nos DTOs (mas ignorado no input)
- `/ledgers/{ledgerId}/me` para role/perms
- `expand=ledger` so se precisar

E a seguranca fica no guard + repo scoped.

---

## I) Matriz de policies por endpoint (roles minimas)
### Autenticacao + preferencias
| Metodo | Path | Role |
|---|---|---|
| POST | /auth/signup | - |
| POST | /auth/login | - |
| POST | /auth/refresh | - |
| POST | /auth/logout | viewer |
| GET | /auth/me | viewer |
| GET | /auth/oauth/{provider}/start | - |
| GET | /auth/oauth/{provider}/callback | - |
| POST | /auth/forgot-password | - |
| POST | /auth/reset-password | - |
| POST | /auth/verify-email | - |
| POST | /auth/resend-verification | - |
| GET | /auth/sessions | viewer |
| DELETE | /auth/sessions/{sessionId} | viewer |
| POST | /auth/logout-all | viewer |
| GET | /auth/providers | viewer |
| POST | /auth/link/{provider}/start | viewer |
| POST | /auth/unlink/{provider} | viewer |
| GET | /me/preferences | viewer |

## J) Frontend ledger context UX
- The frontend uses `useLedgerContext` as the canonical store for the active ledger, role, and errors.
- The header, sidebar, and ledger selector show the active role badge (owner/editor/viewer) and the ledger name.
- Editor-only actions (new transactions, posting cards, ledger writes) must be disabled when `hasRole('editor')` is false and show a tooltip referencing `ledgers.roleRestrictions.transactions`.
- `useApiClient` normalizes errors into `{ code, message, details }`, so `RATE_LIMITED`/`AUTH_RATE_LIMITED` responses can power retry messaging when the UI gets rate-limited.
| PUT | /me/preferences | viewer |

### Ledgers e membros
| Metodo | Path | Role |
|---|---|---|
| GET | /ledgers | viewer |
| POST | /ledgers | viewer |
| GET | /ledgers/{ledgerId} | viewer |
| PATCH | /ledgers/{ledgerId} | editor |
| DELETE | /ledgers/{ledgerId} | owner |
| GET | /ledgers/{ledgerId}/me | viewer |
| GET | /ledgers/{ledgerId}/members | owner |
| POST | /ledgers/{ledgerId}/members | owner |
| PATCH | /ledgers/{ledgerId}/members/{userId} | owner |
| DELETE | /ledgers/{ledgerId}/members/{userId} | owner |

### Accounts
| Metodo | Path | Role |
|---|---|---|
| GET | /ledgers/{ledgerId}/accounts | viewer |
| GET | /ledgers/{ledgerId}/accounts/{accountId} | viewer |
| POST | /ledgers/{ledgerId}/accounts | editor |
| PATCH | /ledgers/{ledgerId}/accounts/{accountId} | editor |
| DELETE | /ledgers/{ledgerId}/accounts/{accountId} | editor |

### Categories
| Metodo | Path | Role |
|---|---|---|
| GET | /ledgers/{ledgerId}/categories | viewer |
| GET | /ledgers/{ledgerId}/categories/{categoryId} | viewer |
| POST | /ledgers/{ledgerId}/categories | editor |
| PATCH | /ledgers/{ledgerId}/categories/{categoryId} | editor |
| DELETE | /ledgers/{ledgerId}/categories/{categoryId} | editor |

### Journal
| Metodo | Path | Role |
|---|---|---|
| POST | /ledgers/{ledgerId}/transactions | editor |
| GET | /ledgers/{ledgerId}/transactions | viewer |
| GET | /ledgers/{ledgerId}/transactions/{transactionId} | viewer |
| PATCH | /ledgers/{ledgerId}/transactions/{transactionId} | editor |
| DELETE | /ledgers/{ledgerId}/transactions/{transactionId} | editor |
| POST | /ledgers/{ledgerId}/transactions:bulk | editor |

### Budget
| Metodo | Path | Role |
|---|---|---|
| GET | /ledgers/{ledgerId}/budget/plan | viewer |
| POST | /ledgers/{ledgerId}/budget/plan | editor |
| PATCH | /ledgers/{ledgerId}/budget/plan | editor |
| DELETE | /ledgers/{ledgerId}/budget/plan | owner |
| POST | /ledgers/{ledgerId}/budget/versions | editor |
| GET | /ledgers/{ledgerId}/budget/versions | viewer |
| GET | /ledgers/{ledgerId}/budget/versions/{versionId} | viewer |
| PATCH | /ledgers/{ledgerId}/budget/versions/{versionId} | owner |
| DELETE | /ledgers/{ledgerId}/budget/versions/{versionId} | owner |
| POST | /ledgers/{ledgerId}/budget/versions/{versionId}/lines | editor |
| PATCH | /ledgers/{ledgerId}/budget/versions/{versionId}/lines/{lineId} | editor |
| DELETE | /ledgers/{ledgerId}/budget/versions/{versionId}/lines/{lineId} | editor |
| GET | /ledgers/{ledgerId}/budget/monthly | viewer |
| GET | /ledgers/{ledgerId}/budget/period | viewer |

### Investimentos
| Metodo | Path | Role |
|---|---|---|
| POST | /ledgers/{ledgerId}/investments/contributions | editor |
| POST | /ledgers/{ledgerId}/investments/redemptions | editor |
| POST | /ledgers/{ledgerId}/investments/earnings | editor |
| GET | /ledgers/{ledgerId}/investments/summary | viewer |

### Cartoes
| Metodo | Path | Role |
|---|---|---|
| GET | /card-networks | viewer |
| POST | /card-networks | owner |
| PATCH | /card-networks/{code} | owner |
| DELETE | /card-networks/{code} | owner |
| GET | /ledgers/{ledgerId}/credit-cards | viewer |
| GET | /ledgers/{ledgerId}/credit-cards/{cardAccountId} | viewer |
| POST | /ledgers/{ledgerId}/credit-cards | editor |
| PATCH | /ledgers/{ledgerId}/credit-cards/{cardAccountId} | editor |
| DELETE | /ledgers/{ledgerId}/credit-cards/{cardAccountId} | editor |
| POST | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans | editor |
| GET | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans | viewer |
| GET | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans/{planId} | viewer |
| PATCH | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans/{planId} | editor |
| DELETE | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans/{planId} | editor |
| GET | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/installments | viewer |
| PATCH | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/installments/{installmentId} | editor |
| POST | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/post | editor |
| GET | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements | viewer |
| GET | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/{statementId} | viewer |
| POST | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/close | editor |
| POST | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/pay | editor |
| PATCH | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/{statementId} | owner |

### Relatorios
| Metodo | Path | Role |
|---|---|---|
| GET | /ledgers/{ledgerId}/reports/balances | viewer |
| GET | /ledgers/{ledgerId}/reports/categories | viewer |
| GET | /ledgers/{ledgerId}/reports/cashflow | viewer |

---

## Tarefas (concluidas)
- LedgerGuard (middleware/policy) resolve membership/role por ledgerId do path e bloqueia por role minima.
- Todas as rotas `/ledgers/:ledgerId/*` usam o guard com role correto (viewer/editor/owner).
- Repos acessam recursos sempre com ledger_id + id (proibido get/update/delete so por id).
- LedgerId do body ignorado em creates/updates (sempre usar o do path).
- Testes de cross-ledger + role matrix para endpoints principais.
- Politica transparente consistente (403 nao-membro, 404 inexistente).
- Matriz de roles por endpoint documentada (tabela de policies).
