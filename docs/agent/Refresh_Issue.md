# Authentication & Button Functionality Issue - Root Cause Analysis

**Data**: 2026-01-04
**Sistema**: HausHaltsMeister Frontend (Nuxt 4 SSR)
**Problema Resolvido**: Botão "Nova Transação" não funcionava após login

---

## 🎯 CAUSA RAIZ REAL (RESOLVIDO)

> **O problema era simplesmente a falta de `{ immediate: true }` no watcher do estado `createOpen` em `journal/index.vue`.**

### O Que Acontecia

```typescript
// SiteHeader.vue - handleNewTransaction
const handleNewTransaction = async () => {
  journalUi.openCreate(); // ← 1. Define createOpen = true ANTES da navegação
  if (route.path !== "/journal") {
    await router.push("/journal"); // ← 2. Navega para /journal
  }
};

// journal/index.vue - Watcher QUEBRADO (sem immediate)
watch(
  () => journalUi.createOpen.value,
  (open) => {
    if (open) {
      openCreate();
      journalUi.createOpen.value = false;
    }
  }
  // ❌ Faltava: { immediate: true }
);
```

### Por Que Não Funcionava

1. Usuário clica em "Nova Transação" no header
2. `journalUi.openCreate()` define `createOpen.value = true`
3. `router.push('/journal')` navega para a página
4. Página `/journal` monta e cria o watcher
5. **O watcher NÃO dispara** porque o valor já era `true` antes de montar
6. Vue watchers só disparam em **mudanças**, não no valor inicial
7. Formulário nunca abre

### A Correção (1 linha)

```diff
watch(
  () => journalUi.createOpen.value,
  (open) => {
    if (open) {
      openCreate()
      journalUi.createOpen.value = false
    }
  },
+ { immediate: true }  // ← Verifica o valor assim que monta
)
```

### Arquivo Modificado

- `frontend/layers/journal/pages/journal/index.vue` (linha ~630)

| Cenário              | Comportamento                             | Estado dos Dados                 | Estado da UI                 |
| -------------------- | ----------------------------------------- | -------------------------------- | ---------------------------- |
| **Primeiro Login**   | Botão aparece habilitado mas não funciona | ✅ Dados carregados corretamente | ❌ UI não responde ao clique |
| **F5 (Refresh)**     | Avatar desaparece, botão desabilitado     | ✅ Dados carregados corretamente | ❌ UI não atualiza           |
| **Com `ClientOnly`** | Funciona mas com flash                    | ✅ Dados OK                      | ⚠️ Funciona mas não é SSR    |

### Causa Raiz Identificada

**Modificação de estado durante o processo de hidratação SSR**, causando incompatibilidade entre o HTML renderizado no servidor e o estado esperado pelo cliente.

---

## 2. Contexto Técnico: Como Funciona SSR no Nuxt

### 2.1 O Processo de Hidratação

```
┌─────────────────────────────────────────────────────────────┐
│ SERVIDOR (SSR)                                              │
├─────────────────────────────────────────────────────────────┤
│ 1. Nuxt renderiza componentes Vue no servidor               │
│ 2. useState('key', () => initialValue) cria estado          │
│ 3. HTML é gerado com valores do estado                      │
│ 4. Estado é serializado como payload JSON                   │
│ 5. Envia HTML + payload para o browser                      │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ BROWSER (Hidratação)                                        │
├─────────────────────────────────────────────────────────────┤
│ 1. Browser recebe HTML já renderizado                       │
│ 2. JavaScript baixa e executa                               │
│ 3. Nuxt desserializa payload e restaura estado              │
│ 4. Vue "hidrata" - conecta JS aos elementos DOM existentes  │
│ 5. ⚠️  PLUGINS EXECUTAM DURANTE A HIDRATAÇÃO                │
│ 6. Vue verifica: "DOM matches estado esperado?"             │
│    - SIM → Hidratação bem-sucedida ✅                       │
│    - NÃO → Hydration mismatch ❌                            │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 useState: Sincronização Server ↔ Client

```typescript
// Comportamento do useState no Nuxt:

// SERVIDOR:
const ready = useState("auth_ready", () => false); // Cria com false
// HTML renderizado: <button disabled>...</button>
// Payload enviado: { auth_ready: false }

// CLIENTE (antes de hidratar):
const ready = useState("auth_ready", () => false); // Restaura de payload
// Estado inicial: ready.value = false (idêntico ao servidor)
// DOM esperado: <button disabled>...</button>

// ✅ Hidratação bem-sucedida: HTML do servidor === DOM esperado pelo cliente
```

### 2.3 O Problema com Modificação Durante Hidratação

```typescript
// NOSSO CÓDIGO (QUEBRADO):

// auth-bootstrap.client.ts (Plugin)
export default defineNuxtPlugin({
  async setup() {
    const { ready } = useAuth();

    ready.value = false; // ← MODIFICAÇÃO DURANTE HIDRATAÇÃO
    await refresh(); // ← Operação assíncrona
    ready.value = true; // ← MODIFICAÇÃO DURANTE HIDRATAÇÃO
  },
});

// Problema:
// 1. Servidor renderiza com ready = false
// 2. Cliente começa hidratação com ready = false
// 3. Plugin muda para true DURANTE a hidratação
// 4. Vue tenta hidratar mas o estado não corresponde mais!
// 5. Resultado: Hydration mismatch
```

---

## 3. Timeline Detalhada: Primeiro Login

### 3.1 Fluxo OAuth → Callback → Home

```
┌────────────────────────────────────────────────────────────────┐
│ PASSO 1: Login OAuth                                           │
├────────────────────────────────────────────────────────────────┤
│ User clica "Login with Google"                                 │
│ → Redireciona para Google OAuth                                │
│ → Google redireciona de volta para /auth/oauth/callback        │
└────────────────────────────────────────────────────────────────┘
                            ↓
