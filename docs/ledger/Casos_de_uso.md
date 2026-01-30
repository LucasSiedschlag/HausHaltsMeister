# Casos de Uso — HausHaltsMeister (Fluxo de Caixa + Orçamento % + Investimentos + Cartão)

Este documento lista e descreve os principais **casos de uso** do sistema, com:

- atores
- pré-condições
- fluxo principal
- fluxos alternativos
- pós-condições
- dados envolvidos (tabelas)
- observações de validação

> Convenções:
>
> - “Mês” = `YYYY-MM-01` (date no 1º dia do mês)
> - Journal = `transactions` + `entries`
> - IN/OUT é definido por `categories.direction`
> - orçamento consome apenas categorias OUT com `is_budget_relevant=true`
> - renda base considera apenas IN com `is_budget_base=true`

---

## Sumário

1. Usuário e Ledger
2. Accounts
3. Categorias
4. Journal (lançamentos)
5. Orçamento
6. Investimentos
7. Cartão de crédito (parcelas + fatura)
8. Relatórios e consultas por período
9. Administração e segurança

---

## 1) Usuário e Ledger

### UC-001 — Criar conta (Sign Up)

**Ator:** Visitante  
**Pré-condições:** email não cadastrado  
**Fluxo principal:**

1. usuário informa email/senha/nome
2. sistema cria `users`
3. sistema cria ledger padrão (UC-002)  
   **Pós-condições:** usuário autenticável; ledger básico criado  
   **Tabelas:** `users`, `ledgers`, `ledger_members`, `accounts` (seed opcional), `categories` (seed opcional)  
   **Validações:** email único; senha com política mínima

---

### UC-002 — Criar ledger padrão para novo usuário

**Ator:** Sistema (pós UC-001)  
**Pré-condições:** user criado  
**Fluxo principal:**

1. cria `ledgers` (owner=user)
2. cria `ledger_members` (role=owner) (opcional)
3. cria contas padrão:
   - “Conta Corrente” (current)
   - “Wallet/Pessoal” (wallet)
   - “Investimentos” (investment) (opcional no onboarding)
4. cria categorias técnicas seed (opcional)  
   **Pós-condições:** ledger pronto para uso  
   **Tabelas:** `ledgers`, `ledger_members`, `accounts`, `categories`  
   **Validações:** nomes únicos por ledger

---

### UC-003 — Compartilhar ledger com outro usuário (futuro)

**Ator:** Owner  
**Pré-condições:** usuário convidado existe; owner tem acesso ao ledger  
**Fluxo principal:**

1. owner seleciona usuário e role (viewer/editor)
2. sistema cria `ledger_members`  
   **Fluxos alternativos:**

- usuário já é membro → retornar “já existe”  
  **Pós-condições:** membro passa a ter acesso ao ledger  
  **Tabelas:** `ledger_members`  
  **Validações:** owner-only; unique(ledger_id,user_id)

---

## 2) Accounts

### UC-010 — Criar conta interna (Account)

**Ator:** Owner/Editor  
**Pré-condições:** acesso ao ledger  
**Fluxo principal:**

1. usuário informa nome e type (current/business/investment/exchange/wallet)
2. sistema cria `accounts`
   **Pós-condições:** conta disponível para lançamentos  
   **Tabelas:** `accounts`  
   **Validações:** unique(ledger_id,name)

---

### UC-011 — Cadastrar cartao (credit_cards)

**Ator:** Owner/Editor  
**Pré-condições:** conta pai (`parent_account_id`) existe  
**Fluxo principal:**

1. usuário informa conta pai, bandeira/brand, label, last4, holder_name, closing_day, due_day, estilo/cores
2. sistema cria conta passivo (`liability_account_id`) automaticamente (`type=current`, `nature=liability`)
3. sistema salva `credit_cards` (1:N por conta)  
   **Pós-condições:** cartao configurado para faturas e ciclos  
   **Tabelas:** `accounts`, `credit_cards`  
   **Validações:** ledger consistente; `liability_account_id` com `nature=liability`; dias válidos

---

## 3) Categorias

### UC-020 — Criar categoria

**Ator:** Owner/Editor  
**Pré-condições:** acesso ao ledger  
**Fluxo principal:**

1. usuário define nome e direction (in/out)
2. define flags (ou aceita defaults):
   - se IN: is_budget_base default true, is_budget_relevant forçado false
   - se OUT: is_budget_relevant default true, is_budget_base forçado false
