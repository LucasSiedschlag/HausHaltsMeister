# Plano de Implementação do Backend — HausHaltsMeister

Este plano descreve passo a passo a implementação do backend alinhada ao modelo final do ledger leve. Cada etapa referencia o documento de regras correspondente e explicita decisões, validações e entregáveis.

Referências principais:
- `docs/ledger/Reestruturação_Completa.md`
- `docs/ledger/Documento_de_Arquitetura.md`
- `docs/ledger/Regras_Core_Ledger.md`
- `docs/ledger/Regras_Categorias_e_Orçamento.md`
- `docs/ledger/Regras_Investimentos.md`
- `docs/ledger/Regras_Cartão_de_crédito.md`
- `docs/ledger/Regras_Segurança.md`
- `docs/ledger/Casos_de_uso.md`
- `docs/ledger/Requisitos_Funcionais.md`

---

## Etapa 0 — Fundamentos e alinhamento técnico

**Documento base:** `docs/ledger/Documento_de_Arquitetura.md`.

1. Confirmar a estrutura de pastas do backend (domínios + adapters).
2. Validar a configuração de migrations e sqlc.
3. Padronizar a forma de autenticação e sessão (ainda que simples, precisa existir para `created_by_user_id`).
4. Definir o padrão de datas e meses no backend:
   - Mês = `YYYY-MM-01` para orçamento e cartões.
5. Fixar regras de banco:
   - `varchar` + `CHECK` para valores enumerados.
   - `created_at` com default; `updated_at` sem default (setado pelo app).

**Entregáveis:**
- Estrutura mínima do backend criada.
- Migrations aplicáveis via `make migrate`.
- Convenções documentadas no código (README interno ou guia de arquitetura).

---

## Etapa 1 — Autenticação e sessões (MVP+)

**Documento base:** `docs/ledger/Regras_Core_Ledger.md` + `docs/ledger/Regras_Segurança.md`.

1. Implementar modelo de usuário (`users`).
2. Fluxo de criação de usuário:
   - normalizar email (lowercase + trim).
   - validar email único.
   - validar força mínima de senha.
   - persistir `password_hash`.
   - criar ledger padrão após signup.
3. Separar identidade de credencial:
   - `users` para identidade.
   - `auth_secrets` para senha (login local).
   - `auth_identities` para providers (password/google/github).
4. Implementar **access token curto + refresh token**:
   - access token JWT (ex.: 15 min).
   - refresh token opaco (ex.: 30–90 dias), rotacionável.
5. Implementar tabela `auth_sessions` (refresh tokens):
   - `refresh_token_hash`, `expires_at`, `revoked_at`.
   - `user_agent`, `ip`, `device_name` (opcional).
6. Endpoints de sessão:
   - `/auth/login` (gera access + refresh).
   - `/auth/refresh` (rotaciona refresh e emite novo access).
   - `/auth/logout` (revoga sessão atual).
   - `/auth/me` (retorna dados do usuário).
7. Armazenamento seguro (web):
   - refresh em cookie HttpOnly + Secure + SameSite.
   - não usar localStorage para refresh token.
8. Segurança de entrada:
   - respostas de login genéricas.
   - rate limit em login, signup e refresh.
9. Preparação para OAuth (Modelo A):
   - criar `auth_identities` e `oauth_states`.
   - endpoints:
     - `GET /auth/oauth/:provider/start` (gera state + PKCE, redireciona).
     - `GET /auth/oauth/:provider/callback` (valida state, troca code, cria sessão).
   - criar/vincular usuário conforme política de linking.
   - registrar `last_login_at` na identidade.

**Fluxo detalhado (Modelo A)**  
**Start (`/auth/oauth/:provider/start`)**  
1. Gerar `state` e `code_verifier` (PKCE).  
2. Persistir em `oauth_states` com `expires_at` curto e `used_at=null`.  
3. Montar URL do provider com `state` e `code_challenge`.  
4. Redirecionar o usuário para o provider.

**Callback (`/auth/oauth/:provider/callback`)**  
1. Validar `state` (existe, não expirado, `used_at` null).  
2. Trocar `code` por tokens no provider usando `code_verifier`.  
3. Buscar profile + email (e flag de verificação quando existir).  
4. Aplicar política de linking:  
   - se identidade existe, usar o user dela;  
   - senão, se email verificado e existe user, linkar;  
   - senão, criar user + identity;  
   - se email ausente (GitHub), exigir completar cadastro.  
5. Criar `auth_session`, emitir access token e setar cookie de refresh.  
6. Marcar `oauth_states.used_at` e registrar `last_login_at`.

**Validações obrigatórias:**
- `users.is_active=false` bloqueia acesso.

