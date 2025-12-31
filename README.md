# HausHaltsMeister

**Sistema de Controle Financeiro de Alta Precisao**

---

### Gestao Financeira Baseada em Dados e Fundamentos Contabeis

A verdadeira gestao financeira nao se resume ao aglutinamento de transacoes bancarias. Ela exige uma compreensao clara do fluxo de caixa, segregacao de patrimonio e planejamento estrategico. O **HausHaltsMeister** foi desenvolvido para oferecer uma estrutura profissional para as financas pessoais, focando na integridade dos dados e na clareza para a tomada de decisao.

Este sistema elimina o ruido das automacoes imprecisas e devolve ao usuario o controle total sobre a interpretacao de cada movimentacao financeira.

---

## Pilares do Sistema

### 1. Integridade do Fluxo de Caixa

O nucleo da saude financeira e a liquidez. O sistema prioriza o registro manual e consciente de cada entrada e saida, garantindo que o fluxo de caixa registrado corresponda fielmente a realidade operacional, sem dependencia de integracoes bancarias falhas ou categorizacoes genericas.

### 2. Gestao de Recebiveis e Obrigacoes

Uma falha comum em sistemas pessoais e misturar patrimonio proprio com recursos transitorios (emprestimos, compras para terceiros, divisao de contas). O HausHaltsMeister trata essas movimentacoes como **Contas a Receber** e **Contas a Pagar**, garantindo que o saldo disponivel reflita seu patrimonio real, isolando o que e dinheiro de terceiros.

### 3. Planejamento Orcamentario Estrategico

O orcamento nao e apenas um limite de gastos, mas uma ferramenta de alocacao de recursos. O sistema permite a definicao de orcamentos mensais e a analise de variancia (Orcado vs. Realizado), permitindo ajustes taticos na estrategia financeira familiar antes que o periodo se encerre.

### 4. Rastreabilidade e Auditoria

Cada decisao financeira deve ser rastreavel. O sistema mantem um historico imutavel e estruturado, permitindo analises longitudinais de evolucao patrimonial e correcao de rota baseada em dados historicos consolidados.

---

## Fundacao do Modelo (Ledger Leve)

O HausHaltsMeister e estruturado sobre um **ledger leve** (journal):

- **Ledger como fronteira de dados**: tudo pertence a um `ledger_id`, com caminho natural para multiusuario.
- **Journal como fonte da verdade**: `transactions` (evento) + `entries` (linhas) registram todo o fluxo.
- **Direcao IN/OUT definida pela categoria**: `categories.direction` define a semantica; valores sao sempre positivos.
- **Orcamento flexivel por percentual**: calculado sobre a renda base do mes e versionado no tempo.
- **Contas internas**: Pessoal, Investimentos e Cartao refletem o destino do dinheiro, sem modelar bancos externos.
- **Cartao por parcela**: o consumo entra no orcamento apenas quando a parcela e postada; pagamento de fatura e transferencia tecnica.

---

## Beneficios da Arquitetura

- **Consistencia**: regras unificadas para entradas, saidas e orcamento.
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
- Categorias e orcamento: `docs/ledger/Regras_Categorias_e_Orçamento.md`
- Investimentos: `docs/ledger/Regras_Investimentos.md`
- Cartao de credito: `docs/ledger/Regras_Cartão_de_crédito.md`
- Seguranca e validacoes: `docs/ledger/Regras_Segurança.md`
- Casos de uso: `docs/ledger/Casos_de_uso.md`

---

## Proposito

O HausHaltsMeister destina-se a individuos que tratam suas financas pessoais com o rigor de uma operacao empresarial. E a ferramenta ideal para quem busca:

- **Precisao**: dados confiaveis e categorizados corretamente na origem.
- **Clareza**: segregacao estrita entre despesa, investimento e movimentacao de terceiros.
- **Controle**: responsabilidade do registro manual reforca a consciencia financeira a cada transacao.

---

_Uma abordagem tecnica e disciplinada para a administracao de recursos domesticos._
