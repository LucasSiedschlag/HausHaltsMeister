# Documento de Casos de Uso

## Sistema de Controle Financeiro Pessoal (Haus Halts Meister)

---

## 1. Introdução

Este documento descreve **os casos de uso do sistema**, ou seja, **como o usuário interage com o sistema para atingir objetivos reais** do seu controle financeiro.

Os casos de uso estão organizados por **áreas funcionais**, refletindo os módulos do domínio:

- Fluxo de Caixa
- Categorias
- Gastos Fixos e Variáveis
- Cartões e Parcelamentos
- Orçamento
- Picuinhas
- Relatórios e Dashboards

Cada caso de uso descreve:

- Objetivo
- Ator
- Pré-condições
- Fluxo principal
- Fluxos alternativos
- Pós-condições
- Regras relevantes

---

## 2. Atores

### Ator Principal

- **Usuário**  
  Único usuário do sistema, responsável por todos os lançamentos, configurações e análises.

Não há outros atores (sistema é single-user e manual).

---

## 3. Casos de Uso — Fluxo de Caixa

---

### UC-01 — Registrar Entrada de Ganho

**Objetivo:**  
Registrar uma entrada simples de dinheiro (salário, freela, venda pontual).

**Ator:**  
Usuário

**Pré-condições:**

- Existir ao menos uma categoria de entrada ativa (“Ganho”).

**Fluxo Principal:**

1. Usuário acessa a tela de “Nova Entrada”.
2. Informa:
   - Data (ou aceita o mês atual)
   - Categoria = Ganho
   - Nome do ganho
   - Valor
3. Confirma o lançamento.
4. Sistema cria um registro de fluxo de caixa de entrada.

**Fluxos Alternativos:**

- Valor inválido (≤ 0) → sistema rejeita.
- Categoria inativa → sistema bloqueia seleção.

**Pós-condições:**

- Entrada registrada no mês correspondente.
- Fluxo de caixa do mês é atualizado.

---

### UC-02 — Registrar Entrada de Investimento

**Objetivo:**  
Registrar dinheiro proveniente de resgate de investimento.

**Pré-condições:**

- Categoria “Investimento” (entrada) ativa.

**Fluxo Principal:**

1. Usuário acessa “Nova Entrada”.
2. Seleciona categoria “Investimento”.
3. Informa nome e valor.
4. Confirma.

**Regras Específicas:**

- Entrada é automaticamente considerada no cálculo de reinvestimento.
- Não exige detalhes do investimento.

**Pós-condições:**

- Entrada registrada.
- Sistema atualiza indicador “quanto deve ser reinvestido”.

---

### UC-03 — Registrar Saída Simples

**Objetivo:**  
Registrar uma saída simples de dinheiro.

**Fluxo Principal:**

1. Usuário acessa “Nova Saída”.
2. Informa:
   - Data
   - Categoria de saída
   - Nome do gasto
   - Valor
3. Confirma.

**Fluxos Alternativos:**

- Categoria incompatível com saída → bloqueio.
- Valor inválido → rejeição.

**Pós-condições:**

- Saída registrada no fluxo de caixa.

---

## 4. Casos de Uso — Categorias

---

### UC-04 — Criar Nova Categoria

**Objetivo:**  
Criar uma nova categoria de entrada ou saída.

**Fluxo Principal:**

1. Usuário acessa “Gerenciar Categorias”.
2. Cria nova categoria informando:
   - Nome
   - Direção (IN ou OUT)
   - Relevante para Orçamento (Sim/Não)
3. Salva.

**Pós-condições:**

- Categoria disponível para lançamentos futuros.
- Não afeta dados históricos.

---

### UC-05 — Desativar Categoria

**Objetivo:**  
Impedir uso futuro de uma categoria sem perder histórico.

**Fluxo Principal:**

1. Usuário desativa uma categoria existente.
2. Sistema marca categoria como inativa.

**Regras:**

- Categoria não pode ser excluída se houver uso histórico.

**Pós-condições:**

- Categoria não aparece em novos lançamentos.
- Histórico permanece íntegro.

---

## 5. Casos de Uso — Gastos Fixos e Variáveis

---

### UC-06 — Registrar Gasto Fixo

**Objetivo:**  
Registrar um gasto recorrente.

**Fluxo Principal:**

1. Usuário registra uma saída.
2. Marca como “Gasto Fixo”.
3. Confirma.

**Pós-condições:**

- Gasto pode ser sugerido automaticamente no mês seguinte.

---

### UC-07 — Copiar Gastos Fixos para Novo Mês

**Objetivo:**  
Facilitar preenchimento do mês.

**Fluxo Principal:**

1. Usuário inicia novo mês.
2. Sistema sugere copiar gastos fixos do mês anterior.
3. Usuário confirma.
4. Sistema cria novos lançamentos ajustáveis.

**Regras:**

- Cópia nunca ocorre automaticamente sem confirmação.

---

