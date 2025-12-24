# Documentação da API HausHaltsMeister

Este documento descreve todos os endpoints da API do sistema, alinhados com a implementação atual (Go + Echo + DTOs).

**Observações Gerais:**

- Todas as datas devem ser enviadas no formato `YYYY-MM-DD` (exeto quando especificado timestamp).
- Os nomes dos campos no JSON seguem o padrão `snake_case`.
- Valores monetários são `float` (ex: `1250.50`).

---

## 0. Segurança e Erros

### 0.1 Autenticação

Todos os endpoints requerem o header:

- `X-App-Token`: `<token_configurado_na_env>`

Respostas comuns:

- `401 Unauthorized`: Token ausente ou inválido.

### 0.2 Tratamento de Erros e Limites

- **Timeouts**: Requests acima de 30s são abortados (`504 Gateway Timeout`).
- **Rate Limit**: Excesso de requisições retorna `429 Too Many Requests`.
- **Formato de Erro Padrão**:

```json
{
  "error": "Descrição do erro",
  "request_id": "req-123456"
}
```

---

## 1. Domínio: Categorias (`category`)

### 1.1 Criar Categoria

**Endpoint:** `POST /categories`

**Payload (JSON):**

```json
{
  "name": "Educação",
  "direction": "OUT",
  "is_budget_relevant": true
}
```

- `direction`: "IN" ou "OUT".
- `is_budget_relevant`: Define se aparece no planejamento.

**Response (201 Created):**

```json
{
  "id": 15,
  "name": "Educação",
  "direction": "OUT",
  "is_budget_relevant": true,
  "is_active": true
}
```

### 1.2 Listar Categorias

**Endpoint:** `GET /categories`

**Query Params:**

- `active` (bool): `true` para apenas ativas.

**Response (200 OK):**

```json
[
  {
    "id": 1,
    "name": "Salário",
    "direction": "IN",
    "is_budget_relevant": true,
    "is_active": true
  }
]
```

---

## 2. Domínio: Fluxo de Caixa (`cashflow`)

### 2.1 Criar Lançamento

**Endpoint:** `POST /cashflows`

**Payload (JSON):**

```json
{
  "date": "2024-03-15",
  "category_id": 10,
  "direction": "OUT",
  "title": "Jantar Especial",
  "amount": 250.0,
  "is_fixed": false
}
```

**Response (201 Created):**

```json
{
  "id": 42,
  "date": "2024-03-15",
  "category_id": 10,
  "direction": "OUT",
  "title": "Jantar Especial",
  "amount": 250.0,
  "is_fixed": false
}
```

### 2.2 Listar Lançamentos (Extrato)

**Endpoint:** `GET /cashflows`

**Query Params:**

- `month` (string): `YYYY-MM-DD` (ex: `2024-03-01`).

**Response (200 OK):** List of CashFlow objects.

### 2.3 Copiar Gastos Fixos

**Endpoint:** `POST /cashflows/copy-fixed`

**Payload (JSON):**

```json
{
  "from_month": "2024-02-01",
  "to_month": "2024-03-01"
}
```

**Response (200 OK):** `{"copied_count": 5}`

### 2.4 Resumo Mensal (Financial Summary)

**Endpoint:** `GET /cashflows/summary`

**Query Params:**

- `month` (string): `YYYY-MM-DD`.

**Response (200 OK):**

```json
{
  "total_income": 5000.0,
  "total_expense": 3500.0,
  "balance": 1500.0
}
```

### 2.5 Resumo por Categoria

**Endpoint:** `GET /cashflows/category-summary`

**Query Params:**

- `month` (string): `YYYY-MM-DD`.

**Response (200 OK):**

```json
[
  {
    "category_name": "Alimentação",
    "direction": "OUT",
    "total_amount": 1200.0
  }
]
```

---

## 3. Domínio: Orçamento (`budget`)

### 3.1 Definir Item de Orçamento

**Endpoint:** `POST /budgets/:month/items`

**Path Params:** `month` (YYYY-MM-DD).

**Payload (JSON):**

```json
{
  "category_id": 10,
  "planned_amount": 2000.0
}
```

**Response (200 OK):** BudgetItem object.

### 3.2 Definir Orçamento em Lote (Batch)

**Endpoint:** `POST /budgets/batch`

Use para aplicar o mesmo valor a vários meses futuros.

**Payload (JSON):**

```json
{
  "start_month": "2024-04-01",
  "end_month": "2024-12-01",
  "category_id": 10,
  "planned_amount": 2000.0
}
```

**Response (200 OK):** `{"status": "success"}`

### 3.3 Visualizar Orçamento (Planned vs Actual)

**Endpoint:** `GET /budgets/:month/summary`

**Response (200 OK):**

```json
{
  "month": "2024-03-01",
  "items": [
    {
      "id": 5,
      "budget_period_id": 10,
      "category_id": 10,
      "category_name": "Alimentação",
      "mode": "ABSOLUTE",
      "planned_amount": 2000.0,
      "actual_amount": 1200.0
    }
  ]
}
```

---

## 4. Domínio: Picuinhas (`picuinha`)

### 4.1 Cadastrar Pessoa

**Endpoint:** `POST /picuinhas/persons`

**Payload (JSON):**

```json
{
  "name": "João",
  "notes": "Amigo do trabalho"
}
```

### 4.2 Listar Pessoas (com Saldo)

**Endpoint:** `GET /picuinhas/persons`

**Response (200 OK):**

```json
[
  {
    "id": 1,
    "name": "João",
    "notes": "Amigo do trabalho",
    "balance": 150.0
  }
]
```

_Note: Balance > 0 means they ensure you (você tem crédito)._

### 4.3 Registrar Entrada (Empréstimo/Pagamento)

**Endpoint:** `POST /picuinhas/entries`

**Payload (JSON):**

```json
{
  "person_id": 1,
  "amount": 100.0,
  "kind": "PLUS",
  "auto_create_flow": true
}
```

- `kind`: "PLUS" (aumenta dívida dela/meu crédito) ou "MINUS" (diminui dívida dela/ela pagou).
- `auto_create_flow`: Se `true`, cria fluxo de caixa correspondente na categoria "Picuinhas".

**Response (201 Created):**

```json
{
  "id": 10,
  "person_id": 1,
  "amount": 100.0,
  "kind": "PLUS",
  "created_at": "2024-03-15T10:00:00Z"
}
```

---

## 5. Domínio: Cartões e Parcelamentos (`payment/installment`)

### 5.1 Criar Meio de Pagamento

**Endpoint:** `POST /payment-methods`

**Payload (JSON):**

```json
{
  "name": "Nubank",
  "kind": "CREDIT_CARD",
  "bank_name": "Nu Pagamentos",
  "closing_day": 1,
  "due_day": 7
}
```

### 5.2 Criar Compra Parcelada

**Endpoint:** `POST /installments`

**Payload (JSON):**

```json
{
  "description": "Notebook",
  "total_amount": 3000.0,
  "count": 10,
  "category_id": 15,
  "payment_method_id": 1,
  "purchase_date": "2024-03-15"
}
```

**Response (201 Created):** Includes `installment_amount` and `start_month`.

### 5.3 Visualizar Fatura

**Endpoint:** `GET /payment-methods/:id/invoice`

**Query Params:**

- `month` (string): `YYYY-MM-DD`.

**Response (200 OK):**

```json
{
  "total": 300.0,
  "entries": [
    {
      "title": "Notebook (1/10)",
      "amount": 300.0,
      "date": "2024-04-07"
    }
  ]
}
```
