package categories

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
	ErrNotFound         = errors.New("not_found")
	ErrAccessDenied     = NewError("LEDGER_ACCESS_DENIED", "Acesso negado ao ledger", nil)
	ErrLedgerNotFound   = NewError("LEDGER_NOT_FOUND", "Ledger nao encontrado", nil)
	ErrCategoryNotFound = NewError("CATEGORY_NOT_FOUND", "Categoria nao encontrada", nil)
	ErrDuplicateName    = NewError("CONFLICT_DUPLICATE_NAME", "Nome duplicado", nil)
	ErrValidation       = NewError("VALIDATION_ERROR", "Validacao falhou", nil)
	ErrNotImplemented   = NewError("NOT_IMPLEMENTED", "Funcionalidade nao disponivel", nil)
)
