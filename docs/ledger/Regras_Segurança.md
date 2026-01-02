# Segurança, Integridade e Validações — Regras de negócio + Pseudo-fluxo

Este documento define as regras para:

- Isolamento multiusuário por `ledger_id`
- Autorização (roles) e checagens obrigatórias
- Integridade referencial e consistência entre tabelas
- Validações de domínio (categorias, entries, transfers, budget, cartão, investimentos)
- Políticas de edição, exclusão e auditoria
- Estratégias anti-erro (idempotência, concorrência, duplicidade)

> Observação: o foco aqui é **o comportamento** (regras e fluxo).
> A implementação pode começar só no service layer (Go) e evoluir para RLS/constraints/triggers depois.

---

## 0) Princípios fundamentais

1. **Ledger é a fronteira de dados (data boundary).**  
   Nenhuma operação pode ler/escrever dados de outro ledger.

2. **Autorização sempre antecede acesso a dados.**  
   Toda rota/serviço precisa validar membership/role antes de consultar o recurso.

3. **Transações do journal são atômicas.**  
   Criar/editar transaction e entries deve ser feito em transação SQL para não deixar “meio estado”.

4. **Uma única fonte de verdade para semântica.**

   - IN/OUT é definido por `categories.direction`
   - `entries.amount_cents` sempre positivo
   - pagamento/transferência usa `entries.kind=transfer` + categorias técnicas

5. **Preferir idempotência e segurança contra duplicidade.**  
   Em especial para import e geração automática (postar parcelas).

---

## 1) Autenticação e sessão (alto nível)

### 1.1. Identidade do usuário

Toda requisição autenticada deve fornecer:

- `user_id` (derivado do token/sessão)
- (opcional) `session_id` para auditoria

### 1.2. Regras mínimas

- usuário inativo (`users.is_active=false`) não pode operar
- endpoints internos (admin) exigem role/flags específicas (futuro)

Pseudo:

1. `user = load_user(user_id)`
2. if `!user.is_active` => deny

---

### 1.3. Tokens e sessões (MVP+)

- **Access token (JWT)**: curta duração (ex.: 15 min).
- **Refresh token (opaco)**: longa duração (ex.: 30–90 dias), rotacionável.
- Refresh token deve ser revogado no logout.
- Para web: armazenar refresh em cookie HttpOnly + Secure + SameSite.
- Não armazenar refresh token em localStorage.

### 1.4. Tabela de sessões (refresh tokens)

Tabela sugerida `auth_sessions`:

- `id` (uuid)
- `user_id`
- `refresh_token_hash`
- `created_at`, `expires_at`, `revoked_at`
- `user_agent`, `ip`, `device_name` (opcional)
- `rotated_from_session_id` (opcional)

Uso:

- logout de um dispositivo (revogar uma sessão)
- logout global (revogar todas as sessões)

### 1.5. Rate limit e brute force

Aplicar rate limit e backoff em:

- `/auth/login`
- `/auth/signup`
- `/auth/refresh`
- `/auth/forgot-password`

Respostas de login devem ser genéricas:

- “Credenciais inválidas” (não revelar se email existe).

### 1.6. Recuperação de senha (fase 2)

- `POST /auth/forgot-password` gera token temporário.
- `POST /auth/reset-password` valida token e troca senha.
- Token deve expirar rapidamente e ser invalidado uma única vez.

### 1.7. Conteúdo do JWT

- Incluir apenas claims estáveis: `sub`, `exp`, `iat`, `jti` (e `sid` opcional).
- Não incluir roles ou dados mutáveis (sempre validar no banco).

### 1.8. Identidade vs credencial (OAuth pronto)

- `users` representa **identidade** (email, nome, avatar, status).
- `auth_secrets` guarda senha (se houver login local).
- `auth_identities` guarda vínculos por provider (`password`, `google`, `github`).

Regras:

- `auth_identities` deve ser único por `(provider, provider_user_id)`.
- Cada usuário pode ter no máximo 1 identidade por provider (se desejar limitar).

### 1.9. OAuth (PKCE + anti-CSRF)