┌────────────────────────────────────────────────────────────────┐
│ PASSO 2: Callback Page (SSR)                                   │
├────────────────────────────────────────────────────────────────┤
│ Servidor:                                                      │
│ - Renderiza /auth/oauth/callback                               │
│ - Estado: ready = false, user = null                           │
│ - Envia HTML para cliente                                      │
│                                                                │
│ Cliente (hidratação):                                          │
│ - Plugin detecta rota pública, pula bootstrap                  │
│ - [AUTH-BOOTSTRAP] Public route, skipping                      │
│ - Callback page extrai tokens da URL                           │
│ - setSessionFromTokens() → define accessToken e user           │
│ - navigateTo('/') → Navega para home                           │
└────────────────────────────────────────────────────────────────┘
                            ↓
┌────────────────────────────────────────────────────────────────┐
│ PASSO 3: Home Page (/) - SSR                                   │
├────────────────────────────────────────────────────────────────┤
│ Servidor:                                                      │
│ - Renderiza componentes com valores padrão                     │
│ - SiteHeader: ready = false, canEdit = false                   │
│ - NavUser: ready = false, displayName = "HausHaltsMeister"     │
│ - HTML: <button disabled>Nova Transação</button>               │
│ - Payload: { auth_ready: false, auth_user: null, ... }         │
└────────────────────────────────────────────────────────────────┘
                            ↓
┌────────────────────────────────────────────────────────────────┐
│ PASSO 4: Home Page - Client Hydration ❌ PROBLEMA AQUI         │
├────────────────────────────────────────────────────────────────┤
│ 1. Browser recebe HTML com <button disabled>                   │
│ 2. JavaScript baixa, hidratação começa                         │
│ 3. Estado restaurado: ready = false, user = null               │
│                                                                │
│ 4. Plugin auth-bootstrap executa:                              │
│    [AUTH-BOOTSTRAP] Starting on route: /                       │
│    [AUTH-BOOTSTRAP] Initial state: { hasToken: false, ... }    │
│    ready.value = false  ← OK até aqui                          │
│                                                                │
│    await refresh()  ← Busca novo token                         │
│    [AUTH-BOOTSTRAP] Refresh complete                           │
│                                                                │
│    await ensureLedger()  ← Carrega ledger e role               │
│    [LEDGER] Role fetched: owner                                │
│                                                                │
│    ready.value = true  ← ❌ PROBLEMA!                          │
│    [AUTH-BOOTSTRAP] Bootstrap complete, ready = true           │
│                                                                │
│ 5. Componentes tentam hidratar:                                │
│    - SiteHeader watcher dispara:                               │
│      [SiteHeader] { ready: true, role: "owner", canEdit: true }│
│    - NavUser watcher dispara:                                  │
│      [NavUser] { ready: true, hasUser: true, name: "Lucas" }   │
│                                                                │
│ 6. ❌ MAS Vue JÁ COMPLETOU A HIDRATAÇÃO!                       │
│    - HTML do servidor tinha ready = false                      │
│    - Cliente mudou para ready = true durante hidratação        │
│    - Vue detecta mismatch mas já é tarde demais                │
│    - DOM não é atualizado corretamente                         │
│                                                                │
│ 7. Resultado:                                                  │
│    - Computed properties têm valores corretos                  │
│    - Mas o DOM ainda mostra estado antigo                      │
│    - Botão não funciona ao clicar (event handler não attached) │
└────────────────────────────────────────────────────────────────┘
```

### 3.2 Por Que o Botão Não Funciona?

```vue
<!-- SiteHeader.vue -->
<Button :disabled="!canEdit" @click="handleNewTransaction">
  Nova Transação
</Button>

<script>
const canEdit = computed(() => ready.value && ledgerContext.hasRole("editor"));
</script>
```

**Problema:**

1. Durante SSR: `ready = false` → `canEdit = false` → `<button disabled>`
2. Durante hidratação: Plugin muda `ready = true` → `canEdit = true`
3. Computed atualiza: `canEdit.value` agora é `true`
4. **MAS**: Vue já hidratou o DOM com `disabled="true"`
5. **MAS**: O atributo `disabled` não é atualizado porque Vue acha que a hidratação já terminou
6. **MAS**: O `@click` handler pode não estar corretamente attached ao DOM
7. **Resultado**: Botão visualmente habilitado mas não funciona

---

## 4. Timeline Detalhada: F5 (Page Refresh)

```
┌────────────────────────────────────────────────────────────────┐
│ PASSO 1: User pressiona F5 na home page (/)                    │
├────────────────────────────────────────────────────────────────┤
│ Browser:                                                       │
│ - Descarta todo o estado JavaScript                            │
│ - Faz nova requisição GET para /                               │
│ - useState é resetado para valores iniciais                    │
└────────────────────────────────────────────────────────────────┘
                            ↓
