# Refatoracao de contas e cartoes (v2)

## Objetivo

- Remover tipos antigos `cash` e `credit_card`.
- Adotar `accounts.type`: `current`, `business`, `investment`, `exchange`, `wallet`.
- Adicionar `accounts.nature`: `asset` | `liability` (default `asset`).
- Permitir varios cartoes por conta (1:N) com `credit_cards.id` como entidade principal.
- Faturas, planos e parcelas passam a usar `credit_card_id`.

## Decisoes principais

- `credit_cards` possui:
  - `parent_account_id` (conta "pai" do cartao, natureza `asset`)
  - `liability_account_id` (conta passivo obrigatoria, `nature=liability`)
- `transactions.credit_card_id` registra compras feitas no cartao.
- Pagamento da fatura usa transferencia:
  - OUT na conta que paga
  - IN na conta de passivo do cartao
- Conta de passivo criada automaticamente para cada cartao:
  - `type=current`, `nature=liability`

## Seeds default

- Conta Corrente (current, asset)
- Wallet/Pessoal (wallet, asset)
- Opcional no onboarding: "Criar Investimentos" => cria conta Investimentos (investment, asset)

## Impactos

- Atualizar docs, migrations, SQLC, services/handlers, reports, frontend e testes.
- Remover rotas antigas por `cardAccountId`; tudo passa a usar `card_id`.

## Breaking changes

- Remocao de `accounts.type=credit_card` e `accounts.type=cash`.
- Substituicao de `card_account_id` por `credit_card_id` em faturas/planos/parcelas.
- Remocao de endpoints legados por `cardAccountId` (tudo por `card_id`).
