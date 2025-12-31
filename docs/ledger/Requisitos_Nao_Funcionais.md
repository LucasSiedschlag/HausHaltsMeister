# Requisitos Não Funcionais — HausHaltsMeister (Fluxo de Caixa + Orçamento % + Investimentos + Cartão)

> Versão: 1.0  
> Escopo: requisitos não funcionais (qualidades do sistema)  
> Contexto-alvo: aplicação **local-first/offline-first**, rodando em ambiente pessoal, com evolução possível para multiusuário e, futuramente, SaaS.

---

## Sumário

1. Qualidade e metas do produto
2. Performance e eficiência
3. Confiabilidade e consistência de dados
4. Disponibilidade e operação
5. Segurança (não funcional)
6. Privacidade
7. Usabilidade e acessibilidade
8. Observabilidade e suporte
9. Manutenibilidade e evolutividade
10. Portabilidade e implantação
11. Backup e recuperação
12. Testabilidade e qualidade
13. Compatibilidade e versionamento
14. Restrições e trade-offs assumidos

---

# 1) Qualidade e metas do produto

## RNF-001 — Simplicidade operacional (local-first)
O sistema deve poder ser operado localmente com o mínimo de dependências externas, preservando as principais funcionalidades sem internet.

**Critérios**
- as operações de criação/consulta de dados não dependem de serviços externos
- o banco PostgreSQL é local ou acessível na rede local (conforme setup do usuário)

---

## RNF-002 — Previsibilidade e transparência
O sistema deve ser previsível na forma de calcular:
- orçamento mensal
- renda base
- consumo por categoria
- saldo por conta

**Critérios**
- os cálculos devem ser reprodutíveis e auditáveis via consultas ao journal

---

# 2) Performance e eficiência

## RNF-010 — Tempo de resposta (UI/API)
O sistema deve responder rapidamente para as operações comuns (em ambiente local).

**Metas sugeridas**
- CRUD de lançamento (transaction+entries): p95 < 150ms (local)
- painel mensal de orçamento: p95 < 300ms (local) para até ~10k entries
- listagem de transações paginada: p95 < 250ms

> Valores são guias; a meta real depende do hardware. Em cenário local, a experiência deve ser “instantânea” na maioria dos casos.

---

## RNF-011 — Eficiência de consultas
O sistema deve usar índices adequados e consultas orientadas a `ledger_id` + intervalos de data.

**Critérios**
- queries críticas devem evitar full scan para uso normal
- índices recomendados:
  - `transactions(ledger_id, occurred_at)`
  - `entries(ledger_id, account_id)`
  - `entries(ledger_id, category_id)`
  - `installments(ledger_id, due_month, status)`
  - `credit_card_statements(card_account_id, statement_month)`

---

## RNF-012 — Escalabilidade de dados (pessoal)
O sistema deve suportar sem degradação severa um volume típico pessoal:

**Referência**
- até 100k entries ao longo dos anos
- até 20 categorias principais com subcategorias
- até 5 cartões e dezenas de planos parcelados ativos

---

# 3) Confiabilidade e consistência de dados

## RNF-020 — Atomicidade do journal
Operações que criam/alteram uma transaction com suas entries devem ser atômicas.

**Critérios**
- se falhar qualquer parte, nada deve persistir (rollback)

---

## RNF-021 — Integridade referencial
O sistema deve garantir consistência de referências entre ledger/account/category/transaction/entry.

**Critérios**
- não pode existir entry apontando para account/category de outro ledger
- validações no service layer são obrigatórias; constraints podem reforçar

---

## RNF-022 — Idempotência em rotinas automáticas
Operações automatizadas (ex.: posting de parcelas) devem ser idempotentes.

**Critérios**
- executar 2x não deve duplicar entries
- uso de status (`scheduled -> posted`) e `posted_transaction_id` é obrigatório

---

## RNF-023 — Auditabilidade (baseline)
O sistema deve manter rastreabilidade suficiente para entender “quem fez o quê”.

**Critérios mínimos**
- `created_by_user_id` em transaction
- `created_at/updated_at`
- `memo` e/ou `notes` para ajustes/estornos

**Evolução desejável**
- `audit_events` (tabela) ou logs estruturados persistentes

---

# 4) Disponibilidade e operação

## RNF-030 — Operação offline/local
A aplicação deve funcionar totalmente em ambiente local, inclusive relatórios e orçamento.

---

## RNF-031 — Resiliência a falhas
Falhas durante operação (ex.: queda de energia durante gravação) não devem corromper dados.

**Critérios**
- transações SQL garantem consistência
- recomenda-se usar `fsync`/config padrão do Postgres apropriada

---

# 5) Segurança (não funcional)

## RNF-040 — Autenticação segura
O sistema deve armazenar senhas de forma segura.

**Critérios**
- hash forte (Argon2id/bcrypt) com parâmetros adequados
- nunca armazenar senha em texto

---

## RNF-041 — Autorização por ledger (anti-vazamento)
O sistema deve impedir acesso a dados de outro ledger.

