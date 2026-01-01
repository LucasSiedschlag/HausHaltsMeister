package reports

import "time"

type AccountBalance struct {
	AccountID    string
	AccountName  string
	AccountType  string
	BalanceCents int64
}

type BalanceReport struct {
	Month time.Time
	Items []AccountBalance
}

type CategorySummary struct {
	CategoryID string
	Name       string
	Direction  string
	TotalCents int64
}

type CategoryReport struct {
	From  time.Time
	To    time.Time
	Items []CategorySummary
}

type CashflowEntry struct {
	Month         time.Time
	TotalInCents  int64
	TotalOutCents int64
	NetCents      int64
}

type CashflowReport struct {
	From  time.Time
	To    time.Time
	Items []CashflowEntry
}
