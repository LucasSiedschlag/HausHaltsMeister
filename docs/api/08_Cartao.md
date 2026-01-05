# API — Cartao de Credito

Este modulo cobre bandeiras, cartoes, planos de parcelamento, parcelas e faturas.

## Padroes globais aplicados

- JSON em `snake_case`.
- Datas e timestamps: ISO-8601 em UTC.
- Meses: `YYYY-MM-01`.
- IDs: UUID (string).
- Auth: `Authorization: Bearer <access_token>`.
- Ledger boundary obrigatorio em todos os endpoints.

## Inventario de endpoints

| Metodo | Path | Descricao | Auth | Role |
|---|---|---|---|---|
| GET | /card-networks | Listar bandeiras | Sim | viewer |
| POST | /card-networks | Criar bandeira | Sim | owner |
| PATCH | /card-networks/{code} | Atualizar bandeira | Sim | owner |
| DELETE | /card-networks/{code} | Remover bandeira | Sim | owner |
| GET | /ledgers/{ledgerId}/credit-cards | Listar cartoes | Sim | viewer |
| GET | /ledgers/{ledgerId}/credit-cards/{cardAccountId} | Detalhe do cartao | Sim | viewer |
| POST | /ledgers/{ledgerId}/credit-cards | Criar cartao | Sim | editor |
| PATCH | /ledgers/{ledgerId}/credit-cards/{cardAccountId} | Atualizar cartao | Sim | editor |
| DELETE | /ledgers/{ledgerId}/credit-cards/{cardAccountId} | Remover cartao | Sim | editor |
| POST | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans | Criar plano | Sim | editor |
| GET | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans | Listar planos | Sim | viewer |
| GET | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans/{planId} | Detalhe do plano | Sim | viewer |
| PATCH | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans/{planId} | Atualizar plano | Sim | editor |
| DELETE | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans/{planId} | Remover plano | Sim | editor |
| GET | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/installments | Listar parcelas | Sim | viewer |
| PATCH | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/installments/{installmentId} | Atualizar parcela | Sim | editor |
| POST | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/post | Postar parcelas do mes | Sim | editor |
| GET | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements | Listar faturas | Sim | viewer |
| GET | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/{statementId} | Detalhe da fatura | Sim | viewer |
| POST | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/close | Fechar fatura | Sim | editor |
| POST | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/pay | Pagar fatura | Sim | editor |
| PATCH | /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/{statementId} | Ajuste manual | Sim | owner |

---

## Erros globais (5xx)

- 500 `INTERNAL_SERVER_ERROR`
- 503 `SERVICE_UNAVAILABLE`
- 504 `GATEWAY_TIMEOUT`

## Contratos detalhados

### GET /card-networks
1) Summary / Purpose
- Listar bandeiras.

2) Auth & Authorization
- Role: viewer+.

3) Request
- n/a.

4) Response
- 200
```json
[ { "code": "visa", "name": "Visa", "is_active": true } ]
```

5) Errors
- 401 `AUTH_INVALID_CREDENTIALS`

6) Semantics / Notes
- Catalogo global.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### POST /card-networks
1) Summary / Purpose
- Criar bandeira.

2) Auth & Authorization
- Role: owner.

3) Request
- Body:
```json
{ "code": "visa", "name": "Visa", "is_active": true }
```

4) Response
- 201 (bandeira).

5) Errors
- 409 `CONFLICT_DUPLICATE_NAME`

6) Semantics / Notes
- `code` unico.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### PATCH /card-networks/{code}
1) Summary / Purpose
- Atualizar bandeira.

2) Auth & Authorization
- Role: owner.

3) Request
- Path: `code`.
- Body parcial.

4) Response
- 200.

5) Errors
- 404 `VALIDATION_ERROR`

6) Semantics / Notes
- Permite ativar/desativar.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### DELETE /card-networks/{code}
1) Summary / Purpose
- Remover bandeira.

2) Auth & Authorization
- Role: owner.

3) Request
- Path: `code`.

4) Response
- 204.

5) Errors
- 404 `VALIDATION_ERROR`

6) Semantics / Notes
- Restrito.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### POST /ledgers/{ledgerId}/credit-cards
1) Summary / Purpose
- Criar cartao.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body:
```json
{
  "account_id": "uuid",
  "issuer_name": "Banco X",
  "network": "visa",
  "nickname": "Cartao Principal",
  "last4": "1234",
  "credit_limit_cents": 500000,
  "closing_day": 25,
  "due_day": 10
}
```

4) Response
- 201 (cartao).

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- `account_id` deve ser tipo credit_card.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### GET /ledgers/{ledgerId}/credit-cards
1) Summary / Purpose
- Listar cartoes.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Path: `ledgerId`.

4) Response
- 200 (lista).

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Ledger boundary.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### GET /ledgers/{ledgerId}/credit-cards/{cardAccountId}
1) Summary / Purpose
- Detalhe do cartao.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Path: `cardAccountId`.

4) Response
- 200 (cartao).

5) Errors
- 404 `CREDITCARD_CARD_NOT_FOUND`

6) Semantics / Notes
- n/a.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### PATCH /ledgers/{ledgerId}/credit-cards/{cardAccountId}
1) Summary / Purpose
- Atualizar cartao.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body parcial.

4) Response
- 200.

5) Errors
- 404 `CREDITCARD_CARD_NOT_FOUND`

