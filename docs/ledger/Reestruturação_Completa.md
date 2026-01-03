# Reestruturação Completa — HausHaltsMeister (Ledger leve + Orçamento % + Investimentos + Cartão)

> Documento base da reestruturação geral do projeto. Consolida o modelo final do domínio, o schema e os fluxos essenciais.
> Referências detalhadas vivem em `docs/ledger/` e são citadas por seção.

---

## Sumário

1. Escopo, objetivos e fontes
2. Princípios e decisões do modelo
3. Core do domínio (Users/Ledgers/Journal/Accounts/Categories)
4. Orçamento (% flexível e versionado)
5. Investimentos
6. Cartão de crédito
7. DBML — Schema atualizado (modelo final)
8. Guia das tabelas (resumo)
9. Fluxos operacionais (resumo)
10. Categorias técnicas e seeds recomendadas
11. Segurança, integridade e validações
12. Relatórios e consultas principais
13. Plano de migração e próximos passos

---

## 1) Escopo, objetivos e fontes

Este documento é a referência principal para a reestruturação do projeto. Ele consolida as decisões do modelo e aponta para documentos específicos:

- Arquitetura geral e visão de produto: `docs/ledger/Documento_de_Arquitetura.md`.
- Regras do core (ledger/journal/accounts): `docs/ledger/Regras_Core_Ledger.md`.
- Categorias e orçamento: `docs/ledger/Regras_Categorias_e_Orçamento.md`.
- Investimentos: `docs/ledger/Regras_Investimentos.md`.
- Cartão de crédito: `docs/ledger/Regras_Cartão_de_crédito.md`.
- Segurança e validações: `docs/ledger/Regras_Segurança.md`.
- Casos de uso: `docs/ledger/Casos_de_uso.md`.
- Requisitos: `docs/ledger/Requisitos_Funcionais.md` e `docs/ledger/Requisitos_Nao_Funcionais.md`.
- Plano de migração: `docs/ledger/Plano_de_Migracao_Backend.md`.
- Blueprint de implementação (estrutura/stack): `docs/ledger/Backend_Blueprint.md`.

Objetivo central: migrar para um modelo **journal leve** (transactions + entries) com orçamento por % e módulos de investimentos e cartão, mantendo consistência, auditabilidade e evolução futura (multiusuário).

---

## 2) Princípios e decisões do modelo

Referência: `docs/ledger/Documento_de_Arquitetura.md`.

- **Ledger é a fronteira de dados**: tudo pertence a um `ledger_id`.
- **Journal é a fonte da verdade**: `transactions` (evento) + `entries` (linhas).
- **Direção IN/OUT vem da categoria** (`categories.direction`), não de entries.
- **Orçamento é percentual** e calculado a partir da **renda base do mês**.
- **Orçamento é versionado** e não reescreve o passado.
- **Cartão consome orçamento na parcela postada**, não no ato da compra.
- **Contas são internas** (cash/investment/credit_card), não bancos externos.

---

## 3) Core do domínio (Users/Ledgers/Journal/Accounts/Categories)

Referência: `docs/ledger/Regras_Core_Ledger.md`.

### 3.1 Ledger e membership

- `ledgers` representa o “universo financeiro” do usuário.
- `ledger_members` prepara o caminho para compartilhamento (roles: owner/editor/viewer).

### 3.2 Journal (transactions + entries)

- `transactions` guarda data, descrição, autor e metadados de import.
- `entries` guarda valor, conta, categoria, memo e kind.
- Sempre criar **transaction** antes de **entries**, em transação SQL.

### 3.3 Accounts (contas internas)

- `cash` (Pessoal), `investment` (Investimentos), `credit_card` (Cartão).
- Saldos são derivados do journal, não armazenados.

### 3.4 Categories

- `direction` define IN/OUT.
- `is_budget_base` (IN) e `is_budget_relevant` (OUT) controlam orçamento.
- `parent_id` permite hierarquia (pai/filho) e orçamento com `include_children`.

---

## 4) Orçamento (% flexível e versionado)

Referência: `docs/ledger/Regras_Categorias_e_Orçamento.md`.

### 4.1 Como calcular o mês

