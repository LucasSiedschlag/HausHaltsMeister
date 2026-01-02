# Security (Objective Summary)

## Tokens

- Access token TTL: 15 minutes (default).
- Refresh token TTL: 30–90 days (persistente).
- Refresh token TTL (sessao curta): 7 dias (default).
- Refresh tokens are rotated on every `/auth/refresh`.

## Cookies

- Refresh token stored in cookie:
  - `HttpOnly`: true
  - `Secure`: true
  - `SameSite`: `None` (when frontend is on another domain) or `Lax` (same-site)

## OAuth redirect

- Validar `redirect_uri` via allowlist (`OAUTH_REDIRECT_ALLOWLIST`).
- Suporte por ambiente: `OAUTH_REDIRECT_ALLOWLIST_{ENV}` substitui o valor base.

## Rate limiting

Apply rate limits and backoff to:
- `/auth/login`
- `/auth/signup`
- `/auth/refresh`
- `/auth/forgot-password`

Responses must be generic for login failures.

## Session policy

- `/auth/logout` revokes the current session.
- `/auth/logout-all` revokes all sessions for the user.
- Sessions are stored in `auth_sessions` with `revoked_at`.

## RBAC by ledger

- Roles: `owner`, `editor`, `viewer`.
- Access checks must happen before any data fetch.

## Anti-leak rule

- Every query must filter by `ledger_id`.
- Fetch-by-id must validate `(ledger_id, id)`.
- Never expose endpoints without ledger context.
