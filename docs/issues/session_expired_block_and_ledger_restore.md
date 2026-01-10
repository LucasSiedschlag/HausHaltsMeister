# Issue: Sessao expirada sem bloqueio total e ledger nao restaurado apos F5

**Data**: 2026-01-09  
**Status**: Resolvido  
**Contexto**: Ao expirar a sessao, o usuario permanecia navegando sem dados. Ao dar F5 em paginas internas, header/sidebar mostravam "Select a ledger" ate abrir `/ledgers`.

## Sintomas

- Sessao expirada exibia dialogo, mas a UI continuava navegavel.
- Em refresh de paginas como `/categories`, o ledger ativo nao aparecia no header/sidebar.

## Causa provavel

- Ausencia de bloqueio global quando `sessionExpired` ativa.
- `ledger.global.ts` nao roda sem token e o `auth-bootstrap` nao garantia `ensureLedger` apos refresh.

## Correcao aplicada

- Overlay fullscreen de sessao expirada para bloquear interacao.
- Plugin client + middleware global para manter bloqueio enquanto `sessionExpired` estiver ativo.
- `auth-bootstrap` agora usa `ensureLedger` e observa o token para restaurar ledger apos refresh.

## Arquivos tocados

- `frontend/app/app.vue`
- `frontend/app/plugins/session-expired.client.ts`
- `frontend/layers/core/middleware/session-expired.global.ts`
- `frontend/app/plugins/auth-bootstrap.client.ts`
- `frontend/layers/shared/composables/useSessionBlock.ts`

## Validacao

1. Forcar expiracao e confirmar overlay bloqueando a UI ate "Ir para o login".
2. Fazer login e dar F5 em `/categories` e validar ledger ativo no header/sidebar.
