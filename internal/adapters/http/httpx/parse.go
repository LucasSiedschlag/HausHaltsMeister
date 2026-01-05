package httpx

import (
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func ParseDateTime(value string) (time.Time, error) {
	if strings.Contains(value, "T") {
		return time.Parse(time.RFC3339, value)
	}
	return time.Parse("2006-01-02", value)
}

func ParseMonth(value string) (time.Time, error) {
	return time.Parse("2006-01-02", value)
}

func IsUUID(value string) bool {
	_, err := uuid.Parse(value)
	return err == nil
}

func RequireUUIDParam(c echo.Context, name string) (string, error) {
	value := strings.TrimSpace(c.Param(name))
	field := paramField(name)
	if value == "" {
		return "", WriteError(c, 422, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{field: "required"})
	}
	if !IsUUID(value) {
		return "", WriteError(c, 422, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{field: "invalid"})
	}
	return value, nil
}

func paramField(name string) string {
	if strings.Contains(name, "_") {
		return name
	}
	var out []rune
	for i, r := range name {
		if i > 0 && unicode.IsUpper(r) {
			out = append(out, '_', unicode.ToLower(r))
			continue
		}
		out = append(out, unicode.ToLower(r))
	}
	return string(out)
}
