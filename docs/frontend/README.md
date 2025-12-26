# Frontend - HausHaltsMeister

Este diretório contém o frontend da aplicação HausHaltsMeister, construído com **Nuxt 3** seguindo uma **Arquitetura Modular (Nuxt Layers)**.

## Visão Geral

O projeto utiliza o conceito de [Nuxt Layers](https://nuxt.com/docs/getting-started/layers) para separar funcionalidades em domínios distintos, permitindo um desenvolvimento escalável que mantém as características de um monólito (repositório único, deploy único), mas com organização de módulos independentes.

O Sistema de Design (UI) é padronizado utilizando **shadcn-vue** + **Tailwind CSS**.

## Como rodar localmente

Certifique-se de ter o `pnpm` instalado.

1.  Acesse a pasta do frontend:

    ```bash
    cd frontend
    ```

2.  Instale as dependências:

    ```bash
    pnpm install
    ```

3.  Inicie o servidor de desenvolvimento:
    ```bash
    pnpm dev
    ```
    O app estará disponível em `http://localhost:3000`.

## Estrutura de Módulos (Layers)

Os módulos estão localizados em `frontend/layers/`. A ordem de carregamento é definida no `nuxt.config.ts` raiz.

- `layers/shared`: UI Kit (shadcn), utilitários globais, estilos base.
- `layers/core`: Layouts principais, páginas base (Home), navegação.
- `layers/dashboard`: Exemplo de feature/domínio específico.
- `layers/payment-methods`: Cadastro de meios de pagamento.
- `layers/picuinhas`: Pessoas e lançamentos de picuinhas.

## Rotas de Picuinhas

- `/picuinhas/pessoas`: cadastro e listagem de pessoas.
- `/picuinhas/lancamentos`: lançamentos vinculados às pessoas (depende do cadastro em Pessoas).

## Rotas de Meios de Pagamento

- `/meios-de-pagamento`: cadastro e listagem de cartões e outros meios.

## Adicionando um novo Módulo (Layer)

1.  Crie uma nova pasta em `frontend/layers/<nome-do-modulo>`.
2.  Adicione um arquivo `nuxt.config.ts` dentro dessa pasta para defini-la como um layer.
    ```typescript
    // frontend/layers/<nome-do-modulo>/nuxt.config.ts
    export default defineNuxtConfig({
      // Configurações específicas do layer
    });
    ```
3.  Registre o layer no `frontend/nuxt.config.ts` principal:
    ```typescript
    export default defineNuxtConfig({
      extends: [
        "./layers/<nome-do-modulo>",
        // ... outros layers
      ],
    });
    ```
    _Nota: A ordem no array `extends` importa. Layers listados primeiro podem ser sobrescritos pelos subsequentes, mas geralmente organizamos da base (shared) para o topo (features)._

## Convenções

- **Imports**: Use aliases automáticos do Nuxt. Componentes em `layers/shared/components` estão disponíveis globalmente se configurados corretamente.
- **Nomes de Componentes**: PascalCase. Ex: `BaseButton`.
- **Rotas**: As rotas são geradas automaticamente baseadas na estrutura de pastas `pages/` dentro de cada layer. Evite conflitos de nomes de arquivos entre layers.
- **Estilos**: Use classes utilitárias do Tailwind sempre que possível. Estilos globais ficam no layer `shared`.