- Usar Authorization Code + PKCE para SPA.
- `oauth_states` guarda `state`, `code_verifier`, `expires_at` e `used_at`.
- `state` é uso único e expira em poucos minutos (5–10 min).
- Sempre validar `state` e `code_verifier` no callback.

### 1.10. Regras de linking (Google/GitHub)

Política recomendada:

1. Se existe `auth_identities(provider, provider_user_id)` → login nesse `user_id`.
2. Senão, se provider retornou email:
   - se existe `users.email = email` **e** email verificado (ou confiança explícita no provider):
     - vincular identidade ao usuário existente.
   - senão: criar novo usuário e identidade.
3. Se provider **não** retorna email (caso possível no GitHub):
   - exigir etapa de completar cadastro com email válido **ou**
   - rejeitar login e orientar usuário a liberar email no provider.

Nunca criar dois usuários diferentes para a mesma conta externa.

## 2) Autorização por Ledger (RBAC)

### 2.1. Regra de acesso ao ledger

Usuário tem acesso ao ledger se:

- `ledgers.owner_user_id == user_id` OR
- existe `ledger_members` com (ledger_id, user_id)

Pseudo:

- `access = is_owner(user_id, ledger_id) OR is_member(user_id, ledger_id)`

### 2.2. Regras por role

- viewer:
  - read-only (listar ledger, contas, categorias, transactions, relatórios)
- editor:
  - tudo do viewer
  - criar/editar/deletar: transactions/entries, plans/cartão (dependendo do produto)
- owner:
  - tudo do editor
  - gerenciar membros e configurações sensíveis

### 2.3. Pseudo-funções padrão

- `RequireLedgerRole(user_id, ledger_id, min_role)`
  - se não atende => 403

### 2.4. Regra de ouro (para evitar vazamento)

**Nunca** aceitar endpoints do tipo:

- `/transactions/:id` sem ledger
  Preferir:
- `/ledgers/:ledger_id/transactions/:id`

E sempre validar:

- transaction.ledger_id == ledger_id

---

## 3) Isolamento de dados (queries seguras)

### 3.1. Qualquer SELECT deve filtrar por ledger_id

Exemplos:

- categories: `WHERE ledger_id = $ledger`
- accounts: `WHERE ledger_id = $ledger`
- transactions: `WHERE ledger_id = $ledger`
- entries: `WHERE ledger_id = $ledger`

### 3.2. Regra extra para recursos por ID

Para qualquer recurso R com id:

- sempre buscar por `(ledger_id, id)` ou validar que o recurso pertence ao ledger.

Pseudo:

1. `r = get_by_id(id)`
2. if `r.ledger_id != ledger_id` => deny (404 preferível)

---

## 4) Integridade e consistência cross-table

> Essas regras evitam inconsistências como:
>
> - entry apontando para category de outro ledger
> - entry apontando para account de outro ledger
> - entry apontando para transaction de outro ledger

### 4.1. Regras obrigatórias (service layer)

Para cada `entry`:

- `entry.ledger_id == transaction.ledger_id`
- `account.ledger_id == entry.ledger_id`
- se `category_id` não nulo:
  - `category.ledger_id == entry.ledger_id`

Pseudo em CreateTransaction:
for each entry:

1. `account = load_account(entry.account_id)`
2. assert `account.ledger_id == ledger_id`
3. if `entry.category_id != null`:
   - `cat = load_category(entry.category_id)`
   - assert `cat.ledger_id == ledger_id`

### 4.2. Estratégia opcional (DB constraints/triggers)

Para fortalecer sem depender do app:

- triggers para checar ledger_id
- ou modelar chaves compostas (ledger_id, id) (mais complexo)
- ou RLS (quando virar SaaS)

Recomendação:

- começar no service layer
- migrar para RLS/constraints se publicar multi-tenant

---

## 5) Validações de domínio: Journal (transactions/entries)

### 5.1. Transaction

- `description` obrigatório, tamanho máximo recomendado (ex. 120)
- `occurred_at` obrigatório
- `created_by_user_id` deve ser membro/editor do ledger
- `ledger_id` obrigatório

