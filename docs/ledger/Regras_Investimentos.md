# Investimentos (Accounts + Categories + Journal) — Regras de negócio + Pseudo-fluxo

Este documento define como o módulo de investimentos funciona dentro do modelo atual:

- `accounts` (wallet/current + investment)
- `transactions` + `entries` (journal)
- `categories` com `direction` + flags (`is_budget_base`, `is_budget_relevant`)
- Padrões de lançamento para: aporte, resgate, rendimento, ajuste
- Regras de orçamento relacionadas (ex.: “10% dos ganhos para investimentos”)
- Consultas e métricas por período (aportes, rendimentos, evolução)

---

## 0) Objetivos e princípios

### Objetivo funcional

1. Separar “Wallet/Pessoal” e “Investimentos” como contas internas (`accounts`).
2. Registrar aportes e resgates como transferências entre contas internas.
3. Registrar rendimentos (e perdas) dentro da conta de investimentos.
4. Permitir relatórios por período:
   - aportes por mês/ano
   - rendimentos por mês/ano
   - saldo/variação da conta de investimentos
5. Integrar com orçamento mensal por % quando desejado (aportes como categoria orçada).

### Princípios do modelo

- Investimentos não precisam ser “multi banco”: são uma conta interna.
- O journal é a fonte da verdade; saldos são derivados.
- `entries.amount_cents` é sempre positivo; IN/OUT vem de `categories.direction`.
- Transferência (`kind=transfer`) é a forma padrão de mover dinheiro entre contas.
- Rendimentos e perdas podem ser modelados como `kind=adjust` para indicar variação.

---

## 1) Entidades e conceitos

### 1.1. Accounts

- `Wallet/Pessoal` (`type=wallet`) ou `Conta Corrente` (`type=current`)
- `Investimentos` (`type=investment`)

### 1.2. Categorias de investimentos + investment_action

- `Investimentos (Entrada)`:
  - direction: in
  - is_budget_base: false
  - usada para aportes e rendimentos
- `Investimentos (Saída)`:
  - direction: out
  - is_budget_relevant: true (quando quiser orcar aportes)
  - usada para resgates e perdas
- Classificacao do movimento via `transactions.investment_action`:
  - `contribution` (aporte)
  - `redemption` (resgate)
  - `earnings` (rendimento)
  - `loss` (perda)

> As categorias separam IN/OUT, e o detalhe continua em `investment_action`.

---

## 2) Regras de orçamento relacionadas a investimentos

### 2.1. “10% dos ganhos vai para investimentos”

Há dois modos:

#### Modo recomendado (com orçamento)

- Definir uma `budget_plan_line` para a categoria `Investimentos (Saída)` com `percent=10`.
- Marcar `Investimentos (Saída).is_budget_relevant=true`.

Resultado:

- Limite mensal de aportes = renda_base_mes \* 10%
- Realizado = soma das entradas OUT na categoria Investimentos (Saída) no mês

#### Modo “fora do orçamento”

- `Investimentos (Saída).is_budget_relevant=false`
- Acompanhamento por relatórios de aportes, não por orçamento
  (útil se você quer orçamento apenas de consumo e tratar aportes separadamente)

---

## 3) Regras de integridade (consistência)

### 3.1. Ledger boundary

- Todas as contas e categorias devem pertencer ao mesmo `ledger_id` do journal.

### 3.2. Transferência (aporte/resgate) deve ser balanceada

Para transações `kind=transfer`:

- sem `investment_action`: total OUT == total IN (por direction)
- com `investment_action`: todas as entries devem ter o mesmo `amount_cents`
- recomendação prática: 2 entries (uma de cada lado)

### 3.3. Rendimentos/perdas não devem mexer na renda base

- a categoria `Investimentos (Entrada)` deve manter `is_budget_base=false`.
- `investment_action=earnings|loss` nao entra na renda base.

---

## 4) Fluxos de negócio (pseudo-fluxos)

## 4.1. Fluxo — Criar as contas padrão (setup)

Quando criar um ledger:

1. criar account `Wallet/Pessoal` (wallet) ou `Conta Corrente` (current)
2. criar account `Investimentos` (investment)
3. criar categorias `Investimentos (Entrada)` e `Investimentos (Saída)` (ou permitir criar depois)

---

## 4.2. Fluxo — Aporte (Wallet/Pessoal -> Investimentos)

Entrada:

- ledger_id
- user_id
- amount
- date
- opcional: memo/notes

Regras:

- este lançamento representa mover dinheiro para investimentos
- deve reduzir saldo de Wallet/Pessoal e aumentar saldo de Investimentos

Persistência (transação SQL):

1. criar `transactions`:
   - occurred_at = date
   - description = "Aporte investimentos"
   - investment_action = `contribution`
2. criar 2 `entries` (kind=transfer):
   A) OUT em Wallet/Pessoal:
   - account_id = Wallet/Pessoal
   - category_id = `Investimentos (Saída)` (OUT)
   - amount_cents = amount
   - kind = transfer
     B) IN em Investimentos:
   - account_id = Investimentos
   - category_id = `Investimentos (Entrada)` (IN)
   - amount_cents = amount
   - kind = transfer

