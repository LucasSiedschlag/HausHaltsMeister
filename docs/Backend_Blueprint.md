# Backend Blueprint — HausHaltsMeister (Go + Echo + pgx + sqlc + tern)

> Objetivo: construir um backend web (single-user) para controle financeiro pessoal, com núcleo mínimo de movimentações (`cash_flows`) e detalhes opcionais (cartões/parcelas/picuinhas/orçamento), usando boas práticas de mercado: Clean/Hexagonal leve.

---

## 1) Stack e decisões

- Linguagem: Go
- HTTP: Echo
- DB: PostgreSQL
- Driver/Pool: pgx + pgxpool
- SQL: sqlc (queries em `.sql` → código Go)
- Migrations: tern
- Padrão arquitetural:
  - `domain/` (regras + ports)
  - `adapters/postgres` (repos com sqlc)
  - `adapters/http` (handlers Echo)
  - `cmd/api` (bootstrap)

---

## 2) Estrutura de pastas (obrigatória)

```txt
haushaltsmeister/
  cmd/
    api/
      main.go

  internal/
    config/
      config.go

    db/
      db.go

    domain/
      cashflow/
        model.go
        ports.go
        service.go
      budget/
        model.go
        ports.go
        service.go
      picuinhas/
        model.go
        ports.go
        service.go
      cards/
        model.go
        ports.go
        service.go

    adapters/
      postgres/
        cashflow_repo.go
        budget_repo.go
        picuinhas_repo.go
        cards_repo.go
        // sqlc generated:
        sqlc/

      http/
        dto/
          category.go
          cashflow.go
          budget.go
          picuinha.go
          payment.go
        router.go
        cashflow_handlers.go
        budget_handlers.go
        picuinhas_handlers.go
        cards_handlers.go
        middleware.go

  db/
    queries/
      cash_flows.sql
      categories.sql
      budgets.sql
      picuinhas.sql
      cards.sql

  migrations/
    001_init.sql
    002_*.sql

  sqlc.yaml
  go.mod
  go.sum
  Makefile
  README.md
```

⸻

### 3) Convenções e guidelines

3.1 Padrões de código
• Sempre usar context.Context em tudo (service + repo).
• Erros do domínio devem ser “sentinels” (var ErrX = errors.New(...)) e tratados no handler.
• Validar inputs no domínio (service), não no handler (handler só faz parse/bind).
• internal/ deve conter todo código não exportável.
• Não usar ORMs.
• **DTO Pattern (HTTP):**

- Todo handler deve usar structs específicas para Request/Response (`internal/adapters/http/dto`).
- Nomes de campos JSON em `snake_case`.
- Conversão DTO <-> Domain deve ser explícita no handler ou em helpers.

  3.2 Nomes e datas
  • Datas: JSON sempre YYYY-MM-DD.
  • Mês de referência: usar o primeiro dia (ex: 2026-01-01) para queries de “por mês”.

⸻

### 4) Modelo de dados (núcleo + extensões)

4.1 Núcleo mínimo (primeira entrega)
• flow_categories
• cash_flows

4.2 Extensões (incrementais)
• payment_methods
• expense_details
• installment_plans
• budget_periods + budget_items
• picuinha_persons + picuinha_entries

Princípio: entradas (casos 1/2/3) não devem exigir colunas irrelevantes.
Entradas só usam cash_flows (núcleo). Detalhes opcionais ficam em outras tabelas.

⸻

### 5) Migrations (tern)

5.1 Regra
• Sempre criar/alterar schema via migrations/\*.sql.
• Nunca “editar banco na mão” em dev.

5.2 Comandos esperados (Makefile)
• make migrate → aplica migrations com tern
• make migrate-status → status
• make sqlc → sqlc generate

⸻

### 6) sqlc

6.1 sqlc.yaml
• Engine postgres
• sql_package: pgx/v5
• output em internal/adapters/postgres/sqlc

6.2 Regras de queries
• Cada domínio possui arquivo de query próprio em db/queries.
• Queries devem ser pequenas e objetivas.
• Sempre nomear com -- name: QueryName :one|:many|:exec.

⸻

### 7) Domínios e responsabilidades

7.1 Domain: CashFlow

Objetivo: núcleo de entradas/saídas.
• Entidades:
• CashFlow { id, date, categoryId, direction, title, amount }
• Ports:
• Repository.CreateCashFlow
• Repository.ListByMonth
• Regras:
• amount > 0
• direction válido
• category.direction deve bater com cash_flow.direction (validar via repo ou join)

7.2 Domain: Category

Objetivo: classificação de lançamentos.
• Entidades: Category { id, name, direction, isBudgetRelevant, isActive }
• Regras:
• direction imutável
• isBudgetRelevant define se aparece no planejamento orçamentário

7.3 Domain: Budget

