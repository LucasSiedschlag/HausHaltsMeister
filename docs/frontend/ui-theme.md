# UI & Tema (shadcn + Tailwind)

O sistema de design utiliza **shadcn-vue** com Tailwind. O tema e definido por tokens CSS no layer `shared` e deve ser compativel com SSR.

## Tokens e cores

- Tokens ficam em `layers/shared/assets/css/tailwind.css`.
- Formato HSL (numeros, sem `hsl()`), compatível com shadcn.
- Deve existir uma paleta clara e uma escura (classe `.dark`).
- O tema e predominantemente azul, seguindo os defaults do shadcn (blue).

## Diretrizes de tema

- Priorizar contraste alto para leitura financeira.
- Base visual: azul (primary) e neutros shadcn.
- Cores de entrada/saida devem ser consistentes:
  - entrada: verde
  - saida: vermelho
  - neutro/transferencia: cinza
- Estados de risco (ex.: fatura vencida) usam amarelo/laranja.

## SSR e modo escuro

- A classe `dark` deve ser definida de forma consistente no SSR para evitar flash.
- Preferir `color-mode` com `preference` persistida em cookie.
- Componentes que leem `window` devem usar `<ClientOnly>`.

## Componentes shadcn

- Componentes ficam em `layers/shared/components/ui`.
- Wrappers de dominio ficam no layer da feature.
- `cn()` fica em `layers/shared/utils/cn.ts`.

## Personalizacao

1. Ajuste os tokens em `layers/shared/assets/css/tailwind.css`.
2. Mantenha o primary azul nos dois modos (claro/escuro).
3. Atualize `tailwind.config.ts` se adicionar novas cores utilitarias.
4. Garanta paridade entre modo claro e escuro.

## Referencias

- Consulte `docs/frontend/references.md` para links de shadcn themes e Tailwind.
