# API — Autenticacao

Este modulo cobre sessao com access token curto + refresh token, alem de OAuth preparado para Google/GitHub.

## Padroes globais aplicados

- JSON em `snake_case`.
- Datas e timestamps: ISO-8601 em UTC.
- IDs: UUID (string).
- Auth: `Authorization: Bearer <access_token>`.
- Refresh token: cookie HttpOnly.

## Inventario de endpoints

| Metodo | Path | Descricao | Auth | Role |
|---|---|---|---|---|
| POST | /auth/signup | Criar usuario e ledger padrao | Nao | - |
| POST | /auth/login | Login | Nao | - |
| POST | /auth/refresh | Rotacionar sessao | Nao (cookie) | - |
| POST | /auth/logout | Encerrar sessao | Sim | viewer |
| GET | /auth/me | Usuario autenticado | Sim | viewer |
| GET | /auth/oauth/{provider}/start | OAuth start (futuro) | Nao | - |
| GET | /auth/oauth/{provider}/callback | OAuth callback (futuro) | Nao | - |
| POST | /auth/forgot-password | Reset (fase 2) | Nao | - |
| POST | /auth/reset-password | Reset (fase 2) | Nao | - |
| POST | /auth/verify-email | Verificacao (fase 2) | Nao | - |
| POST | /auth/resend-verification | Reenvio (fase 2) | Nao | - |
| GET | /auth/sessions | Listar sessoes (opcional) | Sim | viewer |
| DELETE | /auth/sessions/{sessionId} | Encerrar sessao (opcional) | Sim | viewer |
| POST | /auth/logout-all | Encerrar todas as sessoes (opcional) | Sim | viewer |
| GET | /auth/providers | Providers ativos (opcional) | Sim | viewer |
| POST | /auth/link/{provider}/start | Linkar provider (opcional) | Sim | viewer |
| POST | /auth/unlink/{provider} | Desvincular provider (opcional) | Sim | viewer |
| GET | /me/preferences | Preferencias do usuario | Sim | viewer |
| PUT | /me/preferences | Atualizar preferencias do usuario | Sim | viewer |

---

## Erros globais (5xx)

- 500 `INTERNAL_SERVER_ERROR`
- 503 `SERVICE_UNAVAILABLE`
- 504 `GATEWAY_TIMEOUT`

## Contratos detalhados

### POST /auth/signup
1) Summary / Purpose
- Criar usuario e ledger padrao.

2) Auth & Authorization
- Token: nao.
- Role: n/a.

3) Request
- Path params: n/a.
- Query params: n/a.
- Headers: `Content-Type: application/json`.
- Body:
```json
{
  "email": "user@example.com",
  "password": "min8chars",
  "display_name": "Nome"
}
```

4) Response
- 201
```json
{
  "access_token": "...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "display_name": "Nome"
  }
}
```

5) Errors
- 409 `CONFLICT_DUPLICATE_EMAIL`
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Cria ledger padrao e associa role owner.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### POST /auth/login
1) Summary / Purpose
- Autenticar usuario.

2) Auth & Authorization
- Token: nao.

3) Request
- Body:
```json
{ "email": "user@example.com", "password": "...", "remember": true }
```

4) Response
- 200
```json
{
  "access_token": "...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "display_name": "Nome"
  }
}
```

5) Errors
- 401 `AUTH_INVALID_CREDENTIALS`
- 403 `AUTH_USER_INACTIVE`

6) Semantics / Notes
- Refresh token rotacionavel e criado.
- Campo `remember` controla a duracao da sessao de refresh (curta vs longa).

---

### GET /me/preferences
1) Summary / Purpose
- Retorna preferencias do usuario autenticado.

2) Auth & Authorization
- Token: sim.
- Role: viewer.

3) Request
- Headers: `Authorization: Bearer <token>`.

