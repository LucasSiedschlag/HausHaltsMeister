# Issue: Ledger state leak across users (frontend cache)

## Context
When logging in as a second user in the same browser session, the UI showed ledgers/accounts from the previous user. This violates the ledger boundary rules in:
- `docs/ledger/Regras_Core_Ledger.md`
- `docs/ledger/Regras_Seguranca.md`
- `docs/agent/DECISIONS.md`

## Impact
- User B can see User A's ledgers and data in the frontend UI (even if the backend would deny access).
- High severity because it breaks the perception of isolation and can lead to wrong edits.

## Root cause
Frontend state was cached globally and **not cleared** on logout or user switch:
- `useLedgerContext.loadLedgers()` returns cached `ledgers_list` if non-empty.
- `useAccounts`, `useCategories`, `useJournal`, `useBudget` store data in `useState` without user/ledger scoping.
- `hhm_ledger_id` cookie persisted across users and could re-select an old ledger.

## Fix (implemented)
- Added a user-change watcher in `frontend/app/plugins/auth-bootstrap.client.ts` that resets **all user/ledger scoped `useState` keys** and clears `hhm_ledger_id`.
- This forces fresh `/ledgers` and ledger-scoped fetches for the new user.

## Follow-ups
- Add backend tests for ledger access denial (handler + service + integration).
- Consider scoping state by `user_id`/`ledger_id` to prevent future leakage.
- Confirm with network traces that `/ledgers` returns only the current user's ledgers.

## Status
In progress (state reset added; remaining validation + tests pending).

---

# Issue: Ledger member invite validation + member list visibility

## Context
Invite flow for ledger members accepts emails with uppercase characters without validation feedback, and member list for a shared ledger shows as empty for the invited user. This breaks the shared-ledger experience and roles display.

## Reported behavior
- Invite input accepted uppercase email without any validation error.
- On invited user login, the shared ledger appears but the members section shows “Nenhum membro ainda”.
- The invited user cannot view or edit anything inside the shared ledger.

## Expected behavior
- Invite input validates email format consistently (case-insensitive acceptable, but should still validate).
- Members list should include the owner and invited members.
- Invited user should be able to access ledger data according to their role.

## Suspected areas
- Frontend invite validator (missing or not enforcing email format).
- Ledger members list endpoint or membership role fetch for non-owner.
- Ledger guard or membership query for invited users.

## Status
In progress.

## Changes applied
- Frontend: invite email validation added (inline errors + trim/lowercase before submit).
- Backend: members list now allows viewer access (GET /ledgers/:ledgerId/members).
