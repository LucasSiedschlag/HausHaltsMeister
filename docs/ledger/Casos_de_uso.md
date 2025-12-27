# Casos de uso no modelo ledger (veredito)

Este documento mapeia cada caso de uso do `docs/Casos_de_uso.md` para o modelo ledger
proposto (core + camada hibrida de cartoes/parcelamentos).

## Premissas do modelo

- Core ledger:
  - `ledger_accounts` (ASSET, LIABILITY, EQUITY, INCOME, EXPENSE).
  - `ledger_transactions` (cabecalho).
  - `ledger_postings` (partidas; devem balancear).
- Analise e orcamento:
  - `ledger_categories` com `direction` e `is_budget_relevant`.
  - `category_id` fica no posting ligado a INCOME/EXPENSE.
- Picuinhas:
  - `ledger_parties` (kind = PERSON).
  - Postings podem ter `party_id`.
  - Conta padrao: `Receivables:Picuinhas` (ASSET).
- Camada hibrida (automacao):
  - `payment_methods` ligado a `ledger_accounts`.
  - `installment_plans` + `installment_plan_items` para gerar postings futuros.

## UC-01 — Registrar Entrada de Ganho

Veredito: Direto no ledger core.

Como fazer:
- Criar `ledger_transaction`.
- Postings:
  - DEBIT em conta ASSET (Banco/Dinheiro).
  - CREDIT em conta INCOME (Ganho).
  - `category_id` no posting de INCOME.

## UC-02 — Registrar Entrada de Investimento

Veredito: Direto no ledger core.

Como fazer:
- Mesmo fluxo da UC-01.
- `category_id` = Investimento.

## UC-03 — Registrar Saida Simples

Veredito: Direto no ledger core.

Como fazer:
- `ledger_transaction`.
- Postings:
  - DEBIT em EXPENSE (categoria OUT).
  - CREDIT em ASSET (Banco/Dinheiro).
  - `category_id` no posting de EXPENSE.

## UC-04 — Criar Nova Categoria

Veredito: Direto no ledger core.

Como fazer:
- Inserir em `ledger_categories`.

## UC-05 — Desativar Categoria

Veredito: Direto no ledger core.

Como fazer:
- `ledger_categories.is_active = false`.

## UC-06 — Registrar Gasto Fixo

Veredito: Hibrido (automacao).

Como fazer:
- Criar `installment_plan` com `plan_type = RECURRING`.
- Gerar `installment_plan_items` por periodo.
- Cada item gera um `ledger_transaction` quando confirmado.

## UC-07 — Copiar Gastos Fixos para Novo Mes

Veredito: Hibrido (automacao).

Como fazer:
- Em vez de copiar, gerar itens do plano para o mes.
- Criar postings do item quando o usuario confirmar.

## UC-08 — Registrar Compra Parcelada no Cartao

Veredito: Hibrido (automacao).

Como fazer:
- Criar `installment_plan` + `installment_plan_items`.
- Cada item gera:
  - DEBIT em EXPENSE (categoria).
  - CREDIT em LIABILITY (cartao).
  - `payment_method_id` no plano.

## UC-09 — Visualizar Fatura do Cartao

Veredito: Hibrido (consulta).

Como fazer:
- Buscar `installment_plan_items` por cartao e mes de vencimento.
- Somar valores e listar transacoes relacionadas.
- Limite restante = `credit_limit` - saldo atual da conta do cartao.

## UC-10 — Criar Orcamento Mensal

Veredito: Camada de orcamento fora do core (mas usando categorias do ledger).

Como fazer:
- Tabelas de orcamento referenciam `ledger_categories`.
- Total do mes vem da soma de INCOME (`direction = IN`, `is_budget_relevant = true`).
- Validacao de 100% permanece na camada de orcamento.

## UC-11 — Alterar Orcamento de Um Mes

Veredito: Camada de orcamento fora do core.

Como fazer:
- Atualizar `budget_items` do mes.

## UC-12 — Alterar Orcamento em Lote

Veredito: Camada de orcamento fora do core.

Como fazer:
- Aplicar update nos meses do intervalo.

## UC-13 — Cadastrar Pessoa de Picuinha

Veredito: Direto no ledger core.

Como fazer:
- Inserir em `ledger_parties` (kind = PERSON).

## UC-14 — Registrar Emprestimo (Eu Emprestei)

Veredito: Direto no ledger core (com regra de negocio).

Como fazer:
- `ledger_transaction`.
- Postings:
  - DEBIT em `Receivables:Picuinhas` (party_id).
  - CREDIT em ASSET (Banco/Dinheiro) se houve saida real.
- Se sem saida real, usar conta de contrapartida (EQUITY/Adjustments).

## UC-15 — Registrar Recebimento (Ela Pagou)

Veredito: Direto no ledger core.

Como fazer:
- `ledger_transaction`.
- Postings:
  - DEBIT em ASSET (Banco/Dinheiro).
  - CREDIT em `Receivables:Picuinhas` (party_id).

## UC-16 — Registrar Compra no Cartao para Outra Pessoa

Veredito: Hibrido + ledger core.

Como fazer:
- Criar `installment_plan` para o cartao.
- Cada item gera postings:
  - DEBIT em `Receivables:Picuinhas` (party_id).
  - CREDIT em LIABILITY (cartao).
- Evita contaminar despesas pessoais e orcamento.

## UC-17 — Consultar Saldo de Picuinha por Pessoa

Veredito: Direto no ledger core.

Como fazer:
- Somar postings da conta `Receivables:Picuinhas` por `party_id`.

## UC-18 — Visualizar Dashboard Mensal

Veredito: Direto no ledger core.

Como fazer:
- Somar postings por tipo de conta (INCOME/EXPENSE) no mes.
- Saldo = INCOME - EXPENSE.

## UC-19 — Visualizar Fluxo de Caixa Acumulado

Veredito: Direto no ledger core.

Como fazer:
- Calcular saldo acumulado de contas ASSET - LIABILITY por periodo.

## UC-20 — Visualizar Analise por Categoria

Veredito: Direto no ledger core.

Como fazer:
- Agrupar postings por `category_id` no mes e somar valores.

## UC-99 — Backup e Recuperacao

Veredito: Sem mudanca no modelo.

Como fazer:
- Dump do banco ledger e do banco legado (se existirem em paralelo).