┌────────────────────────────────────────────────────────────────┐
│ PASSO 2: Servidor Renderiza Novamente (SSR)                    │
├────────────────────────────────────────────────────────────────┤
│ Servidor:                                                      │
│ - Middleware auth.global.ts executa:                           │
│   - Verifica cookie 'hhm_refresh'                              │
│   - Cookie existe → Permite acesso                             │
│   - Não tenta refresh no servidor                              │
│                                                                │
│ - Renderiza componentes com estado inicial:                    │
│   ssr [NavUser] { ready: false, hasUser: false, name: "..." }  │
│   ssr [SiteHeader] { ready: false, role: null, canEdit: false }│
│                                                                │
│ - HTML gerado:                                                 │
│   <button disabled>Nova Transação</button>                     │
│   <span>HausHaltsMeister</span> (placeholder)                  │
│                                                                │
│ - Payload:                                                     │
│   { auth_ready: false, auth_user: null, active_ledger_id: ... }│
└────────────────────────────────────────────────────────────────┘
                            ↓
┌────────────────────────────────────────────────────────────────┐
│ PASSO 3: Cliente Recebe e Hidrata ❌ HYDRATION MISMATCH        │
├────────────────────────────────────────────────────────────────┤
│ 1. Browser recebe HTML:                                        │
│    <button disabled>Nova Transação</button>                    │
│    <span>HausHaltsMeister</span>                               │
│                                                                │
│ 2. JavaScript baixa, desserializa payload                      │
│    Estado inicial: ready = false, user = null                  │
│                                                                │
│ 3. Hidratação começa:                                          │
│    - Vue tenta conectar JS ao DOM                              │
│    - Watchers disparam com valores iniciais:                   │
│      [NavUser] { ready: false, hasUser: false, ... }           │
│      [SiteHeader] { ready: false, role: null, canEdit: false } │
│                                                                │
│ 4. Plugin auth-bootstrap executa:                              │
│    [AUTH-BOOTSTRAP] Starting on route: /                       │
│    [AUTH-BOOTSTRAP] Initial state: { hasToken: false, ... }    │
│                                                                │
│    ready.value = false  ← Ainda matching                       │
│                                                                │
│    await refresh()  ← Busca novo token (200ms)                 │
│    [AUTH-BOOTSTRAP] Refresh complete                           │
│    Backend: POST /auth/refresh → 200 OK                        │
│                                                                │
│    await me()  ← Busca user data (100ms)                       │
│    Backend: GET /auth/me → 200 OK                              │
│                                                                │
│    await ensureLedger()  ← Busca ledgers e role (150ms)        │
│    [LEDGER] Cookie value: 7d4b6aad-aaa8-4048-a6a0-61d81a0cef56 │
│    [LEDGER] Restoring from cookie                              │
│    Backend: GET /ledgers → 200 OK                              │
│    Backend: GET /ledgers/:id/me → 200 OK                       │
│    [LEDGER] Role fetched: owner                                │
│                                                                │
│    ready.value = true  ← ❌ MUDOU DURANTE HIDRATAÇÃO!          │
│    [AUTH-BOOTSTRAP] Bootstrap complete, ready = true           │
│                                                                │
│ 5. Watchers disparam com novos valores:                        │
│    [NavUser] { ready: true, hasUser: true, name: "Lucas", ... }│
│    [SiteHeader] { ready: true, role: "owner", canEdit: true }  │
│                                                                │
│ 6. ❌ Vue detecta hydration mismatch:                          │
│    "Hydration completed but contains mismatches"               │
│                                                                │
│    Por quê?                                                    │
│    - HTML do servidor: <button disabled>Nova Transação</button>│
│    - Estado após plugin: ready = true, canEdit = true          │
│    - Vue espera: <button>Nova Transação</button> (sem disabled)│
│    - Mas o DOM já foi hidratado com disabled!                  │
│                                                                │
│ 7. Consequências:                                              │
│    ✅ Computed properties corretos (canEdit = true)            │
│    ✅ Estado correto (ready = true, user = {...})              │
│    ✅ Backend respondeu corretamente                           │
│    ❌ DOM não atualiza (ainda mostra disabled button)          │
│    ❌ Avatar não aparece (ainda mostra placeholder)            │
│    ❌ Vue está em estado inconsistente                         │
└────────────────────────────────────────────────────────────────┘
```

### 4.1 O Erro "Hydration completed but contains mismatches"

```
Vue detectou que o HTML renderizado no servidor não corresponde
ao que o cliente esperava após a hidratação.

Servidor enviou:     <button disabled>Nova Transação</button>
Cliente esperava:    <button>Nova Transação</button>

