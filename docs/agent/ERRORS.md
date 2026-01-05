# Error Catalog (API)

This catalog defines stable error codes and their HTTP mapping. All API errors must use these codes and the standard error payload format defined in `docs/agent/CONVENTIONS.md`.

Payload shape:
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Human readable message",
  "details": {
    "field": "reason"
  }
}
```

---

| Code | HTTP | Quando | Body.details |
|------|------|--------|-------------|
| AUTH_INVALID_CREDENTIALS | 401 | Login falhou (email/senha invalido) | `{ "email": "invalid" }` |
| AUTH_USER_INACTIVE | 403 | Usuario inativo tentou operar | `{ "user_id": "uuid" }` |
| AUTH_REFRESH_REVOKED | 401 | Refresh token revogado/expirado | `{ "session_id": "uuid" }` |
| AUTH_RATE_LIMITED | 429 | Tentativas excederam limite | `{ "retry_after": "seconds" }` |
| RATE_LIMITED | 429 | Limite de requisicoes excedido | `{ "retry_after": "seconds" }` |
| AUTH_EMAIL_NOT_VERIFIED | 403 | Email nao verificado (quando exigido) | `{ "email": "user@example.com" }` |
| AUTH_OAUTH_STATE_INVALID | 401 | State OAuth invalido/expirado | `{ "state": "..." }` |
| AUTH_OAUTH_PROVIDER_ERROR | 502 | Provider retornou erro | `{ "provider": "google" }` |
| AUTH_OAUTH_EMAIL_REQUIRED | 422 | Provider nao retornou email e e obrigatorio | `{ "provider": "github" }` |
| LEDGER_ACCESS_DENIED | 403 | Usuario sem acesso ao ledger | `{ "ledger_id": "uuid" }` |
| LEDGER_NOT_FOUND | 404 | Ledger nao encontrado | `{ "ledger_id": "uuid" }` |
| MEMBER_ALREADY_EXISTS | 409 | Membro ja existe no ledger | `{ "user_id": "uuid" }` |
| ACCOUNT_NOT_FOUND | 404 | Conta nao encontrada no ledger | `{ "account_id": "uuid" }` |
| CATEGORY_NOT_FOUND | 404 | Categoria nao encontrada no ledger | `{ "category_id": "uuid" }` |
| TRANSACTION_NOT_FOUND | 404 | Transacao nao encontrada no ledger | `{ "transaction_id": "uuid" }` |
| TRANSACTION_REFERENCED | 409 | Transacao referenciada por parcelas/fatura | `{ "transaction_id": "uuid" }` |
| IDEMPOTENCY_KEY_CONFLICT | 409 | Chave reutilizada com payload diferente | `{ "idempotency_key": "..." }` |
| IDEMPOTENCY_KEY_EXPIRED | 409 | Chave expirada (fora da janela) | `{ "idempotency_key": "..." }` |
| VALIDATION_ERROR | 422 | Validacao de dominio falhou | `{ "field": "reason" }` |
| CONFLICT_DUPLICATE_EMAIL | 409 | Email ja cadastrado | `{ "email": "user@example.com" }` |
| CONFLICT_DUPLICATE_NAME | 409 | Nome duplicado (category/account) | `{ "name": "..." }` |
| BUDGET_VERSION_CONFLICT | 409 | Versao de budget ja existe no mes | `{ "effective_from_month": "YYYY-MM-01" }` |
| BUDGET_CATEGORY_NOT_ELIGIBLE | 422 | Categoria invalida para budget | `{ "category_id": "uuid" }` |
| TRANSFER_NOT_BALANCED | 422 | Transferencia nao balanceada | `{ "out_total": 100, "in_total": 80 }` |
| CREDITCARD_CARD_NOT_FOUND | 404 | Cartao nao encontrado | `{ "card_account_id": "uuid" }` |
| CREDITCARD_INSTALLMENT_ALREADY_POSTED | 409 | Parcela ja postada | `{ "installment_id": "uuid" }` |
| CREDITCARD_STATEMENT_NOT_FOUND | 404 | Fatura nao encontrada | `{ "statement_id": "uuid" }` |
| CREDITCARD_STATEMENT_ALREADY_PAID | 409 | Fatura ja paga | `{ "statement_id": "uuid" }` |
| CREDITCARD_PAYMENT_EXCEEDS_TOTAL | 422 | Pagamento maior que total | `{ "pay_amount_cents": 0, "total_cents": 0 }` |
| NOT_IMPLEMENTED | 501 | Endpoint ou fluxo nao disponivel | `{ "feature": "..." }` |
| INTERNAL_SERVER_ERROR | 500 | Erro inesperado no servidor | `{ "request_id": "..." }` |
| SERVICE_UNAVAILABLE | 503 | Servico indisponivel (manutencao/dependencia) | `{ "retry_after": "seconds" }` |
| GATEWAY_TIMEOUT | 504 | Timeout em dependencia externa | `{ "request_id": "..." }` |