- Mês é sempre `YYYY-MM-01`.
- **Renda base do mês** = soma de IN com `is_budget_base=true`.
- **Limite por categoria** = renda_base_mes \* (percentual / 100).
- **Realizado** = soma de OUT com `is_budget_relevant=true`.

### 4.2 Versionamento

- `budget_plans` (1 por ledger).
- `budget_plan_versions` com `effective_from_month`.
- `budget_plan_lines` com percentuais e `include_children`.
- Alterar orçamento = criar nova versão, sem reescrever meses anteriores.

---

## 5) Investimentos

Referência: `docs/ledger/Regras_Investimentos.md`.

### 5.1 Conta dedicada

- `accounts.type=investment` separa caixa de investimentos.

### 5.2 Fluxos padrão

- Aporte: transferência Pessoal -> Investimentos.
- Resgate: transferência Investimentos -> Pessoal.
- Rendimento: entry IN na conta de investimentos (`kind=adjust` recomendado).

### 5.3 Flags recomendadas

- Entradas técnicas e rendimentos **não** devem inflar renda base.

---

## 6) Cartão de crédito

Referência: `docs/ledger/Regras_Cartão_de_crédito.md`.

### 6.1 Entidades

- `credit_cards` (metadados do cartão).
- `installment_plans` + `installments` (parcelamento).
- `credit_card_statements` (fatura mensal).

### 6.2 Regra-chave

- **Orçamento é consumido apenas quando a parcela é postada**.
- Pagamento de fatura é transferência com categorias técnicas (fora do orçamento).

---

## 7) DBML — Schema atualizado (modelo final)

> Este DBML é o modelo consolidado do domínio. Ajuste nomes de campos se as migrations evoluírem.
> A referência de comportamento está nos documentos listados nas seções anteriores.
> Observação: os enums abaixo são conceituais. Na implementação, foram trocados por `varchar` + `CHECK` para permitir evolução futura sem refatoração de tipos.

