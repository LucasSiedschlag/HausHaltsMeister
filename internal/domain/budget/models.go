package budget

import "time"

type Plan struct {
	ID        string
	LedgerID  string
	Name      string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type Version struct {
	ID                 string
	PlanID             string
	EffectiveFromMonth time.Time
	CreatedByUserID    string
	CreatedAt          time.Time
	UpdatedAt          *time.Time
	Lines              []Line
}

type Line struct {
	ID              string
	VersionID       string
	CategoryID      string
	Percent         float64
	IncludeChildren bool
	CreatedAt       time.Time
	UpdatedAt       *time.Time
}

type MonthlyLine struct {
	CategoryID       string
	Percent          float64
	IncludeChildren  bool
	BudgetLimitCents int64
	SpentActualCents int64
	DeltaCents       int64
	UsagePct         float64
}

type MonthlySummary struct {
	Month              time.Time
	Version            *Version
	IncomeBaseCents    int64
	Lines              []MonthlyLine
	OutsideBudgetCents int64
}
