# Documentação da API HausHaltsMeister

Este documento descreve todos os endpoints da API do sistema HausHaltsMeister, organizados por domínio.

---

## 1. Domínio: Categorias (`category`)

Gerencia as classificações de fluxo de caixa (ex: Salário, Custos fixos, Prazeres).

### 1.1 Criar Categoria

**Endpoint:** `POST /categories`  
**Descrição:** Registra uma nova categoria de fluxo no sistema.

**Inputs:**

- `name` (string, obrigatório): Nome descritivo da categoria.
- `direction` (string, obrigatório): Direção fixa da categoria (`IN` ou `OUT`).

**Exemplo de Requisição (JSON):**

```json
{
  "name": "Educação",
  "direction": "OUT"
}
```

**Exemplo de Resposta (JSON - 201 Created):**

```json
{
  "ID": 15,
  "Name": "Educação",
  "Direction": "OUT",
  "IsBudgetRelevant": true,
  "IsActive": true
}
```

---

### 1.2 Listar Categorias

**Endpoint:** `GET /categories`  
**Descrição:** Retorna a lista de todas as categorias cadastradas (incluindo as pré-definidas no sistema).

**Parâmetros de Consulta (Query Params):**

- `active` (boolean, opcional): Se `true`, retorna apenas categorias ativas. Default: `false`.

**Exemplo de Requisição:**
`GET /categories?active=true`

**Exemplo de Resposta (JSON - 200 OK):**

```json
[
  {
    "ID": 1,
    "Name": "Salário",
    "Direction": "IN",
    "IsBudgetRelevant": true,
    "IsActive": true
  },
  {
    "ID": 5,
    "Name": "Investimentos",
    "Direction": "OUT",
    "IsBudgetRelevant": true,
    "IsActive": true
  },
  {
    "ID": 10,
    "Name": "Prazeres",
    "Direction": "OUT",
    "IsBudgetRelevant": true,
    "IsActive": true
  }
]
```

---

## 2. Domínio: Fluxo de Caixa (`cashflow`)

Gerencia as entradas e saídas reais de dinheiro.

### 2.1 Criar Lançamento

**Endpoint:** `POST /cashflows`  
**Descrição:** Registra uma nova movimentação financeira baseada nas categorias existentes.

**Inputs:**

- `date` (string, obrigatório): Data da movimentação no formato `YYYY-MM-DD`.
- `category_id` (int, obrigatório): ID da categoria associada (ex: 10 para 'Prazeres').
- `direction` (string, obrigatório): Direção da movimentação (`IN` ou `OUT`). Deve coincidir com a direção da categoria.
- `title` (string, obrigatório): Título ou descrição curta do lançamento.
- `amount` (float, obrigatório): Valor da movimentação (maior que zero).

**Exemplo de Requisição (JSON):**

```json
{
  "date": "2025-12-24",
  "category_id": 10,
  "direction": "OUT",
  "title": "Jantar Especial",
  "amount": 250.0
}
```

**Exemplo de Resposta (JSON - 201 Created):**

```json
{
  "ID": 42,
  "Date": "2025-12-24T00:00:00Z",
  "CategoryID": 10,
  "Direction": "OUT",
  "Title": "Jantar Especial",
  "Amount": 250.0
}
```

---

### 2.2 Listar Lançamentos por Mês

**Endpoint:** `GET /cashflows`  
**Descrição:** Retorna todos os lançamentos de um mês específico.

**Parâmetros de Consulta (Query Params):**

- `month` (string, obrigatório): Primeiro dia do mês desejado no formato `YYYY-MM-DD`.

**Exemplo de Requisição:**
`GET /cashflows?month=2025-12-01`

**Exemplo de Resposta (JSON - 200 OK):**

```json
[
  {
    "ID": 42,
    "Date": "2025-12-24T00:00:00Z",
    "CategoryID": 10,
    "Direction": "OUT",
    "Title": "Jantar Especial",
    "Amount": 250.0
  }
]
```

---

## 3. Domínio: Orçamento (`budget`)

Permite planejar gastos mensais por categoria e comparar com o realizado.

### 3.1 Definir Item de Orçamento

