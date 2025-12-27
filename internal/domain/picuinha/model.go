package picuinha

import (
	"errors"
	"time"
)

var (
	ErrPersonNameRequired    = errors.New("person name is required")
	ErrAmountRequired        = errors.New("amount must be greater than zero")
	ErrPersonNotFound        = errors.New("person not found")
	ErrPersonHasEntries      = errors.New("person has entries")
	ErrCaseTitleRequired     = errors.New("case title is required")
	ErrCaseTypeInvalid       = errors.New("invalid case type")
	ErrInstallmentCount      = errors.New("installment count must be greater than zero")
	ErrStartDateRequired     = errors.New("start date is required")
	ErrInstallmentNotFound   = errors.New("installment not found")
	ErrCaseNotFound          = errors.New("case not found")
	ErrPaymentMethodRequired = errors.New("payment method is required for card cases")
	ErrInterestRateUnit      = errors.New("invalid interest rate unit")
	ErrRecurrenceInterval    = errors.New("recurrence interval must be greater than zero")
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

const (
	CaseTypeOneOff        = "ONE_OFF"
	CaseTypeInstallment   = "INSTALLMENT"
	CaseTypeRecurring     = "RECURRING"
	CaseTypeCardInstall   = "CARD_INSTALLMENT"
	InterestRateMonthly   = "MONTHLY"
	InterestRateAnnual    = "ANNUAL"
	StatusOpen            = "OPEN"
	StatusPaid            = "PAID"
	StatusRecurringActive = "RECURRING"
)

type Person struct {
	ID      int32
	Name    string
	Notes   string
	Balance float64 // Calculated field
}

type Case struct {
	ID                       int32
	PersonID                 int32
	Title                    string
	CaseType                 string
	TotalAmount              *float64
	InstallmentCount         *int32
	InstallmentAmount        *float64
	StartDate                time.Time
	PaymentMethodID          *int32
	InstallmentPlanID        *int32
	CategoryID               *int32
	InterestRate             *float64
	InterestRateUnit         string
	RecurrenceIntervalMonths *int32
	CreatedAt                time.Time
}

type CaseSummary struct {
	Case
	InstallmentsTotal int32
	InstallmentsPaid  int32
	AmountPaid        float64
	AmountRemaining   float64
	Status            string
}

type CaseInstallment struct {
	ID                int32
	CaseID            int32
	InstallmentNumber int32
	DueDate           time.Time
	Amount            float64
	ExtraAmount       float64
	IsPaid            bool
	PaidAt            *time.Time
}

type CreateCaseRequest struct {
	PersonID                 int32
	Title                    string
	CaseType                 string
	TotalAmount              float64
	InstallmentCount         int32
	InstallmentAmount        float64
	StartDate                time.Time
	PaymentMethodID          *int32
	InstallmentPlanID        *int32
	CategoryID               *int32
	InterestRate             *float64
	InterestRateUnit         string
	RecurrenceIntervalMonths *int32
}

type UpdateInstallmentRequest struct {
	IsPaid      bool
	ExtraAmount float64
}
