# API — Budget (% flexivel)

Este modulo cobre plano de budget, versoes, linhas e paines de acompanhamento (mensal e periodo).

## Padroes globais aplicados

- JSON em `snake_case`.
- Datas e timestamps: ISO-8601 em UTC.
- Meses: `YYYY-MM-01`.
- IDs: UUID (string).
- Auth: `Authorization: Bearer <access_token>`.
- Ledger boundary obrigatorio em todos os endpoints.
- Respostas de versoes/linhas/paineis incluem `ledger_id`.

## Inventario de endpoints

| Metodo | Path | Descricao | Auth | Role |
|---|---|---|---|---|
| GET | /ledgers/{ledgerId}/budget/plan | Obter plano | Sim | viewer |
| POST | /ledgers/{ledgerId}/budget/plan | Criar plano | Sim | editor |
| PATCH | /ledgers/{ledgerId}/budget/plan | Atualizar plano | Sim | editor |
| DELETE | /ledgers/{ledgerId}/budget/plan | Reset administrativo | Sim | owner |
| POST | /ledgers/{ledgerId}/budget/versions | Criar versao | Sim | editor |
| GET | /ledgers/{ledgerId}/budget/versions | Listar versoes | Sim | viewer |
| GET | /ledgers/{ledgerId}/budget/versions/{versionId} | Detalhe da versao | Sim | viewer |
| PATCH | /ledgers/{ledgerId}/budget/versions/{versionId} | Atualizar (restrito) | Sim | owner |
| DELETE | /ledgers/{ledgerId}/budget/versions/{versionId} | Remover versao | Sim | owner |
| POST | /ledgers/{ledgerId}/budget/versions/{versionId}/lines | Criar linha | Sim | editor |
| PATCH | /ledgers/{ledgerId}/budget/versions/{versionId}/lines/{lineId} | Atualizar linha | Sim | editor |
| DELETE | /ledgers/{ledgerId}/budget/versions/{versionId}/lines/{lineId} | Remover linha | Sim | editor |
| GET | /ledgers/{ledgerId}/budget/monthly | Painel mensal | Sim | viewer |
| GET | /ledgers/{ledgerId}/budget/period | Painel por periodo | Sim | viewer |

---

## Erros globais (5xx)

- 500 `INTERNAL_SERVER_ERROR`
- 503 `SERVICE_UNAVAILABLE`
- 504 `GATEWAY_TIMEOUT`

## Contratos detalhados

### GET /ledgers/{ledgerId}/budget/plan
1) Summary / Purpose
- Retornar plano atual.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Path params: `ledgerId`.

4) Response
- 200
```json
{ "id": "uuid", "ledger_id": "uuid", "name": "Padrao", "is_active": true, "created_at": "...", "updated_at": "..." }
```

5) Errors
- 404 `LEDGER_NOT_FOUND`

6) Semantics / Notes
- Pode ser criado automaticamente no signup.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### POST /ledgers/{ledgerId}/budget/plan
1) Summary / Purpose
- Criar plano.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body:
```json
{ "name": "Padrao", "is_active": true }
```

4) Response
- 201 (plano).

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Apenas um plano ativo por ledger.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### PATCH /ledgers/{ledgerId}/budget/plan
1) Summary / Purpose
- Atualizar plano.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body parcial.

4) Response
- 200.

5) Errors
- 404 `LEDGER_NOT_FOUND`

6) Semantics / Notes
- Pode desativar.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### DELETE /ledgers/{ledgerId}/budget/plan
1) Summary / Purpose
- Reset administrativo.

2) Auth & Authorization
- Role: owner.

3) Request
- Body: n/a.

4) Response
- 204.

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Uso restrito.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### POST /ledgers/{ledgerId}/budget/versions
1) Summary / Purpose
- Criar versao de budget.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body:
```json
{
  "effective_from_month": "2026-01-01",
  "lines": [ { "category_id": "uuid", "percent": 10, "include_children": false } ]
}
```

4) Response
- 201 (versao).

5) Errors
- 409 `BUDGET_VERSION_CONFLICT`
- 422 `BUDGET_CATEGORY_NOT_ELIGIBLE`

6) Semantics / Notes
- Percentual baseado em renda base.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### GET /ledgers/{ledgerId}/budget/versions
1) Summary / Purpose
- Listar versoes.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Query: `from`, `to` (month).

4) Response
- 200 (lista de versoes).

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Pode filtrar intervalo.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### GET /ledgers/{ledgerId}/budget/versions/{versionId}
1) Summary / Purpose
- Detalhe de versao.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Path: `versionId`.

4) Response
- 200 (versao + linhas).

5) Errors
- 404 `BUDGET_VERSION_CONFLICT` (usar NOT_FOUND quando aplicavel).

6) Semantics / Notes
- Inclui linhas.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### PATCH /ledgers/{ledgerId}/budget/versions/{versionId}
1) Summary / Purpose
- Atualizar versao (restrito).

2) Auth & Authorization
- Role: owner.

3) Request
- Body parcial.

4) Response
- 200.

5) Errors
- 501 `NOT_IMPLEMENTED`

6) Semantics / Notes
- Normalmente evitar em meses fechados.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### DELETE /ledgers/{ledgerId}/budget/versions/{versionId}
1) Summary / Purpose
- Remover versao.

2) Auth & Authorization
- Role: owner.

3) Request
- Body: n/a.

4) Response
- 204.

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Evitar em meses passados.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### POST /ledgers/{ledgerId}/budget/versions/{versionId}/lines
1) Summary / Purpose
- Criar linha.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body:
```json
{ "category_id": "uuid", "percent": 10, "include_children": false }
```

4) Response
- 201 (linha).

5) Errors
- 422 `BUDGET_CATEGORY_NOT_ELIGIBLE`

6) Semantics / Notes
- Percent > 0.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### PATCH /ledgers/{ledgerId}/budget/versions/{versionId}/lines/{lineId}
1) Summary / Purpose
- Atualizar linha.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body parcial.

4) Response
- 200.

5) Errors
- 404 `VALIDATION_ERROR` (quando linha nao existir).

6) Semantics / Notes
- Ajusta percentuais.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### DELETE /ledgers/{ledgerId}/budget/versions/{versionId}/lines/{lineId}
1) Summary / Purpose
- Remover linha.

2) Auth & Authorization
- Role: editor+.

3) Request
- Body: n/a.

4) Response
- 204.

5) Errors
- 403 `LEDGER_ACCESS_DENIED`

6) Semantics / Notes
- Nao afeta historico mensal.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### GET /ledgers/{ledgerId}/budget/monthly
1) Summary / Purpose
- Resumo mensal.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Query: `month` (YYYY-MM-01, required).

4) Response
- 200 (resumo + linhas).

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Mostra orcado vs realizado.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### GET /ledgers/{ledgerId}/budget/period
1) Summary / Purpose
- Resumo por periodo.

2) Auth & Authorization
- Role: viewer+.

3) Request
- Query: `from`, `to` (YYYY-MM-01, required).

4) Response
- 200 (agregacao + totais).

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Retorna totais por categoria e geral.

7) Pagination
- n/a.

8) Idempotency
- n/a.
