package auth

import "errors"

type Error struct {
	Code    string
	Message string
	Details map[string]string
}

func (e *Error) Error() string {
	return e.Code
}

func NewError(code, message string, details map[string]string) *Error {
	return &Error{Code: code, Message: message, Details: details}
}

var (
	ErrNotFound          = errors.New("not_found")
	ErrInvalidCredentials = NewError("AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", map[string]string{"email": "invalid"})
	ErrUserInactive       = NewError("AUTH_USER_INACTIVE", "Usuario inativo", nil)
	ErrRefreshRevoked     = NewError("AUTH_REFRESH_REVOKED", "Refresh token revogado", nil)
	ErrRateLimited        = NewError("AUTH_RATE_LIMITED", "Muitas tentativas", nil)
	ErrEmailNotVerified   = NewError("AUTH_EMAIL_NOT_VERIFIED", "Email nao verificado", nil)
	ErrOAuthStateInvalid  = NewError("AUTH_OAUTH_STATE_INVALID", "State OAuth invalido", nil)
	ErrOAuthProvider      = NewError("AUTH_OAUTH_PROVIDER_ERROR", "Erro no provider OAuth", nil)
	ErrOAuthEmailRequired = NewError("AUTH_OAUTH_EMAIL_REQUIRED", "Email necessario para OAuth", nil)
	ErrDuplicateEmail     = NewError("CONFLICT_DUPLICATE_EMAIL", "Email ja cadastrado", nil)
	ErrValidation         = NewError("VALIDATION_ERROR", "Validacao falhou", nil)
	ErrNotImplemented     = NewError("NOT_IMPLEMENTED", "Funcionalidade nao disponivel", nil)
)
