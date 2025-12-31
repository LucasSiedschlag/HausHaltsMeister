# Categorias + Orçamento (% flexível mensal) — Regras de negócio + Pseudo-fluxo

Este documento define como funcionam:

- Categorias (com direção IN/OUT e flags de orçamento)
- Orçamento mensal baseado em **% da renda base**
- Plano de orçamento versionado (replica indefinidamente)
- Cálculo do “orçado vs realizado” (mês e períodos)
- Integração com contas (Pessoal/Investimentos/Cartão) e com cartão parcelado (orçamento por parcela postada)

---

## 0) Objetivos e princípios

### Objetivo funcional

1. Permitir classificar entradas e saídas por categoria.
2. Permitir orçamento mensal por categoria em **percentual**.
3. Orçamento é definido uma vez e replicado indefinidamente.
4. Alterações no orçamento valem “daqui pra frente” e preservam meses anteriores.
5. Orçamento é **flexível**: limites do mês variam com a renda real do mês.
6. Permitir categorias “fora do orçamento” sem perder o controle do fluxo de caixa.

### Princípios do modelo

- A **direção** (IN/OUT) é definida pela categoria (`categories.direction`).
- `entries.amount_cents` é sempre positivo; o sinal vem da categoria.
- Orçamento é calculado (não persistido por mês).
- Orçamento é “limite/metas”, mas pode ser ultrapassado; o sistema deve mostrar:
  - abaixo do esperado / acima do esperado

---

## 1) Categorias: definição e comportamento

### 1.1. Campos relevantes (categories)

- `direction`: `in` ou `out` (obrigatório)
- `is_budget_base`: define se uma categoria `in` conta para a base (renda base) do orçamento.
- `is_budget_relevant`: define se uma categoria `out` consome e aparece no orçamento.
- `parent_id`: hierarquia (pai/filho)
- `is_active`: ativa/desativa sem apagar histórico

### 1.2. Regra: categoria define direção

- Se categoria é IN:
  - ao somar entradas, ela entra como positiva
- Se categoria é OUT:
  - ao somar saídas, ela entra como consumo (gasto)

> O sistema não depende de `entries.direction`. Se existir por legado, deve ser redundante e validado contra a categoria.

### 1.3. Flags de orçamento — significado

#### `is_budget_base` (para categorias IN)

Define se a entrada conta para formar o “bolo” de 100% do orçamento no mês.

Exemplos típicos:

- Salário: `in`, `is_budget_base=true`
- Comissão recorrente: `in`, `is_budget_base=true`
- Reembolso: `in`, `is_budget_base=false` (não é renda “real”)
- Resgate de investimento: `in`, `is_budget_base=false`
- Entrada técnica de transferência: `in`, `is_budget_base=false`
- Rendimentos de investimento: `in`, `is_budget_base=false` (para não inflar orçamento)

#### `is_budget_relevant` (para categorias OUT)

Define se a saída consome e aparece no orçamento.

Exemplos típicos:

- Mercado: `out`, `is_budget_relevant=true`
- Custos fixos: `out`, `is_budget_relevant=true`
- Pagamento fatura cartão: `out`, `is_budget_relevant=false` (evitar dupla contagem)
- Transferência interna (Pessoal -> Investimentos): `out`, `is_budget_relevant=depende`
  - se você quer “orçar aportes”, então `true`
  - se você quer “fora do orçamento”, então `false`

### 1.4. Regras de consistência (recomendadas)

Na aplicação (service layer), aplique:

- Se `direction=in`:
  - `is_budget_relevant` deve ser `false` (ou ignorado)
- Se `direction=out`:
  - `is_budget_base` deve ser `false` (ou ignorado)

Motivo:

- evita combinações sem sentido
- simplifica relatórios

### 1.5. Hierarquia de categorias (parent/child)

Categorias podem ser agrupadas:

- “Custos Fixos” (pai)
  - “Aluguel” (filha)
  - “Internet” (filha)
  - etc.

Uso:

- UI/relatório agrupado
- orçamento por pai com `include_children=true` (nas linhas do orçamento)

---

## 2) Orçamento: modelo versionado por % (plans, versions, lines)

### 2.1. Tabelas relevantes

- `budget_plans`: o “plano de verdade” (container)
- `budget_plan_versions`: versões com `effective_from_month`
- `budget_plan_lines`: percentuais por categoria para aquela versão

### 2.2. Conceito: orçamento padrão que replica indefinidamente

- Você cria 1 plano por ledger.
- Você cria uma versão inicial válida a partir de um mês:
  - ex.: `2026-01-01`
- Essa versão vale para todos os meses seguintes, até existir uma nova versão.

### 2.3. Alteração do orçamento preservando histórico

Quando quiser mudar:

