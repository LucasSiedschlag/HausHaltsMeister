# API — Cartao de Credito

## Catalogo

- `GET /card-networks`
- `POST /card-networks`
- `PATCH /card-networks/{code}`
- `DELETE /card-networks/{code}`

## Cartoes

- `GET /ledgers/{ledgerId}/credit-cards`
- `GET /ledgers/{ledgerId}/credit-cards/{cardAccountId}`
- `POST /ledgers/{ledgerId}/credit-cards`
- `PATCH /ledgers/{ledgerId}/credit-cards/{cardAccountId}`
- `DELETE /ledgers/{ledgerId}/credit-cards/{cardAccountId}`

### POST /ledgers/{ledgerId}/credit-cards

Body:
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

## Parcelamento

- `POST /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans`
- `GET /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans?status=`
- `GET /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans/{planId}`
- `PATCH /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans/{planId}`
- `DELETE /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans/{planId}` (apenas se nenhuma parcela postada)

### POST /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans

Body:
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

## Parcelas (installments)

- `GET /ledgers/{ledgerId}/credit-cards/{cardAccountId}/installments?month=&status=`
- `PATCH /ledgers/{ledgerId}/credit-cards/{cardAccountId}/installments/{installmentId}` (ex.: marcar skipped)

## Posting mensal

- `POST /ledgers/{ledgerId}/credit-cards/{cardAccountId}/post?month=YYYY-MM-01`

## Faturas

- `GET /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements?month=`
- `GET /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/{statementId}`
- `POST /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/close?month=YYYY-MM-01`
- `POST /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/pay`
- `PATCH /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/{statementId}` (uso restrito)

### POST /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/pay

Body:
```json
{
  "statement_id": "uuid",
  "payment_date": "2026-03-10",
  "pay_amount_cents": 300000,
  "cash_account_id": "uuid"
}
```
