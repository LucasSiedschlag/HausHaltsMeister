# Requisitos Funcionais — HausHaltsMeister (Fluxo de Caixa + Orçamento % + Investimentos + Cartão)

> Versão: 1.0  
> Escopo: requisitos funcionais (o que o sistema deve fazer)  
> Não inclui requisitos não-funcionais (performance, disponibilidade, etc.), exceto quando necessário para clareza de comportamento.

---

## Sumário

1. Usuários e autenticação
2. Ledgers e multiusuário
3. Contas internas (Accounts)
4. Categorias
5. Journal (lançamentos)
6. Orçamento (% flexível mensal)
7. Investimentos
8. Cartão de crédito (parcelas + fatura)
9. Relatórios e consultas
10. Segurança funcional e validações
11. Seeds e configuração inicial
12. Administração básica (futuro)

---

## 1) Usuários e autenticação

### RF-001 — Cadastro de usuário

O sistema deve permitir cadastrar um usuário com:

- email (único)
- senha
- nome/apelido (opcional)

Critérios de aceitação

- não permite email duplicado
- senha é armazenada com hash

---

### RF-002 — Login/Autenticação

O sistema deve permitir que o usuário faça login e obtenha uma sessão válida.

Critérios de aceitação

- usuários inativos não conseguem autenticar
- o sistema identifica o `user_id` em todas as requisições autenticadas

---

### RF-003 — Perfil básico

O sistema deve permitir visualizar e atualizar o nome/apelido do usuário.

---

## 2) Ledgers e multiusuário

### RF-010 — Criação automática de ledger padrão

Ao criar um novo usuário, o sistema deve criar automaticamente um ledger padrão.

Critérios de aceitação

- o ledger deve pertencer ao usuário (owner)
- o ledger deve ter moeda padrão (BRL)
- o ledger deve ser utilizável imediatamente para lançamentos

---

### RF-011 — Acesso por ledger (isolamento)

O sistema deve manter os dados isolados por ledger:

- um usuário deve ver apenas os dados do(s) ledger(s) aos quais tem acesso.

---

### RF-012 — Compartilhamento de ledger (futuro)

O sistema deve permitir que o owner convide outros usuários para um ledger, atribuindo role:

- viewer, editor

Critérios de aceitação

- apenas owner pode gerenciar membros
- não deve existir duplicidade de membership por ledger

---

## 3) Contas internas (Accounts)

### RF-020 — Gerenciar contas internas

O sistema deve permitir criar, listar, editar e desativar contas internas dentro de um ledger.

Tipos obrigatórios

- current
- business
- investment
- exchange
- wallet

Critérios de aceitação

- nome único por ledger
- `nature` obrigatório (`asset`/`liability`)
- contas desativadas não podem ser usadas em novos lançamentos

---

### RF-021 — Metadados de cartão de crédito

O sistema deve permitir cadastrar e manter dados específicos do cartão como entidade própria, incluindo:

- conta pai (`parent_account_id`)
- conta passivo (`liability_account_id`)
- bandeira/brand
- label
- final (last4)
- cvv
- holder_name
- cor/estilo (color/style)
- dia de fechamento
- dia de vencimento

Critérios de aceitação

- deve ser 1:N por conta
- `liability_account_id` deve ter `nature=liability`

---

## 4) Categorias

### RF-030 — Criar e manter categorias

O sistema deve permitir criar, listar, editar e desativar categorias dentro do ledger.

Campos obrigatórios**

- name
- direction: in|out

Campos adicionais**

- parent (hierarquia)
- flags:
  - is_budget_base (para IN)
  - is_budget_relevant (para OUT)

Critérios de aceitação

- nome único por ledger
- categoria inativa não pode ser usada em novos lançamentos

---

### RF-031 — Direção definida pela categoria

O sistema deve determinar se um lançamento é IN ou OUT exclusivamente pela categoria.

Critérios de aceitação

- `entries.amount_cents` deve ser sempre positivo
- totalizações devem usar `categories.direction`

---

### RF-032 — Hierarquia de categorias

O sistema deve suportar categorias pai/filho para fins de agrupamento e (opcionalmente) orçamento.

---

## 5) Journal (lançamentos)