Objetivo: orçamento mensal por categoria de SAÍDA.
• Entidades: BudgetPeriod, BudgetItem
• Regras:
• budget_items.category.direction deve ser OUT
• mode = ABSOLUTE | PERCENT_OF_INCOME
• PERCENT_OF_INCOME = percentual sobre a soma das entradas (IN) com isBudgetRelevant = true no mês
• permitir alterar % para um mês ou em lote

7.3 Domain: Picuinhas

Objetivo: razão por pessoa + saldo.
• Entidades: Person, Entry
• Regras:
• saldo por pessoa calculado por soma/subtração conforme kind
• vincular Entry a cash_flow quando mexer no fluxo real
• Entry pode referenciar payment_method_id e card_owner (SELF/THIRD)

7.4 Domain: Cards

Objetivo: parcelamentos e faturas (simples por enquanto).
• Entidades: PaymentMethod (CREDIT_CARD), InstallmentPlan, ExpenseDetails
• Regras:
• compra parcelada gera installment_plan + cash_flows mensais (OUT)
• cada parcela cria expense_detail com payment_method_id e installment_plan_id

⸻

### 8) API (Echo)

8.1 Rotas mínimas (primeira entrega)

CashFlows
• POST /cashflows
• GET /cashflows?month=YYYY-MM-DD

Categories
• POST /categories
• GET /categories?direction=IN|OUT&active=true

8.2 Rotas posteriores (iterativas)

Budget
• POST /budgets/:month/items (upsert)
• GET /budgets/:month/summary
• POST /budgets/batch (alterar em lote para vários meses)

Picuinhas
• POST /picuinhas/persons
• PUT /picuinhas/persons/:id
• DELETE /picuinhas/persons/:id
• GET /picuinhas/persons
• POST /picuinhas/entries
• GET /picuinhas/entries?person_id=:id
• PUT /picuinhas/entries/:id
• DELETE /picuinhas/entries/:id

Cards
• POST /cards/installments (cria parcelamento)
• GET /cards/invoices?month=YYYY-MM-DD (total por cartão)

8.3 Payloads (exemplos)

POST /cashflows

{
"date": "2026-01-01",
"category_id": 1,
"direction": "IN",
"title": "Freela site X",
"amount": 1200.50
}

GET /cashflows?month=2026-01-01
Retorna lista do mês.

⸻

### 9) Bootstrap (cmd/api/main.go)

9.1 Responsabilidades
• Load config
• Run migrations (tern) fora do processo (via Makefile) — ou opcional via wrapper
• Create pgxpool
• Create sqlc queries
• Wire repositories → services → handlers
• Start Echo server

⸻

### 10) Segurança e Observabilidade (RNFs)

> **Ver Guia Completo**: [`docs/Seguranca.md`](./Seguranca.md) para detalhes de implementação passo a passo.

10.1 Autenticação e Hardening
• **Autenticação**: Header `X-App-Token` validado via middleware. Token definido na ENV `APP_TOKEN`.
• **Rate Limit**: Store em memória (x req/s) para evitar loops acidentais.
• **Timeouts**: Middleware de timeout (30s) em todas as rotas.
• **Body Limit**: 1MB max payload.
• **CORS**: Restritivo (apenas domínios confiáveis ou \* em dev).

10.2 Auditoria (Rastreabilidade)
• Tabela: `audit_logs`
• id, entity (table_name), entity_id, action (CREATE/UPDATE/DELETE), diff (JSONB), created_at.
• Service Layer é responsável por despachar log de auditoria em operações críticas.

10.3 Backup e Recuperação
• Script `make backup` que executa `pg_dump` do container.
• Script `make restore` para recuperação de desastres.
• Documentar processo no README.

⸻

### 11) Checklist de implementação (ordem sugerida)

    1.	[x] Criar go.mod, estrutura de pastas
    2.	[x] Criar migrations/001_init.sql (categories + cash_flows)
    3.	[x] make migrate
    4.	[x] Criar sqlc.yaml e db/queries/cash_flows.sql
    5.	[x] make sqlc
    6.	[x] Implementar internal/db (pgxpool)
    7.	[x] Implementar domain/cashflow + repo Postgres + handlers
    8.	[x] Subir API e testar:
    •	POST cashflow
    •	GET cashflows month
    9.	[x] Depois: categories endpoints + seed categories
    10.	[x] Depois: budget, picuinhas, cards (iterativo)
    11. [ ] **Fase 5: Segurança e Operação**
        • [ ] Implementar Middleware Auth (X-App-Token)
        • [ ] Criar tabela audit_logs e AuditService
        • [ ] Configurar Timeouts, BodyLimit e CORS
        • [ ] Scripts de Backup/Restore

⸻

### 12) Critérios de qualidade

    •	Nenhum handler acessa sqlc diretamente: sempre via service + repo.
    •	Domínio não importa Echo nem pgx.
    •	Todas queries sqlc têm testes de integração (posterior).
    •	Migrations versionadas e reprodutíveis.
    •	**Segurança mínima**: Nenhuma rota de escrita exposta sem token.
