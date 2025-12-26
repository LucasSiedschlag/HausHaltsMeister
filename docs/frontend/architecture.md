# Arquitetura do Frontend

## Estrutura de Pastas (Árvore)

A estrutura do projeto dentro de `frontend/` segue o padrão de **Monólito Modular** com Nuxt Layers.

```
frontend/
├── app/                  # App shell (opcional, se não usar layer core para tudo)
├── layers/               # Módulos da aplicação
│   ├── shared/           # Camada Base (UI, Utils, Configs)
│   │   ├── components/
│   │   │   └── ui/       # Componentes shadcn-vue (Button, Sheet, Dropdown, etc)
│   │   ├── utils/        # Ex: cn.ts, formatters
│   │   ├── composables/  # Hooks genéricos
│   │   ├── assets/       # CSS global (Tailwind + Tema Azul)
│   │   └── nuxt.config.ts
│   │
│   ├── core/             # Camada de Aplicação (Layouts, Home)
│   │   ├── layouts/      # default.vue (Shell com Sidebar + Header)
│   │   ├── components/
│   │   │   └── layout/   # AppSidebar.vue, AppHeader.vue, ThemeToggle.vue
│   │   ├── utils/        # navigation.ts
│   │   ├── pages/        # index.vue (Home)
│   │   ├── plugins/      # Plugins globais
│   │   └── nuxt.config.ts
│   │
│   └── dashboard/        # Feature Layer: Dashboard
│       ├── components/   # Componentes específicos do dashboard
│       ├── pages/        # Rotas /dashboard...
│       ├── composables/  # Lógica específica do domínio
│       └── nuxt.config.ts
│
│   └── picuinhas/        # Feature Layer: Picuinhas (Pessoas + Lançamentos)
│       ├── components/   # Componentes específicos do domínio
│       ├── pages/        # Rotas /picuinhas/...
│       ├── services/     # Integração com API de picuinhas
│       ├── types/        # DTOs do domínio
│       ├── validation/   # Validação de formulários
│       └── nuxt.config.ts
│
├── public/               # Assets estáticos
├── nuxt.config.ts        # Ponto de entrada e registro de layers
├── tailwind.config.ts    # Configuração do Tailwind (estende shared se necessário)
└── components.json       # Config do shadcn-vue (aponta para layers/shared)
```

## Layers e Responsabilidades

### 1. Layers Base (`layers/shared`)

- **Responsabilidade**: Prover blocos de construção para todos os outros layers. Deve ser agnóstico de regras de negócio complexas.
- **Conteúdo**:
  - Componentes de UI (shadcn).
  - Configurações base do Tailwind.
  - Utilitários de formatação (datas, números).
  - Tipos globais.
- **Regras de Dependência**: Não pode depender de nenhum outro layer.

### 2. Layers Core (`layers/core`)

- **Responsabilidade**: A "cola" da aplicação. Define a estrutura de navegação principal, layouts globais e páginas institucionais ou de aterrissagem.
- **Conteúdo**:
  - Layout `default.vue`.
  - Página `index.vue` (Home).
  - Menu de navegação principal.
- **Regras de Dependência**: Depende de `shared`.

### 3. Layers de Feature (`layers/dashboard`, `layers/auth`, etc.)

- **Responsabilidade**: Conter a lógica e visualização de um domínio de negócio específico.
- **Conteúdo**:
  - Páginas específicas (`dashboard/index.vue`).
  - Componentes "smart" (com lógica de negócio).
  - Composables de domínio (ex: `useRevenueStats`).
- **Regras de Dependência**: Depende de `shared`. Idealmente não deve depender diretamente de outro layer de feature (acoplamento horizontal), mas pode ser orquestrado pelo `core` ou comunicar via eventos/estado global se estritamente necessário.

#### Domínios atuais

- `layers/categories`: cadastro de categorias.
- `layers/budget`: orçamento mensal.
- `layers/picuinhas`: pessoas e lançamentos de picuinhas.

## Estratégias

### Rotas

O Nuxt mescla as pastas `pages/` de todos os layers.

- Rotas globais (`/`) ficam em `layers/core/pages/index.vue`.
- Rotas de feature (`/dashboard`) ficam em `layers/dashboard/pages/index.vue` (que mapeia para `/dashboard` se configurado assim, ou usar estrutura de pastas `layers/dashboard/pages/dashboard/index.vue`).
- **Trade-off**: Para evitar colisão, prefira prefixar a estrutura de pastas dentro de `pages` com o nome da feature (ex: `layers/dashboard/pages/dashboard/...`) a menos que queira injetar uma rota na raiz propositalmente.

### UI e shadcn-vue

- **Componentes**: Instalados em `layers/shared/components/ui`.
- **Tokens/Tema**: Definidos no `layers/shared/assets/css/tailwind.css` (ou similar) e referenciados no `tailwind.config.ts`.
- **Utils**: A função `cn()` (class merger) fica em `layers/shared/utils/cn.ts`.

### Prioridade de Layers

No `nuxt.config.ts` raiz:

```typescript
extends: [
  './layers/shared',    // Base
  './layers/core',      // Shell
  './layers/dashboard'  // Features (podem sobrescrever core se colidirem, mas idealmente são aditivas)
]
```

A ordem de `extends` define a prioridade de override. O último da lista tem maior precedência para sobrescrever arquivos de mesmo nome.