### RF-040 — Criar transação (transaction) com entradas (entries)

O sistema deve permitir registrar lançamentos financeiros através de:

- 1 transaction (evento)
- 1 ou mais entries (linhas)

Campos transaction**

- occurred_at
- description
- notes (opcional)

Campos entry**

- account_id
- category_id (recomendado obrigatório)
- amount_cents
- kind: normal|transfer|adjust
- memo (opcional)

Critérios de aceitação

- criação deve ser atômica (transaction SQL)
- entries devem pertencer ao mesmo ledger da transaction
- account/category devem pertencer ao ledger

---

### RF-041 — Split de lançamento

O sistema deve permitir que uma única transaction tenha múltiplas entries para dividir valores em categorias diferentes.

---

### RF-042 — Transferências internas

O sistema deve permitir registrar transferências internas entre contas do mesmo ledger:

- usando `entries.kind=transfer`

Critérios de aceitação

- deve existir pelo menos uma entry IN e uma OUT
- total IN deve ser igual ao total OUT

---

### RF-043 — Ajustes e estornos

O sistema deve permitir registrar ajustes/estornos:

- usando `entries.kind=adjust`

Critérios de aceitação

- deve permitir incluir memo/notes explicativos
- não exige balanceamento

---

### RF-044 — Listar e consultar lançamentos

O sistema deve permitir listar transactions e seus detalhes, com filtros por:

- período (from/to)
- conta
- categoria
- busca textual (description/memo)

---

### RF-045 — Edição/Exclusão (política inicial)

O sistema deve permitir uma política inicial (MVP) de edição e exclusão de lançamentos.

Recomendação funcional

- permitir editar dados da transaction e suas entries (com validações)
- impedir remover transações vinculadas a parcelas/faturas (cartão) sem tratamento especial

---

## 6) Orçamento (% flexível mensal)

### RF-050 — Plano de orçamento por ledger

O sistema deve permitir que cada ledger tenha um plano de orçamento padrão.

Critérios de aceitação

- 1 plano por ledger

---

### RF-051 — Orçamento versionado por mês de vigência

O sistema deve permitir criar versões do orçamento que entram em vigor a partir de um mês (effective_from_month).

Critérios de aceitação

- versões não devem sobrescrever o passado
- cada versão tem suas linhas (% por categoria)

---

### RF-052 — Linhas do orçamento em percentual

O sistema deve permitir definir percentuais por categoria (normalmente OUT).

Critérios de aceitação

- percent > 0
- categoria deve pertencer ao ledger
- `include_children` deve ser suportado para considerar subcategorias

---

### RF-053 — Cálculo de renda base do mês

O sistema deve calcular a renda base do mês como:

- soma de entries IN em categorias com `is_budget_base=true`

---

### RF-054 — Cálculo do limite orçado por categoria

O sistema deve calcular o limite orçado do mês como:

- `income_base_month * percent`

---

### RF-055 — Realizado por categoria no mês

O sistema deve calcular o realizado por categoria do mês como:

- soma de entries OUT em categorias com `is_budget_relevant=true`

---

### RF-056 — Painel mensal do orçamento

O sistema deve apresentar para um mês:

- renda base
- limite por categoria
- realizado por categoria
- delta (acima/abaixo)
- (opcional) gastos fora do orçamento

---

### RF-057 — Visão por período (anual/custom)

O sistema deve permitir consolidar orçamento e realizado em um período (ex.: anual), respeitando:

- renda variável mês a mês
- versões diferentes de orçamento dentro do período

---

## 7) Investimentos

### RF-060 — Conta de investimentos

O sistema deve suportar uma conta interna do tipo investment para separar saldo e lançamentos.

---

### RF-061 — Aporte (Pessoal -> Investimentos)

O sistema deve permitir registrar aportes como transferência interna:

- OUT em Pessoal
- IN em Investimentos
  com categorias técnicas apropriadas.

Critérios de aceitação

- transferência balanceada
- opcionalmente, aporte pode consumir orçamento (se categoria relevante=true)

---

### RF-062 — Resgate (Investimentos -> Pessoal)

O sistema deve permitir registrar resgates como transferência interna inversa.

Critérios de aceitação

- a entrada no Pessoal não deve inflar renda base (budget_base=false)