- criar nova `budget_plan_version` com `effective_from_month` no mês da mudança
- inserir novas `budget_plan_lines`
  Meses anteriores continuam usando a versão anterior automaticamente.

### 2.4. Linhas do orçamento (`budget_plan_lines`)

Cada linha define:

- `category_id` (normalmente OUT)
- `percent` (ex.: 10.0000)
- `include_children` (se deve somar subcategorias)

> Não existe “valor fixo” no banco. Só percent.

---

## 3) Cálculo do orçamento mensal (visão flexível)

### 3.1. Entrada: mês alvo

O mês é representado por `month = date(YYYY, MM, 01)`.

### 3.2. Passo 1 — Encontrar a versão aplicável no mês

Regra:

- para `month M`, escolher a versão:
  - `max(effective_from_month)` tal que `effective_from_month <= M`

Pseudo:

1. `version = SELECT * FROM budget_plan_versions
WHERE plan_id = ...
  AND effective_from_month <= M
ORDER BY effective_from_month DESC
LIMIT 1`

Se não houver versão:

- orçamento “não configurado” para o período.

### 3.3. Passo 2 — Calcular renda base do mês (income_base)

Regra:

- somar entries do mês onde:
  - categoria `direction=in`
  - `is_budget_base=true`

Pseudo:

- `income_base = SUM(entries.amount_cents)`
  join categories, transactions:
  - `categories.direction = in`
  - `categories.is_budget_base = true`
  - `transactions.occurred_at` dentro do mês

> A renda base é o “bolo” do orçamento (100%).

### 3.4. Passo 3 — Calcular limite orçado por categoria

Para cada linha do orçamento:

- `budget_limit(category) = income_base * percent / 100`

Se `include_children=true`:

- o gasto realizado deve considerar filhos no cálculo (ver passo 4).

### 3.5. Passo 4 — Calcular realizado (spent_actual) por categoria

Regra:

- somar entries do mês onde:
  - categoria `direction=out`
  - `is_budget_relevant=true`
  - categoria corresponde à linha do orçamento
  - se `include_children=true`: inclui subcategorias

Pseudo:

- `spent_actual = SUM(entries.amount_cents)`
  join categories, transactions:
  - `categories.direction = out`
  - `categories.is_budget_relevant = true`
  - `transactions.occurred_at` dentro do mês
  - filtro de categoria (com ou sem filhos)

### 3.6. Passo 5 — Comparação e indicadores

Para cada categoria orçada:

- `delta = spent_actual - budget_limit`
- `status`:
  - delta <= 0 => abaixo / dentro do limite
  - delta > 0 => acima do limite
- `usage_pct = spent_actual / budget_limit` (se limit > 0)

### 3.7. Categorias fora do orçamento

Para auditoria e transparência:

- exibir também um bloco:
  - “Gastos fora do orçamento”
  - soma de OUT com `is_budget_relevant=false`

---

## 4) Fluxos de negócio (CRUD e operações)

## 4.1. Fluxo — Criar categoria

Entrada:

- ledger_id
- name
- direction (in/out)
- flags (dependendo da direção)
- parent_id opcional

Validações:

- name único por ledger
- se direction=in:
  - default: is_budget_base=true
  - force/override: is_budget_relevant=false
- se direction=out:
  - default: is_budget_relevant=true
  - force/override: is_budget_base=false

Persistência:

- INSERT categories

---

## 4.2. Fluxo — Alterar categoria (flags e direção)

Regras:

- Evitar mudar direction de categoria já usada (pode quebrar histórico). Preferir:
  - desativar categoria antiga (`is_active=false`)
  - criar nova categoria

Se permitir mudar:

- precisa de migração de dados (custoso). Não recomendado no início.

Alterar flags:

- permitido a qualquer momento
- impacto:
  - histórico permanece, mas relatórios antigos podem mudar de interpretação
  - para evitar isso, você pode optar por não retroagir (complicado).
    Recomendação inicial:
- aceitar retroatividade dos flags (simples)
- se no futuro quiser “flags versionados”, isso vira um upgrade.

---

## 4.3. Fluxo — Criar orçamento inicial (plano + versão + linhas)

Entrada:

- ledger_id
- month_start (ex.: 2026-01-01)
- lista de percentuais por categoria OUT

Validações:

- somatório de percentuais:
  - pode ser <= 100 (recomendado) ou permitir >100 (usuário decide)
- categorias devem ser direction=out (por padrão `apply_to=out_only`)
- se categoria `is_budget_relevant=false`, normalmente não faz sentido orçar (bloquear ou avisar)

Persistência (transação DB):

1. create `budget_plans` (se não existir)
2. create `budget_plan_versions` com `effective_from_month=month_start`
3. insert `budget_plan_lines` (uma por categoria)

---

