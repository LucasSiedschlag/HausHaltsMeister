# Core do sistema (Users, Ledgers, Journal, Entries, Accounts) — Regras de negócio + Pseudo-fluxo

Este documento define as regras e os fluxos centrais do sistema:

- Autenticação/Usuários (conceitual)
- Ledgers (multiusuário presente/futuro)
- Membership e permissões
- Accounts (Conta Corrente / Wallet / Investimentos)
- Journal (transactions + entries)
- Regras de integridade e consistência (ledger isolation)
- Padrões de lançamento (normal / transfer / adjust)
- Pseudo-fluxos para os serviços principais

---

## 0) Objetivos e princípios

### Objetivo funcional

1. Permitir que múltiplos usuários usem o sistema, cada um com seus dados isolados.
2. Permitir (no futuro) que um ledger seja compartilhado (múltiplos membros).
3. Registrar todo fluxo financeiro no journal (transactions + entries), com histórico auditável.
4. Dar suporte a:
   - lançamentos simples
   - lançamentos split (uma transação com várias linhas)
   - transferências entre contas internas
   - ajustes/estornos
5. Evitar inconsistência e vazamento de dados entre usuários/ledgers.

### Princípios do modelo

- **Ledger é o boundary de dados**: tudo relevante aponta para `ledger_id`.
- **Transaction é o evento** (cabeçalho); **Entry é a linha de valor**.
- Criação de transaction e entries deve ocorrer em uma **transação SQL** (atomicidade).
- `entries.amount_cents` é sempre positivo; sinal é inferido de `categories.direction` (IN/OUT).
- `accounts` são “caixinhas internas” do ledger (não equivalem necessariamente a bancos).
- `entries.kind` classifica o papel da linha: `normal`, `transfer`, `adjust`.

---

## 1) Entidades do core e responsabilidades

### 1.1. Users

Representa o usuário autenticado.

Responsabilidades:

- autenticação (email/senha ou outro método)
- identidade do criador de eventos (created_by_user_id)
- status ativo/inativo

### 1.2. Ledgers

Representa o “espaço financeiro” (domínio de dados) de um usuário (ou grupo no futuro).

Responsabilidades:

- isolamento de dados
- definição da moeda (BRL)
- possuir contas internas, categorias, budget plans, journal

### 1.3. Ledger Members

Permite múltiplos usuários em um ledger com papéis.

Papéis (role):

- owner: gerencia membros e configurações
- editor: cria/edita lançamentos
- viewer: somente leitura

### 1.4. Accounts

“Caixinhas” internas dentro do ledger.

Tipos (`accounts.type`):

- current (Conta Corrente)
- business (Empresarial)
- investment (Investimentos)
- exchange (Exchange)
- wallet (Wallet/Pessoal)

Natureza (`accounts.nature`):

- asset (padrao)
- liability

Responsabilidades:

- segmentar saldo e relatórios
- permitir transferências internas
- definir sinal/normal balance via `nature`

### 1.5. Journal (Transactions + Entries)

A fonte da verdade para o histórico financeiro.

- `transactions`: evento (data, descrição, notas, autor, origem/import)
- `transactions.investment_action`: classificador opcional para investimentos
- `entries`: linhas de valor (conta, categoria, valor, memo, kind)

---

## 2) Regras de isolamento e permissão (multiusuário)

### 2.1. Princípio: Tudo é escopado por ledger

Qualquer consulta/ação deve sempre filtrar por `ledger_id`.

### 2.2. Regra de acesso

Um usuário pode acessar um ledger se:

- é owner do ledger (`ledgers.owner_user_id = user_id`) OU
- está em `ledger_members` com role adequado

Pseudo:

- `has_access(user_id, ledger_id) = owner || member`

### 2.3. Regras por role (mínimo recomendado)

- viewer:
  - listar ledger, contas, categorias, relatórios
  - listar transactions/entries
- editor:
  - criar/editar/deletar lançamentos (transactions/entries)
  - criar categorias (dependendo do produto)
- owner:
  - tudo do editor
  - gerenciar membros
  - ajustar configurações (moeda, etc.)

> Implementação inicial pode considerar apenas owner/editor para simplificar.

---

## 3) Regras do Journal (transactions + entries)

### 3.1. Ordem de criação

**Sempre**: criar `transaction` primeiro, depois as `entries`.

Motivos:

