package categories

import "time"

type Category struct {
	ID               string
	LedgerID         string
	ParentID         *string
	Name             string
	Direction        string
	IsBudgetBase     bool
	IsBudgetRelevant bool
	IsActive         bool
	CreatedAt        time.Time
	UpdatedAt        *time.Time
}
