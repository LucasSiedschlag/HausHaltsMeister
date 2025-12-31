# Documento de Arquitetura — HausHaltsMeister (Fluxo de Caixa + Orçamento % + Investimentos + Cartão)

> Versão: 1.0 (baseada no modelo e regras consolidadas na conversa)  
> Linguagem/stack alvo: Go + PostgreSQL + sqlc + migrations  
> Natureza do produto: **offline-first / local-first**, evoluindo para **multiusuário** e potencialmente multi-dispositivo no futuro.

---

## 1) Visão geral

O HausHaltsMeister é uma aplicação de controle financeiro pessoal focada em **fluxo de caixa** e **orçamento mensal por categoria**, com uma base de dados robusta e extensível. O núcleo do domínio é um **journal** (diário de eventos) modelado por `transactions` (evento) e `entries` (linhas de valor). A direção financeira (**IN/OUT**) é definida pela categoria, e o orçamento é calculado mensalmente como **percentual** sobre a **renda base** do mês.

A aplicação foi desenhada para:

- funcionar muito bem para uso individual (1 pessoa, 1 ledger),
- suportar **multiusuário** no futuro (cada pessoa com seu ledger e possibilidade de compartilhamento),
- incluir módulos avançados sem quebrar o core:
  - conta de **investimentos** (separada internamente),
  - **cartão de crédito** com parcelas e fatura,
  - orçamentos versionados (mudanças “daqui pra frente” sem mexer no passado).

---

## 2) Objetivos e não-objetivos

### 2.1 Objetivos

- **Controle de fluxo de caixa** com registro consistente e auditável.
- **Orçamento mensal por categoria** baseado em **% da renda base** (visão flexível).
- Orçamento “padrão” replicado indefinidamente e **versionado por mês de vigência**.
- Separação interna de **contas** (Pessoal, Investimentos, Cartão).
- Cartão de crédito completo:
  - compra parcelada,
  - posting de parcelas por mês,
  - fatura (statement) e pagamento,
  - orçamento consumido **por parcela postada no mês**, não no ato da compra.
- Segurança por isolamento de `ledger_id` e validações fortes.

### 2.2 Não-objetivos (por enquanto)

- Contabilidade completa/double-entry formal.
- Precificação de ativos e carteira por ticker/cotações.
- Integração nativa com bancos (import pode existir no futuro, mas não é requisito inicial).
- Motor de reconciliação bancária avançada (pode ser fase posterior).

---

## 3) Princípios arquiteturais

1. **Ledger como fronteira de dados**

   - Qualquer dado financeiro é escopado por `ledger_id`.
   - Toda consulta/ação valida membership/role e filtra por ledger.

2. **Journal como fonte da verdade**

   - `transactions` registram o evento.
   - `entries` registram os valores.
   - Saldo e relatórios são derivados, não armazenados.

3. **Semântica centralizada em categorias**

   - `categories.direction` define IN/OUT.
   - Flags determinam comportamento no orçamento:
     - `is_budget_base` (para IN),
     - `is_budget_relevant` (para OUT).

4. **Orçamento calculado, não persistido por mês**

   - Baseado em:
     - versão vigente do plano no mês,
     - renda base real do mês,
     - percentuais por categoria.

5. **Evolução sem refatoração traumática**
   - Módulos “acima do core” (cartão, investimentos) são implementados por regras e tabelas auxiliares, mas sempre convergem para o journal.

---

## 4) Visão de alto nível dos componentes

### 4.1 Backend (Go)

Arquitetura recomendada: **monólito modular** (clean-ish, porém pragmático), com módulos internos bem definidos:

- `auth` (usuários, sessão, RBAC/ledger membership)
- `ledger` (ledgers, membros)
- `accounts` (Pessoal/Investimentos/Cartão, metadados de cartão)
- `categories` (direção, flags, hierarquia)
- `journal` (transactions/entries, validações, edição/exclusão)
- `budget` (plan/version/lines, cálculos mensais e por período)
- `investments` (fluxos de aporte/resgate/rendimento — em cima do journal)
- `creditcard` (plans/installments/statements — em cima do journal)
- `reporting` (queries agregadas, visões mensais/períodos)

**Ponto importante**: `journal` é o “core” e os demais módulos dependem dele.