```dbml
// =====================
// Enums
// =====================
Enum account_type {
  cash
  investment
  credit_card
}

Enum category_direction {
  in
  out
}

Enum entry_kind {
  normal
  transfer
  adjust
}

Enum ledger_role {
  owner
  editor
  viewer
}

Enum installment_plan_status {
  active
  cancelled
  finished
}

Enum installment_status {
  scheduled
  posted
  paid
  skipped
}

Enum statement_status {
  open
  closed
  paid
}

// =====================
// Core
// =====================
Table users {
  id            uuid [pk]
  email         varchar [not null, unique]
  display_name  varchar
  avatar_url    varchar
  email_verified_at timestamptz
  is_active     boolean [not null, default: true]
  created_at    timestamptz [not null]
  updated_at    timestamptz
}

// =====================
// Auth
// =====================
Table auth_secrets {
  user_id       uuid [pk, ref: > users.id]
  password_hash varchar [not null]
  created_at    timestamptz [not null]
  updated_at    timestamptz
}

Table auth_identities {
  id               uuid [pk]
  user_id          uuid [not null, ref: > users.id]
  provider         varchar [not null] // validated via CHECK in DB
  provider_user_id varchar [not null]
  email            varchar
  display_name     varchar
  avatar_url       varchar
  email_verified   boolean [not null, default: false]
  created_at       timestamptz [not null]
  updated_at       timestamptz
  last_login_at    timestamptz

  Indexes {
    (provider, provider_user_id) [unique]
    (user_id, provider) [unique]
    (user_id)
    (provider)
  }
}

Table user_preferences {
  user_id         uuid [pk, ref: > users.id]
  theme_mode      varchar [not null, default: system] // light|dark|system
  theme_palette   varchar [not null, default: default]
  theme_tone      varchar [not null, default: vivid]
  locale          varchar [not null, default: pt-BR]
  compact_mode    varchar [not null, default: comfortable]
  font_scale      varchar [not null, default: md]

  notify_card_close  boolean [not null, default: true]
  notify_budget_over boolean [not null, default: true]
  notify_payables    boolean [not null, default: true]

  created_at      timestamptz [not null]
  updated_at      timestamptz
}

Table auth_sessions {
  id                      uuid [pk]
  user_id                 uuid [not null, ref: > users.id]
  refresh_token_hash      varchar [not null]
  created_at              timestamptz [not null]
  updated_at              timestamptz
  expires_at              timestamptz [not null]
  revoked_at              timestamptz
  user_agent              varchar
  ip                      varchar
  device_name             varchar
  rotated_from_session_id uuid [ref: > auth_sessions.id]

  Indexes {
    (refresh_token_hash) [unique]
    (user_id)
    (expires_at)
    (revoked_at)
  }
}

Table oauth_states {
  id            uuid [pk]
  provider      varchar [not null] // validated via CHECK in DB
  state         varchar [not null]
  code_verifier varchar [not null]
  redirect_uri  varchar
  created_at    timestamptz [not null]
  updated_at    timestamptz
  expires_at    timestamptz [not null]
  used_at       timestamptz

  Indexes {
    (state) [unique]
    (provider)
    (expires_at)
  }
}

Table ledgers {
  id              uuid [pk]
  owner_user_id   uuid [not null, ref: > users.id]
  name            varchar [not null]
  currency_code   varchar [not null, default: "BRL"]
  created_at      timestamptz [not null]
  updated_at      timestamptz
}

Table ledger_members {
  ledger_id   uuid [not null, ref: > ledgers.id]
  user_id     uuid [not null, ref: > users.id]
  role        ledger_role [not null, default: viewer]
  created_at  timestamptz [not null]
  updated_at  timestamptz

  Indexes {
    (ledger_id, user_id) [unique]
  }
}

Table audit_log {
  id          uuid [pk]
  ledger_id   uuid [not null, ref: > ledgers.id]
  user_id     uuid [not null, ref: > users.id]
  action      varchar [not null]
  entity_id   uuid
  ip          varchar
  user_agent  varchar
  created_at  timestamptz [not null]
  updated_at  timestamptz

  Indexes {
    (ledger_id, created_at)
  }
}

Table accounts {
  id          uuid [pk]
  ledger_id   uuid [not null, ref: > ledgers.id]
  name        varchar [not null]
  type        account_type [not null]
  is_active   boolean [not null, default: true]
  created_at  timestamptz [not null]
  updated_at  timestamptz

  Indexes {
    (ledger_id, name) [unique]
    (ledger_id)
  }
}

Table categories {
  id                uuid [pk]
  ledger_id          uuid [not null, ref: > ledgers.id]
  parent_id          uuid [ref: > categories.id]
  name              varchar [not null]
  direction         category_direction [not null]
  is_budget_base    boolean [not null, default: false] // meaningful when direction=in
  is_budget_relevant boolean [not null, default: true] // meaningful when direction=out
  is_active         boolean [not null, default: true]
  created_at        timestamptz [not null]
  updated_at        timestamptz

  Indexes {
    (ledger_id, name) [unique]
    (ledger_id)
    (parent_id)
  }
}

Table transactions {
  id                 uuid [pk]
  ledger_id           uuid [not null, ref: > ledgers.id]
  occurred_at         timestamptz [not null]
  description         varchar [not null]
  notes               text
  created_by_user_id  uuid [not null, ref: > users.id]
  external_source     varchar
  external_id         varchar
  created_at          timestamptz [not null]
  updated_at          timestamptz

  Indexes {
    (ledger_id, occurred_at)
    (ledger_id, external_source, external_id) [unique]
  }
}

Table idempotency_keys {
  ledger_id     uuid [not null, ref: > ledgers.id]
  key           varchar [not null]
  resource_type varchar [not null]
  resource_id   uuid [not null]
  request_hash  varchar [not null]
  created_at    timestamptz [not null]
  updated_at    timestamptz
  expires_at    timestamptz [not null]

  Indexes {
    (ledger_id, key) [pk]
  }
}

Table entries {
  id             uuid [pk]
  ledger_id       uuid [not null, ref: > ledgers.id]
  transaction_id  uuid [not null, ref: > transactions.id]
  account_id      uuid [not null, ref: > accounts.id]
  category_id     uuid [ref: > categories.id]
  kind            entry_kind [not null, default: normal]
  amount_cents    bigint [not null]
  memo            text
  created_at      timestamptz [not null]
  updated_at      timestamptz

  Indexes {
    (ledger_id, transaction_id)
    (ledger_id, account_id)
    (ledger_id, category_id)
  }
}

// =====================
// Budget (% )
// =====================
Table budget_plans {
  id         uuid [pk]
  ledger_id  uuid [not null, ref: > ledgers.id]
  name       varchar [not null, default: "Default"]
  created_at timestamptz [not null]
  updated_at timestamptz

  Indexes {
    (ledger_id) [unique]
  }
}

Table budget_plan_versions {
  id                   uuid [pk]
  plan_id              uuid [not null, ref: > budget_plans.id]
  effective_from_month date [not null] // always YYYY-MM-01
  created_by_user_id   uuid [not null, ref: > users.id]
  created_at           timestamptz [not null]
  updated_at           timestamptz

  Indexes {
    (plan_id, effective_from_month) [unique]
  }
}

Table budget_plan_lines {
  id             uuid [pk]
  version_id     uuid [not null, ref: > budget_plan_versions.id]
  category_id    uuid [not null, ref: > categories.id]
  percent        numeric [not null] // e.g. 10.0
  include_children boolean [not null, default: false]
  created_at     timestamptz [not null]
  updated_at     timestamptz

  Indexes {
    (version_id, category_id) [unique]
  }
}

// =====================
// Credit Card metadata (1:1)
// =====================
Table card_networks {
  code        varchar [pk]
  display_name varchar [not null]
  created_at  timestamptz [not null]
  updated_at  timestamptz
}

Table credit_cards {
  account_id         uuid [pk, ref: > accounts.id]
  issuer_name        varchar
  network            varchar [not null, default: "other", ref: > card_networks.code]
  nickname           varchar
  last4              char(4)
  credit_limit_cents bigint
  closing_day        int [not null]
  due_day            int [not null]
  created_at         timestamptz [not null]
  updated_at         timestamptz
}

// =====================
// Credit Card Installments + Statements
// =====================
Table installment_plans {
  id                   uuid [pk]
  ledger_id            uuid [not null, ref: > ledgers.id]
  card_account_id      uuid [not null, ref: > accounts.id] // type=credit_card

  purchase_occurred_at timestamptz [not null]
  merchant             varchar
  description          varchar [not null]

  category_id          uuid [not null, ref: > categories.id] // real OUT category

  total_amount_cents       bigint [not null]
  installments_count       int [not null]
  installment_amount_cents bigint [not null]

  first_due_month      date [not null] // YYYY-MM-01
  status               installment_plan_status [not null, default: active]

  created_by_user_id   uuid [not null, ref: > users.id]
  created_at           timestamptz [not null]
  updated_at           timestamptz

  Indexes {
    (ledger_id)
    (card_account_id)
    (ledger_id, first_due_month)
  }
}

Table credit_card_statements {
  id                   uuid [pk]
  ledger_id            uuid [not null, ref: > ledgers.id]
  card_account_id      uuid [not null, ref: > accounts.id]

  statement_month      date [not null] // YYYY-MM-01
  closing_date         date [not null]
  due_date             date [not null]

  total_charges_cents  bigint [not null, default: 0]
  total_payments_cents bigint [not null, default: 0]

  status               statement_status [not null, default: open]
  payment_transaction_id uuid [ref: > transactions.id]

  created_at           timestamptz [not null]
  updated_at           timestamptz

  Indexes {
    (card_account_id, statement_month) [unique]
    (ledger_id)
  }
}

Table installments {
  id                    uuid [pk]
  ledger_id             uuid [not null, ref: > ledgers.id]
  plan_id               uuid [not null, ref: > installment_plans.id]

  installment_no        int [not null]
  due_month             date [not null] // YYYY-MM-01
  amount_cents          bigint [not null]
  status                installment_status [not null, default: scheduled]

  posted_transaction_id uuid [ref: > transactions.id]
  paid_statement_id     uuid [ref: > credit_card_statements.id]

  created_at            timestamptz [not null]
  updated_at            timestamptz

  Indexes {
    (plan_id, installment_no) [unique]
    (ledger_id, due_month)
    (status)
  }
}
```

