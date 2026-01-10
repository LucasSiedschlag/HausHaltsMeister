package preferences

import (
	"context"
	"errors"
)

var (
	ErrInvalidPreferences = errors.New("invalid preferences")
)

var (
	allowedThemeModes    = map[string]struct{}{"light": {}, "dark": {}, "system": {}}
	allowedThemePalettes = map[string]struct{}{
		"default": {}, "red": {}, "rose": {}, "orange": {}, "green": {}, "yellow": {}, "violet": {}, "monochrome": {},
	}
	allowedThemeTones   = map[string]struct{}{"vivid": {}, "pastel": {}, "muted": {}}
	allowedLocales      = map[string]struct{}{"pt-BR": {}, "en-US": {}}
	allowedCompactModes = map[string]struct{}{"comfortable": {}, "compact": {}, "dense": {}}
	allowedFontScales   = map[string]struct{}{"sm": {}, "md": {}, "lg": {}}
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Get(ctx context.Context, userID string) (UserPreferences, error) {
	return s.repo.GetPreferences(ctx, userID)
}

func (s *Service) Update(ctx context.Context, userID string, input UpdateParams) (UserPreferences, error) {
	current, err := s.repo.GetPreferences(ctx, userID)
	if err != nil {
		return UserPreferences{}, err
	}

	if input.ThemeMode != nil {
		current.ThemeMode = *input.ThemeMode
	}
	if input.ThemePalette != nil {
		current.ThemePalette = *input.ThemePalette
	}
	if input.ThemeTone != nil {
		current.ThemeTone = *input.ThemeTone
	}
	if input.Locale != nil {
		current.Locale = *input.Locale
	}
	if input.CompactMode != nil {
		current.CompactMode = *input.CompactMode
	}
	if input.FontScale != nil {
		current.FontScale = *input.FontScale
	}
	if input.NotifyCardClose != nil {
		current.NotifyCardClose = *input.NotifyCardClose
	}
	if input.NotifyBudgetOver != nil {
		current.NotifyBudgetOver = *input.NotifyBudgetOver
	}
	if input.NotifyPayables != nil {
		current.NotifyPayables = *input.NotifyPayables
	}
	if input.DefaultLedgerIDSet {
		current.DefaultLedgerID = input.DefaultLedgerID
	}

	if err := validatePreferences(current); err != nil {
		return UserPreferences{}, err
	}

	return s.repo.UpsertPreferences(ctx, current)
}

func validatePreferences(prefs UserPreferences) error {
	if _, ok := allowedThemeModes[prefs.ThemeMode]; !ok {
		return ErrInvalidPreferences
	}
	if _, ok := allowedThemePalettes[prefs.ThemePalette]; !ok {
		return ErrInvalidPreferences
	}
	if _, ok := allowedThemeTones[prefs.ThemeTone]; !ok {
		return ErrInvalidPreferences
	}
	if _, ok := allowedLocales[prefs.Locale]; !ok {
		return ErrInvalidPreferences
	}
	if _, ok := allowedCompactModes[prefs.CompactMode]; !ok {
		return ErrInvalidPreferences
	}
	if _, ok := allowedFontScales[prefs.FontScale]; !ok {
		return ErrInvalidPreferences
	}
	return nil
}
