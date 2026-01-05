Picuinhas — Regras de negócio + Pseudo-fluxo (por pessoa, com acesso restrito e sem vazamento)

Este documento descreve as regras e os fluxos para suportar Picuinhas por pessoa, garantindo segurança nível-ledger (sem precisar de ACL por linha).

Modelo recomendado:
	•	1 ledger “Picuinha” por pessoa (ex.: Picuinha — João, Picuinha — Maria)
	•	o “acesso do João” é feito via membership do ledger
	•	lançamentos são feitos como transações dentro do ledger daquela pessoa
	•	suas contas pessoais não entram no ledger do João/Maria (evita vazamento)

⸻

0) Objetivos e princípios

Objetivo funcional
	1.	Registrar “picuinhas” por pessoa (empréstimos, devoluções, acertos).
	2.	Calcular saldo: quanto você deve / quanto a pessoa te deve.
	3.	Permitir dar acesso ao João/Maria para ver somente a relação deles.
	4.	Garantir que João não veja:
	•	seu ledger pessoal
	•	a picuinha de Maria
	•	suas accounts bancárias reais

Princípios do modelo (segurança + simplicidade)
	•	Ledger é a fronteira de segurança.
	•	“Pessoa” não é account global nem “filtro por transação”; é um ledger dedicado.
	•	Nada de “um ledger Picuinhas com várias pessoas” (isso exige permissão por linha e é mais fácil errar).

⸻

1) Definições e convenções

1.1. O que é “Picuinha Ledger”

Um ledger com metadados:
	•	kind = "picuinha" (recomendado como metadata)
	•	counterparty_user_id (se a pessoa tiver usuário no sistema)
	•	counterparty_name (se ainda não tiver usuário)

Exemplos:
	•	Picuinha — João
	•	Picuinha — Maria

1.2. Papéis (roles) no Picuinha Ledger
	•	Você: owner
	•	João/Maria:
	•	viewer (recomendado inicialmente): só lê
	•	editor (opcional): pode lançar/confirmar (se você quiser colaboração)

Regra de ouro: se a pessoa vai “apenas acompanhar”, viewer.

1.3. Direção do saldo (conceito)

Você precisa escolher um padrão único para evitar confusão. Recomendação:
	•	loan_in (entrada): a pessoa te emprestou (você ficou “devendo”)
	•	loan_out (saída): você devolveu/pagou (reduz a dívida)

Saldo sugerido (quanto você deve à pessoa):
	•	balance_owed = sum(loan_in) - sum(loan_out)

Se balance_owed > 0: você deve.
Se balance_owed < 0: a pessoa te deve (situação rara, mas pode acontecer).

⸻

2) Categorias: técnicas vs reais (Picuinhas)

Picuinha é “dívida/acerto”, não é consumo. Então use categorias técnicas (sem orçamento):

Categorias técnicas (recomendadas)
	1.	Picuinha — Empréstimo Recebido

	•	direction: in
	•	is_budget_relevant: false

	2.	Picuinha — Devolução/Pagamento

	•	direction: out
	•	is_budget_relevant: false

Assim o Picuinha Ledger não “polui” orçamento e relatórios de consumo.

⸻

3) Fluxo 1 — Criar Picuinha Ledger por pessoa (com ou sem usuário)

3.1. Caso A: pessoa ainda não tem acesso (sem usuário)
	1.	POST /ledgers (cria o ledger Picuinha — João)
	2.	Não adiciona membro ainda (só você no ledger)

3.2. Caso B: pessoa já tem usuário (ou você criou o usuário)
	1.	POST /ledgers (cria o ledger)
	2.	POST /ledgers/{ledgerId}/members adiciona o usuário como viewer

Regra importante:
	•	Você só lança picuinhas “no nome do João” dentro do ledger do João.
	•	João só consegue ver se for membro daquele ledger.

⸻

4) Fluxo 2 — Registrar empréstimo (João emprestou 200)

4.1. Entrada (UI)
	•	ledger: Picuinha — João
	•	valor: 200
	•	data
	•	observação (opcional): “Pix”, “dinheiro”, etc.

