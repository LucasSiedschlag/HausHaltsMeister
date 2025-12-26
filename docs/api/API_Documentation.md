# Documentação da API HausHaltsMeister

Este documento descreve todos os endpoints da API do sistema, alinhados com a implementação atual (Go + Echo + DTOs).

**Observações Gerais:**

- Todas as datas devem ser enviadas no formato `YYYY-MM-DD` (exeto quando especificado timestamp).
- Os nomes dos campos no JSON seguem o padrão `snake_case`.
- Valores monetários são `float` (ex: `1250.50`).

---

## 0. Segurança e Erros

### 0.1 Autenticação

> Para detalhes de configuração, veja [`docs/Seguranca.md`](../Seguranca.md).

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
  "is_active": true,
  "inactive_from_month": null
}
```

### 1.2 Listar Categorias

**Endpoint:** `GET /categories`

**Query Params:**

- `active` (bool): `true` para apenas ativas.
- `month` (string YYYY-MM-DD): filtra categorias válidas para o mês informado.

**Response (200 OK):**

```json
[
  {
    "id": 1,
    "name": "Salário",
    "direction": "IN",
    "is_budget_relevant": true,
    "is_active": true,
    "inactive_from_month": null
  }
]
```

---

### 1.3 Atualizar Categoria

**Endpoint:** `PUT /categories/{id}`

**Payload (JSON):**

```json
{
  "name": "Educação",
  "direction": "OUT",
  "is_budget_relevant": true,
  "is_active": true
}
```

**Response (200 OK):**

```json
{
  "id": 15,
  "name": "Educação",
  "direction": "OUT",
  "is_budget_relevant": true,
  "is_active": true,
  "inactive_from_month": null
}
```

**Erros (400/404):**

```json
{
  "error": "category name cannot be empty",
  "request_id": "req-123456"
}
```

---

### 1.4 Desativar Categoria

**Endpoint:** `PATCH /categories/{id}/deactivate`

**Query Params:**

- `effective_month` (string YYYY-MM-DD): mês a partir do qual a categoria deixa de aparecer.

**Response (200 OK):**

```json
{
  "status": "deactivated"
}
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

Notas:
- `PERCENT_OF_INCOME` usa como base a soma das entradas (`direction = IN`) cujas categorias estão marcadas com `is_budget_relevant = true` no mês.

### 3.1 Definir Item de Orçamento

**Endpoint:** `POST /budgets/:month/items`

**Path Params:** `month` (YYYY-MM-DD).

**Payload (JSON):**

```json
{
  "category_id": 10,
  "mode": "PERCENT_OF_INCOME",
  "target_percent": 25
}
```

Modos suportados:
- `PERCENT_OF_INCOME`: percentual sobre a soma das entradas (`direction = IN`) com `is_budget_relevant = true`.
- `ABSOLUTE`: valor absoluto (usa `planned_amount`).
`target_percent` deve ser entre 0 e 100.

**Response (200 OK):** BudgetItem object.

### 3.2 Definir Itens de Orçamento em Lote (mês)

**Endpoint:** `PUT /budgets/:month/items`

**Path Params:** `month` (YYYY-MM-DD).

**Payload (JSON):**

```json
[
  {
    "category_id": 10,
    "target_percent": 25
  },
  {
    "category_id": 11,
    "target_percent": 15
  }
]
```

Regras:
- A soma de `target_percent` deve ser 100.

**Response (200 OK):** `{"status": "success"}`

### 3.3 Atualizar Item de Orçamento

**Endpoint:** `PUT /budgets/items/{id}`

**Payload (JSON):**

```json
{
  "mode": "PERCENT_OF_INCOME",
  "target_percent": 30
}
```

**Response (200 OK):** BudgetItem object.

### 3.4 Definir Orçamento em Lote (Batch)

**Endpoint:** `POST /budgets/batch`

Use para aplicar o mesmo valor a vários meses futuros.

**Payload (JSON):**

```json
{
  "start_month": "2024-04-01",
  "end_month": "2024-12-01",
  "category_id": 10,
  "mode": "PERCENT_OF_INCOME",
  "target_percent": 15
}
```

**Response (200 OK):** `{"status": "success"}`

### 3.5 Visualizar Orçamento (Planned vs Actual)

**Endpoint:** `GET /budgets/:month/summary`

**Response (200 OK):**