**Entregáveis:**
- CRUD básico de usuário e fluxo de signup/login.
- Tabela de sessões com revogação.
- Middleware que injeta `user_id` no contexto.
- Rate limiting básico nos endpoints de auth.
- Infra pronta para OAuth (tabela + endpoints + linking).

---

## Etapa 2 — Ledger e membership

**Documento base:** `docs/ledger/Regras_Core_Ledger.md` + `docs/ledger/Regras_Segurança.md`.

1. CRUD de `ledgers`.
2. Fluxo `CreateDefaultLedgerForUser`:
   - cria ledger com moeda BRL.
   - cria `ledger_members` com role `owner`.
3. Serviço de acesso:
   - `RequireLedgerRole(user_id, ledger_id, min_role)`.

**Validações obrigatórias:**
- Toda leitura e escrita filtra por `ledger_id`.
- Qualquer recurso é buscado por `(ledger_id, id)`.

**Entregáveis:**
- Serviços de ledger e membership.
- Testes de isolamento por ledger.

---

## Etapa 3 — Accounts (contas internas)

**Documento base:** `docs/ledger/Regras_Core_Ledger.md` + `docs/ledger/Regras_Segurança.md`.

1. CRUD de `accounts`.
2. Respeitar `accounts.type` com CHECK (cash/investment/credit_card).
3. Regras de criação:
   - nome único por ledger.
   - `is_active=false` bloqueia uso em lançamentos.

**Entregáveis:**
- Endpoints de criação/listagem/edição de contas.
- Validação de tipo na criação de cartão.

---

## Etapa 4 — Categories + flags de orçamento

**Documento base:** `docs/ledger/Regras_Categorias_e_Orçamento.md`.

1. CRUD de `categories` com validações:
   - `direction` obrigatório (`in`/`out`).
   - `is_budget_base` permitido apenas para IN.
   - `is_budget_relevant` permitido apenas para OUT.
2. Política de alteração:
   - bloquear mudança de direction se a categoria já foi usada.
3. Hierarquia:
   - permitir `parent_id` e manter consultas com/sem filhos.

**Entregáveis:**
- Endpoints de categoria com validações completas.
- Testes de flags e direction.

---

## Etapa 5 — Journal (transactions + entries)

**Documento base:** `docs/ledger/Regras_Core_Ledger.md` + `docs/ledger/Regras_Segurança.md`.

1. Endpoint de criação de transação:
   - cria `transactions` e N `entries` na mesma transação SQL.
2. Validações por entry:
   - amount_cents > 0.
   - account e category pertencem ao mesmo ledger.
3. Validações por kind:
   - `normal`: livre.
   - `transfer`: exige IN e OUT e soma balanceada.
   - `adjust`: recomenda memo/notes.
4. Consultas:
   - listagem com filtros por período, conta e categoria.
   - detalhe de transação.

**Entregáveis:**
- Fluxos `CreateTransaction`, `ListTransactions`, `GetTransactionDetails`.
- Testes de transferência balanceada.

---

## Etapa 6 — Orçamento mensal (% flexível)

**Documento base:** `docs/ledger/Regras_Categorias_e_Orçamento.md`.

1. CRUD de `budget_plans`, `budget_plan_versions`, `budget_plan_lines`.
2. Regras:
   - 1 plano por ledger.
   - versões válidas sempre no primeiro dia do mês.
3. Cálculo mensal:
   - renda base = soma IN com `is_budget_base=true`.
   - limite = renda base * percent / 100.
   - realizado = soma OUT com `is_budget_relevant=true`.
4. `include_children`:
   - implementar via CTE recursiva no SQL.

**Entregáveis:**
- Endpoint de painel mensal.
- Testes de versionamento e cálculo.

---

## Etapa 7 — Investimentos

**Documento base:** `docs/ledger/Regras_Investimentos.md`.

1. Fluxo de aporte:
   - transferência Pessoal -> Investimentos (2 entries).
2. Fluxo de resgate:
   - transferência Investimentos -> Pessoal.
3. Fluxo de rendimento:
   - entry IN em Investimentos, `kind=adjust`.

**Validações:**
- Transferências sempre balanceadas.
- Entradas técnicas não entram na renda base.

**Entregáveis:**
- Endpoints de aporte/resgate/rendimento.
- Relatórios de aportes e rendimentos por período.

---

## Etapa 8 — Cartão de crédito (parcelas + fatura)

**Documento base:** `docs/ledger/Regras_Cartão_de_crédito.md` + `docs/ledger/Regras_Segurança.md`.

1. CRUD de `credit_cards`.
2. Catálogo `card_networks`:
   - manter seeds e permitir expansão futura.
