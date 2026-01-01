# Cartão de crédito — Regras de negócio + Pseudo-fluxo (parcelas + fatura + orçamento por parcela)

Este documento descreve as regras e os fluxos para suportar cartão de crédito no modelo atual:

- `accounts` (inclui `credit_card`)
- `credit_cards` (metadados do cartão)
- `installment_plans` + `installments` (parcelamento)
- `credit_card_statements` (fatura)
- `transactions` + `entries` (journal)
- orçamento **só consome** quando a parcela é **postada** no mês

---

## 0) Objetivos e princípios

### Objetivo funcional

1. Registrar compras no cartão, inclusive parceladas.
2. Gerar faturas mensais (statement) com fechamento/vencimento.
3. Registrar pagamento da fatura como transferência de `Pessoal -> Cartão`.
4. **Orçamento e realizado** devem refletir **apenas as parcelas do mês** (não o total da compra no ato).

### Princípios do modelo

- Compra parcelada é um **plano** (`installment_plan`).
- Cada parcela é uma **unidade mensal** (`installment`) com `due_month`.
- Uma parcela só vira gasto (e entra no orçamento) quando é **postada**:
  - isso cria uma `transaction` + `entry` (no cartão) e grava `posted_transaction_id`.
- Pagamento de fatura **não** entra no orçamento (não duplicar gasto).
- Categorias técnicas servem para separar “liquidação de dívida” de “consumo”.

---

## 1) Definições e convenções

### 1.1. Dados de bandeira (catálogo)

- `card_networks` mantém o catálogo de bandeiras.
- `credit_cards.network` referencia `card_networks.code`.

### 1.2. Datas e “mês”

- `statement_month`, `due_month`, `first_due_month` são sempre `date` no 1º dia do mês.
  - Ex.: `2026-02-01` significa “competência fevereiro/2026”.

### 1.3. Conceito de “posted”

- Uma parcela está `scheduled` até ser postada.
- Ao postar:
  - cria `transactions` (evento)
  - cria `entries` (linha de gasto com categoria real)
  - atualiza `installments.status = posted`
  - grava `installments.posted_transaction_id`

### 1.4. Categorias: técnicas vs reais

#### Categorias reais (consumo)

- Ex.: Mercado, Eletrônicos, Lazer…
- `categories.direction = out`
- `categories.is_budget_relevant = true`
- Entram no orçamento quando postadas no mês.

#### Categorias técnicas (liquidação)

Para pagamento da fatura:

1. `Pagamento Fatura Cartão`

- direction: out
- is_budget_relevant: false

2. `Entrada Pagamento Cartão`

- direction: in
- is_budget_base: false

> O pagamento é transferência (não consumo).

---

## 2) Cálculo do ciclo do cartão (fechamento/vencimento)

### 2.1. Dados do cartão (credit_cards)

- `closing_day`: dia do fechamento (ex.: 25)
- `due_day`: dia do vencimento (ex.: 10)

### 2.2. Regras de calendário (robustez)

- Se `closing_day` ou `due_day` cair num dia inexistente do mês (ex.: 31 em fevereiro),
  ajustar para o último dia do mês.
- Mínimo recomendado: permitir somente 1..28 na UI para evitar edge-cases
  (mas o banco pode aceitar 1..31).

### 2.3. Determinar a fatura a partir da data da compra

Entrada:

- `purchase_occurred_at` (timestamp)
- cartão com `closing_day`, `due_day`

Saída:

- `statement_month` em que a compra cai (competência)
- `due_month` (competência do vencimento) — normalmente o mês do vencimento

**Regra padrão (comum no Brasil):**

- Compras feitas **antes ou no fechamento** entram na fatura que vence no mês seguinte do fechamento.
- Compras feitas **após o fechamento** entram na fatura do próximo ciclo (vencimento do mês seguinte).

Pseudo:

1. compute `closing_date` do ciclo atual:
   - `closing_date = date(year(purchase), month(purchase), closing_day_adjusted)`
2. se `purchase_date <= closing_date`:
   - compra entra na fatura cujo `due_date` é no próximo mês (ou conforme o emissor)
3. senão:
   - entra na fatura do ciclo seguinte (due mais adiante)

