-- Auth: users

-- name: CreateUser :one
INSERT INTO users (email, display_name, avatar_url, email_verified_at, is_active, updated_at)
VALUES ($1, $2, $3, $4, COALESCE($5, true), $6)
RETURNING id, email, COALESCE(display_name, ''), COALESCE(avatar_url, ''), email_verified_at, is_active, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT id, email, COALESCE(display_name, ''), COALESCE(avatar_url, ''), email_verified_at, is_active, created_at, updated_at
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, COALESCE(display_name, ''), COALESCE(avatar_url, ''), email_verified_at, is_active, created_at, updated_at
FROM users
WHERE id = $1;

-- name: MarkUserEmailVerified :exec
UPDATE users
SET email_verified_at = now(), updated_at = now()
WHERE id = $1 AND email_verified_at IS NULL;

-- Auth: secrets

-- name: CreateAuthSecret :exec
INSERT INTO auth_secrets (user_id, password_hash, updated_at)
VALUES ($1, $2, $3);

-- name: UpdateAuthSecret :exec
UPDATE auth_secrets
SET password_hash = $2, updated_at = $3
WHERE user_id = $1;

-- name: GetAuthSecretHash :one
SELECT password_hash
FROM auth_secrets
WHERE user_id = $1;

-- Auth: identities

-- name: CreateAuthIdentity :one
INSERT INTO auth_identities (user_id, provider, provider_user_id, email, display_name, avatar_url, email_verified, last_login_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, now(), $8)
RETURNING id, user_id, provider, provider_user_id, COALESCE(email, ''), COALESCE(display_name, ''), COALESCE(avatar_url, ''), email_verified, last_login_at;

-- name: GetAuthIdentityByProvider :one
SELECT id, user_id, provider, provider_user_id, COALESCE(email, ''), COALESCE(display_name, ''), COALESCE(avatar_url, ''), email_verified, last_login_at
FROM auth_identities
WHERE provider = $1 AND provider_user_id = $2;

-- name: UpdateAuthIdentityLogin :exec
UPDATE auth_identities
SET last_login_at = now(), updated_at = now()
WHERE user_id = $1 AND provider = $2;

-- Auth: sessions

-- name: CreateAuthSession :one
INSERT INTO auth_sessions (user_id, refresh_token_hash, is_persistent, expires_at, user_agent, ip, device_name, rotated_from_session_id, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, user_id, expires_at, revoked_at, is_persistent;

-- name: GetAuthSessionByRefreshHash :one
SELECT id, user_id, expires_at, revoked_at, is_persistent
FROM auth_sessions
WHERE refresh_token_hash = $1;

-- name: LockAuthSessionByRefreshHash :one
SELECT id, user_id, expires_at, revoked_at, is_persistent
FROM auth_sessions
WHERE refresh_token_hash = $1
FOR UPDATE;

-- name: RevokeAuthSession :exec
UPDATE auth_sessions
SET revoked_at = now(), updated_at = now()
WHERE refresh_token_hash = $1;

-- name: RevokeAuthSessionByID :exec
UPDATE auth_sessions
SET revoked_at = now(), updated_at = now()
WHERE id = $1;

-- name: RevokeAllAuthSessionsForUser :exec
UPDATE auth_sessions
SET revoked_at = now(), updated_at = now()
WHERE user_id = $1 AND revoked_at IS NULL;

-- name: ListAuthSessionsByUser :many
SELECT id, user_id, refresh_token_hash, expires_at, revoked_at, is_persistent, user_agent, ip, device_name, created_at, updated_at, rotated_from_session_id
FROM auth_sessions
WHERE user_id = $1
ORDER BY created_at DESC;

-- Auth: oauth states

-- name: CreateOAuthState :one
INSERT INTO oauth_states (provider, state, code_verifier, redirect_uri, expires_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, provider, state, code_verifier, redirect_uri, expires_at, used_at;

-- name: GetOAuthState :one
SELECT id, provider, state, code_verifier, redirect_uri, expires_at, used_at
FROM oauth_states
WHERE state = $1;

-- name: MarkOAuthStateUsed :exec
UPDATE oauth_states
SET used_at = now(), updated_at = now()
WHERE id = $1 AND used_at IS NULL;

-- Ledgers: default ledger creation on signup

-- name: CreateLedgerForUser :one
INSERT INTO ledgers (owner_user_id, name, currency_code)
VALUES ($1, $2, $3)
RETURNING id;

-- name: AddOwnerMember :exec
INSERT INTO ledger_members (ledger_id, user_id, role, updated_at)
VALUES ($1, $2, 'owner', $3);
