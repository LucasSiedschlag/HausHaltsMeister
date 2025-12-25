# UI & Tema (Shadcn Blue)

O sistema de design utiliza **shadcn-vue** com uma customização de tema "Azul".

## Tokens e Cores

As variáveis CSS estão definidas globalmente em `layers/shared/assets/css/tailwind.css`.
Seguimos o padrão HSL (Hub, Saturation, Lightness) para compatibilidade com o shadcn.

### Paleta Principal (Azul)

- **Primary**: `hsl(221.2 83.2% 53.3%)` (#2563EB - Blue 600)
- **Dark Mode Primary**: `hsl(217.2 91.2% 59.8%)` (#3B82F6 - Blue 500)

### Modo Claro / Escuro

O projeto utiliza `@nuxtjs/color-mode` configurado com a estratégia `class`.

- **Toggle**: O componente `layers/core/components/layout/ThemeToggle.vue` alterna a classe `.dark` no elemento `<html>`.
- **Persistência**: A preferência do usuário é salva automaticamente em cookie/localStorage.

## Como customizar

Para alterar a cor primária globalmente:

1.  Acesse [shadcn themes](https://ui.shadcn.com/themes).
2.  Escolha uma cor ou customize os valores.
3.  Copie os valores das variáveis CSS (`--primary`, `--ring`, etc).
4.  Substitua em `layers/shared/assets/css/tailwind.css`.

> **Nota**: Mantenha a conversão para HSL (sem `hsl()`, apenas os números) para garantir que as classes de opacidade do Tailwind funcionem (ex: `bg-primary/50`).