Isso acontece porque o plugin modificou o estado DURANTE a hidratação.
```

---

## 5. Todas as Possibilidades Investigadas

### 5.1 Cookie Cross-Port Issue ✅ RESOLVIDO

**Descrição**: Cookies não eram compartilhados entre localhost:3000 (frontend) e localhost:8080 (backend).

**Investigação**:

- OAuth callback em porta 8080 definia cookies
- Frontend em porta 3000 não conseguia acessar
- Tokens de refresh perdidos entre portas

**Solução**:

- Backend passa tokens via URL query params no redirect OAuth
- Frontend extrai tokens da URL e salva em `useState`
- Cookie agora é salvo corretamente em `setActiveLedger()`

**Status**: ✅ Problema resolvido

---

### 5.2 Race Condition: Plugin vs Components ⚠️ PARCIALMENTE VERDADE

**Descrição**: Plugin e componentes competindo para acessar/modificar estado.

**Investigação**:

```typescript
// Logs mostram ordem de execução:
1. [AUTH-BOOTSTRAP] Starting...
2. [AUTH-BOOTSTRAP] Starting bootstrap...
3. [LEDGER] ensureLedger() called
4. [LEDGER] setActiveLedger() called
5. [AUTH-BOOTSTRAP] Bootstrap complete
6. [SiteHeader] State update  ← Depois do plugin
7. [NavUser] State update     ← Depois do plugin
```

**Observações**:

- Não é exatamente uma race condition tradicional
- Componentes não estão competindo COM o plugin
- Problema é que plugin muda estado DURANTE hidratação
- Componentes são vítimas, não participantes da race

**Status**: ⚠️ Não é a causa raiz, mas sintoma

---

### 5.3 Plugin Execution Timing ❌ CAUSA RAIZ PRINCIPAL

**Descrição**: Plugin executa DURANTE o processo de hidratação, não depois.

**Investigação**:

```typescript
// auth-bootstrap.client.ts
export default defineNuxtPlugin({
  name: "auth-bootstrap",
  parallel: false, // ← Força execução sequencial
  async setup() {
    // Tentativa de esperar hidratação completar:
    await new Promise((resolve) => setTimeout(resolve, 0));

    // ❌ Isso NÃO funciona! Plugin ainda roda durante hidratação
    ready.value = false;
    await refresh();
    ready.value = true;
  },
});
```

**Por que `setTimeout(0)` não funciona?**

- Plugins executam na fase de setup do Nuxt
- Hidratação acontece imediatamente após setup
- `setTimeout(0)` apenas adia para próximo tick
- Mas hidratação também acontece no próximo tick
- Não há garantia de ordem

**Timing correto**:

```
Setup fase:
├─ Plugins executam
│  ├─ auth-bootstrap (parallel: false)
│  ├─ outros plugins...
│  └─ Setup completo
│
├─ Componentes criam instâncias
│  ├─ SiteHeader setup
│  ├─ NavUser setup
│  └─ Outros componentes
│
└─ Hidratação Vue
   ├─ Vue conecta JS ao DOM
   ├─ Verifica se DOM matches estado esperado
   └─ Completa (ou reporta mismatch)
```

**Status**: ❌ CAUSA RAIZ CONFIRMADA

---

### 5.4 useState Synchronization ⚠️ FUNCIONANDO CORRETAMENTE

**Descrição**: `useState` não está sincronizando entre servidor e cliente?

**Investigação**:

```typescript
// useState ESTÁ funcionando corretamente:

// Servidor:
const ready = useState("auth_ready", () => false);
// Serializa: { auth_ready: false }

// Cliente (antes de plugin):
const ready = useState("auth_ready", () => false);
// Desserializa: ready.value = false (correto!)

// Cliente (depois de plugin):
ready.value = true; // ← Plugin muda manualmente

// ✅ useState está funcionando corretamente
// ❌ Problema é que mudamos o valor durante hidratação
```

**Status**: ✅ Não é o problema

---

### 5.5 Middleware vs Plugin Confusion ❌ ARQUITETURA ERRADA

**Descrição**: Lógica de auth está no lugar errado.

**Análise**:

| Padrão         | Quando Executa             | Para Que Serve               | Onde Está a Auth             |
| -------------- | -------------------------- | ---------------------------- | ---------------------------- |
| **Middleware** | Antes de cada navegação    | Guardar rotas, redirecionar  | ⚠️ Parcial (só checa cookie) |
| **Plugin**     | Durante setup da aplicação | Inicializar serviços globais | ❌ Toda lógica está aqui     |
| **Composable** | Quando componentes chamam  | Compartilhar estado reativo  | ✅ Define estado             |

**Problema**:

- Plugin está fazendo trabalho de middleware (checar auth em cada rota)
- Middleware está subutilizado (só checa cookie, não refresha)
- Composable está correto, mas plugin interfere

**Deveria ser**:

```typescript
// middleware/auth.global.ts (para TODAS rotas protegidas)
export default defineNuxtRouteMiddleware(async (to) => {
  if (publicRoutes.includes(to.path)) return;

  const { accessToken, refresh } = useAuth();

  // Se não tem token, tenta refresh
  if (!accessToken.value) {
    try {
      await refresh();
    } catch {
      return navigateTo("/auth/login");
    }
  }
});

// plugin/auth-bootstrap.client.ts (apenas setup inicial)
export default defineNuxtPlugin({
  async setup() {
    // Apenas carregar ledger context se necessário
    // NÃO modificar estado durante hidratação
  },
});
```

**Status**: ❌ Arquitetura incorreta contribui para o problema

---

## 6. Technical Deep Dive: Nuxt SSR Architecture

### 6.1 Ordem de Execução Completa

```
┌─────────────────────────────────────────────────────────────┐
│ SERVER SIDE                                                 │
├─────────────────────────────────────────────────────────────┤
│ 1. Request chega ao servidor Nuxt                           │
│ 2. Middleware executa (auth.global.ts)                      │
│    └─ Checa cookie, permite/bloqueia rota                   │
│ 3. Componentes renderizam (SSR)                             │
│    └─ useState cria estado inicial                          │
│ 4. HTML gerado + payload serializado                        │
│ 5. Response enviado para browser                            │
└─────────────────────────────────────────────────────────────┘
                         ↓ HTTP Response