## 4.4. Fluxo — Alterar orçamento “daqui pra frente”

Entrada:

- ledger_id
- effective_from_month (ex.: 2026-04-01)
- nova lista de percentuais

Regras:

- não atualizar version antiga
- criar nova version

Persistência:

1. INSERT `budget_plan_versions`
2. INSERT `budget_plan_lines`

Resultado:

- meses anteriores continuam com a versão anterior
- meses a partir do effective_from_month usam nova versão

---

## 4.5. Fluxo — Gerar painel do mês (budget monthly summary)

Entrada:

- ledger_id
- month (YYYY-MM-01)

Saída:

- income_base
- lista de categorias orçadas:
  - percent
  - budget_limit
  - spent_actual
  - delta
  - usage_pct
- lista de gastos fora do orçamento (opcional)

Pseudo alto nível:

1. `version = get_version_for_month(month)`
2. `income_base = calc_income_base(month)`
3. para cada line:
   - `limit = income_base * percent/100`
   - `actual = calc_spent(month, category, include_children)`
4. compute deltas e status
5. compute “outside budget” totals

---

## 5) Integração com Accounts e com Cartão (parcelas)

### 5.1. Accounts não mudam o orçamento por si só

Orçamento é por categoria e por mês.
`account_id` permite segmentar relatórios (ex.: “gastos do cartão”, “gastos do pessoal”),
mas o orçamento pode considerar tudo, ou filtrar (decisão do produto).

Recomendação inicial:

- orçamento considera gastos relevantes em todas as contas (Pessoal, Cartão etc.)
- cartão entra no orçamento **quando as parcelas são postadas**.

### 5.2. Como o cartão entra no orçamento (regra já alinhada)

- Compra parcelada cria `installment_plan` + `installments` (sem journal)
- Ao “postar” parcelas do mês:
  - cria `transaction + entry` na conta do cartão
  - categoria real OUT (Mercado etc.) com `is_budget_relevant=true`
  - isso consome orçamento do mês

### 5.3. Pagamento de fatura não entra no orçamento

- pagamento é transferência com categorias técnicas:
  - OUT em Pessoal: `Pagamento Fatura Cartão` (budget_relevant=false)
  - IN no Cartão: `Entrada Pagamento Cartão` (budget_base=false)

Assim:

- evita dupla contagem
- orçamento reflete consumo real (parcelas)

---

## 6) Casos especiais e recomendações

### 6.1. Categorias IN que não devem aumentar orçamento

Exemplos:

- reembolso
- estorno recebido
- resgate de investimento
- rendimentos de investimento (se você não quiser “expandir” orçamento)

Regra:

- direction=in
- is_budget_base=false

### 6.2. Categorias OUT que não devem consumir orçamento

Exemplos:

- pagamento de fatura
- transferências internas (se você não quer orçar)
- impostos eventuais (se você preferir fora do orçamento)
- “pass-through” (dinheiro que não é gasto real)

Regra:

- direction=out
- is_budget_relevant=false

### 6.3. Somatório de percentuais

Permitir:

- <= 100 (recomendado)
- < 100 significa “sobra” (reserva livre / poupança / imprevistos)
- > 100 é permitido, mas indica plano inviável (apresentar alerta)

### 6.4. Hierarquia + include_children

Se `include_children=true`:

- gasto realizado da categoria pai = soma dos filhos (e netos)
- isso permite orçar “Custos Fixos” sem detalhar.

Implementação recomendada:

- CTE recursiva (bom para começar)
- upgrade opcional: tabela de fechamento (closure table) para performance.

---

## 7) Métricas e análises por período (mensal, anual, custom)

### 7.1. Visão mensal

- calculada como descrito no capítulo 3

### 7.2. Visão por período (anual ou custom)

Para um período de meses:

- somar mês a mês:
  - `income_base_month`
  - `budget_limit_month(category)`
  - `spent_actual_month(category)`
    Isso lida automaticamente com:
- renda variável
- versões diferentes de orçamento dentro do período

Saídas:

- `budget_total_period`
- `spent_total_period`
- `delta_period`
- “acima/abaixo do esperado” por categoria

### 7.3. Acima/abaixo do esperado

No período:

- se `spent_total > budget_total` => acima
- se `spent_total < budget_total` => abaixo
- exibir ranking das maiores divergências

---

## 8) Regras de validação recomendadas (service layer)

1. Ao criar uma `budget_plan_line`:

- categoria deve ser OUT
- categoria deve ser budget_relevant=true (ou exigir confirmação)

2. Ao calcular renda base:

- só considerar categorias IN com budget_base=true

3. Ao calcular realizado:

- só considerar categorias OUT com budget_relevant=true
- considerar include_children conforme a linha

4. Ao editar flags de categoria:

- alertar que o efeito é retroativo (versão
