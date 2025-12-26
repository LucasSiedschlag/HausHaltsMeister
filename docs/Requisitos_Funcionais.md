# Documento de Requisitos Funcionais e Regras de Negócio

## Sistema de Controle Financeiro Pessoal (HausHaltsMeister)

---

## 1. Visão Geral do Sistema

O sistema tem como objetivo permitir o **controle financeiro pessoal completo**, priorizando:

- Simplicidade de lançamento
- Flexibilidade extrema de modelagem
- Fidelidade ao fluxo real de dinheiro
- Separação clara entre:
  - Dinheiro **que é seu**
  - Dinheiro **que apenas passa pela sua conta** (“picuinhas”)
- Capacidade analítica (orçamento, comparativos, histórico)
- Evolução incremental (começa simples, cresce sem refatorações traumáticas)

O sistema é **single-user**, **web**, **manual**, e será utilizado tanto como ferramenta pessoal quanto como **ambiente de prática de engenharia de software**.

---

## 2. Princípios Fundamentais (Regras Mestras)

Esses princípios **regem todas as decisões técnicas e funcionais**:

1. **O núcleo do sistema é o fluxo de caixa**, não contas bancárias.
2. **Entradas devem ser extremamente simples** de registrar.
3. Nenhum lançamento deve exigir campos irrelevantes.
4. Regras de negócio vivem no domínio, não no frontend.
5. O histórico nunca pode ser quebrado:
   - Categorias não são apagadas, apenas desativadas.
   - Dados antigos nunca mudam significado.
6. O sistema deve permitir **lançamentos incompletos**, desde que coerentes.
7. Tudo que impacta decisões financeiras deve aparecer no fluxo de caixa.
8. O sistema deve permitir análise mensal **e acumulada**.
9. O usuário pode mudar sua estratégia mês a mês sem quebrar o sistema.
10. “Picuinhas” são cidadãos de primeira classe no sistema, não exceções.

---

## 3. Conceitos-Chave do Domínio

### 3.1 Fluxo de Caixa (`cash_flows`)

Representa **qualquer entrada ou saída de dinheiro** que afeta, direta ou indiretamente, a sua vida financeira.

Características:

- Toda movimentação relevante passa por aqui.
- É o núcleo imutável do sistema.
- Não carrega detalhes desnecessários.

Um fluxo de caixa é composto por:

- Data (ou mês de referência)
- Direção (entrada ou saída)
- Categoria
- Nome/Título
- Valor

---

### 3.2 Categorias de Fluxo

Categorias representam **o “porquê” do dinheiro**, não o “como”.

Regras:

- Cada categoria possui uma **direção fixa**:
  - Entrada (`IN`)
  - Saída (`OUT`)
- Uma categoria nunca pode ser usada em direção diferente da definida.
- Categorias podem ser:
  - Ativas
  - Inativas (históricas)
- Categorias podem ser marcadas como **Relevantes para Orçamento** (ex: "Investimentos" pode não ser um gasto orçamentário, mas uma transferência patrimonial).

Exemplos:

- Entrada:
  - Ganho
  - Investimento (entrada vinda de investimento)
- Saída:
  - Custos Fixos
  - Conforto
  - Prazeres
  - Veículo / Transporte
  - Picuinhas

---

## 4. Requisitos Funcionais — Fluxo de Caixa

### RF-01 — Criar entrada de dinheiro (ganho)

O sistema deve permitir o registro de entradas simples de dinheiro, como:

- Salário
- Freela
- Venda pontual
- Ganhos diversos

Campos obrigatórios:

- Data
- Categoria (somente categorias de entrada)
- Nome do ganho
- Valor

Regras:

- Valor deve ser maior que zero
- Nenhuma forma de pagamento é exigida
- Nenhum campo adicional deve ser obrigatório

---

### RF-02 — Criar entrada de investimento

O sistema deve permitir registrar entradas provenientes de investimentos (resgates).

Regras específicas:

- A categoria “Investimento” (entrada) implica automaticamente:
  - Que o valor **deve ser considerado no cálculo de quanto deve ser reinvestido**
- Não é necessário informar:
  - Origem do investimento
  - Tipo de ativo
  - Corretora

Esses detalhes serão tratados em um módulo futuro.

---

### RF-03 — Criar saída simples de dinheiro

O sistema deve permitir registrar saídas simples, como:

- Pagamentos à vista
- Gastos cotidianos
- Despesas sem parcelamento

Campos obrigatórios:

- Data
- Categoria (somente categorias de saída)
- Nome do gasto
- Valor

Campos opcionais:

- Forma de pagamento
- Indicação de gasto fixo

---

## 5. Requisitos Funcionais — Gastos Fixos e Variáveis

### RF-04 — Classificar gasto como fixo

O sistema deve permitir marcar uma saída como **gasto fixo**.

Regras:

- Gastos fixos podem ser copiados automaticamente para meses seguintes
- O valor pode ser ajustado mês a mês
- A cópia nunca deve ocorrer sem confirmação explícita

---

### RF-05 — Gastos variáveis

Qualquer saída que não seja marcada como fixa é considerada variável.

Regras:

- Gastos variáveis não são copiados automaticamente
- Devem aparecer apenas no mês em que foram registrados

---

## 6. Requisitos Funcionais — Cartões e Parcelamentos

### RF-06 — Registrar compra parcelada no cartão

O sistema deve permitir registrar compras parceladas realizadas em cartão de crédito.

Regras:

- O usuário informa:
  - Nome da compra
  - Valor total
  - Número de parcelas
  - Cartão (banco)
  - Mês inicial da fatura
- O sistema gera automaticamente:
  - Um plano de parcelamento
  - Um fluxo de caixa de saída para cada mês/parcela