### 4.2 Banco (PostgreSQL)

- Migrations modulares.
- `sqlc` para gerar repositórios fortemente tipados.
- Regras de integridade primariamente no service layer (primeira fase), com possibilidade futura de:
  - constraints adicionais,
  - triggers de consistência,
  - RLS (se virar SaaS multi-tenant).

### 4.3 Frontend (futuro)

Qualquer UI (web/desktop) pode consumir a API. O domínio foi desenhado para ser independente de UI.

---

## 5) Modelo de domínio (visão conceitual)

### 5.1 Entidades base

- **User**: identidade.
- **Ledger**: espaço financeiro (boundary).
- **LedgerMember**: vínculo e permissões (roles).
- **Account**: caixinha interna do ledger (`cash`, `investment`, `credit_card`).
- **Category**: classificação e semântica (IN/OUT + flags).
- **Transaction**: evento (data/descrição/autor).
- **Entry**: linha de valor (conta + categoria + amount + kind + memo).

### 5.2 Orçamento (calculado)

- **BudgetPlan**: container do orçamento do ledger.
- **BudgetPlanVersion**: versões com vigência a partir de um mês.
- **BudgetPlanLine**: percentuais por categoria (geralmente OUT).

### 5.3 Cartão (auxiliar + converge no journal)

- **CreditCard** (metadados 1:1 com account do tipo `credit_card`)
- **InstallmentPlan** (compra parcelada)
- **Installment** (parcela mensal)
- **CreditCardStatement** (fatura do mês)

**Regra central**: a parcela só “vira gasto” quando é **postada**, criando `transaction + entry`.

---

## 6) Fluxos principais

### 6.1 Criação de um lançamento (fluxo geral)

1. Validar usuário e acesso ao ledger.
2. Criar `transaction`.
3. Criar `N entries` (split permitido).
4. Validar consistência:
   - account/category pertencem ao ledger,
   - transfer balanceada quando `kind=transfer`.

### 6.2 Orçamento mensal

Para um mês M:

1. Determinar versão vigente do orçamento (`effective_from_month <= M`).
2. Calcular renda base do mês:
   - somatório de IN com `is_budget_base=true`.
3. Para cada linha do orçamento:
   - limite = renda_base \* percent.
4. Calcular realizado:
   - somatório de OUT com `is_budget_relevant=true`.
5. Exibir deltas (abaixo/acima).

### 6.3 Investimentos

- Aporte: transferência Pessoal (OUT) -> Investimentos (IN).
- Rendimento reinvestido: 1 entry IN na conta Investimentos em categoria “Rendimentos” com `is_budget_base=false`.
- Resgate: transferência inversa.

### 6.4 Cartão com parcelas e fatura

- Compra parcelada:
  - cria `installment_plan` + `installments`,
  - **não cria journal** (para não consumir orçamento no ato).
- Posting mensal:
  - para parcelas do mês, cria `transaction+entry` no cartão com categoria real (Mercado etc.).
  - **isso consome orçamento do mês** (por parcela).
- Pagamento de fatura:
  - transaction transfer Pessoal -> Cartão com categorias técnicas fora do orçamento.

---

## 7) Contratos de API (alto nível)

A API deve ser organizada por `ledger_id` como primeiro-class:

- `POST /ledgers`
- `GET /ledgers`
- `POST /ledgers/{ledgerId}/accounts`
- `POST /ledgers/{ledgerId}/categories`
- `POST /ledgers/{ledgerId}/transactions`
- `GET /ledgers/{ledgerId}/transactions?from=&to=&account=&category=`
- `GET /ledgers/{ledgerId}/budget/monthly?month=YYYY-MM`
- `POST /ledgers/{ledgerId}/budget/versions` (criar nova versão)
- `POST /ledgers/{ledgerId}/credit-cards/{cardAccountId}/plans` (parcelas)
- `POST /ledgers/{ledgerId}/credit-cards/{cardAccountId}/post?month=YYYY-MM`
- `POST /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/close?month=YYYY-MM`
- `POST /ledgers/{ledgerId}/credit-cards/{cardAccountId}/statements/pay?month=YYYY-MM`

**Regra de segurança**: endpoints sempre incluem `ledgerId` e sempre validam membership.

