# Issue: Membros nao veem dados do ledger compartilhado

**Data**: 2026-01-09  
**Status**: Em andamento  
**Contexto**: Usuario editor/viewer enxerga o ledger em `/ledgers` e aparece em `/members`, mas as telas de dados (contas, categorias, transacoes) ficam vazias.

## Sintoma

- `/ledgers` retorna o ledger compartilhado com `role: viewer/editor`.
- `/ledgers/:id/me` confirma a role correta.
- Nas telas de contas/categorias/transacoes, o usuario nao ve nenhum dado.
 - Chamadas diretas em `/api/ledgers/:id/categories` e `/api/ledgers/:id/accounts` retornam `200` com lista vazia.

## Diagnostico

O ledger compartilhado (`7d4b6aad-aaa8-4048-a6a0-61d81a0cef56`) nao possui dados no banco.
Somente o ledger `eb768765-3d61-4897-a441-3dbb7a2bac0a` contem contas/categorias/transacoes.

Consultas realizadas:

```sql
SELECT ledger_id, count(*) FROM accounts GROUP BY ledger_id;
SELECT ledger_id, count(*) FROM categories GROUP BY ledger_id;
SELECT ledger_id, count(*) FROM transactions GROUP BY ledger_id;
```

Resultado: apenas `eb768765-3d61-4897-a441-3dbb7a2bac0a` retorna contagens > 0.

## Hipotese principal

Selecao de ledger em `/ledgers` navega antes do `selectLedger` concluir, criando corrida com o middleware que ativa o ledger padrao. O usuario acha que selecionou o ledger compartilhado, mas o app permanece no ledger padrao (vazio).

## Mudanca aplicada

- `frontend/layers/ledgers/pages/ledgers/index.vue`: agora aguarda `selectLedger` antes de navegar.

## Proximos passos

1. Validar se o usuario quer mover/copiar dados do ledger `eb768...` para o ledger `7d4b...`.
2. Caso nao, convidar usuarios para o ledger com dados (`eb768...`) ou definir o default ledger correto.
3. Opcional: adicionar bloqueio/feedback de troca de ledger enquanto `ledger_context_loading` estiver ativo.