### 5.2. Entry

- `amount_cents > 0`
- `account_id` obrigatório
- `kind` obrigatório (default normal)
- `category_id` pode ser null (mas recomendado sempre existir para relatórios)
- `memo` opcional, mas recomendado em split/adjust

### 5.3. Regras específicas por kind

#### kind = normal

- não exige nada extra
- pode ter 1..N entries por transaction

#### kind = transfer

Regras recomendadas:

- dentro da mesma transaction:
  - deve existir pelo menos 1 entry com category.direction=in
  - deve existir pelo menos 1 entry com category.direction=out
- `SUM(OUT) == SUM(IN)` (balanceamento)
- todas as entries do transfer devem ter category_id não nulo
- valores não podem ser misturados com normal (ou permitir, mas fica confuso)

Pseudo:

1. `outs = sum(entry.amount where cat.direction=out)`
2. `ins = sum(entry.amount where cat.direction=in)`
3. assert outs == ins
4. assert outs > 0 && ins > 0

#### kind = adjust

- recomendado exigir `memo` na entry ou `notes` na transaction
- não precisa ser balanceado

---

## 6) Validações de domínio: Categories

### 6.1. Direção

- `categories.direction` é a única verdade para IN/OUT.
- Não permitir direction null.

### 6.2. Flags e consistência

Regras recomendadas:

- se direction=in:
  - forçar `is_budget_relevant=false`
- se direction=out:
  - forçar `is_budget_base=false`

### 6.3. Unicidade

- `unique(ledger_id, name)` (já no schema)
- impedir nomes duplicados (case-insensitive se desejar)

### 6.4. Alterar direction (política)

Recomendação:

- bloquear mudança de direction em categorias já usadas por entries
- alternativa: criar nova categoria e desativar a antiga

Pseudo:

1. if exists entries where category_id = X:
   - deny direction change

---

## 7) Validações de domínio: Budget (plans/versions/lines)

### 7.1. Plan

- 1 plano por ledger (`unique ledger_id`)
- somente owner/editor pode alterar

### 7.2. Version

- `effective_from_month` deve ser 1º dia do mês
- `unique(plan_id, effective_from_month)`

Validação:

- normalize `effective_from_month` para o 1º dia do mês no service layer

### 7.3. Lines

- `percent > 0`
- categoria deve ser do ledger
- por padrão, categoria deve ser OUT (quando apply_to=out_only)
- `unique(version_id, category_id)`

### 7.4. Somatório de percentuais

Política:

- permitir <= 100 (recomendado)
- permitir > 100 com warning (UI)
- nunca bloquear no DB (flexibilidade)

---

## 8) Validações de domínio: Accounts + Credit Cards

### 8.1. Accounts

- `unique(ledger_id, name)`
- `type` obrigatório
- `is_active` para soft-disable sem apagar histórico

### 8.2. Credit card metadata (credit_cards)

- apenas 1:1 com account tipo credit_card (account_id é PK)
- `closing_day` e `due_day` devem estar em intervalos válidos (1..31)
  Recomendação de UX:
- limitar para 1..28 para evitar meses curtos

Validação de integridade:

- ao criar `credit_cards`, verificar:
  - `accounts.type == credit_card`

---

## 9) Validações de domínio: Parcelas e Fatura (cartão)

### 9.1. Installment plan

- `installments_count >= 1`
- `installment_amount_cents > 0`
- `total_amount_cents > 0`
- categoria do plano deve ser `direction=out` (consumo)
- `card_account_id` deve ser `type=credit_card`
- `first_due_month` deve ser 1º dia do mês

Se o total não bater exatamente:

- permitir última parcela diferente (futuro) ou exigir bate (MVP)

### 9.2. Installments

- `unique(plan_id, installment_no)`
- `installment_no` em 1..N
- `due_month` sempre 1º dia do mês
- status transitions válidos:
  - scheduled -> posted -> paid
  - scheduled -> skipped
  - posted -> skipped (via ajuste/estorno, com cuidado)

### 9.3. Posting (gerar lançamentos do mês)

Regras de idempotência:

