package payment

import "errors"

var (
	ErrNameRequired      = errors.New("name is required")
	ErrKindRequired      = errors.New("kind is required")
	ErrInvalidClosingDay = errors.New("closing day must be between 1 and 31")
	ErrInvalidDueDay     = errors.New("due day must be between 1 and 31")
)

const (
	KindCreditCard = "CREDIT_CARD"
	KindDebitCard  = "DEBIT_CARD"
	KindCash       = "CASH"
	KindPix        = "PIX"
	KindBankSlip   = "BANK_SLIP"
)

type PaymentMethod struct {
	ID         int32
	Name       string
	Kind       string
	BankName   string // Optional
	ClosingDay *int32 // Optional, specific for Credit Card
	DueDay     *int32 // Optional, specific for Credit Card
	IsActive   bool
}

// EnsureValid checks basic rules
func (p *PaymentMethod) Validate() error {
	if p.Name == "" {
		return ErrNameRequired
	}
	if p.Kind == "" {
		return ErrKindRequired
	}
	if p.Kind == KindCreditCard {
		if p.ClosingDay != nil && (*p.ClosingDay < 1 || *p.ClosingDay > 31) {
			return ErrInvalidClosingDay
		}
		if p.DueDay != nil && (*p.DueDay < 1 || *p.DueDay > 31) {
			return ErrInvalidDueDay
		}
	}
	return nil
}
