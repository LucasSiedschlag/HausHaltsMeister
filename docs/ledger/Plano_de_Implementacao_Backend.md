# Plano de Implementacao do Backend — HausHaltsMeister

Este plano descreve passo a passo a implementacao do backend alinhada ao modelo final do ledger leve. Cada etapa referencia o documento de regras correspondente e explicita decisoes, validacoes e entregaveis. O objetivo e reduzir ambiguidade e evitar refatoracoes futuras.

Referencias principais:
- `docs/ledger/Reestruturacao_Completa.md`
- `docs/ledger/Documento_de_Arquitetura.md`
- `docs/ledger/Regras_Core_Ledger.md`
- `docs/ledger/Regras_Categorias_e_Orcamento.md`
- `docs/ledger/Regras_Investimentos.md`
- `docs/ledger/Regras_Cartao_de_credito.md`
- `docs/ledger/Regras_Seguranca.md`
- `docs/ledger/Casos_de_uso.md`
- `docs/ledger/Requisitos_Funcionais.md`
- `docs/agent/DECISIONS.md`
- `docs/agent/CONVENTIONS.md`
- `docs/agent/ERRORS.md`
- `docs/agent/TEST_STRATEGY.md`
- `docs/agent/IMPLEMENTATION_PLAYBOOK.md`

---

## Etapa 0 — Fundamentos e alinhamento tecnico

**Documento base:** `docs/ledger/Documento_de_Arquitetura.md`.

1. Confirmar a estrutura de pastas do backend (dominios + adapters) e naming.
2. Validar migrations e sqlc:
   - migrations rodando via `make migrate`.
   - sqlc apontando para as queries corretas.
3. Fixar convencoes gerais:
   - datas e meses (YYYY-MM-01).
   - erros padronizados (`docs/agent/ERRORS.md`).
   - `snake_case` em JSON e DB.
4. Fixar regras de banco:
   - `varchar` + `CHECK` para valores enumerados.
   - `created_at` default; `updated_at` setado pela app.
5. Confirmar seeds tecnicas:
   - categorias tecnicas (cartao e investimentos).
   - catalogo de bandeiras (card_networks).

**Entregaveis:**
- Estrutura minima do backend criada.
- Migrations aplicaveis via `make migrate`.
- Documentos base validos.

---

## Etapa 1 — Autenticacao e sessoes (MVP+)

**Documento base:** `docs/ledger/Regras_Seguranca.md`.

1. Separar identidade de credencial:
   - `users` para identidade.
   - `auth_secrets` para senha (login local).
   - `auth_identities` para providers (`password`, `google`, `github`).
2. Fluxo de signup:
   - normalizar email.
   - validar senha.
   - criar `users`, `auth_secrets`, `auth_identities` provider=password.
   - criar ledger padrao.
3. Tokens:
   - access token curto (JWT).
   - refresh token opaco, rotacionavel.
4. Sessoes:
   - `auth_sessions` com revogacao e expiracao.
   - refresh token armazenado em cookie HttpOnly.
5. Endpoints base:
   - `/auth/login`, `/auth/refresh`, `/auth/logout`, `/auth/me`.
6. Rate limit e respostas genericas para login.
7. OAuth (Modelo A) preparado:
   - `oauth_states` com PKCE.
   - `/auth/oauth/:provider/start` e `/auth/oauth/:provider/callback`.
   - politica de linking:
     - identity existe -> usar user.
     - email verificado -> linkar.
     - email ausente (GitHub) -> exigir completar cadastro.

**Entregaveis:**
- Sessao completa com refresh + rotacao.
- Tabela de sessoes operante.
- OAuth pronto sem refatoracao.

---

## Etapa 2 — Ledger e membership

**Documento base:** `docs/ledger/Regras_Core_Ledger.md` + `docs/ledger/Regras_Seguranca.md`.

1. CRUD de `ledgers`.
2. `CreateDefaultLedgerForUser`:
   - cria ledger com BRL.
   - cria ledger_members owner.
3. Guard rails:
   - `RequireLedgerRole`.
   - `RequireActiveUser`.

**Entregaveis:**
- Endpoints de ledger e members.
- Validacao de RBAC em todas as rotas.

---

## Etapa 3 — Accounts

**Documento base:** `docs/ledger/Regras_Core_Ledger.md`.

1. CRUD de `accounts`.
2. Regras:
   - nome unico por ledger.
   - `type` validado via CHECK.
   - `nature` (`asset`/`liability`) com default `asset`.
   - `is_active=false` bloqueia uso.
3. Preparar `credit_cards` como entidade 1:N por conta (com passivo dedicado).

**Entregaveis:**
- Endpoints de contas.
- Validacao de tipo.

---

## Etapa 4 — Categories + flags de orcamento

**Documento base:** `docs/ledger/Regras_Categorias_e_Orcamento.md`.

1. CRUD de `categories`.
2. Validacoes:
   - direction obrigatorio.
   - IN => is_budget_relevant=false.
   - OUT => is_budget_base=false.
