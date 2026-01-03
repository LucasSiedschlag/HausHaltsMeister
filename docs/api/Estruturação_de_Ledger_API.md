Checklist “zero brecha” (Ledger Security)

A) Regras absolutas (não negocie)
	1.	Fonte de verdade do ledger = ledgerId do PATH
	•	Em qualquer POST/PATCH, ignore ledgerId do body, mesmo que exista no DTO.
	2.	Toda query de recurso deve filtrar por ledger_id
	•	WHERE ledger_id = $1 AND id = $2 (sempre)
	3.	Nunca autorize por “existe no banco”
	•	Autorize por: (user_id, ledger_id) membership + role
	4.	Role/perms só do servidor
	•	Nunca aceite role no request (exceto endpoints de owner que alteram membership, e ainda assim validando owner).

⸻

B) Middleware/Policy obrigatório (padrão por endpoint)

1) Middleware de autenticação
	•	Resolve userID da sessão/JWT.
	•	Injeta no context.

2) Middleware de “LedgerGuard”

Aplica em qualquer rota que tenha :ledgerId.

Passos do LedgerGuard:
	1.	Ler ledgerId do path.
	2.	Buscar membership:
	•	SELECT role FROM ledger_members WHERE ledger_id=$1 AND user_id=$2 AND removed_at IS NULL
	3.	Se não existir: 403 (ou 404 se você quiser “não revelar” que o ledger existe).
	4.	Verificar role mínima exigida pela rota:
	•	viewer < editor < owner
	5.	Anexar no context:
	•	ctx.ledgerId
	•	ctx.ledgerRole
	•	ctx.permissions (opcional derivado do role)

Boas práticas:
	•	Use uma função única: RequireLedgerRole(minRole)
	•	Evita “esquecer” check em handler.

⸻

C) Padrão de Service/Repo (para não errar)

1) Handler só extrai input e chama service
	•	Nada de SQL no handler.
	•	Nada de regra de autorização no handler (fica no guard/policy).

2) Service recebe sempre (ctx, ledgerId, ...)

Mesmo que o ledger já esteja no ctx, passe explicitamente pra ficar claro.

3) Repo: funções sempre “scoped”

Exemplos de assinaturas (ideia):
	•	GetCategory(ctx, ledgerId, categoryId)
	•	ListTransactions(ctx, ledgerId, cursor, limit)
	•	UpdateAccount(ctx, ledgerId, accountId, patch)

Proibido:
	•	GetCategoryByID(categoryId) sem ledgerId.

⸻

D) Padrões SQL que evitam vazamento

1) Read de recurso

SELECT ...
FROM categories
WHERE ledger_id = $1
  AND id = $2
  AND deleted_at IS NULL;

2) Update de recurso (garantindo scoping)

UPDATE categories
SET name = $3, updated_at = now()
WHERE ledger_id = $1
  AND id = $2
  AND deleted_at IS NULL
RETURNING ...;

Se não retornou linha: responde 404 (ou 403/404 conforme sua política).

3) Delete (soft delete planejado)

UPDATE categories
SET deleted_at = now()
WHERE ledger_id = $1
  AND id = $2
  AND deleted_at IS NULL;


⸻

E) Padrão de erros (segurança + UX)

Escolha 1: “Não revelar existência” (mais seguro)
	•	Se o usuário não é membro do ledger: 404 em tudo (ledger e recursos).
	•	Se é membro mas sem role: 403.

Escolha 2: “Transparente”
	•	Não membro: 403.
	•	Sem permissão: 403.
	•	Não existe: 404.

Escolha 1 costuma ser melhor pra evitar enumeração.

⸻

F) Proteções anti-brecha comuns

1) Validação de IDs
	•	ledgerId, accountId, etc. com formato único (UUID/ULID).
	•	Rejeitar strings inválidas com 400.

2) Rate limit e lockout
	•	Rate limit por IP + user em rotas sensíveis (login, bulk, reports).
	•	Evita brute force e enumeração.

3) Auditoria (muito recomendado em sistema financeiro)
	•	Registrar eventos:
	•	ledger.member.add/remove/role_change
	•	transaction.create/update/delete
	•	budget.plan.update
	•	Guardar: user_id, ledger_id, entity_id, action, created_at, ip/user-agent (se tiver).

4) Idempotência em POST críticos
	•	Para transactions e bulk: Idempotency-Key (header) por ledger.
	•	Evita duplicar transações por retry.

⸻

G) Testes que garantem que não tem furo

1) Testes de autorização por role

Para cada endpoint:
	•	viewer tenta ação editor → deve falhar
	•	editor tenta ação owner → deve falhar
	•	não-membro → deve falhar (404 ou 403)

2) Testes de “cross-ledger”

Crie:
	•	ledger A com transaction X
	•	ledger B com usuário atacante
Tentar acessar /ledgers/B/transactions/X:
	•	deve falhar sempre, mesmo com ID válido.

3) Teste de query “sem scope”
	•	Uma suite que procura repos que fazem WHERE id = sem ledger_id =.
	•	Isso pega regressão.

⸻

H) Como encaixa no seu contrato atual

Mantém o que você implementou:
	•	Flat por padrão
	•	ledgerId nos DTOs (mas ignorado no input)
	•	/ledgers/{ledgerId}/me para role/perms
	•	expand=ledger só se precisar

E a segurança fica no guard + repo scoped.

⸻

Tranformando isso em tarefas, aqui vai um mini-todo pronto:

Tarefas:
	•	Implementar LedgerGuard (middleware/policy) que resolve membership/role por ledgerId do path e bloqueia por role mínima.
	•	Garantir que todas as rotas /ledgers/:ledgerId/* usem o guard com role correto (viewer/editor/owner).
	•	Atualizar repos para que todo acesso a recursos use ledger_id + id (proibir get/update/delete só por id).
	•	Ignorar ledgerId do body em creates/updates (sempre usar o do path).
	•	Adicionar testes de cross-ledger + role matrix para endpoints principais.
	•	Padronizar erro 404 para não-membro (ou 403 conforme escolhido), consistente em toda API.
	•	Defina uma matriz de roles por endpoint no formato “tabela de policies” pra você colar no docs e implementar 1:1?