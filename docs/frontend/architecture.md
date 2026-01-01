# Arquitetura do Frontend

Este documento descreve a arquitetura do frontend baseada em Nuxt Layers com **SSR como padrao**, alinhada aos modulos do backend (ledger, journal, budget, cartao, investimentos e relatorios).

## Estrutura de pastas (proposta)

```
frontend/
├── app/                    # App shell opcional
├── layers/
│   ├── shared/             # Base UI, tokens, utils
│   │   ├── assets/         # CSS global (Tailwind + tokens)
│   │   ├── components/     # UI base e wrappers
│   │   │   └── ui/          # shadcn-vue
│   │   ├── composables/    # helpers genericos
│   │   ├── utils/          # formatters, cn.ts, money/date
│   │   └── nuxt.config.ts
│   ├── core/               # Layouts, navegacao, shell
│   │   ├── layouts/        # default.vue
│   │   ├── components/     # header/sidebar
│   │   ├── middleware/     # auth gate
│   │   ├── plugins/        # setup global
│   │   └── nuxt.config.ts
│   ├── auth/               # login, signup, session
│   ├── ledgers/            # selecao/gestao de ledger
│   ├── accounts/           # contas
│   ├── categories/         # categorias
│   ├── journal/            # transacoes
│   ├── budget/             # planejamento e paines
│   ├── investments/        # aportes/resgates
│   ├── creditcard/         # cartoes, planos, faturas
│   └── reports/            # relatorios
├── public/
├── nuxt.config.ts
├── tailwind.config.ts
└── components.json         # shadcn-vue
```

## Responsabilidades por layer

### shared
- UI base (shadcn), tokens de tema, utils de formato (moeda/datas).
- Nao conhece dominio e nao depende de outros layers.

### core
- Shell da aplicacao, layouts, navegacao principal.
- Middleware de auth e selecao de ledger.
- Depende de `shared`.

### features (auth, accounts, journal, budget, etc.)
- Pages e componentes de dominio.
- Composables especificos de API e estado local.
- Nao dependem entre si; comunicacao via `core` ou store compartilhada.

## SSR (regras obrigatorias)

- Todas as paginas sao SSR por default.
- Evitar `window`/`document` fora de hooks client-side.
- Qualquer componente que dependa de APIs do browser usa `<ClientOnly>`.
- Side effects de navegacao (localStorage, viewport) devem ser isolados em `onMounted`.

## Rotas e prefixos

- Cada feature layer deve prefixar suas rotas para evitar colisao.
  - Ex: `layers/journal/pages/journal/index.vue` -> `/journal`.
- Rotas globais (home, landing) ficam em `core`.

## Integracao com API

- HTTP client centralizado em `shared` (ex: `useApiClient`).
- Tokens via cookie/headers conforme docs de API.
- Nunca chamar endpoints sem `ledgerId` quando exigido.
- OAuth segue fluxo Authorization Code + PKCE.

### SSR + cookies (obrigatorio)

- Em SSR, repassar cookies do request para o backend ao usar `fetch`/`useFetch`.
- Preferir `useRequestHeaders(['cookie'])` no server.
- No client, usar `credentials: 'include'` quando necessario.
- Recomendado: proxy `/api` no Nuxt para evitar CORS e simplificar cookies.

### Rotas de auth (UI)

- Usar as rotas reais do backend: `/auth/login`, `/auth/signup`, `/auth/forgot-password`.

## Estado e sincronizacao

- `ledgerId` selecionado no `core` e compartilhado (store/composable).
- `auth` controla sessao e refresh.
- `journal` usa paginacao por cursor e evita offset.

## Extensao de layers

No `nuxt.config.ts` raiz:

```ts
export default defineNuxtConfig({
  extends: [
    './layers/shared',
    './layers/core',
    './layers/auth',
    './layers/ledgers',
    './layers/accounts',
    './layers/categories',
    './layers/journal',
    './layers/budget',
    './layers/investments',
    './layers/creditcard',
    './layers/reports'
  ]
})
```

A ordem define prioridade de override (o ultimo tem maior precedencia).
