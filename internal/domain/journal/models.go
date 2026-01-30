package journal

import "time"

type Transaction struct {
	ID               string
	LedgerID         string
	OccurredAt       time.Time
	Description      string
	Notes            *string
	CreatedByUserID  string
	CreditCardID     *string
	InvestmentAction *string
	CreatedAt        time.Time
	UpdatedAt        *time.Time
	Entries          []Entry
}

type Entry struct {
	ID            string
	LedgerID      string
	TransactionID string
	AccountID     string
	CategoryID    *string
	Kind          string
	AmountCents   int64
	Memo          *string
	CreatedAt     time.Time
	UpdatedAt     *time.Time
}

type Cursor struct {
	OccurredAt time.Time
	ID         string
}

type ListResult struct {
	Items      []Transaction
	NextCursor *Cursor
}
