# Guia de Implementação de Segurança Mínima

Este documento define o padrão de segurança e operação para o projeto **HausHaltsMeister**. Ele serve como guia de implementação para garantir que, mesmo sendo um sistema single-user, o projeto tenha robustez, auditoria e proteção adequadas.

> **Status**: Referência Técnica (Não Implementado)  
> **Meta**: Proteger contra erros operacionais, acesso indevido em rede local e perda de dados.

---

## 1. Contexto e Objetivo

### Por que segurança em um sistema pessoal?

1. **Dados Sensíveis**: O sistema contém histórico financeiro completo.
2. **Erro Humano**: Scripts mal feitos ou comandos errados (ex: `curl -X DELETE` sem querer) podem destruir dados.
3. **Exposição Acidental**: Em redes domésticas (LAN), serviços HTTP sem auth ficam expostos a qualquer dispositivo (TVs, Visitantes, IoT).
4. **Disciplina de Engenharia**: O projeto serve como laboratório de boas práticas.

### O que é "Segurança Mínima"?

Não é sobre defesa contra hackers estatais. É sobre:

- **Ninguém apaga dados sem querer.**
- **Ninguém acessa dados só por estar no wi-fi.**
- **Saber o que aconteceu (Auditoria).**
- **Recuperar se tudo der errado (Backup).**

---

## 2. Modelo de Ameaça Mínimo

### Riscos Tratados (Escopo)

- **Acesso não autorizado na LAN**: Dispositivo vizinho acessando a porta da API.
- **Operação destrutiva acidental**: Usuário rodando script de limpeza no banco de produção achando que é teste.
- **Perda de dados por falha de disco/container**: Corrupção do volume Docker.
- **Automação desenfreada**: Script em loop criando milhares de registros.

### Riscos NÃO Tratados (Out of Scope)

- **Ataques Físicos**: Alguém roubando o notebook.
- **Injeção de SQL avançada**: Mitigado pelo uso de sqlc/pgx, mas não foco de pentest.
- **DDoS Volumétrico**: O foco é rate limit para scripts locais, não ataques distribuídos.
- **Compliance Bancário (PCI-DSS, etc)**: Não aplicável.

---

## 3. Checklist de Implementação (Macro)

- [ ] **Auth**: Middleware verificando header `X-App-Token`.
- [ ] **Env**: Token definido apenas via variável de ambiente.
- [ ] **Hardening**: Timeout de 30s em todos os requests.
- [ ] **Hardening**: Max Body Size de 1MB.
- [ ] **Hardening**: Rate Limit simples (ex: 100 req/min).
- [ ] **CORS**: Bloquear tudo, liberar apenas origens conhecidas.
- [ ] **Audit**: Tabela `audit_logs` criada.
- [ ] **Audit**: Trigger ou Service gravando logs para operações de escrita.
- [ ] **Backup**: Script `make backup` funcional.
- [ ] **Restore**: Script `make restore` testado.
- [ ] **Logs**: Request ID em todos os logs.

---

## 4. Implementação Detalhada

### 4.1 Autenticação Mínima

Usaremos **API Key Fixa** por simplicidade e eficácia para single-user/scripts.

**Estratégia**: Header `X-App-Token`.

**Configuração (.env):**

```bash
APP_SECURITY_TOKEN=secreta_super_complexa_v1
```

**Pseudo-código Middleware (Echo):**

```go
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        token := c.Request().Header.Get("X-App-Token")
        expected := os.Getenv("APP_SECURITY_TOKEN")

        if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(expected)) != 1 {
            return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
        }
        return next(c)
    }
}
```

### 4.2 Autorização / Escopo

- **Padrão**: _Deny by default_.
- **Endpoints Públicos**: Apenas `/health` ou `/ready`.
- **Endpoints Protegidos**: Todos os outros (`/api/*`).

### 4.3 Hardening da API

Proteções para garantir estabilidade.

**Timeout:**

- Configurar `http.Server{ ReadTimeout: 30s, WriteTimeout: 30s }`.
- No Echo: `e.Use(middleware.TimeoutWithConfig(...))`.

**Body Limit:**

- No Echo: `e.Use(middleware.BodyLimit("1M"))`.

**Rate Limiting:**

- Usar `golang.org/x/time/rate`.
- Configurar 10-20 req/s (alto para humano, baixo para loop infinito de script).

**CORS:**

- `AllowOrigins`: Mapeado de variável `APP_CORS_ORIGINS` (ex: `http://localhost:3000`).
- Se vazio, negar tudo.

