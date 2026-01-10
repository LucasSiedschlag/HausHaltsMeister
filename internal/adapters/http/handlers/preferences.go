package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/dto"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/httpx"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/preferences"
	"github.com/labstack/echo/v4"
)

type PreferencesHandler struct {
	Service *preferences.Service
	Ledger  PreferencesLedgerService
}

type PreferencesLedgerService interface {
	GetMembership(ctx context.Context, userID, ledgerID string) (string, error)
}

func (h *PreferencesHandler) Register(g *echo.Group) {
	g.GET("/preferences", h.GetPreferences)
	g.PUT("/preferences", h.UpdatePreferences)
}

func (h *PreferencesHandler) GetPreferences(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	prefs, err := h.Service.Get(c.Request().Context(), user.ID)
	if err != nil {
		return httpx.WriteError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Erro interno", nil)
	}
	return c.JSON(http.StatusOK, preferencesResponseFrom(prefs))
}

func (h *PreferencesHandler) UpdatePreferences(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}

	var req dto.UserPreferencesRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	if details := validatePreferencesRequest(req); len(details) > 0 {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", details)
	}

	defaultLedgerIDSet := req.DefaultLedgerID.Set
	var defaultLedgerID *string
	if req.DefaultLedgerID.Set && req.DefaultLedgerID.Value != nil {
		trimmed := strings.TrimSpace(*req.DefaultLedgerID.Value)
		defaultLedgerID = &trimmed
	}

	if defaultLedgerIDSet && defaultLedgerID != nil && h.Ledger != nil {
		if _, err := h.Ledger.GetMembership(c.Request().Context(), user.ID, *defaultLedgerID); err != nil {
			return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"default_ledger_id": "invalid"})
		}
	}

	updated, err := h.Service.Update(c.Request().Context(), user.ID, preferences.UpdateParams{
		ThemeMode:          req.ThemeMode,
		ThemePalette:       req.ThemePalette,
		ThemeTone:          req.ThemeTone,
		Locale:             req.Locale,
		CompactMode:        req.CompactMode,
		FontScale:          req.FontScale,
		NotifyCardClose:    req.NotifyCardClose,
		NotifyBudgetOver:   req.NotifyBudgetOver,
		NotifyPayables:     req.NotifyPayables,
		DefaultLedgerID:    defaultLedgerID,
		DefaultLedgerIDSet: defaultLedgerIDSet,
	})
	if err != nil {
		if err == preferences.ErrInvalidPreferences {
			return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", nil)
		}
		return httpx.WriteError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Erro interno", nil)
	}

	return c.JSON(http.StatusOK, preferencesResponseFrom(updated))
}

func preferencesResponseFrom(prefs preferences.UserPreferences) dto.UserPreferencesResponse {
	return dto.UserPreferencesResponse{
		UserID:           prefs.UserID,
		ThemeMode:        prefs.ThemeMode,
		ThemePalette:     prefs.ThemePalette,
		ThemeTone:        prefs.ThemeTone,
		Locale:           prefs.Locale,
		CompactMode:      prefs.CompactMode,
		FontScale:        prefs.FontScale,
		NotifyCardClose:  prefs.NotifyCardClose,
		NotifyBudgetOver: prefs.NotifyBudgetOver,
		NotifyPayables:   prefs.NotifyPayables,
		DefaultLedgerID:  prefs.DefaultLedgerID,
		CreatedAt:        prefs.CreatedAt,
		UpdatedAt:        prefs.UpdatedAt,
	}
}

func validatePreferencesRequest(req dto.UserPreferencesRequest) map[string]string {
	details := map[string]string{}
	allowed := func(value *string, values map[string]struct{}, field string) {
		if value == nil {
			return
		}
		if _, ok := values[*value]; !ok {
			details[field] = "invalid"
		}
	}

	allowed(req.ThemeMode, map[string]struct{}{"light": {}, "dark": {}, "system": {}}, "theme_mode")
	allowed(req.ThemePalette, map[string]struct{}{
		"default": {}, "red": {}, "rose": {}, "orange": {}, "green": {}, "yellow": {}, "violet": {}, "monochrome": {},
	}, "theme_palette")
	allowed(req.ThemeTone, map[string]struct{}{"vivid": {}, "pastel": {}, "muted": {}}, "theme_tone")
	allowed(req.Locale, map[string]struct{}{"pt-BR": {}, "en-US": {}}, "locale")
	allowed(req.CompactMode, map[string]struct{}{"comfortable": {}, "compact": {}, "dense": {}}, "compact_mode")
	allowed(req.FontScale, map[string]struct{}{"sm": {}, "md": {}, "lg": {}}, "font_scale")
	if req.DefaultLedgerID.Set && req.DefaultLedgerID.Value != nil {
		value := strings.TrimSpace(*req.DefaultLedgerID.Value)
		if value == "" || !httpx.IsUUID(value) {
			details["default_ledger_id"] = "invalid"
		}
	}

	return details
}