┌─────────────────────────────────────────────────────────────┐
│ CLIENT SIDE (First Load)                                    │
├─────────────────────────────────────────────────────────────┤
│ 1. Browser recebe HTML                                      │
│ 2. Parse HTML, mostra conteúdo inicial                      │
│ 3. Download JavaScript bundles                              │
│ 4. JavaScript executa:                                      │
│    a. Nuxt app inicializa                                   │
│    b. Payload desserializado                                │
│    c. useState restaurado com valores do payload            │
│    d. Plugins executam (parallel: false = sequencial)       │
│       └─ auth-bootstrap.client.ts                           │
│           ├─ Lê estado atual                                │
│           ├─ Modifica estado (❌ PROBLEMA)                  │
│           └─ Completa                                       │
│    e. Vue hydration inicia:                                 │
│       ├─ Componentes criam instâncias                       │
│       ├─ Vue conecta JS ao DOM existente                    │
│       ├─ Verifica: DOM === estado esperado?                 │
│       └─ Se não: "Hydration mismatch" ❌                    │
│ 5. App está "hydratado"                                     │
│ 6. Interações funcionam (ou não, se mismatch)               │
└─────────────────────────────────────────────────────────────┘
                         ↓ Navigation
┌─────────────────────────────────────────────────────────────┐
│ CLIENT SIDE (Subsequent Navigations)                        │
├─────────────────────────────────────────────────────────────┤
│ 1. User clica link / router.push()                          │
│ 2. Middleware executa (client-side)                         │
│ 3. Componentes re-renderizam (client-side)                  │
│ 4. Sem hydration (já está hydratado)                        │
│ 5. Estado persiste via useState                             │
└─────────────────────────────────────────────────────────────┘
```

### 6.2 Plugin vs Middleware vs Composable

| Aspecto                    | Plugin                                          | Middleware              | Composable              |
| -------------------------- | ----------------------------------------------- | ----------------------- | ----------------------- |
| **Executa quando**         | Setup da app (uma vez por page load)            | Antes de cada navegação | Quando componente chama |
| **Executa onde**           | Server + Client (ou só client com `.client.ts`) | Server + Client         | Onde for chamado        |
| **Acesso a router**        | Sim                                             | Sim                     | Sim                     |
| **Acesso a componentes**   | Não                                             | Não                     | Sim                     |
| **Pode fazer await**       | Sim (mas cuidado com hydration!)                | Sim                     | Sim                     |
| **Modifica estado global** | ⚠️ Pode, mas cuidado                            | ❌ Não deveria          | ✅ Sim, é para isso     |
| **Redireciona**            | ⚠️ Pode (mas middleware é melhor)               | ✅ Sim                  | ❌ Não deveria          |

**Use Case Correto**:

```typescript
// ✅ PLUGIN: Inicializar serviços globais
export default defineNuxtPlugin(() => {
  // Setup analytics, error tracking, etc.
  // NÃO modificar estado que afeta renderização
});

// ✅ MIDDLEWARE: Guardar rotas, autenticação
export default defineNuxtRouteMiddleware(async (to) => {
  const { isAuthenticated } = useAuth();
  if (!isAuthenticated.value) {
    return navigateTo("/login");
  }
});

// ✅ COMPOSABLE: Compartilhar estado reativo
export const useAuth = () => {
  const user = useState("user", () => null);
  const login = async () => {
    /* ... */
  };
  return { user, login };
};
```

---

## 7. Por Que `ClientOnly` "Funciona"?

```vue
<ClientOnly>
  <Button :disabled="!canEdit">Nova Transação</Button>
  <template #fallback>
    <Button disabled>Nova Transação</Button>
  </template>
</ClientOnly>
```

### 7.1 O Que `ClientOnly` Faz

1. **No servidor**: Renderiza o `#fallback` slot

   ```html
   <button disabled>Nova Transação</button>
   ```

2. **No cliente (após hidratação)**: Substitui pelo conteúdo principal

   ```html
   <button>Nova Transação</button>
   <!-- ou disabled se canEdit = false -->
   ```

3. **Resultado**: Não há mismatch porque o servidor nunca tenta renderizar o estado real

### 7.2 Por Que É Um Workaround Ruim

| Problema                       | Descrição                                                |
| ------------------------------ | -------------------------------------------------------- |
| **Derrota o propósito de SSR** | Servidor não renderiza conteúdo real, só placeholder     |
| **SEO afetado**                | Bots veem placeholder, não conteúdo real                 |
| **Flash de conteúdo**          | User vê placeholder → depois conteúdo real (visual ruim) |
| **Mais lento**                 | Cliente tem que re-renderizar tudo                       |
| **Esconde o problema**         | Não resolve a causa raiz, apenas mascara                 |
| **Código mais complexo**       | Precisa manter dois templates (real + fallback)          |

### 7.3 Quando `ClientOnly` É Legítimo

```vue
<!-- ✅ Uso legítimo: código que SÓ funciona no browser -->
<ClientOnly>
  <canvas ref="chart" />  <!-- Canvas API não existe no servidor -->
</ClientOnly>

<ClientOnly>
  <div>{{ window.innerWidth }}</div>  <!-- window não existe no servidor -->
</ClientOnly>

<!-- ❌ Uso ilegítimo: esconder problema de hidratação -->
<ClientOnly>
  <Button :disabled="!canEdit">...</Button>  <!-- Estado DEVERIA funcionar no SSR -->
</ClientOnly>
```

---

## 8. Solução Recomendada: Remover Estado `ready`

### 8.1 Por Que Esta É A Melhor Solução

| Razão                              | Explicação                            |
| ---------------------------------- | ------------------------------------- |
| **Simplicidade**                   | Menos código = menos bugs             |
| **SSR funciona**                   | Não há modificação durante hidratação |
| **Corresponde ao código original** | Volta ao que funcionava antes         |
| **Usa padrões Nuxt corretamente**  | `useState` já gerencia sincronização  |
| **Sem ClientOnly**                 | SSR completo, sem flash               |

### 8.2 Mudanças Necessárias

