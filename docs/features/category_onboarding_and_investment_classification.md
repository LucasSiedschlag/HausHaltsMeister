# Category onboarding + investimento sem categorias tecnicas

## Contexto
- Hoje um ledger novo pode ficar sem categorias.
- Queremos sugerir um conjunto simples (sem categorias tecnicas).
- Investimentos devem ser classificados via `investment_action` e usar categorias base de entrada/saida.

## Objetivos
- Se o usuario nao tem categorias, oferecer sugestao e opcao de adicionar.
- Sugerir categorias base (preset default_v1):
  - Gastos fixos (OUT)
  - Conforto (OUT)
  - Lazer (OUT)
  - Objetivos (OUT)
  - Educação (OUT)
  - Investimentos (Saída) (OUT)
  - Salário (IN)
  - Extra (IN)
  - Terceiros (IN)
  - Investimentos (Entrada) (IN)
- Manter investimentos classificados por `investment_action` (`contribution`, `redemption`, `earnings`, `loss`).
- Remover dependencia de categorias tecnicas no modulo de investimentos.

## Nao objetivos
- Nao criar categorias automaticamente sem confirmacao.
- Nao migrar categorias existentes.
- Nao introduzir subcategorias tecnicas para investimento.

## Sugestoes para substituir categorias tecnicas em investimentos
Escolher 1 abordagem (ordem sugerida por robustez):
1. Campo estruturado em transacao/entrada (ex.: `investment_action`):
   - Enum: contribution | redemption | earnings | loss.
   - UI grava e relatorios filtram por este campo.
   - Mais robusto e facil de consultar.
2. Tags/labels:
   - Criar tag table e aplicar tag "Aporte", "Resgate", "Rendimento", "Perda".
   - Flexivel, mas exige joins extras e UI de tags.
3. Memo padronizado:
   - UI grava memo com prefixo: "[Aporte] texto...".
   - Simples, mas fragil para relatorios e edicoes manuais.
4. Inferencia por contas:
   - Se entrada/saida envolve conta de investimento e conta asset:
     - Saida para investimentos => aporte
     - Entrada de investimentos => resgate
     - Entradas sem conta asset (ex.: apenas investimento) => rendimento/perda (com sinal)
   - Bom como fallback, mas pode falhar em casos complexos.

Abordagem escolhida: Campo estruturado em transacao/entrada (ex.: `investment_action`)

## UX proposta (categorias sugeridas)
- Quando `categories.count == 0`:
  - Mostrar empty state com lista sugerida e checkboxes.
  - Permitir marcar "Relevante no orcamento" para categorias OUT.
  - Botao primario: "Adicionar categorias sugeridas".
  - Botao secundario: "Pular por agora".
- Quando categorias ja existem:
  - Mostrar acao "Adicionar sugestoes" no header da pagina de categorias.
  - Na modal, selecionar quais nomes adicionar.

## Backend proposto
- Novo endpoint:
  - `POST /ledgers/:ledgerId/categories/seed`
  - Payload: `preset = "default_v1"` e `names?: string[]` ou `items?: []`.
  - Resposta: `created`, `skipped` por nome.
- Regras:
  - Idempotente por nome (ledger-scoped).
  - Direcao e flags default seguem o preset, com override opcional via `items`.

## Frontend proposto
- Categories page:
  - Empty state e modal com checklist.
  - Chamar `POST /ledgers/:ledgerId/categories/seed`.
- Onboarding leve:
  - Se `categories.count == 0`, disparar modal na primeira visita da area de categorias.

## Investimentos (sem categorias tecnicas)
- Atualizar regras do modulo para usar a abordagem escolhida.
- Com `investment_action`:
  - Registrar aportes/resgates/rendimentos/perdas no campo da transacao.
  - Ajustar relatorios de investimentos para filtrar por action.
  - Usar categorias base `Investimentos (Entrada)` e `Investimentos (Saída)`.

## Plano de implementacao
1. Decidir abordagem para classificar investimento (preferencia: `investment_action`).
2. Implementar endpoint de seed de categorias (idempotente).
3. Atualizar UI de categorias com empty state + modal de sugestao.
4. Ajustar modulo de investimentos para nao depender de categorias tecnicas.
5. Atualizar docs e testes (seed idempotente + investimentos).

## Criterios de aceite
- Ledger sem categorias recebe sugestao com acao manual de adicionar.
- Categorias sugeridas sao criadas uma unica vez.
- Investimentos funciona apenas com categorias base e `investment_action`.
- Sem categorias tecnicas criadas automaticamente.