```json
{
  "month": "2024-03-01",
  "total_income": 5000.0,
  "items": [
    {
      "id": 5,
      "budget_period_id": 10,
      "category_id": 10,
      "category_name": "Alimentação",
      "mode": "PERCENT_OF_INCOME",
      "planned_amount": 1250.0,
      "actual_amount": 1200.0,
      "target_percent": 25
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

### 4.3 Atualizar Pessoa

**Endpoint:** `PUT /picuinhas/persons/{id}`

**Payload (JSON):**

```json
{
  "name": "João",
  "notes": "Amigo de infância"
}
```

**Response (200 OK):**

```json
{
  "id": 1,
  "name": "João",
  "notes": "Amigo de infância",
  "balance": 150.0
}
```

### 4.4 Excluir Pessoa

**Endpoint:** `DELETE /picuinhas/persons/{id}`

**Response (200 OK):**

```json
{
  "status": "deleted"
}
```

**Observação:** não é permitido excluir uma pessoa que já tenha lançamentos.

### 4.5 Registrar Entrada (Empréstimo/Pagamento)

**Endpoint:** `POST /picuinhas/entries`

**Payload (JSON):**

```json
{
  "person_id": 1,
  "amount": 100.0,
  "kind": "PLUS",
  "auto_create_flow": true,
  "payment_method_id": 2,
  "card_owner": "SELF"
}
```

- `kind`: "PLUS" (aumenta dívida dela/meu crédito) ou "MINUS" (diminui dívida dela/ela pagou).
- `auto_create_flow`: Se `true`, cria fluxo de caixa correspondente na categoria "Picuinhas".
- `payment_method_id`: Opcional. Cartão usado no lançamento.
- `card_owner`: "SELF" (cartão do usuário) ou "THIRD" (cartão de outra pessoa). Para "THIRD" o `auto_create_flow` é ignorado.

**Response (201 Created):**

```json
{
  "id": 10,
  "person_id": 1,
  "amount": 100.0,
  "kind": "PLUS",
  "cash_flow_id": 33,
  "payment_method_id": 2,
  "card_owner": "SELF",
  "created_at": "2024-03-15T10:00:00Z"
}
```

### 4.6 Listar Entradas

**Endpoint:** `GET /picuinhas/entries`

**Query Params (opcional):**

- `person_id` (int): filtra lançamentos por pessoa.

**Response (200 OK):**

```json
[
  {
    "id": 10,
    "person_id": 1,
    "amount": 100.0,
    "kind": "PLUS",
    "cash_flow_id": 55,
    "created_at": "2024-03-15T10:00:00Z"
  }
]
```

### 4.7 Atualizar Entrada

**Endpoint:** `PUT /picuinhas/entries/{id}`

**Payload (JSON):**

```json
{
  "person_id": 1,
  "amount": 80.0,
  "kind": "PLUS",
  "auto_create_flow": true,
  "payment_method_id": 2,
  "card_owner": "SELF"
}
```

**Response (200 OK):** `PicuinhaEntryResponse`

### 4.8 Excluir Entrada

**Endpoint:** `DELETE /picuinhas/entries/{id}`

**Response (200 OK):**

```json
{
  "status": "deleted"
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
  "credit_limit": 5000.0,
  "closing_day": 1,
  "due_day": 7
}
```

### 5.1.1 Listar Meios de Pagamento

**Endpoint:** `GET /payment-methods`

**Response (200 OK):**

```json
[
  {
    "id": 1,
    "name": "Nubank",
    "kind": "CREDIT_CARD",
    "bank_name": "Nu Pagamentos",
    "credit_limit": 5000.0,
    "closing_day": 1,
    "due_day": 7,
    "is_active": true
  }
]
```

### 5.1.2 Atualizar Meio de Pagamento

**Endpoint:** `PUT /payment-methods/{id}`

**Payload (JSON):**

```json
{
  "name": "Nubank",
  "kind": "CREDIT_CARD",
  "bank_name": "Nu Pagamentos",
  "credit_limit": 5000.0,
  "closing_day": 1,
  "due_day": 7,
  "is_active": true
}
```

**Response (200 OK):**

```json
{
  "id": 1,
  "name": "Nubank",
  "kind": "CREDIT_CARD",
  "bank_name": "Nu Pagamentos",
  "credit_limit": 5000.0,
  "closing_day": 1,
  "due_day": 7,
  "is_active": true
}
```

### 5.1.3 Excluir Meio de Pagamento

**Endpoint:** `DELETE /payment-methods/{id}`

**Obs:** a exclusão é lógica (is_active = false) para preservar histórico.

**Response (200 OK):**

```json
{
  "status": "deleted"
}
```

### 5.2 Criar Compra Parcelada

**Endpoint:** `POST /installments`

**Payload (JSON):**

```json
{
  "description": "Notebook",
  "amount_mode": "TOTAL",
  "total_amount": 3000.0,
  "count": 10,
  "category_id": 15,
  "payment_method_id": 1,
  "purchase_date": "2024-03-15"
}
```

**Obs:** preencha apenas um entre `total_amount` e `installment_amount`.  
`amount_mode` aceita `TOTAL` ou `INSTALLMENT`.

**Exemplo usando valor da parcela:**

```json
{
  "description": "Notebook",
  "amount_mode": "INSTALLMENT",
  "installment_amount": 300.0,
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
  "total_remaining": 2700.0,
  "entries": [
    {
      "title": "Notebook (1/10)",
      "amount": 300.0,
      "date": "2024-04-07"
    }
  ]
}
```
