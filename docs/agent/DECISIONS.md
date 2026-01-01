# Project Decisions (Immutable Baselines)

This document captures a small set of foundational decisions that are treated as immutable unless explicitly revised. These are non-negotiable rules that shape every layer (schema, services, API, and UI).

---

## 1) Ledger boundary is mandatory

**Decision**
All data access must be scoped by `ledger_id`. There are no cross-ledger reads or writes.

**Implications**
- Every query must include `ledger_id`.
- Fetch-by-id must validate ownership: `(ledger_id, id)`.
- No endpoint should accept only `:id` without ledger context.

---

## 2) Transactions pagination is cursor-based

**Decision**
Transaction listing uses cursor pagination with stable ordering: `occurred_at DESC, id DESC`.

**Implementation rule**
Pagination must be done in **two steps**:
1. CTE selecting only the page of transaction IDs.
2. Fetch those IDs with joined children (entries) and aggregate.

**Why**
Offset pagination causes duplicated/missing rows under concurrent inserts and creates “ghost” entries when joining children.

---

## 3) Refresh tokens are stored in HttpOnly cookies

**Decision**
Use short-lived access tokens + long-lived refresh tokens. Refresh tokens are stored in HttpOnly + Secure cookies (for web).

**Implications**
- `/auth/refresh` must rotate refresh tokens.
- `/auth/logout` must revoke refresh tokens (session-based).
- Refresh tokens must not be stored in localStorage.

---

## 4) Credit card model is plan-based

**Decision**
Card purchases are **installment plans**. A purchase creates `installment_plans` + `installments`.

**Budget rule**
The budget is consumed only when installments are **posted** (monthly). Payment is a technical transfer and never counts as spend.

---

## 5) Budget is percentage of income base

**Decision**
Budget is calculated as percentages over the **income base** of the month.

**Income base**
Sum of IN categories where `is_budget_base=true`.

**Consumption**
OUT categories consume budget only when `is_budget_relevant=true`.

---

## 6) Edit/Delete policy for the journal

**Decision (MVP)**
Allow edits but protect auditability and card integrity.

**Rules**
- Editing transactions is allowed, but must preserve ledger/account/category consistency.
- Deletion must be blocked if the transaction is referenced by:
  - `installments.posted_transaction_id`
  - `credit_card_statements.payment_transaction_id`
- When in doubt, use `kind=adjust` to correct history.

**Future option**
Move to append-only policy with adjustments instead of edits.
