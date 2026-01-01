package auth

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
	ErrNotFound           = errors.New("not_found")
	ErrInvalidCredentials = NewError("AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", map[string]string{"email": "invalid"})
	ErrUserInactive       = NewError("AUTH_USER_INACTIVE", "Usuario inativo", nil)
	ErrRefreshRevoked     = NewError("AUTH_REFRESH_REVOKED", "Refresh token revogado", nil)
	ErrRateLimited        = NewError("AUTH_RATE_LIMITED", "Muitas tentativas", nil)
	ErrEmailNotVerified   = NewError("AUTH_EMAIL_NOT_VERIFIED", "Email nao verificado", nil)
	ErrOAuthStateInvalid  = NewError("AUTH_OAUTH_STATE_INVALID", "State OAuth invalido", nil)
	ErrOAuthProvider      = NewError("AUTH_OAUTH_PROVIDER_ERROR", "Erro no provider OAuth", nil)
	ErrOAuthEmailRequired = NewError("AUTH_OAUTH_EMAIL_REQUIRED", "Email necessario para OAuth", nil)
	ErrServiceUnavailable = NewError("SERVICE_UNAVAILABLE", "Servico indisponivel", nil)
	ErrGatewayTimeout     = NewError("GATEWAY_TIMEOUT", "Timeout na dependencia", nil)
	ErrDuplicateEmail     = NewError("CONFLICT_DUPLICATE_EMAIL", "Email ja cadastrado", nil)
	ErrValidation         = NewError("VALIDATION_ERROR", "Validacao falhou", nil)
	ErrNotImplemented     = NewError("NOT_IMPLEMENTED", "Funcionalidade nao disponivel", nil)
)
