package preferences

import (
	"context"
	"errors"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/preferences"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(store *postgres.Store) *Repository {
	return &Repository{pool: store.Pool()}
}

func (r *Repository) GetPreferences(ctx context.Context, userID string) (preferences.UserPreferences, error) {
	var prefs preferences.UserPreferences
	row := r.pool.QueryRow(ctx, `
		SELECT user_id, theme_mode, theme_palette, theme_tone, locale, compact_mode, font_scale,
		       notify_card_close, notify_budget_over, notify_payables, default_ledger_id, created_at, updated_at
		FROM user_preferences
		WHERE user_id = $1
	`, userID)
	if err := scanPreferences(row, &prefs); err != nil {
		if errors.Is(err, preferences.ErrNotFound) {
			return r.createDefaultPreferences(ctx, userID)
		}
		return preferences.UserPreferences{}, err
	}
	return prefs, nil
}

func (r *Repository) UpsertPreferences(ctx context.Context, prefs preferences.UserPreferences) (preferences.UserPreferences, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO user_preferences (
			user_id, theme_mode, theme_palette, theme_tone, locale, compact_mode, font_scale,
			notify_card_close, notify_budget_over, notify_payables, default_ledger_id, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, now())
		ON CONFLICT (user_id)
		DO UPDATE SET
			theme_mode = EXCLUDED.theme_mode,
			theme_palette = EXCLUDED.theme_palette,
			theme_tone = EXCLUDED.theme_tone,
			locale = EXCLUDED.locale,
			compact_mode = EXCLUDED.compact_mode,
			font_scale = EXCLUDED.font_scale,
			notify_card_close = EXCLUDED.notify_card_close,
			notify_budget_over = EXCLUDED.notify_budget_over,
			notify_payables = EXCLUDED.notify_payables,
			default_ledger_id = EXCLUDED.default_ledger_id,
			updated_at = now()
		RETURNING user_id, theme_mode, theme_palette, theme_tone, locale, compact_mode, font_scale,
		          notify_card_close, notify_budget_over, notify_payables, default_ledger_id, created_at, updated_at
	`, prefs.UserID, prefs.ThemeMode, prefs.ThemePalette, prefs.ThemeTone, prefs.Locale, prefs.CompactMode, prefs.FontScale,
		prefs.NotifyCardClose, prefs.NotifyBudgetOver, prefs.NotifyPayables, prefs.DefaultLedgerID)
	if err := scanPreferences(row, &prefs); err != nil {
		return preferences.UserPreferences{}, err
	}
	return prefs, nil
}

func (r *Repository) createDefaultPreferences(ctx context.Context, userID string) (preferences.UserPreferences, error) {
	var prefs preferences.UserPreferences
	row := r.pool.QueryRow(ctx, `
		INSERT INTO user_preferences (user_id)
		VALUES ($1)
		RETURNING user_id, theme_mode, theme_palette, theme_tone, locale, compact_mode, font_scale,
		          notify_card_close, notify_budget_over, notify_payables, default_ledger_id, created_at, updated_at
	`, userID)
	if err := scanPreferences(row, &prefs); err != nil {
		return preferences.UserPreferences{}, err
	}
	return prefs, nil
}

func scanPreferences(row pgx.Row, prefs *preferences.UserPreferences) error {
	if err := row.Scan(
		&prefs.UserID,
		&prefs.ThemeMode,
		&prefs.ThemePalette,
		&prefs.ThemeTone,
		&prefs.Locale,
		&prefs.CompactMode,
		&prefs.FontScale,
		&prefs.NotifyCardClose,
		&prefs.NotifyBudgetOver,
		&prefs.NotifyPayables,
		&prefs.DefaultLedgerID,
		&prefs.CreatedAt,
		&prefs.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return preferences.ErrNotFound
		}
		return err
	}
	return nil
}