Cada parcela:

- Deve impactar o mês correto
- Deve aparecer na fatura do cartão correspondente

---

### RF-07 — Projeção de gastos futuros

O sistema deve:

- Mostrar parcelas futuras como compromissos (saídas previstas)
- Permitir visualização do impacto futuro no fluxo de caixa
- Não misturar gastos futuros com gastos já realizados (distinção visual ou de status)

---

## 7. Requisitos Funcionais — Orçamento

### RF-08 — Criar orçamento mensal

O sistema deve permitir criar um orçamento para um mês específico.

O orçamento é composto por:

- Itens por categoria de saída
- Valores absolutos ou percentuais da renda (percentual calculado sobre entradas com categoria relevante para orçamento)

---

### RF-09 — Modos de orçamento

O sistema deve suportar dois modos de análise:

- **Mensal**: considera apenas o mês atual
- **Somatório**: considera o acumulado dos meses anteriores

---

### RF-10 — Alteração de orçamento por mês

O usuário pode alterar:

- Percentuais
- Valores
- Estratégia de orçamento

Sem impacto retroativo em meses já fechados.

---

### RF-11 — Alteração de orçamento em lote

O sistema deve permitir:

- Aplicar uma nova regra de orçamento a vários meses futuros
- Manter histórico intacto

---

## 8. Requisitos Funcionais — Picuinhas

### RF-12 — Cadastro de pessoas

O sistema deve permitir cadastrar pessoas associadas a picuinhas:

- Familiares
- Amigos
- Terceiros

---

### RF-13 — Registrar empréstimos e dívidas

O sistema deve permitir registrar:

- Dinheiro emprestado
- Compras feitas no cartão para terceiros
- Ajustes de saldo

Regras:

- O saldo da pessoa é calculado dinamicamente
- Não existe “quitação automática”
- Valores podem variar livremente

---

### RF-14 — Pagamentos irregulares

O sistema deve permitir:

- Pagamentos parciais
- Pagamentos fora de ordem
- Novos empréstimos antes da quitação anterior

Tudo deve refletir corretamente no saldo.

---

### RF-15 — Integração com fluxo de caixa

A integração entre Picuinhas e Fluxo de Caixa é **opcional no momento do lançamento**:

- Se o usuário selecionar "Auto-criar Fluxo":
  - Empréstimo (`PLUS`) → cria fluxo de caixa `OUT` na categoria "Picuinhas".
  - Pagamento/Recebimento (`MINUS`) → cria fluxo de caixa `IN` na categoria "Picuinhas".
- Se NÃO selecionar:
  - Apenas o saldo da dívida é ajustado (ex: dívida antiga, compra para terceiro, ajuste de contas).

Picuinhas conceituais (ex.: dívida registrada) **podem não gerar fluxo imediato**.

---

## 9. Requisitos Funcionais — Categorias

### RF-16 — Criar categoria

O usuário pode criar novas categorias a qualquer momento.

Regras:

- Categoria deve ter direção fixa (IN ou OUT)
- Categoria nova só afeta lançamentos futuros

---

### RF-17 — Desativar categoria

O sistema deve permitir desativar categorias.

Regras:

- Categorias desativadas:
  - Não aparecem em formulários novos
  - Permanecem visíveis em dados históricos
- Nunca apagar categorias com uso histórico

---

## 10. Requisitos Funcionais — Relatórios e Dashboards

### RF-18 — Dashboard mensal

O sistema deve exibir:

- Total de entradas
- Total de saídas
- Saldo do mês
- Distribuição por categoria
- Comparativo com mês anterior

---

### RF-19 — Fluxo de caixa acumulado

O sistema deve permitir visualizar:

- Evolução do saldo ao longo do tempo
- Impacto de gastos futuros
- Separação visual entre:
  - Dinheiro próprio
  - Picuinhas

---

### RF-20 — Visões por categoria

O sistema deve permitir visualizar:

- Histórico por categoria
- Orçado x realizado
- Histórico por categoria
- Orçado x realizado
- Tendências de crescimento/redução

---

## 11. Requisitos Funcionais — Manutenção e Sistema

### RF-21 — Execução de Backup

O sistema deve fornecer mecanismos (via CLI ou interface administrativa) para a extração completa dos dados.

- **Formato**: SQL Dump ou formato portável (JSON/CSV) completo.
- **Conteúdo**: Todas as tabelas, incluindo usuários, lançamentos, categorias e configurações.

### RF-22 — Auditoria e Rastreabilidade

O sistema deve garantir a persistência de um rastro de auditoria.

- **Regra**: O sistema deve registrar logs de auditoria para operações críticas.
- **Capacidade**: O sistema deve permitir (ao administrador) a consulta desses logs, seja via banco de dados ou endpoint específico, para resolução de incidentes.

## 11. Regras de Negócio Críticas (Resumo)

- Nenhum lançamento pode contradizer a direção da categoria
- Entradas de investimento sempre impactam o cálculo de reinvestimento
- Orçamento nunca bloqueia gasto; ele apenas indica
- Histórico nunca deve ser alterado retroativamente
- Picuinhas nunca “somem”; tudo é histórico
- Parcelamentos sempre geram múltiplos fluxos de caixa
- Simplicidade de entrada é prioridade absoluta

---

## 12. Considerações Finais

Este sistema **não é um app financeiro comum**.  
Ele é:

- Um **registro fiel da realidade**
- Um **ambiente de decisão**
- Um **laboratório de arquitetura de software**

Qualquer implementação deve respeitar:

- Evolução sem quebra
- Clareza de domínio
- Zero gambiarras
- Zero campos inúteis

---

**Fim do Documento**
