# Frontend — HausHaltsMeister

Este diretório define o frontend em Nuxt 4 com **SSR habilitado** e arquitetura modular (Nuxt Layers) alinhada ao modelo de ledger do backend.

## Visao geral

O frontend sera organizado por dominios equivalentes ao backend:

- auth (sessao e OAuth)
- ledgers (selecao e administracao)
- accounts, categories
- journal (transacoes)
- budget (planejamento e paines)
- investments
- creditcard (cartoes, planos, faturas)
- reports (relatorios agregados)

A UI usa shadcn-vue + Tailwind e tokens centralizados no layer `shared`.

## SSR como padrao

- Todas as paginas rodam em SSR por default.
- Componentes client-only devem ser explicitos (`<ClientOnly>`).
- Evitar acesso a `window`/`document` fora de hooks client-side.

## Rotas de auth (UI)

- A UI usa as mesmas rotas do backend:
  - `/auth/login`
  - `/auth/signup`
  - `/auth/forgot-password`

## Estrutura de layers

- `layers/shared`: UI base, tokens, utils.
- `layers/core`: shell, layouts, navegacao.
- `layers/auth`: login, signup, sessao.
- `layers/ledgers`: selecao e membros.
- `layers/accounts`: contas.
- `layers/categories`: categorias.
- `layers/journal`: transacoes e entradas.
- `layers/budget`: plano e dashboards.
- `layers/investments`: aportes e resumo.
- `layers/creditcard`: cartoes, parcelas, faturas.
- `layers/reports`: relatorios.

## Principios de integracao

- Toda chamada de API respeita `ledgerId` quando exigido.
- Paginacao do journal e por cursor (`cursor_occurred_at`, `cursor_id`).
- Sem dependencia cruzada entre features: comunicacao via `core`.
- Padroes de payload seguem `docs/api/`.

## Design system

- Componentes shadcn-vue ficam em `layers/shared/components/ui`.
- Tokens e CSS global ficam em `layers/shared/assets`.
- `components.json` aponta para o layer `shared`.

## Validacao inline

- Validadores padrao ficam em `layers/shared/validators` e usam Zod.
- Use `emailSchema`, `passwordSchema`, `nameSchema` e `useInlineValidation`.
- Use `getInputClass` para manter o estilo de foco (outline solido + ring com opacidade 0,2).
- Cada campo exibe apenas o primeiro erro (ordem definida no schema).
- Erros de backend permanecem separados do erro de validacao local.

## Eventos de analytics

- Use `useAnalytics().track(event, payload)` para emitir eventos no frontend.
- Eventos sao disparados via `window` com `CustomEvent` (`hhm:analytics`).
- Fluxos de auth ja emitem: `auth.login`, `auth.signup`, `auth.oauth.start`, `auth.refresh`.

## Referencias

Consulte `docs/frontend/references.md` para links de Nuxt Layers, shadcn-vue, OAuth e cookies.