**Endpoint:** `POST /budgets/:month/items`  
**Descrição:** Define ou atualiza o valor planejado para uma categoria em um mês.

**Parâmetros de Caminho (Path Params):**

- `month` (string, obrigatório): Primeiro dia do mês no formato `YYYY-MM-DD`.

**Inputs:**

- `category_id` (int, obrigatório): ID da categoria (deve ser `OUT`, ex: 10 para 'Prazeres').
- `planned_amount` (float, obrigatório): Valor orçado para o mês.

**Exemplo de Requisição (JSON):**

```json
{
  "category_id": 10,
  "planned_amount": 2000.0
}
```

**Exemplo de Resposta (JSON - 200 OK):**

```json
{
  "ID": 1,
  "BudgetPeriodID": 100,
  "CategoryID": 10,
  "CategoryName": "Prazeres",
  "Mode": "ABSOLUTE",
  "PlannedAmount": 2000.0,
  "ActualAmount": 0,
  "TargetPercent": 0,
  "Notes": ""
}
```

---

### 3.2 Resumo do Orçamento (Orçado x Realizado)

**Endpoint:** `GET /budgets/:month/summary`  
**Descrição:** Retorna o plano orçamentário completo do mês com os valores reais gastos em cada categoria.

**Parâmetros de Caminho (Path Params):**

- `month` (string, obrigatório): Primeiro dia do mês no formato `YYYY-MM-DD`.

**Exemplo de Requisição:**
`GET /budgets/2025-12-01/summary`

**Exemplo de Resposta (JSON - 200 OK):**

```json
{
  "ID": 100,
  "Month": "2025-12-01T00:00:00Z",
  "AnalysisMode": "DEFAULT",
  "IsClosed": false,
  "Items": [
    {
      "ID": 1,
      "BudgetPeriodID": 100,
      "CategoryID": 10,
      "CategoryName": "Prazeres",
      "Mode": "ABSOLUTE",
      "PlannedAmount": 2000.0,
      "ActualAmount": 250.0,
      "TargetPercent": 0,
      "Notes": ""
    }
  ]
}
```

---

## 4. Domínio: Picuinhas (`picuinha`)

Gerencia empréstimos, dívidas e dinheiro de terceiros.

### 4.1 Cadastrar Pessoa

**Endpoint:** `POST /picuinhas/persons`  
**Descrição:** Adiciona uma nova pessoa para controle de dívidas.

**Inputs:**

- `name` (string, obrigatório): Nome da pessoa.
- `notes` (string, opcional): Notas adicionais.

**Exemplo de Requisição (JSON):**

```json
{
  "name": "Joãozinho",
  "notes": "Colega de trabalho"
}
```

**Exemplo de Resposta (JSON - 201 Created):**

```json
{
  "ID": 1,
  "Name": "Joãozinho",
  "Notes": "Colega de trabalho",
  "Balance": 0
}
```

---

### 4.2 Listar Pessoas e Saldos

**Endpoint:** `GET /picuinhas/persons`  
**Descrição:** Retorna a lista de pessoas com seus respectivos saldos devedores (positivo = deve para você, negativo = você deve para ela).

**Exemplo de Resposta (JSON - 200 OK):**

```json
[
  {
    "ID": 1,
    "Name": "Joãozinho",
    "Notes": "Colega de trabalho",
    "Balance": 150.0
  }
]
```

---

### 4.3 Registrar Movimentação de Dívida

**Endpoint:** `POST /picuinhas/entries`  
**Descrição:** Registra um novo empréstimo (PLUS) ou recebimento (MINUS).

**Inputs:**

- `person_id` (int, obrigatório): ID da pessoa.
- `kind` (string, obrigatório): Tipo da movimentação (`PLUS` para aumentar a dívida dela, `MINUS` para diminuir).
- `amount` (float, obrigatório): Valor da movimentação.
- `cash_flow_id` (int, opcional): Vínculo com um lançamento manual de fluxo de caixa.
- `auto_create_flow` (boolean, opcional): Se `true`, cria automaticamente um lançamento no Fluxo de Caixa na categoria pré-definida 'Picuinhas'.

**Exemplo de Requisição (JSON):**

