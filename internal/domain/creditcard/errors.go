package creditcard

import "errors"

type Error struct {
	code    string
	message string
	details map[string]string
}

func (e *Error) Error() string {
	return e.code
}

func (e *Error) Code() string {
	return e.code
}

func (e *Error) Message() string {
	return e.message
}

func (e *Error) Details() map[string]string {
	return e.details
}

func NewError(code, message string, details map[string]string) *Error {
	return &Error{code: code, message: message, details: details}
}

var (
	ErrNotFound                 = errors.New("not_found")
	ErrAccessDenied             = NewError("LEDGER_ACCESS_DENIED", "Acesso negado ao ledger", nil)
	ErrLedgerNotFound           = NewError("LEDGER_NOT_FOUND", "Ledger nao encontrado", nil)
	ErrCardNotFound             = NewError("CREDITCARD_CARD_NOT_FOUND", "Cartao nao encontrado", nil)
	ErrCardAlreadyExists        = NewError("CREDITCARD_CARD_ALREADY_EXISTS", "Cartao ja cadastrado", nil)
	ErrStatementNotFound        = NewError("CREDITCARD_STATEMENT_NOT_FOUND", "Fatura nao encontrada", nil)
	ErrStatementAlreadyPaid     = NewError("CREDITCARD_STATEMENT_ALREADY_PAID", "Fatura ja paga", nil)
	ErrInstallmentAlreadyPosted = NewError("CREDITCARD_INSTALLMENT_ALREADY_POSTED", "Parcela ja postada", nil)
	ErrPaymentExceedsTotal      = NewError("CREDITCARD_PAYMENT_EXCEEDS_TOTAL", "Pagamento excede total", nil)
	ErrTransactionReferenced    = NewError("TRANSACTION_REFERENCED", "Transacao referenciada", nil)
	ErrValidation               = NewError("VALIDATION_ERROR", "Validacao falhou", nil)
)
