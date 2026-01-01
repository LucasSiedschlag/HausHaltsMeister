package journal

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
	ErrNotFound              = errors.New("not_found")
	ErrAccessDenied          = NewError("LEDGER_ACCESS_DENIED", "Acesso negado ao ledger", nil)
	ErrLedgerNotFound        = NewError("LEDGER_NOT_FOUND", "Ledger nao encontrado", nil)
	ErrTransactionNotFound   = NewError("TRANSACTION_NOT_FOUND", "Transacao nao encontrada", nil)
	ErrTransactionReferenced = NewError("TRANSACTION_REFERENCED", "Transacao referenciada", nil)
	ErrTransferNotBalanced   = NewError("TRANSFER_NOT_BALANCED", "Transferencia nao balanceada", nil)
	ErrValidation            = NewError("VALIDATION_ERROR", "Validacao falhou", nil)
)
