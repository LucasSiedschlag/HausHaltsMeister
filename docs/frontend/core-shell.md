# Core Layer / App Shell

O módulo `core` é responsável pela estrutura principal da aplicação (Shell), layouts e configurações globais de navegação.

## Componentes do Shell

A interface principal é composta por:

1.  **Sidebar (`layers/core/components/layout/AppSidebar.vue`)**:
    - Menu lateral fixo em desktop.
    - Exibe o logo e lista de navegação.
2.  **Header (`layers/core/components/layout/AppHeader.vue`)**:
    - Barra superior fixa.
    - Contém o botão de "Menu" (hambúrguer) para mobile, que abre um `Sheet` (Drawer) com a navegação.
    - Contém o `ThemeToggle` e menu de usuário.
3.  **Layout (`layers/core/layouts/default.vue`)**:
    - Orquestra o Sidebar e Head e define a área de conteúdo (`<slot />`).

## Diagnóstico rápido (pré-revisão)

Antes dos ajustes, o header estava fixo com offset manual e o layout dependia de `mt`/`pl` para compensar o espaço do header e sidebar, o que gerava alinhamentos inconsistentes; o `ThemeToggle` renderizava ícones via string (podendo ficar invisível) e o `color-mode` estava configurado apenas com `classSuffix`; os componentes de overlay já usavam portal, mas o empilhamento dependia de z-index baixos no shell, aumentando o risco de dropdown/sheet parecerem "atrás" no mobile.

## Navegação

A navegação é centralizada em `layers/core/utils/navigation.ts`.
Para adicionar um novo item de menu:

1.  Edite este arquivo.
2.  Adicione um objeto ao array `mainNavigation`:
    ```typescript
    {
      title: "Nova Feature",
      href: "/nova-feature",
      icon: "NomeDoIconeLucide" // Ex: 'Settings', 'Users'
    }
    ```
    _Nota: Certifique-se que o ícone existe em `lucide-vue-next`._

## Convenções

- **Rotas**: As páginas devem usar o layout padrão (`definePageMeta({ layout: 'default' })`) – que é o comportamento automático se o arquivo layout se chamar `default.vue`.
- **Responsividade**: O breakpoint para sidebar/drawer é `md` (768px).

## Theme Toggle

O `ThemeToggle` fica no header (desktop e mobile) e usa `@nuxtjs/color-mode` para alternar entre `light` e `dark`, persistindo a preferência automaticamente via storage/cookie. A classe `dark` é aplicada no `<html>` para ativar os tokens do shadcn.

## Responsividade do Shell

- Desktop (`md+`): sidebar fixa com largura `w-64`, header `sticky` com altura consistente e conteúdo centralizado em container com `max-w-6xl`.
- Mobile (`< md`): sidebar vira `Sheet` com backdrop; o botão hambúrguer fica sempre visível no header.

## Overlays (Dropdown/Sheet/Popover)

- Overlays usam portal/teleport para o `body` e z-index acima do shell.
- Evite `overflow-hidden`, `transform` e `opacity` no wrapper global, pois criam stacking context e podem cortar overlays.
- Camadas sugeridas: sidebar `z-30`, header `z-40`, overlays `z-60+`.
