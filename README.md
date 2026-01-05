# HausHaltsMeister

**Sistema de Controle Financeiro de Alta Precisao**

---

### Gestão Financeira Baseada em Dados e Fundamentos Contabeis

A verdadeira gestão financeira não se resume ao aglutinamento de transações bancarias. Ela exige uma compreensão clara do fluxo de caixa, segregação de patrimônio e planejamento estratégico. O **HausHaltsMeister** foi desenvolvido para oferecer uma estrutura profissional para as finanças pessoais, focando na integridade dos dados e na clareza para a tomada de decisão.

Este sistema elimina o ruido das automações imprecisas e devolve ao usuário o controle total sobre a interpretação de cada movimentação financeira.

---

## Pilares do Sistema

### 1. Integridade do Fluxo de Caixa

O núcleo da saúde financeira é a liquidez. O sistema prioriza o registro manual e consciente de cada entrada e saída, garantindo que o fluxo de caixa registrado corresponda fielmente a realidade operacional, sem dependência de integrações bancarias falhas ou categorizações genericas.

### 2. Gestão de Recebíveis e Obrigações

Uma falha comum em sistemas pessoais é misturar patrimônio proprio com recursos transitorios (emprestimos, compras para terceiros, divisão de contas). O HausHaltsMeister trata essas movimentações como **Contas a Receber** e **Contas a Pagar**, garantindo que o saldo disponivel reflita seu patrimonio real, isolando o que é dinheiro de terceiros.

### 3. Planejamento Orçamentário Estratégico

O orçamento não é apenas um limite de gastos, mas uma ferramenta de alocação de recursos. O sistema permite a definição de orçamentos mensais e a análise de variancia (Orçado vs. Realizado), permitindo ajustes táticos na estratégia financeira familiar antes que o periodo se encerre.

### 4. Rastreabilidade e Auditoria

Cada decisão financeira deve ser rastreável. O sistema mantém um histórico imutável e estruturado, permitindo análises longitudinais de evolução patrimonial e correção de rota baseada em dados históricos consolidados.

---

## Fundação do Modelo (Ledger Leve)

O HausHaltsMeister é estruturado sobre um **ledger leve** (journal):

- **Ledger como fronteira de dados**: tudo pertence a um `ledger_id`, com caminho natural para multiusuario.
- **Journal como fonte da verdade**: `transactions` (evento) + `entries` (linhas) registram todo o fluxo.
- **Direção IN/OUT definida pela categoria**: `categories.direction` define a semantica; valores sao sempre positivos.
- **Orçamento flexivel por percentual**: calculado sobre a renda base do mes e versionado no tempo.
- **Contas internas**: Pessoal, Investimentos e Cartao refletem o destino do dinheiro, sem modelar bancos externos.
- **Cartao por parcela**: o consumo entra no orçamento apenas quando a parcela e postada; pagamento de fatura e transferencia tecnica.

---

## Beneficios da Arquitetura

- **Consistencia**: regras unificadas para entradas, saidas e orçamento.
- **Evolucao segura**: novos modulos se integram ao journal sem refatorar o core.
- **Clareza patrimonial**: separacao de fluxos evita inflar saldo e renda base.
- **Transparencia**: relatorios sempre rastreaveis a partir das entradas originais.

---

## Seguranca e Integridade

- Isolamento por ledger em todas as operacoes.
- Validacao de consistencia entre transaction, entry, account e category.
- Transferencias balanceadas por regra de negocio.
- Rotinas idempotentes para cartao (posting de parcelas e fatura).

---

## Documentacao do Modelo

A documentacao completa esta em `docs/ledger/`:

- Modelo final e DBML: `docs/ledger/Reestruturação_Completa.md`
- Arquitetura: `docs/ledger/Documento_de_Arquitetura.md`
- Core ledger: `docs/ledger/Regras_Core_Ledger.md`
- Categorias e orçamento: `docs/ledger/Regras_Categorias_e_Orçamento.md`
- Investimentos: `docs/ledger/Regras_Investimentos.md`
- Cartão de crédito: `docs/ledger/Regras_Cartão_de_crédito.md`
- Segurança e validações: `docs/ledger/Regras_Segurança.md`
- Casos de uso: `docs/ledger/Casos_de_uso.md`

---

## Propósito

O HausHaltsMeister destina-se a individuos que tratam suas finanças pessoais com o rigor de uma operação empresarial. É a ferramenta ideal para quem busca:

- **Precisão**: dados confiáveis e categorizados corretamente na origem.
- **Clareza**: segregação estrita entre despesa, investimento e movimentação de terceiros.
- **Controle**: responsabilidade do registro manual reforça a consciência financeira a cada transação.

---

_Uma abordagem técnica e disciplinada para a administração de recursos domésticos._
