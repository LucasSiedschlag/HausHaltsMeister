# Contratos de Integração Front-End

Este documento descreve os contratos de dados (DTOs) e endpoints disponíveis para consumo pelo Front-End.
Ele é derivado da implementação real e deve ser a referência para criação de interfaces TypeScript.

> **OpenAPI**: Veja [`./openapi.yaml`](./openapi.yaml) para a especificação técnica completa.

---

## 1. Padrões Globais

### 1.1 Formatos de Dados

- **Data (Date)**: String no formato `YYYY-MM-DD`. Ex: `"2024-03-25"`.
- **Timestamp**: String ISO 8601. Ex: `"2024-03-25T14:30:00Z"`.
- **Dinheiro**: Number (float/double). O front deve cuidar da formatação de moeda (R$).
- **IDs**: Number (integer).

### 1.2 Tratamento de Erro

Todos os endpoints, em caso de falha (4xx ou 5xx), retornam o seguinte formato:

```typescript
interface ErrorResponse {
  error: string; // Mensagem legível do erro
  request_id?: string; // ID para rastreamento (opcional)
}
```

Exemplo:

```json
{
  "error": "category direction mismatch",
  "request_id": "req-12345"
}
```

---

## 2. Categorias (`/categories`)

### 2.1 Types

```typescript
type Direction = "IN" | "OUT";

interface Category {
  id: number;
  name: string;
  direction: Direction;
  is_budget_relevant: boolean;
  is_active: boolean;
  inactive_from_month?: string | null; // YYYY-MM-DD
}

interface CreateCategoryRequest {
  name: string;
  direction: Direction;
  is_budget_relevant: boolean;
}

interface UpdateCategoryRequest {
  name: string;
  direction: Direction;
  is_budget_relevant: boolean;
  is_active: boolean;
}
```

### 2.2 Endpoints

**GET /categories**

- Query: `?active=true` (opcional), `?month=YYYY-MM-DD` (opcional)
- Response: `Category[]`

**POST /categories**

- Body: `CreateCategoryRequest`
- Response: `Category`

**PUT /categories/:id**

- Body: `UpdateCategoryRequest`
- Response: `Category`

**PATCH /categories/:id/deactivate**

- Query: `?effective_month=YYYY-MM-DD` (opcional)
- Response: `{ status: string }`

---

## 3. Fluxo de Caixa (`/cashflows`)

### 3.1 Types

```typescript
interface CashFlow {
  id: number;
  date: string; // YYYY-MM-DD
  category_id: number;
  direction: Direction;
  title: string;
  amount: number;
  is_fixed: boolean;
}

interface CreateCashFlowRequest {
  date: string;
  category_id: number;
  direction: Direction;
  title: string;
  amount: number;
  is_fixed: boolean;
}

interface MonthlySummary {
  total_income: number;
  total_expense: number;
  balance: number;
}

interface CategorySummary {
  category_name: string;
  direction: Direction;
  total_amount: number;
}
```

### 3.2 Endpoints

**GET /cashflows**

- Query: `?month=YYYY-MM-DD` (Obrigatório, dia 1)
- Response: `CashFlow[]`

**POST /cashflows**

- Body: `CreateCashFlowRequest`
- Response: `CashFlow`

**GET /cashflows/summary**

- Query: `?month=YYYY-MM-DD`
- Response: `MonthlySummary`

**GET /cashflows/category-summary**

- Query: `?month=YYYY-MM-DD`
- Response: `CategorySummary[]`

---

## 4. Orçamento (`/budgets`)

### 4.1 Types

```typescript
type BudgetMode = "ABSOLUTE" | "PERCENT_OF_INCOME";

interface BudgetItem {
  id: number;
  budget_period_id: number;
  category_id: number;
  category_name?: string;
  mode: BudgetMode;
  planned_amount: number;
  actual_amount: number;
  target_percent: number;
}

interface SetBudgetItemRequest {
  category_id: number;
  mode: BudgetMode;
  planned_amount?: number;
  target_percent?: number;
}

interface SetBudgetItemsBulkRequest {
  category_id: number;
  target_percent: number;
}

interface UpdateBudgetItemRequest {
  mode: BudgetMode;
  planned_amount?: number;
  target_percent?: number;
}

interface BudgetSummary {
  month: string;
  total_income: number;
  items: BudgetItem[];
}
```

### 4.2 Endpoints

**GET /budgets/:month/summary**

- Params: `month` (YYYY-MM-DD)
- Response: `BudgetSummary`

**POST /budgets/:month/items**

- Params: `month`
- Body: `SetBudgetItemRequest`
- Response: `BudgetItem`

**PUT /budgets/:month/items**

- Params: `month`
- Body: `SetBudgetItemsBulkRequest[]`
- Response: `{ status: string }`

**PUT /budgets/items/:id**

- Body: `UpdateBudgetItemRequest`
- Response: `BudgetItem`

---

## 5. Picuinhas (`/picuinhas`)

### 5.1 Types

```typescript
interface Person {
  id: number;
  name: string;
  notes: string;
  balance: number; // >0: Você tem crédito (receber). <0: Você deve (pagar).
}

interface CreatePersonRequest {
  name: string;
  notes: string;
}

interface AddEntryRequest {
  person_id: number;
  kind: "PLUS" | "MINUS"; // PLUS: Aumenta dívida dela. MINUS: Diminui.
  amount: number;
  auto_create_flow: boolean; // Se true, cria registro em CashFlow
}
```

### 5.2 Endpoints

**GET /picuinhas/persons**

- Response: `Person[]`

**POST /picuinhas/persons**

- Body: `CreatePersonRequest`
- Response: `Person`

**POST /picuinhas/entries**

- Body: `AddEntryRequest`
- Response: `EntryResponse` (ver OpenAPI)

---

## 6. Cartões e Parcelas (`/cards`)

### 6.1 Types

```typescript
interface PaymentMethod {
  id: number;
  name: string;
  kind: "CREDIT_CARD";
  bank_name: string;
  closing_day?: number;
  due_day?: number;
}

interface CreateInstallmentRequest {
  description: string;
  total_amount: number;
  count: number;
  category_id: number;
  payment_method_id: number;
  purchase_date: string; // YYYY-MM-DD
}

interface Invoice {
  month: string;
  total: number;
  entries: {
    date: string;
    title: string;
    amount: number;
    category_name: string;
  }[];
}
```

### 6.2 Endpoints

**POST /payment-methods**

- Body: `CreatePaymentMethodRequest`
- Response: `PaymentMethod`

**POST /installments**

- Body: `CreateInstallmentRequest`
- Response: `{ installment_amount: number, start_month: string }`

**GET /payment-methods/:id/invoice**

- Query: `?month=YYYY-MM-DD`
- Response: `Invoice`

---

## Notas de Implementação

1. **Estado Inicial**: Comece implementando `Categories` e `CashFlows`.
2. **Segurança**: O cabeçalho `X-App-Token` deve ser enviado se configurado no backend, mas para desenvolvimento local pode estar desativado dependendo do `.env`.
3. **Validação**: O backend valida tudo, mas o front deve impedir envio de valores negativos em `amount` (exceto ajustes específicos) e garantir datas válidas.