**Observação importante:**
Varia por banco/cartão. Por isso, na prática:

- o sistema pode oferecer dois modos configuráveis por cartão:
  - **Modo A (padrão)**: compra após fechamento vai para fatura do mês seguinte
  - **Modo B**: compra após fechamento vai para “duas faturas adiante” (alguns casos)
    Você pode começar com o modo padrão e evoluir.

### 2.4. Derivar `statement_month` e `due_month`

Para simplificar a implementação inicial:

- use `statement_month` = mês do vencimento (primeiro dia do mês do due_date)
- use `due_month` = `statement_month`

Assim:

- tudo que vence em fevereiro fica em `statement_month = 2026-02-01`.

---

## 3) Fluxo 1 — Cadastrar compra no cartão (parcelada)

### 3.1. Input do usuário (UI)

- cartão (account_id)
- data da compra (`purchase_occurred_at`)
- descrição/lojista (`merchant`/`description`)
- categoria real (ex.: Mercado)
- total (ex.: 600)
- número de parcelas (ex.: 2)
- (opcional) “primeira fatura” manualmente (override)

### 3.2. Regras de validação

- `installments_count >= 1`
- `total_amount = installment_amount * count` (se não bater, permitir última parcela diferente)
- categoria escolhida deve ser `direction=out`
- cartão escolhido deve ser `accounts.type=credit_card`

### 3.3. Determinar `first_due_month`

Regras:

- calcular qual `statement_month` a compra cai
- `first_due_month = statement_month`
- permitir override manual (usuário escolhe a primeira fatura)

### 3.4. Criar registros

Em uma transação no banco:

1. `INSERT installment_plans (...) RETURNING id`
2. gerar N parcelas:
   - for i in 1..N:
     - `due_month = add_months(first_due_month, i-1)`
     - `amount_cents = installment_amount_cents`
     - `status = scheduled`
   - `INSERT installments (...)`

> Não cria `transactions/entries` agora (porque orçamento só quando parcela cair no mês).

---

## 4) Fluxo 2 — Gerar/Postar parcelas do mês (o que entra no orçamento)

### 4.1. Quando roda

Opções:

- manual: botão “Gerar lançamentos do cartão para o mês X”
- automático: no 1º dia do mês (ou no dia do fechamento)

### 4.2. Entrada

- `ledger_id`
- `card_account_id`
- `target_month` (date: 1º dia do mês)

### 4.3. Seleção de parcelas a postar

Selecionar `installments`:

- `ledger_id = ...`
- `due_month = target_month`
- `status = scheduled`

### 4.4. Criação do journal (transações e entries)

Para cada parcela selecionada:

1. criar `transactions`:
   - `occurred_at`: recomendação = data do fechamento ou 1º dia do mês (consistência)
   - `description`: `"Parcela {n}/{N} - {merchant/description}"`
   - `notes`: opcional com referência do plano
2. criar `entries` (1 linha):

   - `account_id = card_account_id` (cartão)
   - `category_id = installment_plan.category_id` (categoria real: Mercado, etc.)
   - `kind = normal`
   - `amount_cents = installment.amount_cents`
   - `memo`: opcional `"Plano {id} parcela {n}/{N}"`

3. atualizar `installments`:
   - `status = posted`
   - `posted_transaction_id = transaction.id`

### 4.5. Efeito no orçamento

Como a parcela foi postada com categoria real OUT (`is_budget_relevant=true`):

- ela consome orçamento do mês do `due_month`.

Exemplo:

- orçamento Mercado do mês = 400
- compra 600 em 2x => parcelas 300 em dois meses
- mês 1 consome 300 do orçamento; mês 2 consome 300.

---

## 5) Fluxo 3 — Criar/atualizar fatura (statement)

### 5.1. Quando roda

- ao fechar o mês do cartão (no `closing_day`)
- ou manualmente: “Fechar fatura do mês X”

### 5.2. Entrada

- `card_account_id`
- `statement_month`

### 5.3. Garantir existência de statement

- buscar `credit_card_statements` para (card_account_id, statement_month)
- se não existir, criar com:
  - `closing_date` calculada para aquele mês
  - `due_date` calculada para aquele mês
  - `status = open` inicialmente

### 5.4. Compor totais

