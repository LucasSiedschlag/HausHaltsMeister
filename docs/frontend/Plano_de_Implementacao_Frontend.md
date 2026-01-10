# Plano_de_Implementacao_Frontend

Este plano define as etapas detalhadas para implementar o frontend em Nuxt 4 com SSR e Nuxt Layers, alinhado ao backend ledger-based.

## Etapa 0 — Fundacao do projeto

Objetivo: criar a base tecnica do frontend com SSR e layers.

Passos:
1. Criar `frontend/` com Nuxt 4 (SSR habilitado por padrao).
2. Configurar `nuxt.config.ts` raiz com `extends` dos layers base.
3. Criar `layers/shared` com Tailwind + shadcn-vue e tokens base (tema azul).
4. Configurar `components.json` do shadcn-vue apontando para `layers/shared`.
5. Definir `layers/core` com layout base e shell de navegacao.
6. Adicionar `@nuxtjs/color-mode` com persistencia em cookie (SSR-friendly).
7. Definir `useApiClient` em `layers/shared` com proxy `/api`.
8. Configurar i18n com `no_prefix` e pt-BR/en-US.
9. Integrar Notivue no app shell (tema sincronizado).

Saidas esperadas:
- Estrutura de layers criada.
- Layout base renderizando com SSR.
- Tokens de tema aplicados em modo claro/escuro.
- i18n ativo sem prefixo em URL.

Status: concluida.

---

## Etapa 1 — Auth e sessao

Objetivo: login e signup com refresh token via cookie HttpOnly.

Referencias:
- `docs/api/01_Autenticacao.md`
- `docs/api/00_Convencoes.md`

Passos:
1. Criar `layers/auth` com rotas `/auth/login` e `/auth/signup`.
2. Implementar forms com validacao client-side (senha minima, email).
3. Chamar `POST /auth/login` e `POST /auth/signup`.
4. Salvar access token somente em memoria (store in-memory).
5. Implementar `GET /auth/me` e refresh via `/auth/refresh`.
6. Criar middleware global `auth` em `layers/core` (redireciona anonimos).
7. Implementar OAuth (Google/GitHub) + callback.
8. Implementar expiracao de sessao com aviso no UI.
9. Implementar rotas auxiliares: forgot password, logout, sessions.

Saidas esperadas:
- Sessao funciona em SSR e client.
- Refresh token sempre via cookie.
- Fluxo OAuth funcionando.
- Feedback visual em erros de auth.

Status: concluida.

---

## Etapa 1.1 — Configuracoes do usuario

Objetivo: centralizar preferencias do usuario e aplicar no UI.

Referencias:
- `docs/api/01_Autenticacao.md`
- `docs/agent/CONVENTIONS.md`

Passos:
1. Implementar pagina `/settings` no `layers/core`.
2. Criar `usePreferences` com `GET/PUT /me/preferences`.
3. Aplicar `theme_mode`, `theme_palette`, `theme_tone` e densidade no layout.
4. Trocar idioma via `setLocale()` apos salvar preferencia.
5. Adicionar notificacoes (Notivue) e estados de erro.
6. Listar sessoes e permitir logout de dispositivos.

Saidas esperadas:
- Preferencias persistem no banco e refletem no UI.
- Idioma aplicado sem prefixo de URL.
- Notificacoes padronizadas.

Status: concluida.

---

## Etapa 2 — Selecao de ledger e shell

Objetivo: selecionar ledger ativo e montar shell principal.

Referencias:
- `docs/api/02_Ledgers.md`

Passos:
1. Criar `layers/ledgers` com pagina de selecao.
2. Consumir `GET /ledgers` e armazenar `ledgerId` atual.
3. Adicionar switcher de ledger no header do `core`.
4. Garantir `ledgerId` persistido (cookie SSR-friendly).
5. Adicionar middleware de ledger para bloquear rotas sem selecao.

Saidas esperadas:
- Usuario sempre tem um ledger ativo.
- Rotas dependem de ledger selecionado.

Status: concluida.

---

## Etapa 3 — Accounts

Referencias:
- `docs/api/03_Accounts.md`

Passos:
1. CRUD de contas.
2. Tela de lista com filtros por status.
3. Form de criacao/edicao.

Saidas esperadas:
- Lista de contas com filtro ativo/inativo.
- Criacao, edicao e desativacao via drawer.
- Guards de role (viewer apenas leitura).

Status: concluida.

---

## Etapa 4 — Categories

Referencias:
- `docs/api/04_Categories.md`

Passos:
1. CRUD de categorias.
2. Respeitar regras `direction` e flags de budget.
3. Exibir hierarquia (parent_id) com indentacao.

Saidas esperadas:
- Lista hierarquica com filtros por direcao e status.
- Criacao e edicao com flags de orcamento.
- Desativacao com confirmacao centralizada.

Status: concluida.

---

## Etapa 5 — Journal

Referencias:
- `docs/api/05_Journal.md`

Passos:
1. Listagem com cursor e filtros.
2. Detalhe de transacao (entries agrupadas).
3. Criacao e edicao de transacoes.
4. Bloquear delete/edicao se referenciada.

Saidas esperadas:
- Lista compacta com filtros completos e paginação por cursor.
- Sheet para criar/editar com entries em blocos reutilizaveis.
- Confirmacao centralizada para exclusao.

Status: concluida.

---

## Etapa 6 — Budget

Referencias:
- `docs/api/06_Budget.md`

Passos:
1. Tela de plano e versoes.
2. Editor de linhas por categoria.
3. Painel mensal e painel por periodo.
4. Indicadores de variancia (orcado vs realizado).

---

## Etapa 7 — Investments

Referencias:
- `docs/api/07_Investimentos.md`

Passos:
1. Forms para aporte, resgate e rendimento.
2. Resumo agregado por periodo.

Status: concluida.

---

## Etapa 8 — Cartao

Referencias:
- `docs/api/08_Cartao.md`

Passos:
1. Cadastro de cartoes e bandeiras.
2. Criacao de planos de parcelas.
3. Lista de parcelas por mes.
4. Posting mensal com idempotency key.
5. Gestao de faturas (close, pay).

---

## Etapa 9 — Relatorios

Referencias:
- `docs/api/09_Relatorios.md`

Passos:
1. Relatorio de saldos.
2. Relatorio por categoria.
3. Fluxo de caixa.

---

## Etapa 10 — QA e consistencia

Objetivo: garantir qualidade e padroes.

Passos:
1. Testes E2E basicos para login e ledger selection.
2. Testes de pagina com SSR (render sem window).
3. Validar erros e estados vazios.
4. Alinhar UI com tokens e paleta azul.

---

## Checklist por feature

1. Confirmar contrato em `docs/api/<modulo>.md`.
2. Implementar composables de API no layer do modulo.
3. Implementar page + componentes.
4. Validar SSR (sem window).
5. Estados de loading, erro e empty.
6. Atualizar docs se houver nova decisao.