---

## 8) Guia das tabelas (resumo)

Referência: `docs/ledger/Regras_Core_Ledger.md` e `docs/ledger/Regras_Cartão_de_crédito.md`.

- `users`: identidade e autenticação (base do usuário).
- `auth_secrets`: credenciais locais (senha), separadas da identidade.
- `auth_identities`: vínculos com provedores (password, google, github).
- `user_preferences`: preferências do usuário (tema, idioma, notificações, densidade).
- `auth_sessions`: sessões e refresh tokens.
- `oauth_states`: estados temporários de OAuth (PKCE + anti-CSRF).
- `ledgers`: escopo de dados; moeda e owner.
- `ledger_members`: compartilhamento com roles.
- `accounts`: contas internas (`cash`, `investment`, `credit_card`).
- `categories`: direção IN/OUT e flags do orçamento.
- `transactions`: cabeçalho do evento.
- `entries`: linhas financeiras; kind `normal`, `transfer`, `adjust`.
- `budget_*`: plano, versões e linhas por categoria.
- `card_networks`: catálogo de bandeiras de cartão.
- `credit_cards`: metadados 1:1 com conta de cartão.
- `installment_plans` + `installments`: compras parceladas e parcelas.
- `credit_card_statements`: fatura mensal e pagamento.

---

