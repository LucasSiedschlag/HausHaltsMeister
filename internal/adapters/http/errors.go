package httpapi

import (
	nethttp "net/http"

	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

type CodedError interface {
	error
	Code() string
	Message() string
	Details() map[string]string
}

func WriteError(c echo.Context, status int, code, message string, details map[string]string) error {
	return c.JSON(status, ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	})
}

func WriteAuthError(c echo.Context, err error) error {
	return WriteAppError(c, err)
}

func WriteAppError(c echo.Context, err error) error {
	if appErr, ok := err.(CodedError); ok {
		status := errorStatus(appErr.Code())
		return WriteError(c, status, appErr.Code(), appErr.Message(), appErr.Details())
	}
	return WriteError(c, nethttp.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno", nil)
}

func errorStatus(code string) int {
	switch code {
	case "AUTH_INVALID_CREDENTIALS":
		return nethttp.StatusUnauthorized
	case "AUTH_USER_INACTIVE":
		return nethttp.StatusForbidden
	case "AUTH_REFRESH_REVOKED":
		return nethttp.StatusUnauthorized
	case "AUTH_RATE_LIMITED":
		return nethttp.StatusTooManyRequests
	case "AUTH_EMAIL_NOT_VERIFIED":
		return nethttp.StatusForbidden
	case "AUTH_OAUTH_STATE_INVALID":
		return nethttp.StatusUnauthorized
	case "AUTH_OAUTH_PROVIDER_ERROR":
		return nethttp.StatusBadGateway
	case "AUTH_OAUTH_EMAIL_REQUIRED":
		return nethttp.StatusUnprocessableEntity
	case "CONFLICT_DUPLICATE_EMAIL":
		return nethttp.StatusConflict
	case "LEDGER_ACCESS_DENIED":
		return nethttp.StatusForbidden
	case "LEDGER_NOT_FOUND":
		return nethttp.StatusNotFound
	case "MEMBER_ALREADY_EXISTS":
		return nethttp.StatusConflict
	case "VALIDATION_ERROR":
		return nethttp.StatusUnprocessableEntity
	case "NOT_IMPLEMENTED":
		return nethttp.StatusNotImplemented
	default:
		return nethttp.StatusInternalServerError
	}
}