3. Fluxo de compra parcelada:
   - cria `installment_plans` + `installments` (scheduled).
4. Posting mensal:
   - gera `transactions` + `entries` no cartão.
   - marca `installments.status=posted`.
   - idempotência obrigatória.
5. Fatura:
   - `credit_card_statements` com status open/closed/paid.
6. Pagamento da fatura:
   - transferência Pessoal -> Cartão com categorias técnicas.

**Entregáveis:**
- Endpoints de parcelamento, posting, fechamento e pagamento.
- Testes de idempotência e dupla contagem.

---

## Etapa 9 — Segurança e integridade transversal

**Documento base:** `docs/ledger/Regras_Segurança.md`.

1. Guard rails obrigatórios:
   - RequireActiveUser.
   - RequireLedgerRole.
   - ValidateLedgerOwnershipOfRefs.
2. Garantir que toda query filtra por `ledger_id`.
3. Nunca expor endpoints sem ledger no path.
4. Rate limit e backoff nos endpoints de autenticacao.

**Entregáveis:**
- Middlewares e funções de guard.
- Testes de acesso indevido.

---

## Etapa 10 — Relatórios essenciais

**Documento base:** `docs/ledger/Regras_Core_Ledger.md` + `docs/ledger/Regras_Categorias_e_Orçamento.md`.

1. Saldos por conta:
   - cash/investment: IN - OUT.
   - credit_card: OUT - IN.
2. Resumo por categoria (período).
3. Visão mensal consolidada.

**Entregáveis:**
- Endpoints read-only de relatórios.
- Queries agregadas documentadas.

---

## Etapa 11 — Auditoria e políticas de edição

**Documento base:** `docs/ledger/Regras_Segurança.md`.

1. Política de edição de transaction:
   - permitir editar description/occurred_at/notes.
   - permitir substituir entries (MVP).
2. Política de exclusão:
   - preferir soft delete (se adotado, criar campos).
3. Regras de proteção:
   - não permitir delete de transactions ligadas a parcelas/faturas.

**Entregáveis:**
- Endpoints de update/delete com validações.
- Registro de `updated_at` consistente.

---

## Etapa 12 — Testes e qualidade

**Documento base:** `docs/ledger/Requisitos_Nao_Funcionais.md`.

1. Testes de domínio:
   - transferências balanceadas.
   - cálculo do orçamento mensal.
   - posting idempotente de parcelas.
2. Testes de segurança:
   - isolamento por ledger.
   - viewer não pode escrever.
3. Testes de integração com Postgres.

**Entregáveis:**
- Suite de testes automatizada.
- Cobertura mínima das regras críticas.

---

## Etapa 13 — Observabilidade e operação

**Documento base:** `docs/ledger/Requisitos_Nao_Funcionais.md`.

1. Logs estruturados com `ledger_id` e `user_id`.
2. Métricas básicas:
   - número de transações/dia.
   - tempo de geração do orçamento mensal.
3. Processo de migração controlado:
   - uso de `TERN_CONF` externo em produção.

**Entregáveis:**
- Logging consistente em handlers e services.
- Procedimento de migração documentado.

---

## Encerramento

Este plano serve como roteiro completo para implementação do backend alinhado ao modelo ledger leve. Cada etapa deve ser implementada na ordem proposta, com testes e validações correspondentes, garantindo que a base (journal + ledger boundary) esteja sólida antes das camadas de orçamento, investimentos e cartão.

---

## Apêndice — Ordem de entrega por sprints (sugestão)

**Sprint 1 — Fundamentos e Core**
- Etapas 0 a 3 (base técnica, usuário, ledger, accounts).
- Entregável: criação de ledger e contas funcionais.

**Sprint 2 — Categorias + Journal**
- Etapas 4 e 5 (categorias e transações completas).
- Entregável: criação de lançamentos com split e transferências balanceadas.

**Sprint 3 — Orçamento**
- Etapa 6 (budget versionado + painel mensal).
- Entregável: painel mensal com orçado vs realizado.

**Sprint 4 — Investimentos**
- Etapa 7 (aporte, resgate, rendimento).
- Entregável: fluxos completos e relatórios básicos.

**Sprint 5 — Cartão de crédito**
- Etapa 8 (parcelas, posting e fatura).
- Entregável: ciclo completo do cartão com idempotência.

**Sprint 6 — Segurança + Relatórios + Auditoria**
- Etapas 9, 10 e 11.
- Entregável: segurança consistente, relatórios essenciais e política de edição.

**Sprint 7 — Qualidade e Operação**
- Etapas 12 e 13.
- Entregável: testes críticos, logging e fluxo de operação definido.
