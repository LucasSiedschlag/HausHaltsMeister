# Backend Blueprint — Ledger (HausHaltsMeister)

> Objetivo: backend baseado em ledger (partidas dobradas) com camada hibrida para cartoes/parcelamentos, mantendo o app simples, consistente e auditavel.

---

## 1) Stack e decisoes

- Linguagem: Go
- HTTP: Echo
- DB: PostgreSQL
- Driver/Pool: pgx + pgxpool
- SQL: sqlc (queries em `.sql` -> codigo Go)
- Migrations: tern
- Arquitetura: domain + adapters (postgres/http), clean/hex leve

---

## 2) Estrutura de pastas (obrigatoria)

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
      ledger/
        model.go
        ports.go
        service.go
      category/
        model.go
        ports.go
        service.go
      budget/
        model.go
        ports.go
        service.go
      picuinha/
        model.go
        ports.go
        service.go
      cards/
        model.go
        ports.go
        service.go

    adapters/
      postgres/
        ledger_repo.go
        category_repo.go
        budget_repo.go
        picuinha_repo.go
        cards_repo.go
        // sqlc generated:
        sqlc/

      http/
        dto/
          ledger.go
          category.go
          budget.go
          picuinha.go
          payment.go
        router.go
        ledger_handlers.go
        category_handlers.go
        budget_handlers.go
        picuinha_handlers.go
        cards_handlers.go
        middleware.go

  db/
    queries/
      ledger.sql
      categories.sql
      budgets.sql
      picuinhas.sql
      cards.sql

  migrations/
    001_init.sql
    002_*.sql

  sqlc.yaml
  Makefile
  README.md
