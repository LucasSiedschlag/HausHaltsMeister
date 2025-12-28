# Plano de migracao do backend para o modelo ledger

Objetivo: migrar o backend atual para o novo modelo ledger em etapas pequenas, com validacao ao final de cada fase.

## 1) Base de dados e ferramentas

Escopo:
- Confirmar migrations do ledger como baseline unica.
- Garantir make migrate, make sqlc e conexao com DB padrao.

Tarefas:
1. Rodar `make migrate` no banco ledger.
2. Garantir `sqlc.yaml` apontando para `migrations/` (ledger).
3. Criar queries base em `db/queries/ledger.sql`.

Validacao:
- `make migrate` conclui sem erro.
- `make sqlc` gera codigo sem erro.

---

## 2) Dominio ledger (core)

Escopo:
- Implementar o dominio principal (transacoes + postings).
- Repositorio e handlers basicos.

Tarefas:
1. Criar `internal/domain/ledger` (model, ports, service).
2. Criar repositorio `internal/adapters/postgres/ledger_repo.go`.
3. Criar handlers `internal/adapters/http/ledger_handlers.go`.
4. Criar DTOs em `internal/adapters/http/dto/ledger.go`.
5. Rotas:
   - POST /ledger/transactions
   - GET /ledger/transactions?month=YYYY-MM-01
   - GET /ledger/accounts

Validacao:
- Teste manual via curl para criar transacao balanceada.
- `go test ./...` sem falhas (quando possivel).

---

## 3) Categorias (categories)

Escopo:
- Substituir fluxo antigo por novo categories.

Tarefas:
1. Atualizar repos e services para ler/escrever em `categories`.
2. Ajustar handlers e DTOs existentes (Categories).
3. Atualizar queries sqlc em `db/queries/categories.sql`.

Validacao:
- CRUD de categorias funcionando.
- Filtros direction/active respondendo corretamente.

---

## 4) Budget (planejamento)

Escopo:
- Manter budget como camada de planejamento.
- Fazer os calculos usando postings do ledger.

Tarefas:
1. Atualizar queries de budget para usar `ledger_categories`.
2. Implementar calculo de total IN a partir de postings.
3. Garantir validacao 100% (quando percentual).

Validacao:
- Criar orcamento em % e salvar com total 100%.
- Summary bate com soma de postings do mes.

---

## 5) Picuinhas (parties + receivables)

Escopo:
- Trocar saldo baseado em tabelas antigas por postings em conta Receivables.

Tarefas:
1. Criar/garantir conta `Receivables:Picuinhas`.
2. Mapear pessoas para `ledger_parties`.
3. Criar casos com postings (emprestimo, recebimento, cartao).
4. Atualizar endpoints Picuinhas para ler saldo via postings.

Validacao:
- Saldo por pessoa reflete pagamentos/pendencias.
- Casos parcelados geram postings futuros.

---

## 6) Cartoes e parcelamentos (hibrido)

Escopo:
- Implementar payment_methods e installment_plans do novo schema.

Tarefas:
1. CRUD de payment_methods com account_id.
2. Criacao de installment_plans + items.
3. Geracao de postings para parcelas.
4. Endpoint de fatura por mes (fechamento/vencimento).

Validacao:
- Criar compra parcelada gera items.
- Fatura soma itens do periodo correto.

---

## 7) Lancamentos manuais (cashflow)

Escopo:
- Substituir o antigo cashflow por lancamentos no ledger.
- Separar entradas, variaveis, fixos e estornos.

Tarefas:
1. Criar tabela `cashflow_entries` para metadados (payment_method, fixed, reversals).
2. Ajustar endpoints `POST/GET/PUT/DELETE /cashflows` e `POST /cashflows/:id/reverse`.
3. Garantir regra de fixos (categoria Custos fixos) e data por mes (exceto cartao).
4. Replicacao de fixos via `POST /cashflows/copy-fixed`.

Validacao:
- Criar entrada/saida manual (com payment_method).
- Estorno cria transacao inversa.
- Listagem por mes funciona com filtros direction/is_fixed.

---

## 8) Relatorios e dashboards

Escopo:
- Recalcular dashboards usando postings.

Tarefas:
1. Atualizar queries de dashboard para usar ledger_postings.
2. Comparativos mensais baseados em categorias IN/OUT.

Validacao:
- Totais de IN/OUT batem com ledger.
- Saldo mensal consistente.

---

## 9) Limpeza do legado

Escopo:
- Remover codigo, handlers e queries que referenciam o modelo antigo.

Tarefas:
1. Apagar services/repos obsolete.
2. Remover DTOs antigos.
3. Atualizar docs e swagger.

Validacao:
- `go test ./...` sem falhas (quando possivel).
- Nenhum handler aponta para tabelas antigas.

---

## 10) Testes e qualidade

Escopo:
- Criar casos de teste minimos por dominio no novo modelo.

Tarefas:
1. Testes de transacao balanceada.
2. Testes de budget com 100%.
3. Testes de picuinha com parcelas e pagamento.
4. Testes de fatura de cartao.

Validacao:
- Testes de unidade/integracao passam.

---

## 11) Atualizacao de documentacao

Escopo:
- Documentar o novo comportamento.

Tarefas:
1. Atualizar `docs/ledger/Backend_Blueprint.md` se necessario.
2. Atualizar `docs/ledger/Casos_de_uso.md` com ajustes finais.
3. Atualizar `docs/api/*` com novos endpoints.

Validacao:
- Docs consistentes com o comportamento real.