### 4.4 Validação de Entrada

Validar estritamente no **Domínio**, não apenas no Handler DTO.

- **Handler**: Valida tipos (int, string, bool) e presença de campos (required).
- **Service/Domain**: Valida regras de negócio (ex: "Categoria IN não pode ter value negativo").
- **Padronização de Erro**:
  ```json
  {
    "error": "invalid_input",
    "details": "amount must be positive",
    "request_id": "req-123"
  }
  ```

### 4.5 Logging e Observabilidade

**O que logar:**

- Request ID (gerado no middleware).
- Método, Path, Status Code, Latência.
- Erros 5xx com stack trace (apenas no log do servidor, não no JSON de resposta).

**O que NUNCA logar:**

- O valor do `X-App-Token`.
- Conteúdo de `body` (pode conter PII ou dados sensíveis futuramente).

**Níveis:**

- `DEBUG`: Apenas em dev.
- `INFO`: Start/Stop, Requests normais.
- `ERROR`: Falhas de banco, pânicos recuperados.

### 4.6 Auditoria Mínima

**Schema Sugerido (`audit_logs`)**:

```sql
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    entity_table TEXT NOT NULL,      -- ex: 'cash_flows'
    entity_id INT NOT NULL,          -- ex: 42
    action TEXT NOT NULL,            -- 'CREATE', 'UPDATE', 'DELETE'
    old_value JSONB,                 -- snapshot anterior (UPDATE/DELETE)
    new_value JSONB,                 -- snapshot novo (CREATE/UPDATE)
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    request_id TEXT                  -- para correlacionar com logs da app
);
```

**Estratégia de Gravação:**

- Síncrona (na mesma transação do negócio) para garantir integridade. Se falhar auditoria, falha a operação.

### 4.7 Gestão de Segredos

- **Nunca commitar `.env`**.
- Ter `.env.example` com valores dummy.
- **Rotação**: Para mudar a chave, basta alterar a ENV e reiniciar o serviço.

### 4.8 Backup e Recuperação

**Backup (Diário/Manual):**

```bash
# Exemplo de comando make
pg_dump -h localhost -U user -d cashflow > backups/backup_$(date +%F).sql
```

**Restore (Emergência):**

```bash
# ⚠️ Destrutivo!
psql -h localhost -U user -d cashflow < backups/backup_2024-01-01.sql
```

**Teste de Restore:**

- Deve ser feito quinzenalmente em ambiente de teste para garantir que o backup não está corrompido.

### 4.9 Estratégia de Ambientes

Isolamento físico de dados.

| Ambiente | Banco de Dados  | Configuração            |
| -------- | --------------- | ----------------------- |
| **Dev**  | `cashflow_dev`  | Log DEBUG, CORS \*      |
| **Test** | `cashflow_test` | Reset a cada teste      |
| **Prod** | `cashflow_prod` | Log INFO, CORS restrito |

---

## 5. Plano de Adoção Incremental

**Semana 1: Proteção Básica (Auth & Hardening)**

1. Implementar Middleware de Auth.
2. Configurar Timeouts e Body Limits.
3. Atualizar clientes (scripts/frontend) para enviar Header.

**Semana 2: Resiliência (Backup & Logs)**

1. Padronizar Logs com Request ID.
2. Criar scripts de Makefile para Backup/Restore.
3. Validar restore em banco de teste.

**Semana 3: Rastreabilidade (Audit)**

1. Criar migração `audit_logs`.
2. Implementar gravação de auditoria no `CashFlowService` (o mais crítico).
3. Expandir para outros services gradualmente.

---

## 6. Estratégia de Testes

A segurança deve ser testada como funcionalidade.

**Exemplos de Casos de Teste (Integração):**

1. `TestAuth_NoHeader`: Request sem token -> Esperado `401`.
2. `TestAuth_InvalidToken`: Request com token errado -> Esperado `401`.
3. `TestRateLimit`: 100 requests em 1s -> Esperado `429`.
4. `TestAudit_CreateFlow`: Criar cashflow -> Verificar se linha apareceu em `audit_logs`.

---

## 7. Apêndice — Exemplos

**Chamada Sucesso:**

```bash
curl -X GET http://localhost:8080/cashflows \
  -H "X-App-Token: minha_senha_secreta"
```

**Chamada Falha (Sem Auth):**

```bash
curl -X GET http://localhost:8080/cashflows
# < HTTP/1.1 401 Unauthorized
# {"error": "missing or invalid token"}
```