4.2. Efeito contábil (no Picuinha Ledger)

Criar uma transaction com 1 entry:
	•	entry:
	•	category: Picuinha — Empréstimo Recebido (IN)
	•	amount: 200
	•	memo: opcional

Resultado:
	•	aumenta balance_owed (você deve mais ao João)

⸻

5) Fluxo 3 — Registrar devolução (você devolveu 50)

Criar transaction com 1 entry:
	•	category: Picuinha — Devolução/Pagamento (OUT)
	•	amount: 50

Resultado:
	•	reduz balance_owed

⸻

6) Fluxo 4 — (Opcional) Vincular ao seu ledger pessoal (sem vazamento)

Você pode querer refletir o dinheiro real saindo/entrando no seu ledger pessoal (banco/carteira).

Regra de segurança:
	•	O ledger do João nunca referencia suas accounts reais.
	•	O vínculo é feito só no backend via um link_id interno (não expor detalhes do seu ledger pro João).

6.1. Empréstimo recebido (João te empresta 200)
	•	No Picuinha — João: loan_in 200
	•	No Seu ledger pessoal: entrada na sua account real (ex.: “Banco Inter”) com categoria técnica “Entrada Picuinha”
	•	Salvar transfer_link_id (interno) conectando as duas transações

6.2. Devolução (você devolve 50)
	•	No Picuinha — João: loan_out 50
	•	No Seu ledger pessoal: saída da sua account real com categoria técnica “Saída Picuinha”
	•	Vincular por transfer_link_id

Isso dá rastreabilidade sem quebrar isolamento.

⸻

7) Fluxo 5 — Dar acesso ao João depois (sem mudar histórico)

Quando João ganhar acesso (login):
	1.	POST /ledgers/{ledgerId}/members adiciona João como viewer
	2.	João agora consegue:
	•	GET /ledgers → só verá os ledgers onde é membro (ex.: Picuinha — João)
	•	GET /ledgers/{ledgerId}/transactions → só as transações daquela relação

Garantia de privacidade:
	•	Maria não é membro do ledger do João, então não vê nada.
	•	João não é membro do seu ledger pessoal, então não vê suas accounts.

⸻

8) Consultas essenciais (conceituais)

8.1. Saldo atual da picuinha (quanto você deve)

No ledger da pessoa:
	•	sum(IN empréstimo) - sum(OUT devolução)

8.2. Extrato da picuinha
	•	lista de transactions ordenadas por data
	•	com memo/observação

8.3. Relatório resumido (“Picuinhas abertas”)

No seu usuário:
	•	listar seus ledgers kind=picuinha
	•	para cada ledger, calcular balance_owed
	•	mostrar: João: R$ 150, Maria: R$ 300, etc.

⸻

9) Regras de validação e integridade
	•	valores > 0
	•	categorias do Picuinha Ledger devem ser is_budget_relevant=false
	•	proibir “misturar” accounts reais no Picuinha Ledger (se você usar accounts no modelo)
	•	permissão:
	•	viewer: read-only
	•	editor: pode criar/editar transações
	•	owner: gerencia membros/roles

⸻

10) Extensões futuras (não obrigatórias agora)
	•	Confirmação/aceite: João marca “confirmado” um lançamento
	•	Disputa: status pending/confirmed/disputed
	•	Notificações: quando você lança algo no ledger do João, ele recebe aviso
	•	Importação: anexar comprovante/arquivo
	•	Pagamentos parciais: já suportado naturalmente por múltiplas devoluções

⸻

Decisão-chave (repetindo porque é o que te salva de brecha)

Não faça um único ledger “Picuinhas” com várias pessoas e tente filtrar por pessoa.
Faça um ledger por pessoa, e use membership/role como sua barreira de segurança.

Se você quiser, eu adapto isso para o seu contrato atual e escrevo quais endpoints mínimos você precisa expor na UI (criação de ledger picuinha, add member, lançar empréstimo/devolução, ver saldo).