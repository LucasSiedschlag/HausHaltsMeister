package ledger

import "time"

type Ledger struct {
	ID           string
	OwnerUserID  string
	Name         string
	CurrencyCode string
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

type LedgerWithRole struct {
	Ledger Ledger
	Role   string
}

type Member struct {
	LedgerID  string
	UserID    string
	Role      string
	CreatedAt time.Time
	UpdatedAt *time.Time
}
