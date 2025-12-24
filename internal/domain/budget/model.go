package budget

import (
	"errors"
	"time"
)

var (
	ErrInvalidMonth    = errors.New("invalid month")
	ErrInvalidCategory = errors.New("invalid category for budget")
	ErrInvalidAmount   = errors.New("amount must be non-negative")
	ErrInvalidMode     = errors.New("invalid budget mode")
)

const (
	ModeAbsolute = "ABSOLUTE"
	ModePercent  = "PERCENT"
)

type BudgetPeriod struct {
	ID           int32
	Month        time.Time
	AnalysisMode string
	IsClosed     bool
	Items        []BudgetItem
}

type BudgetItem struct {
	ID             int32
	BudgetPeriodID int32
	CategoryID     int32
	CategoryName   string
	Mode           string // ABSOLUTE or PERCENT
	PlannedAmount  float64
	TargetPercent  float64
	Notes          string
}

func NewPeriod(month time.Time) *BudgetPeriod {
	return &BudgetPeriod{
		Month:        month,
		AnalysisMode: "DEFAULT",
		IsClosed:     false,
	}
}
