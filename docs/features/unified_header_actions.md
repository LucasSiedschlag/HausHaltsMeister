# Plano: Header Unificado com Acoes por Pagina

**Data**: 2026-01-09  
**Status**: Planejado  
**Objetivo**: Manter o header com o mesmo layout em todas as paginas, variando apenas o botao de acao na extrema direita.

## Regras de UX

- Header sempre mostra: **Ledger selector | Seletor de mes/ano | Botao de acao**.
- O botao muda de label e handler conforme a pagina.
- Sem tooltip ou texto extra de role no header.

## Plano de Implementacao

1. **Criar um registro global de acao do header**
   - Novo composable `useHeaderAction` (ou similar) em `frontend/layers/shared/composables/`.
   - Estado global via `useState` com: `label`, `onClick`, `disabled`.
   - Helpers: `setHeaderAction`, `clearHeaderAction`.

2. **Atualizar o `SiteHeader`**
   - Remover condicao do seletor de periodo e exibir sempre.
   - Botao da direita passa a usar o estado global do header.

3. **Registrar a acao em cada pagina**
   - Em cada pagina, chamar `setHeaderAction` no `onMounted`.
   - Limpar no `onBeforeUnmount`.
   - Usar label especifico e handler local da pagina.

4. **Mapear labels e handlers**
   - `/journal`: "Nova transacao" -> `openCreate()`.
   - `/accounts`: "Nova conta" -> `openCreate()`.
   - `/categories`: "Nova categoria" -> `openCreate()`.
   - `/ledgers`: "Novo ledger" -> foco no form ou abrir drawer (se existir).
   - `/budgets`: "Novo orcamento" -> `openCreate()`.
   - `/credit-cards`: "Novo cartao" -> `openCreate()`.
   - `/investments`: definir acao principal (ex.: "Nova movimentacao").

5. **Ajustar i18n**
   - Adicionar strings de labels para cada botao novo.

## Observacoes

- O seletor de mes/ano usa o estado de `useJournalPeriod`; ele deve continuar funcionando no `/journal`.
- Para paginas sem acao clara, definir texto e comportamento antes de implementar.