4) Response
- 200
```json
{
  "user_id": "uuid",
  "theme_mode": "system",
  "theme_palette": "default",
  "theme_tone": "vivid",
  "locale": "pt-BR",
  "compact_mode": "comfortable",
  "font_scale": "md",
  "notify_card_close": true,
  "notify_budget_over": true,
  "notify_payables": true
}
```

5) Errors
- 401 `AUTH_INVALID_CREDENTIALS`
- 500 `INTERNAL_SERVER_ERROR`

---

### PUT /me/preferences
1) Summary / Purpose
- Atualiza preferencias do usuario autenticado.

2) Auth & Authorization
- Token: sim.
- Role: viewer.

3) Request
- Body (campos opcionais):
```json
{
  "theme_mode": "dark",
  "theme_palette": "green",
  "theme_tone": "pastel",
  "locale": "en-US",
  "compact_mode": "compact",
  "font_scale": "lg",
  "notify_card_close": false,
  "notify_budget_over": true,
  "notify_payables": true
}
```

4) Response
- 200 (mesmo formato do GET)

5) Errors
- 401 `AUTH_INVALID_CREDENTIALS`
- 422 `VALIDATION_ERROR`

### GET /auth/oauth/{provider}/start (Frontend)
1) Summary / Purpose
- Inicia o fluxo OAuth e redireciona o usuario.

2) Auth & Authorization
- Token: nao.

3) Request
- Path params: `provider` (google|github).
- Query params: `redirect_uri` (URL do frontend).

4) Response
- 302 Redirect para o provider.

5) Errors
- 501 `NOT_IMPLEMENTED`

6) Semantics / Notes
- Frontend usa `/auth/oauth/callback` como `redirect_uri`.

---

### GET /auth/oauth/{provider}/callback (Frontend)
1) Summary / Purpose
- Finaliza OAuth e cria sessao local.

2) Auth & Authorization
- Token: nao.

3) Request
- Query params: `code`, `state`.

4) Response
- 302 Redirect (quando `redirect_uri` foi informado) ou JSON de sessao.

5) Errors
- 400 `AUTH_OAUTH_STATE_INVALID`
- 504 `GATEWAY_TIMEOUT`

6) Semantics / Notes
- Frontend chama `/auth/refresh` apos o redirect para obter o access token.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### POST /auth/refresh
1) Summary / Purpose
- Rotacionar refresh token e emitir novo access.

2) Auth & Authorization
- Token: refresh em cookie HttpOnly.

3) Request
- Headers: Cookie com refresh.
- Body: vazio.

4) Response
- 200 (mesmo payload do login).

5) Errors
- 401 `AUTH_REFRESH_REVOKED`

6) Semantics / Notes
- Rotacao invalida o refresh anterior.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### POST /auth/logout
1) Summary / Purpose
- Revogar sessao atual.

2) Auth & Authorization
- Token: sim.

3) Request
- Body: vazio.

4) Response
- 204

5) Errors
- 401 `AUTH_INVALID_CREDENTIALS`

6) Semantics / Notes
- Revoga refresh atual.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### GET /auth/me
1) Summary / Purpose
- Retornar perfil basico.

2) Auth & Authorization
- Token: sim.

3) Request
- Path/query: n/a.

4) Response
- 200
```json
{ "id": "uuid", "email": "user@example.com", "display_name": "Nome" }
```

5) Errors
- 401 `AUTH_INVALID_CREDENTIALS`

6) Semantics / Notes
- Pode incluir ledgers no futuro.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### GET /auth/oauth/{provider}/start (futuro)
1) Summary / Purpose
- Inicia OAuth (PKCE).

2) Auth & Authorization
- Token: nao.

3) Request
- Path params: `provider` (google|github).

4) Response
- 302 redirect.

5) Errors
- 502 `AUTH_OAUTH_PROVIDER_ERROR`

6) Semantics / Notes
- Gera state + code_verifier.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### GET /auth/oauth/{provider}/callback (futuro)
1) Summary / Purpose
- Finaliza OAuth e cria sessao.

2) Auth & Authorization
- Token: nao.

