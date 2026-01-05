package preferences

import "context"

type Repository interface {
	GetPreferences(ctx context.Context, userID string) (UserPreferences, error)
	UpsertPreferences(ctx context.Context, prefs UserPreferences) (UserPreferences, error)
}
