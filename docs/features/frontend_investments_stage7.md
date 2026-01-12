# Plano: Frontend Investments (Etapa 7)

**Data**: 2026-01-09  
**Status**: Implementado  
**Referencias**: `docs/frontend/Plano_de_Implementacao_Frontend.md`, `docs/api/07_Investimentos.md`, `docs/ledger/Regras_Investimentos.md`

## Pre-requisitos (antes da UI funcionar corretamente)

1. **Contas obrigatorias no ledger**
   - `Wallet/Pessoal` (`type=wallet`) ou `Conta Corrente` (`type=current`)
   - `Investimentos` (`type=investment`)

2. **Categorias tecnicas recomendadas** (mesmo ledger)
   - `Aportes Investimentos` (out, `is_budget_relevant=true`, `is_budget_base=false`)
   - `Entrada Investimentos (Aporte)` (in, `is_budget_relevant=false`, `is_budget_base=false`)
   - `Resgate Investimentos` (out, `is_budget_relevant=false`, `is_budget_base=false`)
   - `Entrada Resgate (Investimentos)` (in, `is_budget_relevant=false`, `is_budget_base=false`)
   - `Rendimentos` (in, `is_budget_relevant=false`, `is_budget_base=false`)
   - Opcional: `Perdas` (out, `is_budget_relevant=false`, `is_budget_base=false`)

3. **Permissoes**
   - Contribuicoes, resgates e rendimentos: role `editor+`.
   - Resumo: role `viewer+`.

4. **Datas**
   - Utilizar o mesmo seletor global de mes/ano (header).

## Plano de implementacao

1. **Composables**
   - Criar `layers/investments/composables/useInvestments.ts`:
     - `contribute`, `redeem`, `earn` chamando os endpoints POST.
     - `fetchSummary` chamando GET `/investments/summary`.
   - Normalizar erros com `useApiClient` e mensagens padrao.

2. **Setup check da pagina**
   - Verificar se existem as contas e categorias tecnicas.
   - Se faltar algo:
     - Exibir aviso com CTA para criar categoria/conta (link para `/accounts` e `/categories`).
     - Opcional: sugerir nomes exatos e flags esperadas.

3. **Pagina Investments**
   - Criar `frontend/layers/investments/pages/investments/index.vue`.
   - Estrutura:
     - Header: titulo + descricao.
     - Cards de resumo (aporte, resgate, rendimento, variacao liquida).
     - 3 forms: Aporte, Resgate, Rendimento.
   - Envio:
     - `amount_cents`, `occurred_at`, `memo`.
     - Idempotency opcional (se alinhado ao restante do app).
   - Feedback:
     - Notivue para sucesso/erro.
     - Bloqueio para viewer (botao desabilitado com hint).

4. **Integracao com header global**
   - Definir acao: "Novo aporte" (ou "Nova movimentacao") no `useHeaderAction`.
   - Ao clicar, abrir o formulario principal (Aporte).

5. **i18n**
   - Adicionar strings PT/EN para labels, botoes, validacao e cards.

6. **Docs**
   - Atualizar `docs/frontend/Plano_de_Implementacao_Frontend.md` para marcar Etapa 7 como iniciada.
   - Se houver ajustes nos fluxos, refletir em `docs/ledger/Regras_Investimentos.md`.

## Observacoes

- As categorias tecnicas podem ser encontradas por nome + `direction` para evitar conflito.
- O resumo precisa respeitar o filtro de datas do header (usar `useJournalPeriod`).

## Implementacao (2026-01-10)

- `useInvestments` criado com endpoints de aporte, resgate, rendimento e resumo.
- Pagina `/investments` adicionada com cards de resumo e 3 formularios.
- Gate de setup para contas/categorias tecnicas com CTAs para configuracao.
- i18n e validacao de valor adicionadas no layer shared.
