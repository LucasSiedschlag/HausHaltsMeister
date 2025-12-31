# Backend Blueprint — HausHaltsMeister (modelo ledger leve)

> Objetivo: orientar a implementação do backend alinhada ao modelo final definido em `docs/ledger/`.
> Fonte de verdade: `docs/ledger/Reestruturação_Completa.md` e documentos específicos de regras.

---

## 1) Stack e decisões

- Linguagem: Go
- Banco: PostgreSQL
- SQL: sqlc (queries em `.sql` -> codigo Go)
- Migrations: tern
- Arquitetura: monolito modular (dominio + adapters), com journal como core

Referencias:
- `docs/ledger/Documento_de_Arquitetura.md`
- `docs/ledger/Regras_Core_Ledger.md`

---

## 2) Estrutura de pastas (sugerida)

```txt
haushaltsmeister/
  cmd/
    api/
      main.go

  internal/
    config/
    db/

    domain/
      auth/
      ledger/
      accounts/
      categories/
      journal/
      budget/
      investments/
      creditcard/
      reporting/

    adapters/
      postgres/
        sqlc/
      http/
        dto/
        router.go
        handlers/

  db/
    queries/
      ledger.sql
      categories.sql
      journal.sql
      budget.sql
      investments.sql
      creditcard.sql
      reporting.sql

  migrations/
    001_init.sql
    002_*.sql

  sqlc.yaml
  Makefile
```

Referencias:
- `docs/ledger/Documento_de_Arquitetura.md`
- `docs/ledger/Requisitos_Funcionais.md`

---

## 3) Convencoes e regras de dominio

Referencias:
- `docs/ledger/Regras_Core_Ledger.md`
- `docs/ledger/Regras_Categorias_e_Orçamento.md`
- `docs/ledger/Regras_Segurança.md`

- `ledger_id` e obrigatorio em todas as rotas e queries.
- `transactions` e `entries` devem ser criados em uma transacao SQL.
- `entries.amount_cents` e sempre positivo; IN/OUT vem de `categories.direction`.
- `entries.kind` define `normal`, `transfer`, `adjust`.
- Transferencias exigem pelo menos uma IN e uma OUT e devem ser balanceadas.
- Mes de referencia: sempre `YYYY-MM-01`.

---

## 4) Modelo de dados (resumo)

Referencias:
- `docs/ledger/Reestruturação_Completa.md`
- `docs/ledger/Documento_de_Arquitetura.md`

### 4.1 Core
- `users`, `ledgers`, `ledger_members`
- `accounts` (`cash`, `investment`, `credit_card`)
- `categories` (direction + flags)
- `transactions` + `entries`

### 4.2 Orçamento (% flexivel)
- `budget_plans`, `budget_plan_versions`, `budget_plan_lines`
- `effective_from_month` sempre no 1o dia do mes

### 4.3 Cartao
- `credit_cards` (metadados)
- `installment_plans`, `installments`
- `credit_card_statements`

### 4.4 Investimentos
- usa `accounts` + `categories` tecnicas e o journal

---

## 5) Fluxos essenciais (alto nivel)

Referencias:
- `docs/ledger/Casos_de_uso.md`
- `docs/ledger/Regras_Cartão_de_crédito.md`
- `docs/ledger/Regras_Investimentos.md`

### 5.1 Criar lancamento
1. Criar `transactions` (data, descricao, autor).
2. Criar `entries` (conta, categoria, valor, kind).
3. Validar ledger boundary e balanceamento (transfer).

### 5.2 Orçamento mensal
1. Selecionar versao vigente do plano no mes.
2. Calcular renda base (IN com `is_budget_base=true`).
3. Calcular limite por categoria (percentual).
4. Calcular realizado (OUT com `is_budget_relevant=true`).

### 5.3 Cartao (parcelas)
1. Compra cria `installment_plans` + `installments` (scheduled).
2. Posting mensal gera `transactions` + `entries` no cartao.
3. Pagamento da fatura e transferencia com categorias tecnicas (fora do orçamento).

---

## 6) Migrations e sqlc

Referencias:
- `docs/ledger/Plano_de_Migracao_Backend.md`
- `docs/ledger/Requisitos_Nao_Funcionais.md`

- Sempre criar/alterar schema via `migrations/*.sql`.
- `make migrate` aplica migrations; `make sqlc` gera codigo.
- Queries devem sempre filtrar por `ledger_id`.

---

## 7) API (diretriz de rotas)

Referencia: `docs/ledger/Documento_de_Arquitetura.md`.

- `POST /ledgers`
- `POST /ledgers/{ledgerId}/transactions`
- `GET /ledgers/{ledgerId}/transactions?from=&to=&account=&category=`
- `POST /ledgers/{ledgerId}/categories`
- `POST /ledgers/{ledgerId}/accounts`
- `GET /ledgers/{ledgerId}/budget/monthly?month=YYYY-MM`
- `POST /ledgers/{ledgerId}/budget/versions`
- `POST /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans`
- `POST /ledgers/{ledgerId}/credit-cards/{cardAccountId}/post?month=YYYY-MM`
- `POST /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/close?month=YYYY-MM`
- `POST /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/pay?month=YYYY-MM`

---

## 8) Checklist de implementacao (ordem sugerida)

1. Core: users, ledgers, accounts, categories, transactions, entries.
2. Budget: plans, versions, lines + calculo mensal.
3. Investimentos: aportes/resgates/rendimentos via journal.
4. Cartao: credit_cards, installments, statements, posting e pagamento.
5. Reporting: orcado vs realizado, saldos por conta, resumos por periodo.
6. Hardening: validacoes, idempotencia e testes.

---

Fim.