```json
{
  "person_id": 1,
  "kind": "PLUS",
  "amount": 150.0,
  "auto_create_flow": true
}
```

**Exemplo de Resposta (JSON - 201 Created):**

```json
{
  "ID": 5,
  "PersonID": 1,
  "Date": "2025-12-24T00:00:00Z",
  "Kind": "PLUS",
  "Amount": 150.0,
  "CashFlowID": 201
}
```

---

## 5. Domínio: Meios de Pagamento (`payment`)

Gerencia as formas como o dinheiro é transacionado (Cartões, Pix, Dinheiro).

### 5.1 Criar Meio de Pagamento

**Endpoint:** `POST /payment-methods`  
**Descrição:** Registra um novo meio de pagamento.

**Inputs:**

- `name` (string, obrigatório): Nome (ex: "Nubank").
- `kind` (string, obrigatório): Tipo (`CREDIT_CARD`, `DEBIT_CARD`, `CASH`, `PIX`, `BANK_SLIP`).
- `bank_name` (string, opcional): Nome do banco.
- `closing_day` (int, opcional): Dia de fechamento (apenas para `CREDIT_CARD`).
- `due_day` (int, opcional): Dia de vencimento (apenas para `CREDIT_CARD`).

**Exemplo de Requisição (JSON):**

```json
{
  "name": "Nubank",
  "kind": "CREDIT_CARD",
  "bank_name": "Nubank S.A.",
  "closing_day": 1,
  "due_day": 7
}
```

**Exemplo de Resposta (JSON - 201 Created):**

```json
{
  "ID": 1,
  "Name": "Nubank",
  "Kind": "CREDIT_CARD",
  "BankName": "Nubank S.A.",
  "ClosingDay": 1,
  "DueDay": 7,
  "IsActive": true
}
```

---

### 5.2 Visualizar Fatura do Cartão

**Endpoint:** `GET /payment-methods/:id/invoice`  
**Descrição:** Lista todos os lançamentos vinculados a um cartão específico em um determinado mês.

**Parâmetros de Caminho (Path Params):**

- `id` (int, obrigatório): ID do meio de pagamento.

**Parâmetros de Consulta (Query Params):**

- `month` (string, obrigatório): Mês da fatura no formato `YYYY-MM-DD`.

**Exemplo de Requisição:**
`GET /payment-methods/1/invoice?month=2024-01-01`

**Exemplo de Resposta (JSON - 200 OK):**

```json
{
  "PaymentMethodID": 1,
  "Month": "2024-01-01T00:00:00Z",
  "Total": 500.0,
  "Entries": [
    {
      "CashFlowID": 101,
      "Date": "2024-01-07T00:00:00Z",
      "Title": "Compra iPhone (1/10)",
      "Amount": 500.0,
      "CategoryName": "Picuinhas"
    }
  ]
}
```

---

## 6. Domínio: Parcelamentos (`installment`)

Gerencia compras parceladas com geração automática de fluxos de caixa.

### 6.1 Criar Compra Parcelada

**Endpoint:** `POST /installments`  
**Descrição:** Registra uma compra e gera automaticamente todos os lançamentos de fluxo de caixa futuros.

**Inputs:**

- `description` (string, obrigatório): Título da compra.
- `total_amount` (float, obrigatório): Valor total bruto.
- `count` (int, obrigatório): Número de parcelas.
- `category_id` (int, obrigatório): Categoria do gasto (ex: 12 para 'Educação').
- `payment_method_id` (int, obrigatório): ID do cartão ou meio de pagamento.
- `purchase_date` (string, obrigatório): Data da compra (YYYY-MM-DD).

**Exemplo de Requisição (JSON):**

```json
{
  "description": "Compra iPhone",
  "total_amount": 5000.0,
  "count": 10,
  "category_id": 12,
  "payment_method_id": 1,
  "purchase_date": "2023-12-25"
}
```

**Exemplo de Resposta (JSON - 201 Created):**

```json
{
  "ID": 1,
  "Description": "Compra iPhone",
  "TotalAmount": 5000.0,
  "InstallmentCount": 10,
  "InstallmentAmount": 500.0,
  "StartMonth": "2023-12-25T00:00:00Z",
  "PaymentMethodID": 1
}
```