3. (opcional) define parent_id
4. sistema cria `categories`  
   **Pós-condições:** categoria utilizável nos lançamentos e no orçamento  
   **Tabelas:** `categories`  
   **Validações:** unique(ledger_id,name)

---

### UC-021 — Desativar categoria

**Ator:** Owner/Editor  
**Pré-condições:** categoria existe  
**Fluxo principal:**

1. usuário marca categoria como inativa
2. sistema seta `is_active=false`  
   **Pós-condições:** categoria não aparece para novos lançamentos (histórico preservado)  
   **Tabelas:** `categories`  
   **Validações:** não quebrar histórico

---

## 4) Journal (lançamentos)

### UC-030 — Lançar transação simples (1 entry)

**Ator:** Editor  
**Pré-condições:** conta e categoria válidas do ledger  
**Fluxo principal:**

1. usuário informa data + descrição
2. escolhe conta e categoria
3. informa valor
4. sistema cria `transactions`
5. sistema cria `entries` (1 linha)  
   **Pós-condições:** evento registrado; aparece em relatórios e orçamento (se aplicável)  
   **Tabelas:** `transactions`, `entries`, `accounts`, `categories`  
   **Validações:** entry.amount>0; refs do mesmo ledger

---

### UC-031 — Lançar transação com split (N entries)

**Ator:** Editor  
**Pré-condições:** categorias válidas  
**Fluxo principal:**

1. usuário informa data + descrição
2. adiciona N linhas (categoria + valor + memo)
3. sistema cria `transactions`
4. cria N `entries`  
   **Pós-condições:** transação agregada, mas detalhada por categoria  
   **Tabelas:** `transactions`, `entries`  
   **Observações:** memo recomendado em cada linha

---

### UC-032 — Registrar transferência interna (kind=transfer)

**Ator:** Editor  
**Pré-condições:** duas contas do mesmo ledger  
**Fluxo principal:**

1. usuário escolhe conta origem e destino e valor
2. sistema cria transaction
3. cria 2 entries:
   - OUT na conta origem (categoria técnica OUT)
   - IN na conta destino (categoria técnica IN)
4. valida balanceamento OUT==IN  
   **Pós-condições:** saldos ajustados sem afetar orçamento (dependendo flags)  
   **Tabelas:** `transactions`, `entries`, `categories`, `accounts`  
   **Validações:** balanceamento; direction via categoria

---

### UC-033 — Registrar ajuste/estorno (kind=adjust)

**Ator:** Editor  
**Pré-condições:** necessidade de correção  
**Fluxo principal:**

1. usuário cria transaction “Ajuste” com notes/memo explicativo
2. cria entry IN ou OUT na conta relevante  
   **Pós-condições:** histórico preservado e corrigido via ajuste  
   **Tabelas:** `transactions`, `entries`  
   **Validações:** recomendar memo/notes

---

## 5) Orçamento

### UC-040 — Configurar orçamento inicial (plano + versão)

**Ator:** Owner/Editor  
**Pré-condições:** categorias OUT existentes  
**Fluxo principal:**

1. usuário escolhe mês inicial (effective_from_month)
2. define percentuais por categoria OUT
3. sistema cria `budget_plans` (se não existir)
4. cria `budget_plan_versions`
5. cria `budget_plan_lines`  
   **Pós-condições:** orçamento válido daquele mês em diante  
   **Tabelas:** `budget_plans`, `budget_plan_versions`, `budget_plan_lines`  
   **Validações:** categorias OUT; percent>0; effective no 1º dia do mês

---

### UC-041 — Alterar orçamento “daqui pra frente”

**Ator:** Owner/Editor  
**Pré-condições:** plano existente  
**Fluxo principal:**

1. usuário define novo effective_from_month
2. define nova lista de percentuais
3. sistema cria nova version e novas lines  
   **Pós-condições:** meses anteriores preservados; futuros usam nova versão  
   **Tabelas:** `budget_plan_versions`, `budget_plan_lines`  
   **Validações:** não sobrescrever versão antiga

---

### UC-042 — Ver painel do orçamento do mês

**Ator:** Viewer/Editor  
**Pré-condições:** versão de orçamento aplicável existe  
**Fluxo principal:**