---

### RF-063 — Rendimentos reinvestidos

O sistema deve permitir registrar rendimentos como lançamento IN na conta Investimentos.

Critérios de aceitação

- rendimentos não devem inflar renda base (budget_base=false)
- recomendado kind=adjust

---

### RF-064 — Relatórios de investimentos por período

O sistema deve permitir consultar por período:

- total aportado
- total resgatado
- total rendimentos
- saldo derivado da conta de investimentos

---

## 8) Cartão de crédito (parcelas + fatura)

### RF-070 — Compra no cartão como plano de parcelas

O sistema deve permitir cadastrar uma compra no cartão (inclusive parcelada) criando:

- installment_plan
- installments (scheduled)

Critérios de aceitação

- categoria da compra deve ser OUT (consumo real)
- `credit_card_id` deve existir no ledger

---

### RF-071 — Posting mensal das parcelas (gasto real do mês)

O sistema deve permitir “postar” as parcelas do mês, criando para cada parcela:

- transaction + entry no passivo do cartao com a categoria real

Critérios de aceitação

- idempotência: uma parcela não pode ser postada duas vezes
- após posting, a parcela deve guardar `posted_transaction_id`

---

### RF-072 — Orçamento do cartão por parcela

O sistema deve considerar o consumo de orçamento apenas quando a parcela for postada no mês.

Critérios de aceitação

- compra parcelada não consome orçamento no ato
- cada parcela consome orçamento no mês do due_month/posting

---

### RF-073 — Fatura (statement) mensal

O sistema deve suportar criação e fechamento de fatura do cartão por mês:

- statement_month
- closing_date
- due_date
- totais de charges/payments (derivados ou armazenados)

---

### RF-074 — Pagamento de fatura

O sistema deve permitir registrar pagamento de fatura como transferência:

- OUT em Pessoal (categoria técnica fora do orçamento)
- IN no Cartão (categoria técnica não base)

Critérios de aceitação

- não deve duplicar gasto no orçamento
- pagamento deve poder marcar fatura como paid (se total)

---

### RF-075 — Relatórios do cartão

O sistema deve permitir ver por cartão e mês:

- parcelas do mês
- status da fatura (open/closed/paid)
- pagamentos realizados
- total devido

---

## 9) Relatórios e consultas

### RF-080 — Extrato mensal

O sistema deve permitir visualizar extrato por mês:

- transactions com suas entries
- filtros por conta/categoria

---

### RF-081 — Resumo por categoria no período

O sistema deve permitir agrupar e somar valores por categoria em um período.

---

### RF-082 — Saldos por conta (derivado)

O sistema deve permitir exibir saldos por conta, derivados do journal:

- nature=asset: IN - OUT
- nature=liability: OUT - IN (dívida)

---

## 10) Segurança funcional e validações

### RF-090 — Autorização por ledger

O sistema deve validar que o usuário tem acesso ao ledger para qualquer operação.

---

### RF-091 — Integridade ledger/account/category

O sistema deve garantir que:

- entries referenciem apenas accounts e categories do mesmo ledger
- entries pertençam ao ledger da transaction

---

### RF-092 — Validação de transferências

O sistema deve validar que transferências:

- tenham IN e OUT
- sejam balanceadas (IN==OUT)

---

### RF-093 — Idempotência de rotinas automáticas (cartão)

O sistema deve garantir que posting de parcelas e criação de faturas:

- não gere duplicidade de lançamentos

---

## 11) Seeds e configuração inicial

### RF-100 — Seeds mínimas recomendadas

O sistema deve oferecer (opcionalmente) seeds para:

- account Conta Corrente (current)
- account Wallet/Pessoal (wallet)
- account Investimentos (investment) (opcional no onboarding)
- categorias técnicas (pagamento fatura, rendimentos, aportes, etc.)

---

### RF-101 — Configuração inicial guiada (opcional)

O sistema pode oferecer um wizard inicial para:

- criar categorias reais básicas
- configurar orçamento inicial
- configurar cartão(s)

---

## 12) Administração básica (futuro)

### RF-110 — Gerenciar membros e roles (owner-only)

- adicionar membro
- alterar role
- remover membro

---

Fim.
