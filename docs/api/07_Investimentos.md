# API — Investimentos

Este modulo cobre aportes, resgates e rendimentos (atalhos para journal), alem de resumo agregado.

## Padroes globais aplicados

- JSON em `snake_case`.
- Datas e timestamps: ISO-8601 em UTC.
- IDs: UUID (string).
- Auth: `Authorization: Bearer <access_token>`.
- Ledger boundary obrigatorio em todos os endpoints.
- Respostas incluem `ledger_id`.

## Inventario de endpoints

| Metodo | Path | Descricao | Auth | Role |
|---|---|---|---|---|
| POST | /ledgers/{ledgerId}/investments/contributions | Aporte | Sim | editor |
| POST | /ledgers/{ledgerId}/investments/redemptions | Resgate | Sim | editor |
| POST | /ledgers/{ledgerId}/investments/earnings | Rendimento | Sim | editor |
| GET | /ledgers/{ledgerId}/investments/summary | Resumo | Sim | viewer |

---

## Erros globais (5xx)

- 500 `INTERNAL_SERVER_ERROR`
- 503 `SERVICE_UNAVAILABLE`
- 504 `GATEWAY_TIMEOUT`

## Contratos detalhados

### POST /ledgers/{ledgerId}/investments/contributions
1) Summary / Purpose
- Registrar aporte.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body:
```json
{ "amount_cents": 50000, "occurred_at": "2026-02-01", "memo": "Aporte mensal" }
```

4) Response
- 201 (transaction criada).

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Gera entries de investimento.

7) Pagination
- n/a.

8) Idempotency
- Opcional.

---

### POST /ledgers/{ledgerId}/investments/redemptions
1) Summary / Purpose
- Registrar resgate.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body:
```json
{ "amount_cents": 50000, "occurred_at": "2026-03-01", "memo": "Resgate" }
```

4) Response
- 201 (transaction criada).

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Movimenta saldo de investimento.

7) Pagination
- n/a.

8) Idempotency
- Opcional.

---

### POST /ledgers/{ledgerId}/investments/earnings
1) Summary / Purpose
- Registrar rendimento.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body:
```json
{ "amount_cents": 500, "occurred_at": "2026-03-31", "memo": "Rendimento mensal" }
```

4) Response
- 201 (transaction criada).

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Lanca ganho no investimento.

7) Pagination
- n/a.

8) Idempotency
- Opcional.

---

### GET /ledgers/{ledgerId}/investments/summary
1) Summary / Purpose
- Resumo agregado no periodo.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Query params:
  - `from` (date-time, opcional)
  - `to` (date-time, opcional)

4) Response
- 200
```json
{
  "ledger_id": "uuid",
  "from": "2026-02-01T00:00:00Z",
  "to": "2026-03-31T23:59:59Z",
  "total_contributions": 0,
  "total_redemptions": 0,
  "total_earnings": 0,
  "total_losses": 0,
  "net_variation": 0
}
```

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Intervalo fechado se ambos enviados.

7) Pagination
- n/a.

8) Idempotency
- n/a.
