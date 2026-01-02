package dto

import "time"

type UserPreferencesResponse struct {
	UserID           string     `json:"user_id"`
	ThemeMode        string     `json:"theme_mode"`
	ThemePalette     string     `json:"theme_palette"`
	ThemeTone        string     `json:"theme_tone"`
	Locale           string     `json:"locale"`
	CompactMode      string     `json:"compact_mode"`
	FontScale        string     `json:"font_scale"`
	NotifyCardClose  bool       `json:"notify_card_close"`
	NotifyBudgetOver bool       `json:"notify_budget_over"`
	NotifyPayables   bool       `json:"notify_payables"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at,omitempty"`
}

type UserPreferencesRequest struct {
	ThemeMode        *string `json:"theme_mode"`
	ThemePalette     *string `json:"theme_palette"`
	ThemeTone        *string `json:"theme_tone"`
	Locale           *string `json:"locale"`
	CompactMode      *string `json:"compact_mode"`
	FontScale        *string `json:"font_scale"`
	NotifyCardClose  *bool   `json:"notify_card_close"`
	NotifyBudgetOver *bool   `json:"notify_budget_over"`
	NotifyPayables   *bool   `json:"notify_payables"`
}
