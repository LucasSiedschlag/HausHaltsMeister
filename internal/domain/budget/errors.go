package budget

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
	ErrNotFound        = errors.New("not_found")
	ErrAccessDenied    = NewError("LEDGER_ACCESS_DENIED", "Acesso negado ao ledger", nil)
	ErrLedgerNotFound  = NewError("LEDGER_NOT_FOUND", "Ledger nao encontrado", nil)
	ErrPlanNotFound    = NewError("BUDGET_PLAN_NOT_FOUND", "Plano de orçamento não encontrado", map[string]string{"plan": "not_found"})
	ErrVersionNotFound = NewError("BUDGET_VERSION_NOT_FOUND", "Versão do orçamento não encontrada", map[string]string{"version": "not_found"})
	ErrLineNotFound    = NewError("BUDGET_LINE_NOT_FOUND", "Linha do orçamento não encontrada", map[string]string{"line": "not_found"})
	ErrValidation      = NewError("VALIDATION_ERROR", "Erro de validação", nil)
	ErrVersionConflict = NewError("BUDGET_VERSION_CONFLICT", "Já existe uma versão de orçamento para este mês", nil)
)