---

## 8) Segurança, integridade e validações

### 8.1 Segurança

- Autenticação: sessão/token.
- Autorização: RBAC por ledger:
  - viewer/editor/owner.

### 8.2 Integridade (service layer)

- `entry.ledger_id == transaction.ledger_id`
- `account.ledger_id == ledger_id`
- `category.ledger_id == ledger_id`
- Transferência:
  - `SUM(OUT) == SUM(IN)` por transaction quando `kind=transfer`

### 8.3 Idempotência

- Importações futuras: `transactions.external_source/external_id`.
- Rotinas automáticas (posting parcelas):
  - postar apenas parcelas `scheduled`,
  - gravar `posted_transaction_id`.

---

## 9) Estratégia de dados, migrations e sqlc

### 9.1 Migrations

- Migrations por módulos (core, budget, investments, creditcard).
- Seeds opcionais:
  - accounts padrão (Pessoal, Investimentos),
  - categorias técnicas (pagamento fatura, rendimentos etc.),
  - budget plan inicial (se desejar).

### 9.2 sqlc

- Queries pequenas e previsíveis.
- Preferir queries por ledger para evitar vazamentos.

### 9.3 Modelos calculados (views no app)

- Budget mensal é calculado em runtime via queries agregadas.
- Saldos por conta também derivados.

---

## 10) Observabilidade e rastreabilidade

### 10.1 Logs

- log estruturado por request:
  - user_id, ledger_id, endpoint, latency, status
- logs de domínio para ações críticas:
  - posting de parcelas,
  - pagamento de fatura,
  - criação de versões de orçamento.

### 10.2 Métricas (futuro)

- número de transações por dia/mês
- tempo médio de geração do budget mensal
- falhas de validação

---

## 11) Testes

### 11.1 Testes de domínio (unit)

- validações de transfer balanceada
- regras de flags de categoria
- seleção de versão de orçamento por mês
- posting idempotente de parcelas

### 11.2 Testes de integração (DB)

- criação transaction + entries em transação
- queries de orçamento mensal
- rotinas de cartão (plan -> installments -> post -> statement -> pay)

### 11.3 Testes de segurança

- tentativa de acessar recursos de outro ledger (deve falhar)
- role viewer tentando escrever (deve falhar)

---

## 12) Roadmap técnico sugerido (ordem de implementação)

1. Core:
   - users/auth (mínimo)
   - ledger + accounts + categories
   - journal (transactions + entries)
2. Budget:
   - budget plan/version/lines
   - monthly summary
3. Investimentos:
   - aporte/resgate/rendimento (em cima do journal)
4. Cartão:
   - credit_cards metadata
   - installment_plans/installments
   - posting mensal
   - statements + pagamento
5. Reporting:
   - painéis por período, rankings de delta, “fora do orçamento”
6. Evoluções:
   - soft delete
   - import idempotente
   - RLS/constraints avançadas (se virar SaaS)

---

## 13) Decisões finais consolidadas

- Direção IN/OUT definida por `categories.direction`.
- `entries.kind = normal | transfer | adjust`.
- Orçamento mensal em %:
  - baseado em renda base do mês (`is_budget_base=true`).
- Orçamento consome apenas OUT relevantes (`is_budget_relevant=true`).
- Investimentos como account separada:
  - aportes/resgates como transfer.
- Cartão completo:
  - compra cria plano e parcelas,
  - orçamento consome no posting mensal de parcelas,
  - pagamento fatura fora do orçamento.

---

## 14) Apêndice: categorias técnicas recomendadas (resumo)

### Investimentos

- Aportes Investimentos (OUT, budget_relevant = true se quiser orçar 10%)
- Entrada Investimentos (IN, budget_base=false)
- Resgate Investimentos (OUT, budget_relevant=false)
- Entrada Resgate (IN, budget_base=false)
- Rendimentos (IN, budget_base=false)
- (opcional) Perdas (OUT, budget_relevant=false)

### Cartão

- Pagamento Fatura Cartão (OUT, budget_relevant=false)
- Entrada Pagamento Cartão (IN, budget_base=false)
- Categorias reais (Mercado, etc.) continuam OUT, budget_relevant=true

---

Fim.