3) Request
- Query: `code`, `state`.

4) Response
- 302 redirect + cookie refresh.

5) Errors
- 401 `AUTH_OAUTH_STATE_INVALID`
- 422 `AUTH_OAUTH_EMAIL_REQUIRED`

6) Semantics / Notes
- Cria/vincula `auth_identities`.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### POST /auth/forgot-password (fase 2)
1) Summary / Purpose
- Solicitar reset de senha.

2) Auth & Authorization
- Token: nao.

3) Request
- Body:
```json
{ "email": "user@example.com" }
```

4) Response
- 204.

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Resposta generica (nao vazar existencia).

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### POST /auth/reset-password (fase 2)
1) Summary / Purpose
- Definir nova senha com token.

2) Auth & Authorization
- Token: nao.

3) Request
- Body:
```json
{ "token": "...", "new_password": "..." }
```

4) Response
- 204.

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Token de uso unico.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### POST /auth/verify-email (fase 2)
1) Summary / Purpose
- Confirmar email com token.

2) Auth & Authorization
- Token: nao.

3) Request
- Body:
```json
{ "token": "..." }
```

4) Response
- 204.

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Atualiza `email_verified_at`.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### POST /auth/resend-verification (fase 2)
1) Summary / Purpose
- Reenviar verificacao de email.

2) Auth & Authorization
- Token: nao.

3) Request
- Body:
```json
{ "email": "user@example.com" }
```

4) Response
- 204.

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Rate limit aplicado.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### GET /auth/sessions (opcional)
1) Summary / Purpose
- Listar sessoes ativas do usuario.

2) Auth & Authorization
- Token: sim.

3) Request
- Query: n/a.

4) Response
- 200
```json
[ { "id": "uuid", "created_at": "...", "expires_at": "...", "user_agent": "..." } ]
```

5) Errors
- 401 `AUTH_INVALID_CREDENTIALS`

6) Semantics / Notes
- Nao expor refresh.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### DELETE /auth/sessions/{sessionId} (opcional)
1) Summary / Purpose
- Encerrar sessao especifica.

2) Auth & Authorization
- Token: sim.

3) Request
- Path: `sessionId`.

4) Response
- 204.

5) Errors
- 401 `AUTH_INVALID_CREDENTIALS`

6) Semantics / Notes
- Revoga sessao.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### POST /auth/logout-all (opcional)
1) Summary / Purpose
- Encerrar todas as sessoes.

2) Auth & Authorization
- Token: sim.

3) Request
- Body: vazio.

4) Response
- 204.

5) Errors
- 401 `AUTH_INVALID_CREDENTIALS`

6) Semantics / Notes
- Revoga todas as sessoes do usuario.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### GET /auth/providers (opcional)
1) Summary / Purpose
- Listar providers ativos.

2) Auth & Authorization
- Token: sim.

3) Request
- n/a.

4) Response
- 200
```json
[ { "provider": "google", "active": true } ]
```

5) Errors
- 401 `AUTH_INVALID_CREDENTIALS`

6) Semantics / Notes
- Util para UI.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### POST /auth/link/{provider}/start (opcional)
1) Summary / Purpose
- Iniciar fluxo de link.

2) Auth & Authorization
- Token: sim.

3) Request
- Path: `provider`.

4) Response
- 302 redirect.

5) Errors
- 502 `AUTH_OAUTH_PROVIDER_ERROR`

6) Semantics / Notes
- Usa OAuth state dedicado ao linking.

7) Pagination
- n/a.

8) Idempotency
- n/a.

---

### POST /auth/unlink/{provider} (opcional)
1) Summary / Purpose
- Desvincular provider.

2) Auth & Authorization
- Token: sim.

3) Request
- Path: `provider`.

4) Response
- 204.

5) Errors
- 422 `VALIDATION_ERROR`

6) Semantics / Notes
- Nao permitir desvincular se for unico metodo.

7) Pagination
- n/a.

8) Idempotency
- n/a.