1. usuário seleciona mês M
2. sistema encontra versão vigente no mês M
3. calcula renda base do mês (IN com is_budget_base=true)
4. calcula limite por categoria (percentual)
5. calcula realizado por categoria (OUT com is_budget_relevant=true)
6. exibe delta (acima/abaixo) e “fora do orçamento”  
   **Pós-condições:** nenhuma (read-only)  
   **Tabelas:** `budget_*`, `transactions`, `entries`, `categories`  
   **Observações:** cartão entra no orçamento apenas por parcelas postadas (UC-063)

---

## 6) Investimentos

### UC-050 — Aportar para investimentos (planejado no orçamento)

**Ator:** Editor  
**Pré-condições:** accounts Wallet/Pessoal e Investimentos existem  
**Fluxo principal:**

1. usuário informa valor e data
2. sistema cria transaction “Aporte”
3. cria 2 entries kind=transfer:
   - OUT em Wallet/Pessoal (categoria Investimentos (Saída))
   - IN em Investimentos (categoria Investimentos (Entrada))  
     **Pós-condições:** saldo migra para investimentos; orçamento pode refletir o aporte  
     **Tabelas:** `transactions`, `entries`, `accounts`, `categories`  
     **Validações:** balanceamento por amount + `investment_action=contribution`

---

### UC-051 — Registrar rendimento reinvestido

**Ator:** Editor  
**Pré-condições:** conta Investimentos existe  
**Fluxo principal:**

1. usuário informa valor do rendimento e data (ex.: último dia do mês)
2. cria transaction “Rendimentos”
3. cria entry na conta Investimentos (categoria Investimentos (Entrada)) com kind=adjust e `investment_action=earnings`  
   **Pós-condições:** saldo investimento aumenta sem inflar orçamento  
   **Tabelas:** `transactions`, `entries`, `categories`

---

### UC-052 — Resgatar investimentos

**Ator:** Editor  
**Pré-condições:** conta Investimentos existe  
**Fluxo principal:**

1. usuário informa valor e data
2. transaction “Resgate”
3. 2 entries kind=transfer:
   - OUT em Investimentos (categoria Investimentos (Saída))
   - IN em Wallet/Pessoal (categoria Investimentos (Entrada))  
     **Pós-condições:** saldo volta ao Wallet/Pessoal sem inflar renda base  
     **Tabelas:** `transactions`, `entries`

---

## 7) Cartão de crédito (parcelas + fatura)

### UC-060 — Cadastrar compra parcelada no cartão

**Ator:** Editor  
**Pré-condições:** account cartão configurada; categoria real OUT escolhida  
**Fluxo principal:**

1. usuário informa data da compra, merchant/descrição, total, parcelas, categoria real
2. sistema determina `first_due_month` (pode aceitar override)
3. cria `installment_plans`
4. cria N `installments` (scheduled)  
   **Pós-condições:** compra registrada e parcelas agendadas; ainda sem consumo de orçamento  
   **Tabelas:** `installment_plans`, `installments`  
   **Validações:** category.direction=out; credit_card_id valido no ledger

---

### UC-061 — Postar parcelas do mês (gerar lançamentos do cartão)

**Ator:** Editor (ou job automático)  
**Pré-condições:** existem parcelas scheduled no mês alvo  
**Fluxo principal:**

1. usuário seleciona `target_month` e cartão
2. sistema lista `installments` scheduled desse mês
3. para cada parcela:
   - cria `transactions` “Parcela n/N - merchant”
   - cria `entries` no cartão com categoria real OUT e amount da parcela
   - seta installment status=posted e posted_transaction_id  
     **Pós-condições:** parcelas viram gasto do mês e **entram no orçamento**  
     **Tabelas:** `installments`, `transactions`, `entries`, `categories`  
     **Validações:** idempotência via status; refs do ledger

---

### UC-062 — Fechar fatura do mês (statement)

**Ator:** Editor  
**Pré-condições:** cartão configurado; mês alvo  
**Fluxo principal:**

1. sistema garante statement para (cartão, statement_month)
2. calcula totals (derivados ou snapshot) a partir das parcelas posted do mês
3. marca status=closed  
   **Pós-condições:** fatura pronta para pagamento  
   **Tabelas:** `credit_card_statements`, `installments`  
   **Validações:** unique(card,month); transitions open->closed

---

### UC-063 — Pagar fatura

**Ator:** Editor  
**Pré-condições:** statement closed; conta pagadora (nature=asset) existe  
**Fluxo principal:**

