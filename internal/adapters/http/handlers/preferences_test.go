package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/httpx"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/auth"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/preferences"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type fakePreferencesRepo struct {
	prefs preferences.UserPreferences
}

func (f fakePreferencesRepo) GetPreferences(ctx context.Context, userID string) (preferences.UserPreferences, error) {
	if f.prefs.UserID == "" {
		return defaultPreferences(userID), nil
	}
	return f.prefs, nil
}

func (f fakePreferencesRepo) UpsertPreferences(ctx context.Context, prefs preferences.UserPreferences) (preferences.UserPreferences, error) {
	return prefs, nil
}

func defaultPreferences(userID string) preferences.UserPreferences {
	return preferences.UserPreferences{
		UserID:           userID,
		ThemeMode:        "system",
		ThemePalette:     "default",
		ThemeTone:        "vivid",
		Locale:           "pt-BR",
		CompactMode:      "comfortable",
		FontScale:        "md",
		NotifyCardClose:  true,
		NotifyBudgetOver: true,
		NotifyPayables:   true,
	}
}

func TestUpdatePreferencesRejectsInvalidDefaultLedgerID(t *testing.T) {
	e := echo.New()
	payload := map[string]any{"default_ledger_id": "not-a-uuid"}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/me/preferences", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", auth.User{ID: "user-1"})

	service := preferences.NewService(fakePreferencesRepo{prefs: defaultPreferences("user-1")})
	handler := PreferencesHandler{Service: service}

	err = handler.UpdatePreferences(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var resp httpx.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "VALIDATION_ERROR", resp.Code)
	require.Equal(t, "invalid", resp.Details["default_ledger_id"])
}
