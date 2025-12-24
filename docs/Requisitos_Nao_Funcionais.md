# Documento de Requisitos Não-Funcionais (RNF)

## Segurança, Confiabilidade e Operação

Este documento estabelece os padrões de qualidade e critérios operacionais (arquiteturais) que o sistema **HausHaltsMeister** deve seguir. Enquanto os Requisitos Funcionais descrevem _o que_ o sistema faz, estes requisitos descrevem _como_ ele deve se comportar sob aspectos de segurança e engenharia.

---

### RNF-01 — Autenticação Mínima

O sistema deve implementar um mecanismo de verificação de identidade para impedir execuções acidentais ou não autorizadas (ex: cURL errado, scripts locais, acesso LAN).

- **Mecanismo sugerido**: Header fixo (`X-App-Token` ou similar) ou Basic Auth.
- **Configuração**: Token definido via variáveis de ambiente.

### RNF-02 — Limites Operacionais (Hardening)

Para garantir estabilidade e evitar DoS acidental:

- **Timeouts**: Todos os requests devem ter timeout definido (ex: 30s).
- **Payload**: Limitar tamanho do corpo da requisição (ex: 1MB).
- **Rate Limit**: Proteção básica contra loops infinitos em scripts de automação.

### RNF-03 — Auditoria Financeira

Todas as operações que alteram o estado financeiro (criação, edição, exclusão de lançamentos) devem ser registradas em um log de auditoria persistente.

- **Dados**: Quem (user/system), Quando, O Que (diff), Entidade Afetada.
- **Imutabilidade**: O log de auditoria não deve ser editável via API padrão.

### RNF-04 — Estratégia de Backup e Recuperação

O sistema deve possuir documentação e ferramentas claras para:

- **Backup**: Dump completo do banco de dados (schema + data).
- **Restore**: Capacidade de restaurar o banco em caso de falha catastrófica.
- **Frequência**: Recomendação de backup regular (ex: diário ou pré-deploy).

### RNF-05 — Precisão e Integridade de Dados

O sistema deve garantir a consistência matemática e transacional das operações financeiras.

- **Transações Atômicas (ACID)**: Lançamentos que envolvem múltiplas tabelas (ex: parcelamento, picuinha com fluxo) devem ser nucleares. Falha em uma parte reverte tudo.
- **Precisão Numérica**: O armazenamento e cálculo devem evitar erros de arredondamento comuns em ponto flutuante, garantindo precisão de 2 casas decimais.

### RNF-06 — Isolamento de Ambiente

O sistema deve prever a separação clara de ambientes e configurações.

- **Configuração**: Uso estrito de variáveis de ambiente (12-factor app) para credenciais e conexões.
- **Ambientes**: O banco de dados de testes (`casfhlow_test`) deve ser fisicamente distinto dos dados de produção/uso pessoal.

---

**Fim do Documento**
