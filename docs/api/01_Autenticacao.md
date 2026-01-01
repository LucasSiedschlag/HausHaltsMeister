# API — Autenticacao

Este modulo descreve o fluxo de sessao com access token curto + refresh token, alinhado ao modelo de seguranca do projeto.

## Estrategia de tokens (MVP+)

- **Access token (JWT)**: curta duracao (ex.: 15 min).
- **Refresh token (opaco)**: longa duracao (ex.: 30–90 dias), rotacionavel.
- **Refresh em cookie HttpOnly** (recomendado para web).
- **Nao** armazenar refresh token em localStorage.

## Endpoints recomendados

Base:
- `POST /auth/signup`
- `POST /auth/login`

Sessao:
- `POST /auth/refresh`
- `POST /auth/logout`
- `GET /auth/me`

OAuth (preparado):
- `GET /auth/oauth/{provider}/start`
- `GET /auth/oauth/{provider}/callback`

Recuperacao (fase 2, mas previsto):
- `POST /auth/forgot-password`
- `POST /auth/reset-password`

Opcional (quando houver email):
- `POST /auth/verify-email`
- `POST /auth/resend-verification`

Sessao multi-dispositivo (opcional):
- `GET /auth/sessions`
- `DELETE /auth/sessions/{sessionId}`
- `POST /auth/logout-all`

Vinculo de contas (opcional):
- `GET /auth/providers`
- `POST /auth/link/{provider}/start`
- `POST /auth/unlink/{provider}`

---

## Contratos

### POST /auth/signup

Body:
```json
{
  "email": "user@example.com",
  "password": "...",
  "display_name": "Nome"
}
```

Regras:
- normalizar email (lowercase + trim).
- validar forca minima de senha.
- criar ledger padrao apos signup.

### POST /auth/login

Body:
```json
{ "email": "user@example.com", "password": "..." }
```

Response:
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

### POST /auth/refresh

- Se o refresh token estiver em cookie HttpOnly, o body pode ser vazio.
- Retorna um novo access token (mesmo formato do login).

### POST /auth/logout

- Revoga a sessao atual (refresh token).
- Se houver cookie, deve ser invalidado.

### GET /auth/me

Retorna dados minimos do usuario autenticado:
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "display_name": "Nome"
}
```

### POST /auth/forgot-password (fase 2)

Body:
```json
{ "email": "user@example.com" }
```

### POST /auth/reset-password (fase 2)

Body:
```json
{ "token": "...", "new_password": "..." }
```

### GET /auth/oauth/{provider}/start

- Gera `state` + `code_verifier` (PKCE) e redireciona para o provider.

### GET /auth/oauth/{provider}/callback

- Valida `state` e troca `code` por tokens no provider.
- Cria/vincula `auth_identities` e inicia `auth_session`.

---

## Tabela de sessoes (refresh tokens)

Para logout real e multi-dispositivo, usar uma tabela de sessoes:

`auth_sessions` (sugestao):
- `id` (uuid)
- `user_id`
- `refresh_token_hash`
- `created_at`, `expires_at`, `revoked_at`
- `user_agent`, `ip`, `device_name` (opcional)
- `rotated_from_session_id` (opcional)

## Regras de seguranca

- Respostas de login sempre genericas: “Credenciais invalidas”.
- Rate limit em `/auth/login`, `/auth/signup`, `/auth/refresh`, `/auth/forgot-password`.
- JWT deve conter apenas o basico: `sub`, `exp`, `iat`, `jti` (e `sid` opcional).
- Roles e dados mutaveis nao devem ficar no JWT; sempre validar no banco.