6) Semantics / Notes
- Permite alterar limites e datas.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### DELETE /ledgers/{ledgerId}/credit-cards/{cardAccountId}
1) Summary / Purpose
- Remover cartao.

2) Auth & Authorization
- Role: editor+.

3) Request
- n/a.

4) Response
- 204.

5) Errors
- 404 `CREDITCARD_CARD_NOT_FOUND`

6) Semantics / Notes
- Bloquear se houver planos ativos.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### POST /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans
1) Summary / Purpose
- Criar plano de parcelas.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body:
```json
{
  "purchase_occurred_at": "2026-02-15",
  "merchant": "Loja X",
  "description": "Notebook",
  "category_id": "uuid",
  "total_amount_cents": 600000,
  "installments_count": 6,
  "installment_amount_cents": 100000,
  "first_due_month": "2026-03-01"
}
```

4) Response
- 201 (plano).

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Compra vira plano + parcelas.

7) Pagination
- n/a.

8) Idempotency
- opcional.

---

### GET /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans
1) Summary / Purpose
- Listar planos.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Query: `status` (active|cancelled|finished).

4) Response
- 200 (lista).

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Filtra status.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### GET /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans/{planId}
1) Summary / Purpose
- Detalhe do plano.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Path: `planId`.

4) Response
- 200 (plano + parcelas).

5) Errors
- 404 `CREDITCARD_CARD_NOT_FOUND`

6) Semantics / Notes
- Inclui installments.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### PATCH /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans/{planId}
1) Summary / Purpose
- Atualizar plano.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body parcial.

4) Response
- 200.

5) Errors
- 409 `CREDITCARD_INSTALLMENT_ALREADY_POSTED`

6) Semantics / Notes
- Bloquear se parcelas postadas.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### DELETE /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans/{planId}
1) Summary / Purpose
- Remover plano.

2) Auth & Authorization
- Role: editor+.

3) Request
- n/a.

4) Response
- 204.

5) Errors
- 409 `CREDITCARD_INSTALLMENT_ALREADY_POSTED`

6) Semantics / Notes
- Apenas se nenhuma parcela postada.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### GET /ledgers/{ledgerId}/credit-cards/{cardAccountId}/installments
1) Summary / Purpose
- Listar parcelas.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Query: `month` (YYYY-MM-01), `status`.

4) Response
- 200 (lista).

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Status derivado.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### PATCH /ledgers/{ledgerId}/credit-cards/{cardAccountId}/installments/{installmentId}
1) Summary / Purpose
- Atualizar parcela (ex.: skipped).

2) Auth & Authorization
- Role: editor+.

3) Request
- Body:
```json
{ "status": "skipped" }
```

4) Response
- 200.

5) Errors
- 409 `CREDITCARD_INSTALLMENT_ALREADY_POSTED`

6) Semantics / Notes
- Somente enquanto nao postada.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### POST /ledgers/{ledgerId}/credit-cards/{cardAccountId}/post
1) Summary / Purpose
- Postar parcelas do mes.

2) Auth & Authorization
- Role: editor+.

3) Request
- Query: `month` (YYYY-MM-01).
- Headers: `Idempotency-Key` (recomendado).

4) Response
- 200
```json
{ "posted_count": 3, "transaction_id": "uuid" }
```

5) Errors
- 409 `CREDITCARD_INSTALLMENT_ALREADY_POSTED`

6) Semantics / Notes
- Idempotente por mes.

7) Pagination
- n/a.

8) Idempotency
- Obrigatoria por mes.

---

### GET /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements
1) Summary / Purpose
- Listar faturas.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Query: `month` (YYYY-MM-01, opcional).

4) Response
- 200 (lista).

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Inclui status.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### GET /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/{statementId}
1) Summary / Purpose
- Detalhe da fatura.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Path: `statementId`.

4) Response
- 200 (statement).

5) Errors
- 404 `CREDITCARD_STATEMENT_NOT_FOUND`

6) Semantics / Notes
- n/a.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### POST /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/close
1) Summary / Purpose
- Fechar fatura do mes.

2) Auth & Authorization
- Role: editor+.

3) Request
- Query: `month` (YYYY-MM-01).

4) Response
- 200 (statement fechado).

5) Errors
- 409 `VALIDATION_ERROR`

6) Semantics / Notes
- Fecha periodo e calcula total.

7) Pagination
- n/a.

8) Idempotency
- Recomendada por mes.

---

### POST /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/pay
1) Summary / Purpose
- Pagar fatura.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body:
```json
{
  "statement_id": "uuid",
  "payment_date": "2026-03-10",
  "pay_amount_cents": 300000,
  "cash_account_id": "uuid"
}
```

4) Response
- 200 (statement pago).

5) Errors
- 404 `CREDITCARD_STATEMENT_NOT_FOUND`
- 409 `CREDITCARD_STATEMENT_ALREADY_PAID`
- 422 `CREDITCARD_PAYMENT_EXCEEDS_TOTAL`

6) Semantics / Notes
- Gera transaction de pagamento.

7) Pagination
- n/a.

8) Idempotency
- Recomendada por pagamento.

---

### PATCH /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/{statementId}
1) Summary / Purpose
- Ajuste manual (restrito).

2) Auth & Authorization
- Role: owner.

3) Request
- Body parcial.

4) Response
- 200.

5) Errors
- 501 `NOT_IMPLEMENTED`

6) Semantics / Notes
- Uso administrativo.

7) Pagination
- n/a.

8) Idempotency
- n/a.
