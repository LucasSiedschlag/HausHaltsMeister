# API — Relatorios

## Endpoints

- `GET /ledgers/{ledgerId}/reports/balances?month=YYYY-MM-01`
- `GET /ledgers/{ledgerId}/reports/categories?from=&to=`
- `GET /ledgers/{ledgerId}/reports/cashflow?from=&to=`

Notas:
- Saldos por conta seguem a regra do journal (cash/investment: IN-OUT; credit_card: OUT-IN).
- Relatorios sempre filtram por `ledger_id` e por periodo.