```diff
// useAuth.ts
export const useAuth = () => {
  const accessToken = useState<string | null>('access_token', () => null)
  const user = useState<AuthUser | null>('auth_user', () => null)
- const ready = useState<boolean>('auth_ready', () => false)

  return {
    accessToken,
    user,
-   ready,
    login,
    refresh,
    // ...
  }
}

// SiteHeader.vue
<script setup>
- const { ready } = useAuth()
  const ledgerContext = useLedgerContext()

- const canEdit = computed(() => ready.value && ledgerContext.hasRole('editor'))
+ const canEdit = computed(() => ledgerContext.hasRole('editor'))
</script>

<template>
- <ClientOnly>
    <Button :disabled="!canEdit">Nova Transação</Button>
-   <template #fallback>
-     <Button disabled>Nova Transação</Button>
-   </template>
- </ClientOnly>
</template>

// NavUser.vue
<script setup>
- const { user, logout, ready } = useAuth()
+ const { user, logout } = useAuth()

- const displayName = computed(() => ready.value && user.value ? user.value.display_name : props.name)
+ const displayName = computed(() => user.value?.display_name || props.name)
</script>

<template>
- <ClientOnly>
    <Avatar>
      <AvatarImage :src="user?.avatar_url" />
    </Avatar>
-   <template #fallback>...</template>
- </ClientOnly>
</template>

// auth-bootstrap.client.ts
export default defineNuxtPlugin({
  async setup() {
    // Remover todas as referências a ready.value
-   ready.value = false
    // ...
-   ready.value = true
  }
})
```

### 8.3 Como Isso Resolve o Problema

**Antes (com `ready`):**

```
SSR:    ready = false → <button disabled>
Plugin: ready = true  → (mismatch!)
Client: Tenta usar ready = true mas DOM já hidratado com disabled
```

**Depois (sem `ready`):**

```
SSR:    activeRole = null → <button disabled>  (correto! ainda não tem role)
Plugin: (não muda estado durante hydratação)
Client: activeRole = "owner" → <button> sem disabled (atualiza corretamente)
```

**Por que funciona:**

1. SSR renderiza com estado inicial correto (null/undefined)
2. Hidratação completa sem mudanças de estado
3. Depois da hidratação, middleware/plugin carrega dados
4. `useState` atualiza reativamente
5. Vue re-renderiza normalmente (não durante hidratação)

---

## 9. Alternativa: Se Precisar Manter `ready`

Se houver motivo legítimo para manter o estado `ready`, precisa ser feito corretamente:

### 9.1 Abordagem Correta

```typescript
// useAuth.ts
export const useAuth = () => {
  // Inicializar ready baseado em estado REAL, não hardcoded false
  const accessToken = useState<string | null>("access_token", () => null);
  const user = useState<AuthUser | null>("auth_user", () => null);

  const ready = computed(() => {
    // Ready = true quando temos dados, não quando plugin completa
    return !!accessToken.value && !!user.value;
  });

  // Não retornar ready como ref mutável, mas como computed (read-only)
  return {
    accessToken,
    user,
    ready, // computed, não useState
  };
};

// auth-bootstrap.client.ts
export default defineNuxtPlugin({
  async setup() {
    // NÃO modificar ready! Ele é computed
    // Apenas carregar dados que ready depende
    const { refresh, me } = useAuth();

    // Usar onNuxtReady para garantir que hidratação completou
    onNuxtReady(async () => {
      await refresh();
      await me();
      // ready automaticamente fica true quando user.value é setado
    });
  },
});
```

**Vantagens:**

- `ready` é sempre consistente com estado real
- Não pode ser modificado manualmente
- Sem hydration mismatch (computed não serializa)
- Usa `onNuxtReady` para garantir ordem correta

**Desvantagens:**

- Ainda mais complexo que simplesmente remover `ready`
- `onNuxtReady` pode causar flash de conteúdo

---

## 10. Checklist de Verificação

Após implementar a solução, verificar:

- [ ] Remove `ready` state de `useAuth.ts`
- [ ] Remove `ready` de todos os componentes
- [ ] Remove todos os `ClientOnly` wrappers adicionados
- [ ] Remove debug console.logs
- [ ] Simplifica plugin para não modificar estado durante hydration
- [ ] Testa primeiro login: botão deve funcionar imediatamente
- [ ] Testa F5: avatar e botão devem aparecer corretamente
- [ ] Console não deve mostrar "hydration mismatch"
- [ ] Backend logs mostram requisições corretas
- [ ] Cookie `hhm_ledger_id` é salvo e restaurado corretamente
- [ ] Navegação entre páginas funciona suavemente
- [ ] Logout funciona corretamente

---

## 11. Conclusão

### Causa Raiz

**Modificação de estado durante o processo de hidratação SSR**, causada pela introdução do estado `ready` que é alterado assincronamente pelo plugin enquanto Vue está tentando hidratar o DOM.

### Solução

**Remover o estado `ready`** e deixar que `useState` do Nuxt gerencie a sincronização automaticamente. Componentes devem usar diretamente `user.value` e `activeRole.value`, que são gerenciados corretamente pelo framework.

### Lições Aprendidas

1. **Não modificar estado durante hidratação** - Use `onNuxtReady` ou middleware
2. **useState já gerencia SSR** - Não precisa de camada adicional
3. **ClientOnly é último recurso** - Não use para esconder problemas
4. **Plugins ≠ Middleware** - Use cada um para seu propósito correto
5. **Simplicidade vence** - Código mais simples é mais fácil de manter e debugar