- `transaction` é o cabeçalho do evento: data, descrição, autor, rastreio
- `entries` dependem da `transaction_id`
- permite split natural

### 3.2. Atomicidade

Criação de:

- 1 transaction + N entries
  deve ser feita em **uma transação SQL** (BEGIN/COMMIT).
  Se falhar, rollback.

### 3.3. Número mínimo de entries

- transação normal: >= 1 entry
- transação split: >= 2 entries
- transferência: recomendado exatamente 2 entries (1 OUT + 1 IN)
- ajuste: normalmente 1 entry (mas pode ser múltiplo)

> A regra “exatamente 2 para transfer” pode ser validação de aplicação.

### 3.4. Campo `memo` em entries

Uso:

- detalhe por linha (especialmente em split)
- justificativa em ajustes
- referência humana (“parcela 2/5”, “correção duplicado”)

### 3.5. Campos de import (external_source/external_id)

Uso:

- evitar duplicidade em importações (CSV, OFX etc.)
- permitir reprocessamento idempotente

Regra recomendada:

- uniqueness lógica por (ledger_id, external_source, external_id)
- ao importar, se já existe, não criar novamente

---

## 4) Consistência entre ledger/account/category

### 4.1. Regra: account deve pertencer ao mesmo ledger

- `entries.account_id` deve referenciar uma account cujo `accounts.ledger_id == entries.ledger_id`

### 4.2. Regra: categoria deve pertencer ao mesmo ledger

- `entries.category_id` deve referenciar uma categoria cujo `categories.ledger_id == entries.ledger_id`

### 4.3. Regra: entry deve pertencer ao mesmo ledger da transaction

- `entries.ledger_id == transactions.ledger_id`

### 4.4. Onde validar?

Recomendação:

- validar na aplicação (service layer) sempre
- opcional futuro: constraints/triggers no Postgres

---

## 5) Padrões de lançamento (normal / transfer / adjust)

### 5.1. `kind=normal`

Uso:

- entradas e saídas comuns
- gastos splitados
- lançamentos cotidianos

### 5.2. `kind=transfer`

Uso:

- movimentação entre contas internas do mesmo ledger
  - Wallet/Pessoal -> Investimentos (aporte interno)
  - Conta pagadora -> Passivo do cartão (pagamento fatura)
  - Investimentos -> Wallet/Pessoal (resgate)
    Regras recomendadas:
- sem `investment_action`:
  - 1 entry OUT (categoria direction=out)
  - 1 entry IN (categoria direction=in)
  - total OUT == total IN
- com `investment_action`:
  - todas as entries devem ter o mesmo `amount_cents`
  - direction pode ser diferente (ex.: categorias Investimentos (Entrada) e Investimentos (Saída))

### 5.3. `kind=adjust`

Uso:

- correções e estornos
- rendimento/perda de investimento (via `investment_action`)
- juros/multa (se preferir tratar como ajuste)
  Regras:
- pode ser 1 ou várias linhas
- deve ser rastreável via memo/notes

---

## 6) Cálculo de saldos e totalizações (conceitual)

### 6.1. Saldo por account (visão geral)

Saldo pode ser derivado do journal:

- Para accounts com `nature=asset`:
  - saldo = SUM(IN) - SUM(OUT)
- Para accounts com `nature=liability`:
  - saldo interpretado como “dívida”:
    - saldo = SUM(OUT) - SUM(IN)
      (OUT aumenta a divida; IN reduz)

> Isso é regra de apresentação/relatório. O banco não precisa armazenar saldo.

### 6.2. IN/OUT por categoria

- total IN de categoria = soma entries dessa categoria onde categories.direction=in
- total OUT de categoria = soma entries onde categories.direction=out

---

## 7) Fluxos de negócio (pseudo-fluxos)

## 7.1. `SignUpUser` (conceitual)

Input: email, password, display_name

- validate email unique
- hash password
- create user

Após signup, criar ledger padrão:

- `CreateDefaultLedgerForUser`

---

## 7.2. `CreateDefaultLedgerForUser`

Input: user_id
Em transação:

1. create ledger:
   - owner_user_id = user_id
   - name = "Pessoal"
   - currency_code = "BRL"
2. create ledger_members (opcional) com role owner
3. create accounts padrão:
   - Conta Corrente (current, asset)
   - Wallet/Pessoal (wallet, asset)
   - (opcional no onboarding) Investimentos (investment, asset)
   - (cartões são adicionados depois)
