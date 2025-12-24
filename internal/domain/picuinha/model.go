package picuinha

import (
	"errors"
	"time"
)

var (
	ErrPersonNameRequired = errors.New("person name is required")
	ErrAmountRequired     = errors.New("amount must be greater than zero")
	ErrInvalidKind        = errors.New("invalid kind (must be 'loan' or 'repayment')")
)

const (
	KindLoan      = "LOAN"      // Empréstimo (Eu emprestei ou peguei? Vamos definir: Entries aumentam ou diminuem saldo)
	KindRepayment = "REPAYMENT" // Pagamento/Abatimento da dívida
	// Simplificação: Vamos assumir que 'amount' positivo em LOAN aumenta a dívida da pessoa para comigo?
	// Ou vamos usar um campo de direção? O schema tem apenas 'kind' e 'amount'.
	// Vamos convencionar:
	// LOAN -> Cria dívida (Saldo aumenta)
	// REPAYMENT -> Paga dívida (Saldo diminui)
	// Mas e se EU devo a pessoa?
	// O sistema é "HausHaltsMeister" (Mestre da Casa).
	// Geralmente "Picuinhas" são pequenas dívidas *de terceiros* para com o usuário ou vice-versa.
	// Vamos manter simples: Saldo positivo = A pessoa me deve. Saldo negativo = Eu devo a pessoa.
	// ENTRIES:
	// kind='LENT' (Emprestei/Paguei por ela) -> Aumenta saldo (Pessoa me deve mais)
	// kind='BORROWED' (Peguei emprestado) -> Diminui saldo (Pessoa me deve menos ou eu devo ela)
	// kind='PAID' (Pessoa me pagou) -> Diminui saldo
	// kind='RECEIVED' (Eu paguei ela) -> Aumenta saldo (Estou pagando minha dívida)

	// Para simplificar FASE 5, vamos usar:
	// PLUS (Aumenta o que a pessoa me deve)
	// MINUS (Diminui o que a pessoa me deve)
)

type Person struct {
	ID      int32
	Name    string
	Notes   string
	Balance float64 // Calculated field
}

type Entry struct {
	ID         int32
	PersonID   int32
	Date       time.Time
	Kind       string // PLUS, MINUS
	Amount     float64
	CashFlowID *int32 // Optional link directly to a transaction (e.g. I paid the bar tab)
}