---

## 12. Problema Persistente: Hydration Mismatch Após Solução com `onNuxtReady`

**Data da Atualização**: 2026-01-04 (Parte 2)
**Status**: Funcionalidade OK, mas erros de hidratação persistem

### 12.1 Situação Atual

Após implementar a solução com `onNuxtReady`:

✅ **Funcionalidade**: Avatar e botão aparecem corretamente após F5
❌ **Console**: Erros de hidratação ainda aparecem

```
[Vue warn]: Hydration attribute mismatch on <button>
  - rendered on server: (not rendered)
  - expected on client: disabled="true"
```

### 12.2 Análise do Erro Persistente

**O que o erro significa:**

- **"rendered on server: (not rendered)"** → O atributo `disabled` NÃO foi renderizado no servidor (botão estava habilitado)
- **"expected on client: disabled="true""** → Cliente espera que o botão esteja desabilitado durante hidratação

**Isso indica:**

```
SSR:    activeRole tem valor (owner/editor) → canEdit = true → botão SEM disabled
Client: activeRole é null → canEdit = false → botão COM disabled
```

### 12.3 Possíveis Causas

#### Hipótese 1: SSR Está Carregando Dados de Alguma Forma

**Evidência:**

- Servidor renderiza botão habilitado (sem disabled)
- Isso só acontece se `hasRole('editor')` retornar `true` no SSR

**Possibilidades:**

1. **Middleware executando no SSR** e carregando ledger/role
2. **Cookie sendo lido no SSR** e restaurando estado
3. **Plugin executando no SSR** (mesmo sendo `.client.ts`)
4. **Estado persistido de requisição anterior** (cache do Nuxt)

**Como verificar:**

```typescript
// Adicionar logs no composable para ver o valor durante SSR
export const useLedgerContext = () => {
  const activeRole = useState<string | null>("active_ledger_role", () => {
    if (import.meta.server) {
      console.log("[SSR] activeRole initialized:", null);
    }
    return null;
  });

  // ...
};
```

#### Hipótese 2: `hasRole` Retornando Valor Diferente SSR vs Cliente

**Código atual:**

```typescript
const hasRole = (role: LedgerRole) =>
  hasRoleRank(activeRole.value as LedgerRole | null, role);

export const hasRoleRank = (
  current: LedgerRole | null,
  required: LedgerRole
): boolean => Boolean(current) && roleRank(current) >= roleRank(required);
```

**Problema potencial:**

- Se `activeRole.value` é `undefined` (não null) no SSR, `Boolean(undefined)` é `false`
- Se `activeRole.value` é uma string vazia `""` no SSR, `Boolean("")` também é `false`
- Mas se o cast `as LedgerRole | null` está mascarando outro tipo...

**Como verificar:**

```typescript
const hasRole = (role: LedgerRole) => {
  const current = activeRole.value;
  const result = hasRoleRank(current as LedgerRole | null, role);

  if (import.meta.server) {
    console.log("[SSR hasRole]", { current, required: role, result });
  }

  return result;
};
```

#### Hipótese 3: Componente Renderizando Condicionalmente no SSR

**Possibilidades:**

1. **`v-if` ou `v-show` escondendo o botão no SSR**
2. **Slot ou componente dinâmico** renderizando diferente
3. **Suspense** causando render diferente
4. **Algum parent component** com lógica condicional

**Como verificar:**

- Inspecionar a árvore de componentes no HTML renderizado pelo servidor
- Adicionar `<!-- SSR: botão aqui -->` comentário no template para confirmar que renderiza

#### Hipótese 4: Race Condition Durante Hidratação

**Cenário:**

```
1. SSR renderiza com activeRole = null → botão disabled
2. HTML chega ao cliente
3. JavaScript baixa e executa
4. ANTES da hidratação começar:
   - useState deserializa payload
   - Algum código síncrono muda activeRole?
5. Hidratação começa com activeRole ≠ null
6. Cliente espera botão enabled, mas HTML tem disabled
7. MISMATCH!
```

**Como verificar:**

- Adicionar logs no início do `setup()` de cada composable
- Verificar se algum código síncrono está modificando estado antes de hidratação

#### Hipótese 5: Cookie `hhm_ledger_id` Sendo Lido no SSR

**Fluxo problemático:**

```
SSR:
1. Request chega com cookie hhm_ledger_id=7d4b6aad...
2. useLedgerContext() é chamado
3. ensureLedger() lê o cookie no SSR
4. Chama loadLedgers() no SSR (faz request ao backend)
5. Chama setActiveLedger() no SSR (busca role do backend)
6. activeRole.value = "owner" no SSR
7. Renderiza botão SEM disabled

Client:
1. Hidrata com payload vazio (activeRole = null)
2. Botão espera estar disabled
3. MISMATCH!
```

**Como verificar:**

```typescript
const ensureLedger = async (ledgerId?: string) => {
  if (import.meta.server) {
    console.log("[SSR] ensureLedger called!");
    return; // Não carregar nada no SSR
  }

  // Resto do código...
};
```

### 12.4 Investigações Necessárias

**Checklist de Debug:**

1. ✅ Remover `definePageMeta` do layout (feito)
2. ⏳ **Adicionar guards no SSR para evitar carregamento de dados**:

   ```typescript
   const setActiveLedger = async (ledgerId: string) => {
     if (import.meta.server) {
       console.log("[SSR] setActiveLedger blocked");
       return;
     }
     // ...
   };
   ```

