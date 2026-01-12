package accounts

import "time"

type Account struct {
	ID        string
	LedgerID  string
	Name      string
	Type      string
	Nature    string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt *time.Time
}