## 9) Fluxos operacionais (resumo)

Referência: `docs/ledger/Casos_de_uso.md`.

### 9.1 Lançamento normal

1. Criar `transactions` (data, descrição).
2. Criar `entries` (1 ou N linhas).
3. Categoria define IN/OUT; orçamento usa flags.

### 9.2 Transferência interna

1. Criar `transactions`.
2. Criar 2 `entries` com `kind=transfer` (uma IN e uma OUT).
3. Validar balanceamento (IN == OUT).

### 9.3 Orçamento mensal

1. Identificar `budget_plan_version` vigente do mês.
2. Calcular renda base (IN com `is_budget_base=true`).
3. Calcular limite por categoria (%).
4. Calcular realizado (OUT com `is_budget_relevant=true`).

### 9.4 Parcelas do cartão

1. Compra cria `installment_plan` + `installments` (scheduled).
2. No mês alvo, postar parcelas -> `transactions` + `entries` no cartão.
3. Pagamento de fatura é transferência com categorias técnicas.

---

## 10) Categorias técnicas e seeds recomendadas

Referência: `docs/ledger/Regras_Investimentos.md` e `docs/ledger/Regras_Cartão_de_crédito.md`.

### 10.1 Cartão

- Pagamento Fatura Cartão (OUT, `is_budget_relevant=false`)
- Entrada Pagamento Cartão (IN, `is_budget_base=false`)

### 10.2 Investimentos

- Aportes Investimentos (OUT, `is_budget_relevant=false` por padrão; habilitar se for orçar)
- Entrada Investimentos (Aporte) (IN, `is_budget_base=false`)
- Resgate Investimentos (OUT, `is_budget_relevant=false`)
- Entrada Resgate (Investimentos) (IN, `is_budget_base=false`)
- Rendimentos (IN, `is_budget_base=false`)
- (Opcional) Perdas (OUT, `is_budget_relevant=false`)

---

## 11) Segurança, integridade e validações

Referência: `docs/ledger/Regras_Segurança.md`.

- Sempre validar acesso por `ledger_id` e role.
- Garantir consistência entre entry/account/category/transaction.
- Transferências devem ser balanceadas (IN == OUT).
- Rotinas automáticas (posting de parcelas) devem ser idempotentes.
- Evitar endpoints sem `ledger_id` (anti-vazamento).

---

## 12) Relatórios e consultas principais

Referência: `docs/ledger/Regras_Categorias_e_Orçamento.md` e `docs/ledger/Regras_Investimentos.md`.

- Renda base do mês (IN com `is_budget_base=true`).
- Orçado vs realizado por categoria (mensal e por período).
- Gastos fora do orçamento (`is_budget_relevant=false`).
- Aportes, resgates e rendimentos por período.
- Fatura mensal do cartão e totais do mês.

---

## 13) Plano de migração e próximos passos

Referência principal: `docs/ledger/Plano_de_Migracao_Backend.md`.

- Migrar base do ledger e garantir `make migrate`/`make sqlc`.
- Implementar core (journal + categories) antes de budget/cartão.
- Ajustar serviços para usar postings do novo modelo.
- Atualizar docs e contratos quando cada fase concluir.
- Usar `docs/ledger/Backend_Blueprint.md` para estrutura de pastas e stack, mantendo o modelo deste documento como fonte de verdade.

---

Fim.