3. Bloquear troca de direction se ja usada.
4. Hierarquia via parent_id.

**Entregaveis:**
- Endpoints de categorias com validacao.
- Testes de flags.

---

## Etapa 5 — Journal (transactions + entries)

**Documento base:** `docs/ledger/Regras_Core_Ledger.md` + `docs/ledger/DECISIONS.md`.

1. CreateTransaction:
   - criar transaction e entries na mesma transacao SQL.
   - validar ledger/account/category.
   - validar balanceamento em transferencias.
2. Listagem paginada:
   - ordenar por `occurred_at DESC, id DESC`.
   - paginacao por cursor (cursor_occurred_at + cursor_id).
   - query em 2 etapas (CTE de IDs + join entries).
3. GetTransactionDetails:
   - buscar por (ledger_id, transaction_id) e retornar entries.
4. Update/Delete:
   - bloquear se referenciado em installments/statement.
   - preferir ajuste (`kind=adjust`).
5. Endpoint bulk (opcional):
   - operacoes de import/posting em lote.

**Entregaveis:**
- Endpoints do journal com paginacao segura.
- Testes contra “fantasmas”.

---

## Etapa 6 — Orcamento mensal (% flexivel)

**Documento base:** `docs/ledger/Regras_Categorias_e_Orcamento.md`.

1. CRUD de `budget_plans`, `budget_plan_versions`, `budget_plan_lines`.
2. Versao vigente por mes.
3. Calculo mensal:
   - income_base (IN com is_budget_base=true).
   - limite = percent / 100.
   - realizado = OUT com is_budget_relevant=true.
4. include_children via CTE recursiva.

**Entregaveis:**
- Endpoint de painel mensal.
- Testes de calculo e versionamento.

---

## Etapa 7 — Investimentos

**Documento base:** `docs/ledger/Regras_Investimentos.md`.

1. Fluxos:
   - aporte (transfer).
   - resgate (transfer).
   - rendimento (adjust).
2. Garantir que entradas tecnicas nao aumentam renda base.

**Entregaveis:**
- Endpoints de aporte/resgate/rendimento.
- Relatorios basicos.

---

## Etapa 8 — Cartao de credito

**Documento base:** `docs/ledger/Regras_Cartao_de_credito.md`.

1. CRUD de `credit_cards` (1:N por conta, com passivo dedicado).
2. Catalogo `card_networks`.
3. Fluxos:
   - criar installment_plans + installments.
   - posting mensal (idempotente).
   - statements open/closed/paid.
   - pagamento da fatura via transferencias tecnicas.

**Entregaveis:**
- Endpoints completos de cartao.
- Testes de idempotencia e dupla contagem.

---

## Etapa 9 — Seguranca transversal

**Documento base:** `docs/ledger/Regras_Seguranca.md`.

1. Guard rails globais (ledger boundary).
2. Rate limit nos endpoints de auth.
3. Bloqueio de delete de transactions referenciadas.

**Entregaveis:**
- Middlewares de seguranca e testes.

---

## Etapa 10 — Relatorios essenciais

**Documento base:** `docs/ledger/Regras_Core_Ledger.md`.

1. Saldos por account.
2. Resumo por categoria.
3. Resumo mensal consolidado.

**Entregaveis:**
- Endpoints read-only.
- Queries agregadas.

---

## Etapa 11 — Auditoria e politicas de edicao

**Documento base:** `docs/ledger/Regras_Seguranca.md`.

1. Politica MVP: editar transaction e substituir entries.
2. Bloquear edit/delete de transacoes referenciadas.
3. Preparar caminho para append-only (futuro).

**Entregaveis:**
- Endpoints update/delete com validacoes.

---

## Etapa 12 — Testes e qualidade

**Documento base:** `docs/agent/TEST_STRATEGY.md`.

1. Unit tests para regras criticas.
2. Integration tests para queries e constraints.
3. Coverage obrigatoria para:
   - transfer balance.
   - budget monthly calc.
   - posting idempotente.
   - refresh rotation.
   - ledger access denial.

**Entregaveis:**
- Suite de testes minima.

---

## Etapa 13 — Observabilidade e operacao

**Documento base:** `docs/ledger/Requisitos_Nao_Funcionais.md`.

1. Logging estruturado (ledger_id, user_id).
2. Metricas basicas.
3. Processo de migracao controlado (TERN_CONF externo).

**Entregaveis:**
- Logs e procedimento operacional.

---

## Apice — Ordem de entrega por sprints (sugestao)

**Sprint 1**: Etapas 0–3 (base tecnica, auth, ledger, accounts)

**Sprint 2**: Etapas 4–5 (categorias + journal)

**Sprint 3**: Etapa 6 (budget)

**Sprint 4**: Etapa 7 (investimentos)

**Sprint 5**: Etapa 8 (cartao)

**Sprint 6**: Etapas 9–11 (seguranca, relatorios, auditoria)

**Sprint 7**: Etapas 12–13 (testes, observabilidade)