```

---

## 3) Convencoes e guidelines

3.1 Padrões de codigo
- Sempre usar context.Context em tudo (service + repo).
- Erros de dominio devem ser "sentinels" (var ErrX = errors.New(...)) e tratados no handler.
- Validacao no service, nao no handler (handler so faz parse/bind).
- Nao usar ORM.
- Nao acessar sqlc diretamente em handlers.

3.2 DTO Pattern (HTTP)
- Todo handler deve usar structs especificas para Request/Response (`internal/adapters/http/dto`).
- JSON em snake_case.
- Conversao DTO <-> Domain explicita no handler (ou helpers).

3.3 Datas
- JSON sempre YYYY-MM-DD.
- "Mes de referencia": sempre o primeiro dia (YYYY-MM-01).

---

## 4) Modelo de dados (ledger + hibrido)

4.1 Core ledger (migrations/001_init.sql)
- ledger_accounts
  - type: ASSET | LIABILITY | EQUITY | INCOME | EXPENSE
  - currency: "BRL" por padrao
- ledger_transactions
  - occurred_at, description, reference, notes
- ledger_postings
  - transaction_id, account_id, side (DEBIT/CREDIT), amount
  - category_id (opcional)
  - party_id (opcional)
- ledger_categories
  - direction (IN/OUT), is_budget_relevant, is_active
- ledger_parties
  - kind: PERSON | ORG | SYSTEM
- ledger_tags + ledger_posting_tags (opcional)

4.2 Camada hibrida (migrations/002 e 003)
- payment_methods
  - 1:1 com ledger_accounts (account_id)
  - closing_day, due_day, credit_limit
- installment_plans
  - plan_type: INSTALLMENT | RECURRING
  - total_amount ou installment_amount + installment_count
  - start_date, payment_method_id ou account_id
  - category_id e party_id (vinculo de orcamento e picuinhas)
- installment_plan_items
  - due_date, amount, status
  - transaction_id (quando virar posting no ledger)

4.3 Principios
- Toda movimentacao contabilizavel passa por postings.
- As tabelas hibridas so geram postings ou ajudam no UX (cartao, fatura, parcelamento).
- O ledger e a fonte da verdade para saldo e analises.

---

## 5) Regras de dominio (detalhadas)

5.1 Partidas dobradas (obrigatorio)
- Para cada transacao: soma dos DEBITs == soma dos CREDITs.
- Valor sempre > 0, o "sinal" e dado pelo side.

5.2 Categorias e orcamento
- category_id deve existir apenas em postings de INCOME/EXPENSE.
- Orçamento usa ledger_categories com direction = OUT.
- Percentual calcula sobre IN (direction = IN, is_budget_relevant = true).

5.3 Contas (exemplos base)
- Assets: Banco, Carteira
- Liabilities: Cartao Nubank, Cartao Inter
- Income: Ganhos, Investimentos
- Expense: Moradia, Alimentacao, Transporte, Picuinhas
- Equity: Ajustes (para correcoes sem caixa real)

5.4 Picuinhas
- Pessoa: ledger_parties (kind = PERSON).
- Conta padrao: Receivables:Picuinhas (ASSET).
- Emprestimo:
  - DEBIT Receivables:Picuinhas (party_id)
  - CREDIT Banco (se houve saida real)
- Recebimento:
  - DEBIT Banco
  - CREDIT Receivables:Picuinhas (party_id)
- Compra no cartao para pessoa:
  - DEBIT Receivables:Picuinhas (party_id)
  - CREDIT Cartao (LIABILITY)

5.5 Cartoes e parcelamentos
- Cartao = account LIABILITY + payment_methods.
- Parcela gera posting:
  - DEBIT Expense (categoria)
  - CREDIT Cartao (LIABILITY)
- Fatura = soma dos itens do mes (com fechamento/vencimento).

---

## 6) Migrations (tern)

6.1 Regras
- Sempre criar/alterar schema via migrations/*.sql.
- Nunca editar banco na mao em dev.

6.2 Comandos esperados (Makefile)
- make migrate -> aplica migrations com tern
- make migrate-status -> status
- make sqlc -> sqlc generate

---

## 7) sqlc

7.1 sqlc.yaml
- Engine: postgresql
- sql_package: pgx/v5
- Output: internal/adapters/postgres/sqlc

7.2 Regras de queries
- Cada dominio possui arquivo de query proprio em db/queries.
- Queries pequenas e objetivas.
- Sempre nomear com -- name: QueryName :one|:many|:exec.

---

## 8) Dominios e responsabilidades

8.1 Ledger
- Criar transacoes com postings balanceados.
- Consultas por mes, conta, categoria e party.
- Validacao de integridade (soma, tipos de conta).

8.2 Category
- CRUD de ledger_categories.
- Bloquear alteracao de direction quando houver uso historico.

8.3 Budget
- Periodos e itens por mes.
- Validacao de 100% quando em percentual.
- Replicacao de meses futuros quando necessario.

8.4 Picuinhas
- Pessoas (ledger_parties) + casos.
- Saldo em aberto via postings da conta Receivables.

8.5 Cards
- payment_methods e installment_plans.
- Fatura por mes (somar items + considerar fechamento).

---

## 9) API (diretrizes)

9.1 Regras gerais
- DTOs em internal/adapters/http/dto
- JSON snake_case
- Datas em YYYY-MM-DD
- Validacao no service, handler apenas parse

9.2 Rotas minimas (primeira entrega)

Ledger
- POST /ledger/transactions
- GET /ledger/transactions?month=YYYY-MM-01
- GET /ledger/accounts

Categories
- POST /categories
- GET /categories?direction=IN|OUT&active=true

9.3 Rotas incrementais

Budget
- POST /budgets/:month/items
- GET /budgets/:month/summary
- POST /budgets/batch

Picuinhas
- POST /picuinhas/persons
- GET /picuinhas/persons
- POST /picuinhas/cases
- GET /picuinhas/cases?person_id=:id
- GET /picuinhas/cases/:id/installments

Cards
- POST /cards/installments
- GET /cards/invoices?month=YYYY-MM-01

9.4 Payload exemplo (ledger)

POST /ledger/transactions

{
  "occurred_at": "2026-01-10",
  "description": "Salario",
  "postings": [
    {"account_id": 1, "side": "DEBIT", "amount": 5000.00},
    {"account_id": 10, "side": "CREDIT", "amount": 5000.00, "category_id": 2}
  ]
}

---

## 10) Bootstrap (cmd/api/main.go)

Responsabilidades:
- Load config
- Criar pgxpool
- Criar sqlc queries
- Wire repositories -> services -> handlers
- Start Echo server

---

## 11) Seguranca e observabilidade (RNFs)

11.1 Autenticacao e hardening
- Header `X-App-Token` via middleware
- Timeout global (ex: 30s)
- Body limit (ex: 1MB)
- CORS restritivo

11.2 Auditoria
- Tabela audit_logs (acao, entidade, diff, created_at)
- Service layer dispara logs

11.3 Backup e restore
- `make backup` e `make restore`
- Documentar no README

---

## 12) Checklist de implementacao (ordem sugerida)

1. [x] migrations/001_init.sql (ledger core)
2. [x] make migrate
3. [x] sqlc.yaml + queries base
4. [x] make sqlc
5. [ ] Implementar domain/ledger + repo + handlers
6. [ ] Integrar categories
7. [ ] Integrar budget
8. [ ] Integrar picuinhas
9. [ ] Integrar cards/parcelamentos
10. [ ] Seguranca e operacao (auth, timeout, backup)
