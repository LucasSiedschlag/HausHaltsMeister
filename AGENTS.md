# Repository Guidelines

## Purpose and Source of Truth
- This repository is being rebuilt from the ledger model in `docs/ledger/Reestruturação_Completa.md`.
- `docs/ledger/` is the authoritative spec for domain rules, flows, and data model decisions.
- Keep documentation and schema aligned; do not implement behavior that is not reflected in the docs.

## Ledger Model Baseline
- **Ledger boundary**: every record and query is scoped by `ledger_id`.
- **Journal core**: `transactions` (event) + `entries` (lines) are the source of truth.
- **IN/OUT semantics**: `categories.direction` defines direction; `entries.amount_cents` is always positive.
- **Budget**: percentage-based, calculated from monthly income base, versioned by `effective_from_month`.
- **Credit card**: spending enters the budget only when installments are posted; payments are technical transfers.

## Frontend ledger context
- `useLedgerContext` is the single source of truth for the active ledger and role; `useLedger` simply wraps it and keeps the `hhm_ledger_id` cookie synchronized.
- The header/sidebar must show the active role badge, surface ledger errors (`ledger.error`), and display the ledger name/role near the navigation.
- Actions that change ledger state (e.g., “Nova transação”) must disable for `viewer` roles and explain the restriction via tooltip.
- The ledger selector must show the role badge, and all ledger-specific screens run `ensure-ledger` before rendering.

## Frontend error handling
- Always use `useApiClient` so errors arrive normalized as `{ code, message, details }`.
- Rate-limit scenarios must expose `RATE_LIMITED` or `AUTH_RATE_LIMITED` so the UI can show a retry timer/message.

## Project Layout (current baseline)
- `docs/ledger/`: domain model, rules, and use cases.
- `cmd/`: backend entrypoint (when code is reintroduced).
- `db/`: SQL assets and sqlc queries (when present).
- `sqlc.yaml`, `sqlc-ledger.yaml`: sqlc generation configs.
- `Makefile`, `docker-compose.yaml`: build and local infra helpers.

## Layout Rules (mandatory)
- HTTP handlers live in `internal/adapters/http/handlers`; middleware in `internal/adapters/http/middleware`.
- HTTP DTOs live in `internal/adapters/http/dto`; shared HTTP helpers live in `internal/adapters/http/httpx`.
- Domain packages expose interfaces; adapters implement them. Domain never imports adapters or sqlc.
- sqlc generated code lives in `internal/adapters/postgres/sqlc`; SQL files are organized per module.

## Change Rules
- Keep table and field names consistent with the DBML in `docs/ledger/Reestruturação_Completa.md`.
- Avoid reintroducing legacy double-entry or posting-based models; the baseline is a ledger light model.
- Any change to flows (budget, investments, credit card) must update the corresponding docs.
- Represent enumerated values as `varchar` + `CHECK` constraints (no PostgreSQL ENUM types).
- Every table must have `created_at` and `updated_at`; `updated_at` has no default and is set by the app.
- Card networks are a catalog table (`card_networks`) referenced by `credit_cards.network`.
- Seeded technical categories must match the docs naming (including accents) and default flags.
- Identity is split from credentials: use `auth_identities` for providers and `auth_secrets` for passwords.
- Sessions use refresh tokens stored in `auth_sessions`; access tokens are short-lived JWTs.
- OAuth uses PKCE with `oauth_states` (state + code_verifier, short TTL, single-use).
- All feature work must follow `docs/agent/IMPLEMENTATION_PLAYBOOK.md`.
- Keep the policy matrix in `docs/api/Estruturação_de_Ledger_API.md` updated when roles change.

## Integrity and Safety
- Validate ledger ownership for accounts, categories, transactions, and entries.
- Enforce transfer balance (IN == OUT) at the service layer.
- Keep card posting routines idempotent to avoid duplicate entries.
- Normalize month fields to the first day (`YYYY-MM-01`) for budget and card cycles.
- Keep category flag semantics consistent: IN => `is_budget_relevant=false`, OUT => `is_budget_base=false`.
- Ledger scope scan (`TestQueriesScopedByLedgerID`) must stay green.
- Reports endpoints must remain rate-limited (and bulk endpoints when implemented).

## Autonomia e fluxo de trabalho

- Evite pedir confirmação do tipo "posso prosseguir?"
- Assuma que deve prosseguir até entregar a tarefa completa.
- Faça perguntas quando houver ambiguidade real que impeça a implementação (ex.: duas opções incompatíveis e nenhuma preferência indicada).
- Quando houver escolhas razoáveis, escolha a opção mais segura e alinhada às convenções do projeto, e documente rapidamente no PR/commit message.

## Regras de decisão (quando houver dúvida)

- Preferir consistência com padrões existentes no código.
- Preferir segurança e integridade (ledger boundary, idempotência, validações).

## Quando perguntar

Pergunte apenas se:
- há decisão que muda o contrato público (API) e não está documentada
- há risco de perda de dados (alteração destrutiva)
- há conflito direto com um documento existente
- falta um segredo/config (client id OAuth etc.) que impede rodar/testar

## Definition of Done (DoD)

Uma tarefa só está concluída quando:
1) Código implementado conforme docs
2) Testes criados/atualizados
3) Documentação atualizada (docs/api, docs/ledger, etc.)

## Relatório final

Ao finalizar, responda sempre com:
- Resumo do que foi feito
- Arquivos alterados
- Como rodar/testar
- TODOs (se existirem)