- totais podem ser derivados (recomendado) ou armazenados (cache).
  Para derivar:
- total_charges = soma `installments.amount_cents` para:
  - `due_month = statement_month`
  - `status in (posted, paid)` (ou `posted` se quiser “aberto”)
- total_payments = soma de pagamentos associados (`payment_transaction_id` ou múltiplos pagamentos no futuro)

### 5.5. Fechamento

Ao fechar:

- `status = closed`
- snapshot dos totais (opcional)
- a fatura está pronta para pagamento.

---

## 6) Fluxo 4 — Pagar fatura (transferência Pessoal -> Cartão)

### 6.1. Entrada (UI)

- `statement_id`
- `payment_date`
- `pay_amount` (normalmente total da fatura, mas permitir parcial)
- `cash_account_id` (Pessoal)

### 6.2. Criar transaction de pagamento

Em uma transação de banco:

1. criar `transactions`:
   - description: `"Pagamento fatura {cartão} {statement_month}"`
   - occurred_at = payment_date
2. criar 2 `entries` (transfer):
   A) OUT em Pessoal:

   - account_id = cash_account_id
   - category = `Pagamento Fatura Cartão` (OUT, budget_relevant=false)
   - kind = transfer
   - amount = pay_amount
     B) IN no Cartão:
   - account_id = card_account_id
   - category = `Entrada Pagamento Cartão` (IN, budget_base=false)
   - kind = transfer
   - amount = pay_amount

3. atualizar `credit_card_statements`:

   - `payment_transaction_id = transaction.id`
   - `status`:
     - se pagou total => `paid`
     - se parcial => manter `closed` (e registrar pagamentos múltiplos futuramente)

4. marcar `installments` como `paid` (se pagamento total):
   - selecionar installments com `due_month = statement_month` e `status=posted`
   - set `status=paid`, `paid_statement_id = statement.id`

### 6.3. Por que pagamento não entra no orçamento?

Porque o gasto já entrou quando as parcelas foram postadas nas categorias reais (Mercado, etc.).
Pagamento é liquidação da dívida, não “novo consumo”.

---

## 7) Fluxo 5 — Cancelar compra / estornar parcela

### 7.1. Cancelar plano antes de postar

Se nenhuma parcela foi postada:

- `installment_plans.status = cancelled`
- `installments.status = skipped` (ou deletar — recomendado manter histórico)

### 7.2. Estornar após postar

Se parcelas já foram postadas, você não deve apagar.
Você faz ajuste:

- criar `transaction` “Estorno parcela X”
- criar `entry` IN no cartão na mesma categoria real (ou categoria “Estorno” IN)
- `kind=adjust`
  E atualizar a parcela como `skipped` ou manter `paid` com marcação adicional (depende do seu reporting).

Recomendação:

- manter histórico e usar ajustes para reversões.

---

## 8) Consultas essenciais (conceituais)

### 8.1. Parcelas do mês (para orçamento e fatura)

- listar installments por `due_month` e status

### 8.2. Gastos no cartão por categoria no mês

- somar `entries.amount_cents` onde:
  - entry.account = cartão
  - categoria OUT real
  - mês do occurred_at (ou do due_month via join em installments/posting)

### 8.3. Aderência do orçamento

- orçamento calcula limite por % da renda base
- realizado = soma das entries OUT relevantes (incluindo parcelas postadas no mês)

### 8.4. Fatura do mês

- statement totals derivados de installments do mês
- status open/closed/paid

---

## 9) Decisões importantes já alinhadas com o requisito #3 (orçamento por parcela)

Compra parcelada **não** consome orçamento no mês da compra.  
Cada parcela consome orçamento **no mês em que é devida/postada**.  
Pagamento da fatura **não** consome orçamento.

---

## 10) Próximas extensões possíveis (não obrigatórias agora)

- compras não parceladas no cartão (count=1) seguem o mesmo fluxo
- juros/IOF/multa:
  - tratados como parcelas extras ou entries `kind=adjust` em categoria OUT específica (ex.: “Juros cartão”)
- pagamento parcial:
  - permitir múltiplos pagamentos por statement (tabela `statement_payments`)
- conciliação/importação:
  - usar `transactions.external_source/external_id` para evitar duplicidade