- nunca postar a mesma parcela duas vezes
- condição:
  - só postar se status == scheduled
- após postar:
  - set status=posted
  - set posted_transaction_id

Pseudo:

1. select installments where due_month=M and status=scheduled
2. for each:
   - begin tx
   - create transaction + entry (cartão, categoria real, amount)
   - update installment status=posted, posted_transaction_id=...
   - commit

### 9.4. Statements (faturas)

- `unique(card_account_id, statement_month)`
- statement_month sempre 1º dia do mês

Fechamento:

- open -> closed
  Pagamento:
- closed -> paid (se pagamento total)

### 9.5. Pagamento de fatura

Validações:

- pagamento não pode exceder total (se bloquear) ou permitir parcial (mais realista)
- criar transaction transfer com categorias técnicas
- não afetar orçamento (flags garantem)

---

## 10) Políticas de edição e exclusão (auditabilidade)

### 10.1. Edição de transaction/entries (MVP)

Recomendação:

- permitir editar:
  - description, notes, occurred_at
  - entries (substituir lista) desde que preserve integridade ledger/account/category
- registrar `updated_at`

### 10.2. Alternativa futura: append-only

- correções feitas com `kind=adjust`
- proibir editar entries antigas
- melhora auditabilidade

### 10.3. Exclusão

Recomendação:

- soft delete (campos `deleted_at` ou `is_deleted`)
- esconder em relatórios e UI
- manter histórico

Se hard delete no MVP:

- nunca permitir deletar transaction vinculada a:
  - installment.posted_transaction_id
  - statement.payment_transaction_id
    pois quebra a integridade do cartão
    (pelo menos bloquear nesses casos)

---

## 11) Idempotência, duplicidade e concorrência

### 11.1. Imports

- usar `transactions.external_source/external_id`
- antes de inserir, buscar existente
- se existir, retornar “já importado”

### 11.2. Rotinas automáticas (post installments / close statement)

- proteger com locks lógicos (por card + month) se necessário
- idempotência por status:
  - só posta installments scheduled
  - só fecha statement open
  - só marca paid se ainda não paid

### 11.3. Concorrência (double-submit)

Para endpoints que criam transaction:

- gerar `client_request_id` (futuro) ou usar `external_id` interno
- ou bloquear duplicidade por (ledger_id, occurred_at, description) com heurística (frágil)
  Recomendação MVP:
- confiar em UI + logs, e evoluir conforme necessidade

---

## 12) Logging e auditoria (recomendado)

### 12.1. Campos já existentes

- `created_by_user_id`
- `created_at`, `updated_at`

### 12.2. Logs de domínio (opcional)

Criar tabela `audit_events` no futuro:

- user_id, ledger_id, action, entity, entity_id, payload, created_at

---

## 13) Pseudo-fluxos “guard rails” (checagens obrigatórias)

### 13.1. Guard: `RequireActiveUser`

Input: user_id

- load user
- if !is_active => deny

### 13.2. Guard: `RequireLedgerAccess`

Input: user_id, ledger_id, min_role

- check owner or member
- check role >= min_role
- deny if not

### 13.3. Guard: `ValidateLedgerOwnershipOfRefs`

Input: ledger_id, account_ids[], category_ids[]

- for each account: assert account.ledger_id == ledger_id
- for each category: assert category.ledger_id == ledger_id

### 13.4. Guard: `ValidateTransferBalance`

Input: entries[]

- outs = sum where category.direction=out
- ins = sum where category.direction=in
- assert outs == ins && outs > 0

### 13.5. Guard: `ValidateBudgetVersionMonth`

Input: effective_from_month

- normalize to first day of month
- assert uniqueness per plan

---

## 14) Resultado final (o que este conjunto garante)

Isolamento por ledger em todas as operações.  
RBAC consistente (viewer/editor/owner).  
Integridade entre transaction/entry/account/category.  
Transferências balanceadas e sem duplicidade.  
Budget versionado coerente e com validações previsíveis.  
Cartão (parcelas/fatura) idempotente e sem dupla contagem no orçamento.  
Base pronta para evoluir para constraints/RLS sem refatorar o domínio.
