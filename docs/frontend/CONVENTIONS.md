# Padrões do Frontend

Este documento consolida as decisões e padrões aplicados no frontend para evitar divergências entre módulos.

## Stack e princípios base

- **Nuxt 4 + SSR** por padrão; nada deve assumir ambiente client-only sem `onMounted` ou `<ClientOnly>`.
- **Nuxt Layers** por domínio, alinhados ao backend (auth, ledger, journal, budget, creditcard, etc.).
- **Sem dependência cruzada entre features**: cada módulo fala com o backend via `shared`/`core`.

## Estrutura e importações

- UI base sempre em `layers/shared/components/ui` (shadcn-vue).
- Utilitários em `layers/shared/utils` (ex.: `cn.ts`).
- Composables compartilhados em `layers/shared/composables`.
- Imports preferenciais: `@shared/...` ou `#layers/<layer>/...` (evitar relativos longos).

## Rotas e auth

- Rotas de auth da UI devem bater com o backend: `/auth/login`, `/auth/signup`, `/auth/forgot-password`.
- Middleware obrigatório: `layers/core/middleware/auth.global.ts`.
- OAuth usa `/auth/oauth/callback` e `useAuth().startOAuth`.

## API e SSR

- Usar `useApiClient` (em `shared`) para todas as chamadas.
- Tokens: access token em header `Authorization`, refresh via cookie HttpOnly.
- Para SSR, repasse cookies via `useRequestHeaders(['cookie'])` quando usar `fetch`.

## Preferências do usuário

- Preferências vivem no backend e são acessadas por `usePreferences`.
- `usePreferences` é a fonte de verdade para tema, idioma e notificações.
- **Mudança de idioma**: só via `setLocale()` após salvar no backend.

## i18n

- Locales oficiais: `pt-BR` (default) e `en-US`.
- Sempre usar `t('...')` para textos; nada hardcoded em componentes.
- Para placeholders com `@`, use interpolacao (ex.: `exemplo{at}dominio.com` + `t(..., { at: '@' })`).

## Tema e UI

- Tokens globais em `layers/shared/assets/css/tailwind.css`.
- Paleta/tom via `data-theme-palette` e `data-theme-tone` no `<html>`.
- `color-mode` usa cookie; respeitar `theme_mode`.
- Componentes shadcn adicionados via `npx shadcn-vue@latest add ...`.

## Formulários e validação

- Validadores em `layers/shared/validators` (Zod).
- Use `useInlineValidation` + `getInputClass` para estilo padrão.
- Mostrar **apenas o primeiro erro** por campo.
- Campos obrigatórios com `*` no label.

## Notificações (Notivue)

- Configuração global em `frontend/app/app.vue`.
- Usar `push.success` / `push.error` nos módulos.
- Tema do Notivue segue preferência de tema do usuário.

## Tipagem e qualidade

- `make typecheck` roda `nuxi typecheck` no frontend.
- Tipos de enum devem usar **union types** (ex.: `'pt-BR' | 'en-US'`).

## Checklist para novas telas

1. Puxar textos via i18n (pt-BR/en-US).
2. Usar UI do shadcn (shared/ui).
3. Validar inputs com Zod + inline validation.
4. Evitar efeitos client-only fora de hooks.
5. Respeitar ledger boundary em todas as chamadas API.