Efeito:

- no orçamento (se relevante=true): o aporte consome o orçamento da categoria Investimentos (Saída)
- no saldo:
  - Wallet/Pessoal diminui
  - Investimentos aumenta

---

## 4.3. Fluxo — Resgate (Investimentos -> Wallet/Pessoal)

Entrada:

- ledger_id, user_id, amount, date, memo/notes

Persistência:

1. criar `transactions` "Resgate investimentos"
   - investment_action = `redemption`
2. 2 entries (kind=transfer):
   A) OUT em Investimentos:
   - account_id = Investimentos
   - category_id = `Investimentos (Saída)` (OUT)
   - amount = amount
     B) IN em Wallet/Pessoal:
   - account_id = Wallet/Pessoal
   - category_id = `Investimentos (Entrada)` (IN)
   - amount = amount

Efeito:

- nao deve inflar renda base (category Investimentos (Entrada) com is_budget_base=false)
- saldo:
  - Investimentos diminui
  - Wallet/Pessoal aumenta

---

## 4.4. Fluxo — Registrar rendimento (reinvestido)

Cenário:

- aporte de R$500
- rendimento de R$5 no mês
- você não transfere para Wallet/Pessoal, fica em Investimentos

Entrada:

- ledger_id, user_id
- amount (rendimento)
- date (geralmente último dia do mês)
- opcional: referência/nota (ex.: "1% no mês")

Persistência:

1. criar `transactions`:
   - description = "Rendimento investimentos (mês/ano)"
   - occurred_at = date
   - investment_action = `earnings`
2. criar 1 entry:
   - account_id = Investimentos
   - category_id = `Investimentos (Entrada)` (IN)
   - kind = adjust (recomendado) ou normal
   - amount_cents = rendimento

Efeito:

- saldo de investimentos aumenta
- orçamento não aumenta (base=false)

---

## 4.5. Fluxo — Registrar perda (opcional)

Se quiser acompanhar quedas:

1. transaction "Perda investimentos"
   - investment_action = `loss`
2. entry:
   - account_id = Investimentos
   - category_id = `Investimentos (Saída)` (OUT)
   - kind = adjust
   - amount = perda

Efeito:

- saldo de investimentos diminui
- orçamento não é afetado (relevant=false)

---

## 4.6. Fluxo — Ajuste manual (correções)

Exemplo: corrigir rendimento lançado errado

- criar transaction "Ajuste investimentos"
- entry IN ou OUT conforme a correção
- kind=adjust, memo explicando

---

## 5) Consultas e métricas (conceituais)

### 5.1. Aportes por período

Soma de entries:

- transactions.investment_action = `contribution`
- account = Investimentos
- filtrar por transactions.occurred_at no período

Saída:

- total aportado no período
- aportes por mês (group by mês)

### 5.2. Resgates por período

Soma de entries:

- transactions.investment_action = `redemption`
- account = Investimentos
- período

### 5.3. Rendimentos por período

Soma de entries:

- transactions.investment_action = `earnings`
- account = Investimentos
- período

### 5.4. Evolução do saldo de investimentos (por período)

Saldo derivado:

- para a conta Investimentos:
  - saldo = soma considerando `investment_action` para definir o sinal
- pode ser calculado:
  - até uma data (saldo atual)
  - por mês (saldo acumulado por competência)

### 5.5. Performance simplificada (sem preço de ativos)

Sem detalhar ativos e cotações, você ainda consegue:

- aporte total
- resgate total
- rendimento total
- variação líquida = rendimentos - perdas (se existir)
- saldo final

> Isso não é uma rentabilidade de carteira “real” com cotações, mas é ótimo para controle pessoal.

---

## 6) Integração com orçamento (exemplo completo)

### Exemplo:

- renda base do mês: 5.000
- orçamento de aporte: 10% => limite 500
- você aportou 500 (Wallet/Pessoal -> Investimentos)
- rendimento do mês: 5

No mês:

- orçamento:
  - limite de aportes = 500
  - realizado de aportes = 500
  - delta = 0 (dentro)
- saldo Investimentos:
  - +500 (investment_action=contribution)
  - +5 (rendimentos)
- renda base:
  - nao inclui movimentos com investment_action=earnings/loss

---

## 7) Regras de UX / consistência recomendadas (opcionais)

1. Assistente de “Aporte” na UI:

- formulário com valor/data
- gera automaticamente a transaction e as 2 entries com investment_action

2. Assistente de “Rendimento mensal”:

- sugere lançar no último dia do mês
- permite memo “x%”
- cria 1 entry `kind=adjust`

3. Bloquear erro comum:

- não permitir que usuário crie “Aporte” como gasto normal (kind=normal)
- sempre usar transfer para manter consistência (Wallet/Pessoal OUT + Investimentos IN)

---

## 8) Resultado final (o que estas regras garantem)

Separação clara de caixa vs investimentos.  
Aportes/resgates como transferências balanceadas.  
Rendimentos como eventos dentro de investimentos sem inflar orçamento.  
Métricas por período (aportes, rendimentos, saldo).  
Orçamento opcional para aportes (ex.: 10% dos ganhos).