**Critérios**
- todas as rotas exigem `ledger_id`
- todas as queries filtram por `ledger_id`
- validações de ownership/membership antes de retornar dados

---

## RNF-042 — Princípio do menor privilégio
O sistema deve usar papéis (viewer/editor/owner) e garantir que cada ação exija a permissão apropriada.

---

## RNF-043 — Proteção contra CSRF/XSS (se web)
Se houver UI web:
- deve haver proteção CSRF quando aplicável
- saída deve ser escapada para prevenir XSS
- cookies de sessão devem ser `HttpOnly`, `Secure` (quando TLS)

---

## RNF-044 — Criptografia em trânsito e em repouso
**Local-first:** pode ser opcional, mas recomendado:
- TLS para acesso remoto
- criptografia de disco (responsabilidade do ambiente) ou opção de DB encryption (futuro)

---

# 6) Privacidade

## RNF-050 — Minimização de dados
O sistema deve coletar apenas dados necessários ao funcionamento.

---

## RNF-051 — Exportação/portabilidade dos dados
O sistema deve permitir exportar dados do usuário (por ledger) em formato portátil.

**Critérios**
- exportar CSV/JSON no mínimo
- (opcional) backup completo via dump do Postgres

---

# 7) Usabilidade e acessibilidade

## RNF-060 — UX consistente para lançamentos
A UI deve:
- permitir lançamentos rápidos (1-3 cliques)
- suportar split facilmente
- diferenciar categorias reais vs técnicas (cartão, transferências)

---

## RNF-061 — Feedback e prevenção de erro
O sistema deve avisar quando:
- percentual do orçamento passa de 100% (warning)
- tentativa de postar parcelas já postadas (idempotência)
- tentativa de usar categoria inativa

---

## RNF-062 — Acessibilidade (se UI web)
Metas sugeridas:
- contraste adequado
- navegação por teclado
- labels em inputs

---

# 8) Observabilidade e suporte

## RNF-070 — Logging estruturado
O sistema deve gerar logs com:
- timestamp
- user_id (quando aplicável)
- ledger_id
- action/endpoint
- status/erro
- latência

---

## RNF-071 — Diagnóstico de falhas
Erros devem ser:
- rastreáveis com IDs de correlação (request_id)
- retornados ao cliente sem vazar informações sensíveis (stack traces)

---

# 9) Manutenibilidade e evolutividade

## RNF-080 — Modularidade do código
O código deve ser organizado por módulos de domínio (journal, budget, creditcard, etc.).

**Critérios**
- dependências claras: módulos “de cima” dependem do journal
- evitar acoplamento circular

---

## RNF-081 — Separação de camadas (pragmática)
O backend deve separar:
- HTTP handlers
- services (regras/validações)
- repositories/sqlc (acesso a dados)

---

## RNF-082 — Compatibilidade de schema e migrations
O schema deve evoluir por migrations com versionamento.

**Critérios**
- migrations imutáveis após publicadas
- rollback não é obrigatório, mas forward-only deve ser suportado

---

# 10) Portabilidade e implantação

## RNF-090 — Deploy local simples
A aplicação deve poder rodar com:
- um binário Go (recomendado)
- Postgres local (instalado no host) ou container (opcional)

---

## RNF-091 — Configuração por arquivo/variáveis
O sistema deve suportar configuração via:
- `.env` / variáveis de ambiente
- arquivo de configuração (opcional)

---

# 11) Backup e recuperação

## RNF-100 — Backup manual e automático (recomendado)
O sistema deve suportar:
- backup manual (dump)
- agendamento de backup (futuro)

**Critérios**
- procedimento documentado
- restauração testável

---

## RNF-101 — Recuperação após falha
Deve ser possível restaurar o banco e a aplicação voltar a operar sem perda de integridade.

---

# 12) Testabilidade e qualidade

## RNF-110 — Cobertura mínima de regras críticas
Deve existir suite de testes para:
- validações do journal (transfer balance)
- cálculo do budget mensal
- posting idempotente de parcelas
- cálculo de renda base

---

## RNF-111 — Testes de integração com Postgres
Deve existir testes que validem:
- transações atômicas
- constraints essenciais
- queries agregadas principais

---

# 13) Compatibilidade e versionamento

## RNF-120 — Versionamento semântico
O sistema deve usar versionamento semântico para releases (ex.: semver).

---

## RNF-121 — Compatibilidade com dados antigos
Migrations devem preservar dados existentes; mudanças destrutivas exigem migração explícita.

---

# 14) Restrições e trade-offs assumidos

1) O sistema não é contabilidade formal — é um journal leve voltado a fluxo pessoal.
2) Orçamento é flexível por renda real; não há “orçamento fixo” persistido por mês.
3) Flags de categoria são retroativas (MVP). Versionar flags é uma evolução futura.
4) Cálculo de saldo do cartão é uma interpretação (dívida = OUT - IN), não um campo persistido.
5) A idempotência e segurança dependem inicialmente do service layer; hardening no DB pode ser fase 2.

---

Fim.