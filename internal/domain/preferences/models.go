package preferences

import "time"

type UserPreferences struct {
	UserID           string
	ThemeMode        string
	ThemePalette     string
	ThemeTone        string
	Locale           string
	CompactMode      string
	FontScale        string
	NotifyCardClose  bool
	NotifyBudgetOver bool
	NotifyPayables   bool
	DefaultLedgerID  *string
	CreatedAt        time.Time
	UpdatedAt        *time.Time
}

type UpdateParams struct {
	ThemeMode          *string
	ThemePalette       *string
	ThemeTone          *string
	Locale             *string
	CompactMode        *string
	FontScale          *string
	NotifyCardClose    *bool
	NotifyBudgetOver   *bool
	NotifyPayables     *bool
	DefaultLedgerID    *string
	DefaultLedgerIDSet bool
}
