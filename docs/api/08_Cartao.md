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
| GET | /accounts/{accountId}/credit-cards | Listar cartoes da conta | Sim | viewer |
| POST | /accounts/{accountId}/credit-cards | Criar cartao para a conta | Sim | editor |
| GET | /credit-cards/{cardId} | Detalhe do cartao | Sim | viewer |
| PATCH | /credit-cards/{cardId} | Atualizar cartao | Sim | editor |
| DELETE | /credit-cards/{cardId} | Remover cartao | Sim | editor |
| POST | /credit-cards/{cardId}/plans | Criar plano | Sim | editor |
| GET | /credit-cards/{cardId}/plans | Listar planos | Sim | viewer |
| GET | /credit-cards/{cardId}/plans/{planId} | Detalhe do plano | Sim | viewer |
| PATCH | /credit-cards/{cardId}/plans/{planId} | Atualizar plano | Sim | editor |
| DELETE | /credit-cards/{cardId}/plans/{planId} | Remover plano | Sim | editor |
| GET | /credit-cards/{cardId}/installments | Listar parcelas | Sim | viewer |
| PATCH | /credit-cards/{cardId}/installments/{installmentId} | Atualizar parcela | Sim | editor |
| POST | /credit-cards/{cardId}/post | Postar parcelas do mes | Sim | editor |
| GET | /credit-cards/{cardId}/statements | Listar faturas | Sim | viewer |
| GET | /credit-cards/{cardId}/statements/{statementId} | Detalhe da fatura | Sim | viewer |
| POST | /credit-cards/{cardId}/statements/close | Fechar fatura | Sim | editor |
| POST | /credit-cards/{cardId}/statements/pay | Pagar fatura | Sim | editor |
| PATCH | /credit-cards/{cardId}/statements/{statementId} | Ajuste manual | Sim | owner |

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

### POST /accounts/{accountId}/credit-cards
1) Summary / Purpose
- Criar cartao.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body:
```json
{
  "label": "Cartao Principal",
  "brand": "visa",
  "last4": "1234",
  "cvv": "123",
  "holder_name": "Fulano da Silva",
  "active": true,
  "color": "slate",
  "style": "gradient",
  "closing_day": 25,
  "due_day": 10
}
```

4) Response
- 201 (cartao).

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- `accountId` vem do path e deve pertencer ao ledger.
- o passivo do cartao e criado automaticamente (`type=current`, `nature=liability`).

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### GET /accounts/{accountId}/credit-cards
1) Summary / Purpose
- Listar cartoes.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Path: `accountId`.

4) Response
- 200 (lista).

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Ledger boundary via account.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### GET /credit-cards/{cardId}
1) Summary / Purpose
- Detalhe do cartao.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Path: `cardId`.

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

### PATCH /credit-cards/{cardId}
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

### DELETE /credit-cards/{cardId}
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

### POST /credit-cards/{cardId}/plans
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

### GET /credit-cards/{cardId}/plans
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

### GET /credit-cards/{cardId}/plans/{planId}
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

### PATCH /credit-cards/{cardId}/plans/{planId}
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

### DELETE /credit-cards/{cardId}/plans/{planId}
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

### GET /credit-cards/{cardId}/installments
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

### PATCH /credit-cards/{cardId}/installments/{installmentId}
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

### POST /credit-cards/{cardId}/post
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

### GET /credit-cards/{cardId}/statements
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

### GET /credit-cards/{cardId}/statements/{statementId}
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

### POST /credit-cards/{cardId}/statements/close
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

### POST /credit-cards/{cardId}/statements/pay
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
  "paying_account_id": "uuid"
}
```

4) Response
- 200 (statement pago).

5) Errors
- 404 `CREDITCARD_STATEMENT_NOT_FOUND`
- 409 `CREDITCARD_STATEMENT_ALREADY_PAID`
- 422 `CREDITCARD_PAYMENT_EXCEEDS_TOTAL`

6) Semantics / Notes
- Gera transaction de pagamento (transfer para o passivo do cartao).

7) Pagination
- n/a.

8) Idempotency
- Recomendada por pagamento.

---

### PATCH /credit-cards/{cardId}/statements/{statementId}
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