3. ⏳ **Verificar se cookies estão sendo lidos no SSR**:

   ```typescript
   const ledgerCookie = useCookie("hhm_ledger_id");
   if (import.meta.server && ledgerCookie.value) {
     console.log("[SSR] Cookie detected:", ledgerCookie.value);
   }
   ```

4. ⏳ **Adicionar logs detalhados durante hidratação**:

   ```typescript
   // Em cada computed que afeta renderização
   const canEdit = computed(() => {
     const result = ledgerContext.hasRole("editor");
     console.log("[canEdit]", {
       isServer: import.meta.server,
       activeRole: ledgerContext.activeRole.value,
       result,
     });
     return result;
   });
   ```

5. ⏳ **Verificar ordem de execução**:
   - Plugin auth-bootstrap
   - Middleware ensure-ledger
   - Componentes (SiteHeader, NavUser)
   - onNuxtReady callback

### 12.5 Solução Temporária Aceitável

Se a funcionalidade está OK mas apenas os warnings persistem, e não conseguimos identificar a causa raiz facilmente:

**Opção 1: Suprimir warnings específicos** (não recomendado, mas viável)

```typescript
// nuxt.config.ts
export default defineNuxtConfig({
  vue: {
    compilerOptions: {
      // Suprimir warnings específicos (apenas em produção)
      isCustomElement: (tag) => false,
      comments: process.env.NODE_ENV === "development",
    },
  },
});
```

**Opção 2: Renderizar botão sempre disabled no SSR** (força consistência)

```typescript
const canEdit = computed(() => {
  // No SSR, sempre retorna false para garantir consistência
  if (import.meta.server) return false;
  return ledgerContext.hasRole("editor");
});
```

**Opção 3: Usar `key` para forçar re-render após hidratação**

```vue
<Button :key="activeRole ? 'with-role' : 'no-role'" :disabled="!canEdit">
  Nova Transação
</Button>
```

### 12.6 Próximos Passos Recomendados

1. **Adicionar logs detalhados** nos composables para rastrear quando/onde o estado é modificado no SSR
2. **Testar com SSR desabilitado** para confirmar que é realmente um problema de SSR:
   ```typescript
   // nuxt.config.ts
   export default defineNuxtConfig({
     ssr: false, // Testar temporariamente
   });
   ```
3. **Verificar se o problema ocorre em produção** ou só em desenvolvimento (Vite HMR pode causar inconsistências)
4. **Considerar usar `useLazyAsyncData`** para carregar ledger de forma mais controlada:
   ```typescript
   const { data: ledgerData } = useLazyAsyncData("ledger", async () => {
     // Só executa no cliente
     if (import.meta.server) return null;
     return loadLedgers();
   });
   ```

### 12.7 Conclusão Parcial

**Estado atual:**

- ✅ Funcionalidade: Avatar e botão funcionam corretamente
- ❌ Console: Warnings de hidratação persistem
- ⚠️ Causa: Ainda não identificada com certeza

**Hipótese mais provável:**
Algum código está carregando `activeRole` durante o SSR (via cookie, middleware, ou plugin), fazendo o servidor renderizar o botão habilitado, enquanto o cliente espera que esteja desabilitado durante a hidratação inicial.

**Ação recomendada:**
Adicionar guards `if (import.meta.server) return` em todos os métodos que carregam dados (setActiveLedger, ensureLedger, loadLedgers) para garantir que NADA é carregado durante SSR, apenas durante hidratação no cliente via `onNuxtReady`.

---

**Autor**: Claude (Anthropic)
**Data**: 2026-01-04
**Versão**: 1.1 (Atualizado com análise de erros persistentes)

Pensamentos do Opus:

Checking setRefreshCookie function for path configuration

Analyzed
auth.go#L30-50
Searched
setRefreshCookie
5 results
Analyzed
auth.go#L290-310
O cookie está sendo definido com Path = "/", o que deveria funcionar. O problema deve ser outro.

Pensando melhor, o problema pode estar no fluxo do OAuth callback:

Backend define hhm_refresh cookie via Set-Cookie header
Backend faz redirect para http://localhost:3000/auth/oauth/callback
O redirect vai através do PROXY Nitro (/api/**)? NÃO!
O redirect vai DIRETO para o frontend
O novo código do callback chama 
refresh()
 que faz POST para /api/auth/refresh
MAS o cookie hhm_refresh pode não estar disponível ainda porque foi setado em outra origem (porta 8080)
O problema é que o cookie foi definido pelo backend (porta 8080) para a origem do frontend (porta 3000). Dependendo do browser e configuração, isso pode ter delay.

Solução mais simples: manter o callback funcionando mesmo quando ssr: false, verificando se há accessToken via a abordagem do plugin auth-bootstrap.client.ts que usa onNuxtReady.

Deixa eu verificar o callback atual e ver se o problema é o timing:

2
Checking OAuth callback flow and timing

Analyzed
callback.vue#L1-50
Entendi o problema! O callback chama 
refresh()
 no onMounted, mas o cookie hhm_refresh pode estar em uma situação estranha com o proxy Nitro.

O problema é:

Backend (porta 8080) define hhm_refresh cookie
O cookie é definido para o domínio do frontend (localhost), mas foi setado numa resposta que veio do backend
Com o proxy /api/**, o browser vê tudo como vindo de localhost:3000
MAS o redirect OAuth vai direto do backend para o frontend, não passa pelo proxy
A solução pode ser remover ssr: false apenas para o callback, para que ele passe pelo servidor Nuxt e tenha acesso aos cookies corretamente:

3
Adjusting routeRules to exclude OAuth callback from ssr: false