4. create categorias seed (mínimo) (opcional)
5. create budget_plan inicial (opcional)

Output: ledger_id

---

## 7.3. `AddLedgerMember` (futuro)

Input: owner_user_id, ledger_id, new_user_id, role

- verify requester is owner
- insert ledger_members unique(ledger_id, user_id)

---

## 7.4. `CreateAccount`

Input: user_id, ledger_id, name, type

- verify user has role owner/editor
- validate unique name per ledger
- insert accounts

---

## 7.5. `CreateTransaction` (core)

Input:

- user_id, ledger_id
- occurred_at, description, notes?
- entries[]: (account_id, category_id, kind, amount_cents, memo?)

Validar:

1. user has editor access to ledger
2. description not empty
3. entries length >= 1
4. para cada entry:
   - amount_cents > 0
   - account belongs to ledger
   - category belongs to ledger (se category_id não nulo)
   - kind válido
5. regras adicionais por kind:
   - se houver entries kind=transfer:
     - total OUT == total IN (usando categories.direction)
     - pelo menos uma IN e uma OUT
   - se kind=adjust:
     - exigir memo/notes (opcional, mas recomendado)

Persistir (transação SQL):

1. INSERT transactions (...) RETURNING id
2. INSERT entries (transaction_id, ledger_id, ...) para cada entry
3. COMMIT

Output: transaction_id

---

## 7.6. `UpdateTransaction` (recomendação de política)

Existem 2 políticas possíveis:

### Política A — Editável (mais simples)

- permitir editar description/occurred_at/notes
- permitir substituir entries (delete+insert) ou atualizar individualmente
- manter updated_at

### Política B — Append-only (auditável)

- não alterar entries antigas
- fazer correções com novas transactions kind=adjust
- permitir editar apenas notas superficiais

Recomendação inicial:

- Política A para MVP (prático)
- Política B pode ser adotada depois, especialmente para import e auditoria

---

## 7.7. `DeleteTransaction` (recomendação)

Políticas:

### A) Delete hard (simples)

- deletar entries e transaction (com cascata)
- perde auditabilidade

### B) Soft delete (recomendado)

- adicionar `is_deleted` em transactions/entries ou `deleted_at`
- esconder do UI/relatórios
- manter rastreio

Recomendação:

- soft delete para evitar “sumir” com histórico sem querer

> Este é um ponto de schema opcional. Se não quiser agora, comece com hard delete.

---

## 7.8. `ListTransactions`

Input: user_id, ledger_id, filters (date range, account, category, search)

- verify access
- query transactions + aggregate entries (sum, count)
- return paginated list

---

## 7.9. `GetTransactionDetails`

Input: user_id, ledger_id, transaction_id

- verify access
- fetch transaction header
- fetch entries
- return details

---

## 8) Seeds e padrões mínimos recomendados

### 8.1. Accounts seed

- Conta Corrente (current, asset)
- Wallet/Pessoal (wallet, asset)
- Investimentos (investment, asset) (opcional no onboarding)

### 8.2. Categorias sugeridas (onboarding opcional)

- Gastos fixos
- Conforto
- Lazer
- Investimentos
- Objetivos
- Educação
- Salário (in, budget_base=true)
- Extra (in, budget_base=true)
- Terceiros (in, budget_base=true)

> A classificacao de investimentos e feita via `transactions.investment_action`.

---

## 9) Observações de implementação (sem código, mas guia prático)

### 9.1. Sempre passar ledger_id

Todas as rotas/queries devem exigir `ledger_id` e validar membership.
Não aceitar “buscar por transaction_id” sem ledger_id.

### 9.2. Denormalização de ledger_id em entries

`entries.ledger_id` existe para:

- facilitar filtros
- reduzir risco de inconsistência
- acelerar consultas

Mas a aplicação deve garantir consistência com transaction/account/category.

### 9.3. Índices essenciais

- transactions: (ledger_id, occurred_at)
- entries: (ledger_id, account_id), (ledger_id, category_id)
- ledger_members: unique(ledger_id, user_id)

---

## 10) Resultado final (o que este core garante)

Multiusuário por ledger com possibilidade de compartilhamento (roles).  
Journal robusto (transactions + entries) com suporte a split.  
Transferências e ajustes como padrões formais (kind).  
Isolamento forte por ledger (sem vazamento de dados).  
Base consistente para orçamento, investimentos e cartão.
