# SQL Queries

Store sqlc query files for the module here.

- One file per concern (e.g., `list_accounts.sql`, `create_account.sql`).
- Prefer deterministic ordering and explicit column lists.
- Avoid `SELECT *` and always filter by `ledger_id` when applicable.
