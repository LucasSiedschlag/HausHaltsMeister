# Referencias do Frontend

Lista de referencias tecnicas e conceituais utilizadas na construcao e evolucao do frontend.

## Arquitetura & Nuxt

- [Authoring Nuxt Layers (Docs Oficiais)](https://nuxt.com/docs/guide/going-further/layers)
  - Relevancia: estrutura e prioridade de layers.
- [Nuxt Config Reference](https://nuxt.com/docs/api/nuxt-config)
  - Relevancia: configuracao de `extends`, alias e composables globais.
- [Nuxt Route Rules](https://nuxt.com/docs/guide/going-further/route-rules)
  - Relevancia: cache e headers por rota (futuro).
- [Nuxt Rendering Modes](https://nuxt.com/docs/getting-started/going-further#rendering-modes)
  - Relevancia: SSR como padrao e uso de client-only quando necessario.

## UI & Design System

- [shadcn-vue (Nuxt)](https://www.shadcn-vue.com/docs/installation/nuxt.html)
  - Relevancia: instalacao correta do shadcn-vue.
- [shadcn UI Themes](https://ui.shadcn.com/themes)
  - Relevancia: tokens base para o tema azul.
- [Radix Vue](https://www.radix-vue.com/)
  - Relevancia: acessibilidade e primitives headless.
- [Tailwind CSS](https://tailwindcss.com/docs)
  - Relevancia: utilitarios e tokens.

## Auth & OAuth

- [OAuth 2.0 for Browser-Based Apps (IETF BCP)](https://www.rfc-editor.org/rfc/rfc8252)
  - Relevancia: PKCE e fluxo seguro para SPAs/SSR.
- [OpenID Connect Core](https://openid.net/specs/openid-connect-core-1_0.html)
  - Relevancia: claims e userinfo.
- [Google OAuth 2.0](https://developers.google.com/identity/protocols/oauth2)
  - Relevancia: endpoints e parametros do provider.
- [GitHub OAuth](https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps)
  - Relevancia: fluxo de autorizacao do provider.

## Seguranca & Cookies

- [SameSite Cookies](https://web.dev/samesite-cookies-explained/)
  - Relevancia: configuracao de refresh token em cookie.
- [OWASP: JSON Web Token](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html)
  - Relevancia: boas praticas de JWT (claims, expiracao).
