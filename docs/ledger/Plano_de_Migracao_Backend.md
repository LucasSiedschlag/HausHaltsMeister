# Plano de migracao do backend para o modelo ledger (atualizado)

Objetivo: migrar o backend para o modelo ledger leve (transactions + entries), alinhado aos documentos atuais em `docs/ledger/`.

Referencias principais:
- `docs/ledger/Reestruturação_Completa.md`
- `docs/ledger/Regras_Core_Ledger.md`
- `docs/ledger/Regras_Categorias_e_Orçamento.md`
- `docs/ledger/Regras_Investimentos.md`
- `docs/ledger/Regras_Cartão_de_crédito.md`
- `docs/ledger/Regras_Segurança.md`
- `docs/ledger/Casos_de_uso.md`
- `docs/ledger/Documento_de_Arquitetura.md`

---

## 1) Base de dados e ferramentas

Escopo:
- Garantir migrations e sqlc alinhados ao schema final.

Tarefas:
1. Confirmar `migrations/` como baseline do modelo final.
2. Validar `make migrate`, `make migrate-status` e `make sqlc`.
3. Revisar `sqlc.yaml` para apontar as migrations do ledger.

Validacao:
- Migrations aplicam sem erro.
- sqlc gera codigo sem erro.

---

## 2) Core ledger (users, ledgers, ledger_members, accounts, categories, journal)

Escopo:
- Implementar o core com journal leve.

Tarefas:
1. Criar dominio de `users`, `ledgers`, `ledger_members`, `accounts`, `categories` e `journal`.
2. Implementar repositorios e services para:
   - CRUD de accounts e categories.
   - Criacao de transactions + entries (atomicidade).
3. Implementar validacoes de integridade (ledger boundary, roles e transferencias).
4. Ajustar handlers e DTOs para rotas com `ledger_id`.

Validacao:
- Criacao de lancamento simples e split funciona.
- Transferencia interna balanceada (IN == OUT).
- Acesso por ledger e roles validado (owner/editor/viewer).

---

## 3) Orçamento (% flexivel)

Escopo:
- Planos, versoes e linhas de orçamento, calculo mensal.

Tarefas:
1. Implementar `budget_plans`, `budget_plan_versions`, `budget_plan_lines`.
2. Criar queries para:
   - versao vigente do mes
   - renda base (IN com `is_budget_base=true`)
   - realizado (OUT com `is_budget_relevant=true`)
3. Implementar endpoint de painel mensal.

Validacao:
- Percentuais aplicados sobre renda base real.
- Mudanca de versao nao altera meses anteriores.

---

## 4) Investimentos

Escopo:
- Fluxos de aporte, resgate e rendimentos via journal.

Tarefas:
1. Garantir account `investment` e categorias tecnicas.
2. Implementar fluxos:
   - Aporte: transferencia Pessoal -> Investimentos
   - Resgate: transferencia Investimentos -> Pessoal
   - Rendimento: entry IN com `kind=adjust`

Validacao:
- Entradas tecnicas nao entram na renda base.
- Aporte pode ser orcado se `is_budget_relevant=true`.

---

## 5) Cartao de credito (parcelas + fatura)

Escopo:
- Implementar ciclo completo do cartao no modelo final.

Tarefas:
1. CRUD de `credit_cards` (1:1 com account credit_card).
2. Criar `installment_plans` + `installments` (scheduled).
3. Implementar posting mensal:
   - cria `transactions` + `entries` no cartao
   - marca parcelas como `posted`
4. Implementar `credit_card_statements`:
   - fechamento e pagamento
   - pagamento como transferencia com categorias tecnicas

Validacao:
- Parcela so entra no orcamento quando postada.
- Pagamento de fatura nao duplica gasto.
- Rotinas sao idempotentes (nao duplicam lancamentos/entries).

---

## 6) Relatorios e dashboards

Escopo:
- Atualizar relatorios para usar journal + categorias.

Tarefas:
1. Resumo mensal (IN, OUT, saldo por conta).
2. Orcado vs realizado por categoria.
3. Relatorios de investimentos (aportes, resgates, rendimentos).
4. Fatura do cartao por mes.

Validacao:
- Totais batem com o journal.
- Meses com versoes diferentes de budget sao calculados corretamente.

---

## 7) Migracao de dados legados (se existir)

Escopo:
- Mapear dados antigos para o journal leve.

Tarefas:
1. Mapear entradas/saidas antigas para `transactions` + `entries`.
2. Normalizar categorias e flags (`direction`, `is_budget_base`, `is_budget_relevant`).
3. Garantir consistencia de `ledger_id`.

Validacao:
- Amostragem de dados migrados com comparacao de totais.

---

## 8) Limpeza do legado

Escopo:
- Remover caminhos e tabelas antigas apos migracao.

Tarefas:
1. Remover handlers/repos/queries antigas.
2. Atualizar swagger/contratos.
3. Remover tabelas antigas apenas apos validar migracao.

Validacao:
- `go test ./...` sem falhas.
- Nenhuma rota aponta para tabelas antigas.

---

## 9) Testes e qualidade

Escopo:
- Cobrir regras criticas do dominio.

Tarefas:
1. Testes de transferencias balanceadas.
2. Testes do calculo de budget mensal.
3. Testes de posting idempotente de parcelas.
4. Testes de seguranca (ledger boundary).

Validacao:
- Suites passam localmente.

---

## 10) Atualizacao de documentacao

Escopo:
- Manter docs sincronizados com o modelo implementado.

Tarefas:
1. Revisar `docs/ledger/Reestruturação_Completa.md` se houver mudancas.
2. Atualizar `docs/api/*` com endpoints reais.
3. Garantir consistencia com `docs/ledger/Casos_de_uso.md`.

Validacao:
- Docs coerentes com o comportamento real.

---

Fim.
