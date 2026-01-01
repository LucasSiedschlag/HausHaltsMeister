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

## Project Layout (current baseline)
- `docs/ledger/`: domain model, rules, and use cases.
- `cmd/`: backend entrypoint (when code is reintroduced).
- `db/`: SQL assets and sqlc queries (when present).
- `sqlc.yaml`, `sqlc-ledger.yaml`: sqlc generation configs.
- `Makefile`, `docker-compose.yaml`: build and local infra helpers.

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

## Integrity and Safety
- Validate ledger ownership for accounts, categories, transactions, and entries.
- Enforce transfer balance (IN == OUT) at the service layer.
- Keep card posting routines idempotent to avoid duplicate entries.
- Normalize month fields to the first day (`YYYY-MM-01`) for budget and card cycles.
- Keep category flag semantics consistent: IN => `is_budget_relevant=false`, OUT => `is_budget_base=false`.
