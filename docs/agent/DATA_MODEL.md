# Data Model (Source of Truth)

This document summarizes the database model, key invariants, and seeds. The full DBML lives in `docs/ledger/Reestruturação_Completa.md`.

---

## High-level diagram (ASCII)

```
users ──┐
        ├─ auth_secrets
        ├─ auth_identities
        ├─ auth_sessions
        ├─ user_preferences
        └─ ledgers ── ledger_members
                ├─ accounts ── credit_cards ── credit_card_statements
                │                 └─ installment_plans ── installments
                ├─ categories
                ├─ transactions ── entries
                └─ budget_plans ── budget_plan_versions ── budget_plan_lines

card_networks ── credit_cards.network

oauth_states (standalone, short-lived)
```

---

## Table overview (condensed)

**Identity & Auth**
- `users`: identity profile (email, display_name, avatar_url, email_verified_at, is_active).
- `auth_secrets`: password hash for local login.
- `auth_identities`: provider bindings (password/google/github), unique per provider user.
- `user_preferences`: user-level settings (theme, locale, density, notifications).
- `auth_sessions`: refresh token sessions (revocation + rotation).
- `oauth_states`: PKCE state store for OAuth (short-lived).

**Ledger core**
- `ledgers`: data boundary (owner + currency).
- `ledger_members`: roles per ledger.
- `accounts`: internal accounts (cash/investment/credit_card).
- `categories`: IN/OUT semantics + budget flags.
- `transactions`: event header.
- `entries`: transaction lines (amount + account + category + kind).

**Budget**
- `budget_plans`: one plan per ledger.
- `budget_plan_versions`: versioning by `effective_from_month`.
- `budget_plan_lines`: category percentages and include_children.

**Credit card**
- `card_networks`: catalog of card brands.
- `credit_cards`: card metadata (linked to account).
- `installment_plans`: purchase plans.
- `installments`: monthly items.
- `credit_card_statements`: monthly statements.

---

## Invariants (must always hold)

**Ledger boundary**
- Every table with `ledger_id` must be queried by `ledger_id`.
- Cross-ledger references are invalid (account/category/transaction/entry must share ledger).

**Journal semantics**
- `entries.amount_cents` is always positive.
- IN/OUT is defined only by `categories.direction`.
- `entries.kind` is `normal`, `transfer`, or `adjust` (validated via CHECK).

**Transfers**
- For `kind=transfer`, totals must balance: SUM(OUT) == SUM(IN).
- At least one IN and one OUT entry per transfer transaction.

**Budget**
- `effective_from_month` and all budget/card “month” fields are first day of month (`YYYY-MM-01`).
- Income base: IN categories with `is_budget_base=true`.
- Budget consumption: OUT categories with `is_budget_relevant=true`.

**Credit card**
- Purchase creates `installment_plans` + `installments`.
- Budget is consumed only when installments are posted.
- Statement payment is a technical transfer (does not count as spend).

**Auth**
- `auth_identities` is unique by `(provider, provider_user_id)`.
- Sessions are revocable and rotated via `auth_sessions`.

---

## Seeds (technical categories)

Seeded per ledger (from `migrations/002_seed_technical_categories.sql`):

**Cartão**
- Pagamento Fatura Cartão (OUT, `is_budget_relevant=false`)
- Entrada Pagamento Cartão (IN, `is_budget_base=false`)

**Investimentos**
- Aportes Investimentos (OUT, `is_budget_relevant=false` por padrão)
- Entrada Investimentos (Aporte) (IN, `is_budget_base=false`)
- Resgate Investimentos (OUT, `is_budget_relevant=false`)
- Entrada Resgate (Investimentos) (IN, `is_budget_base=false`)
- Rendimentos (IN, `is_budget_base=false`)
