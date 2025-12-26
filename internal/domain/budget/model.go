package budget

import (
	"errors"
	"time"
)

var (
	ErrInvalidMonth       = errors.New("invalid month")
	ErrInvalidCategory    = errors.New("invalid category for budget")
	ErrInvalidAmount      = errors.New("amount must be non-negative")
	ErrInvalidPercent     = errors.New("percent must be between 0 and 100")
	ErrInvalidMode        = errors.New("invalid budget mode")
	ErrBudgetItemNotFound = errors.New("budget item not found")
)

const (
	ModeAbsolute        = "ABSOLUTE"
	ModePercentOfIncome = "PERCENT_OF_INCOME"
)

type BudgetPeriod struct {
	ID           int32
	Month        time.Time
	AnalysisMode string
	IsClosed     bool
	TotalIncome  float64
	Items        []BudgetItem
}

type BudgetItem struct {
	ID             int32
	BudgetPeriodID int32
	CategoryID     int32
	CategoryName   string
	Mode           string // ABSOLUTE or PERCENT
	PlannedAmount  float64
	ActualAmount   float64 // Calculated at runtime
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