1. usuário informa data e valor (normalmente total)
2. cria transaction “Pagamento fatura”
3. cria 2 entries kind=transfer:
   - OUT na conta pagadora (Pagamento Fatura Cartão, budget_relevant=false)
   - IN no Cartão (Entrada Pagamento Cartão, budget_base=false)
4. marca statement como paid (se pagamento total)
5. marca installments do mês como paid (paid_statement_id)  
   **Pós-condições:** dívida reduz; orçamento não duplica gasto  
   **Tabelas:** `transactions`, `entries`, `credit_card_statements`, `installments`  
   **Validações:** balanceamento; evitar estourar total (política)

---

### UC-064 — Cancelar compra parcelada

**Ator:** Editor  
**Pré-condições:** plano ativo  
**Fluxo alternativo A (nenhuma parcela postada):**

1. marcar plan status=cancelled
2. marcar installments scheduled como skipped  
   **Fluxo alternativo B (já postou):**
3. criar ajuste/estorno via journal (kind=adjust)
4. manter histórico e reequilibrar  
   **Pós-condições:** parcelas não futuras não serão postadas; histórico preservado  
   **Tabelas:** `installment_plans`, `installments`, `transactions`, `entries`

---

## 8) Relatórios e consultas por período

### UC-070 — Extrato do mês (journal)

**Ator:** Viewer  
**Pré-condições:** ledger acessível  
**Fluxo principal:**

1. usuário escolhe mês
2. sistema lista transactions no período com suas entries
3. permite filtros por conta/categoria/busca  
   **Tabelas:** `transactions`, `entries`, `accounts`, `categories`

---

### UC-071 — Resumo por categoria (período)

**Ator:** Viewer  
**Pré-condições:** dados no journal  
**Fluxo principal:**

1. usuário escolhe período
2. sistema agrupa entries por categoria e direction
3. exibe totais e ranking  
   **Tabelas:** `transactions`, `entries`, `categories`

---

### UC-072 — Aportes por período

**Ator:** Viewer  
**Fluxo:** somar entries com `investment_action=contribution` por período  
**Tabelas:** `entries`, `transactions`, `categories`

---

### UC-073 — Rendimentos por período

**Ator:** Viewer  
**Fluxo:** somar entries com `investment_action=earnings` na conta Investimentos por período  
**Tabelas:** `entries`, `transactions`, `categories`, `accounts`

---

### UC-074 — Orçado vs realizado anual (ou período custom)

**Ator:** Viewer  
**Fluxo principal:**

1. sistema percorre meses do período
2. para cada mês:
   - encontra versão vigente
   - calcula renda base e limites
   - calcula realizado
3. acumula por categoria e total
4. exibe acima/abaixo do esperado  
   **Tabelas:** `budget_*`, `transactions`, `entries`, `categories`

---

### UC-075 — Fatura do cartão do mês

**Ator:** Viewer  
**Fluxo principal:**

1. usuário seleciona cartão e mês
2. sistema mostra:
   - parcelas do mês (installments)
   - total de charges
   - status open/closed/paid
   - link para transaction de pagamento (se existir)  
     **Tabelas:** `credit_card_statements`, `installments`, `transactions`, `entries`

---

## 9) Administração e segurança

### UC-090 — Trocar role de um membro (futuro)

**Ator:** Owner  
**Fluxo:** atualizar `ledger_members.role`  
**Validações:** owner-only

---

### UC-091 — Remover membro do ledger (futuro)

**Ator:** Owner  
**Fluxo:** remover linha em `ledger_members`  
**Validações:** não remover owner; evitar ledger sem owner

---

### UC-092 — Tentativa de acesso indevido (segurança)

**Ator:** Usuário mal-intencionado  
**Fluxo esperado:**

1. usuário tenta acessar recurso de outro ledger
2. sistema valida ledger_id e membership
3. retorna 404/403  
   **Pós-condições:** nenhum dado vazado; evento logado (opcional)  
   **Tabelas:** todas (consulta negada)

---

### Anexo A — Mapeamento rápido de tabelas por módulo

#### Core

- users
- ledgers
- ledger_members
- accounts
- categories
- transactions
- entries

#### Budget

- budget_plans
- budget_plan_versions
- budget_plan_lines

#### Investimentos

- (usa core + categorias técnicas)

#### Cartão

- credit_cards
- installment_plans
- installments
- credit_card_statements
- (usa journal para posting e pagamento)

---

Fim.