## 6. Casos de Uso — Cartões e Parcelamentos

---

### UC-08 — Registrar Compra Parcelada no Cartão

**Objetivo:**  
Registrar compra parcelada em cartão de crédito.

**Fluxo Principal:**

1. Usuário acessa “Nova Compra Parcelada”.
2. Informa:
   - Nome da compra
   - Cartão
   - Valor total
   - Número de parcelas
   - Mês inicial
3. Confirma.

**Pós-condições:**

- Sistema cria:
  - Um plano de parcelamento
  - Um fluxo de caixa mensal por parcela
- Parcelas futuras aparecem como compromissos.

---

### UC-09 — Visualizar Fatura do Cartão

**Objetivo:**  
Ver total da fatura por mês.

**Fluxo Principal:**

1. Usuário seleciona cartão e mês.
2. Sistema soma todas as parcelas do mês.
3. Exibe total e lista de compras.

---

## 7. Casos de Uso — Orçamento

---

### UC-10 — Criar Orçamento Mensal

**Objetivo:**  
Planejar distribuição do dinheiro do mês.

**Fluxo Principal:**

1. Usuário acessa “Orçamento”.
2. Seleciona mês.
3. Define valores ou percentuais por categoria.
4. Confirma.

---

### UC-11 — Alterar Orçamento de Um Mês

**Objetivo:**  
Ajustar estratégia financeira pontual.

**Fluxo Principal:**

1. Usuário edita orçamento do mês.
2. Altera valores ou percentuais.
3. Salva.

**Regras:**

- Alterações não afetam meses fechados.

---

### UC-12 — Alterar Orçamento em Lote

**Objetivo:**  
Aplicar nova estratégia para vários meses futuros.

**Fluxo Principal:**

1. Usuário seleciona múltiplos meses.
2. Aplica novos percentuais.
3. Confirma.

---

## 8. Casos de Uso — Picuinhas

---

### UC-13 — Cadastrar Pessoa de Picuinha

**Objetivo:**  
Registrar alguém com quem há movimentações recorrentes.

**Fluxo Principal:**

1. Usuário cadastra nova pessoa.
2. Informa nome e observações.

---

### UC-14 — Registrar Empréstimo

**Objetivo:**  
Registrar dinheiro emprestado a alguém.

**Fluxo Principal:**

1. Usuário cria entrada de picuinha:
   - Pessoa
   - Tipo: Empréstimo
   - Valor
2. Sistema registra saldo da pessoa.

**Pós-condições:**

- Saldo da pessoa aumenta.

---

### UC-15 — Registrar Pagamento de Picuinha

**Objetivo:**  
Registrar dinheiro devolvido ao usuário.

**Fluxo Principal:**

1. Usuário registra pagamento.
2. Sistema cria:
   - Entrada no fluxo de caixa
   - Entrada de picuinha vinculada

**Pós-condições:**

- Saldo da pessoa diminui.

---

### UC-16 — Registrar Compra no Cartão para Outra Pessoa

**Objetivo:**  
Registrar compra parcelada no cartão para terceiro.

**Fluxo Principal:**

1. Usuário registra compra parcelada.
2. Associa a uma pessoa.
3. Sistema:
   - Cria parcelamento
   - Registra dívida total da pessoa

---

### UC-17 — Consultar Saldo de Picuinha por Pessoa

**Objetivo:**  
Saber quanto alguém deve ou tem crédito.

**Fluxo Principal:**

1. Usuário seleciona pessoa.
2. Sistema exibe:
   - Saldo atual
   - Histórico completo

---

## 9. Casos de Uso — Relatórios e Dashboards

---

### UC-18 — Visualizar Dashboard Mensal

**Objetivo:**  
Obter visão geral do mês.

**Fluxo Principal:**

1. Usuário seleciona mês.
2. Sistema exibe:
   - Entradas
   - Saídas
   - Saldo
   - Comparativo com mês anterior

---

### UC-19 — Visualizar Fluxo de Caixa Acumulado

**Objetivo:**  
Analisar evolução financeira no tempo.

**Fluxo Principal:**

1. Usuário acessa “Fluxo Acumulado”.
2. Sistema exibe gráfico de saldo ao longo dos meses.

---

### UC-20 — Visualizar Análise por Categoria

**Objetivo:**  
Analisar comportamento financeiro por categoria.

**Fluxo Principal:**

1. Usuário seleciona categoria.
2. Sistema exibe:
   - Histórico mensal
   - Orçado x realizado
   - Tendência

---

## 10. Considerações Finais

Este conjunto de casos de uso cobre:

- Uso diário (lançamentos rápidos)
- Uso estratégico (orçamento e análise)
- Situações complexas (parcelamentos, picuinhas)

O sistema deve ser capaz de evoluir adicionando novos casos de uso **sem alterar os existentes**, respeitando sempre:

- Integridade histórica
- Simplicidade operacional
- Clareza de domínio

---

**Fim do Documento**
