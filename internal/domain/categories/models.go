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

type SeedCategoryItem struct {
	Name             string
	Direction        string
	IsBudgetRelevant *bool
	IsBudgetBase     *bool
}

type SeedCategoriesParams struct {
	Preset string
	Names  []string
	Items  []SeedCategoryItem
}

type SeedCategoriesResult struct {
	Created []string
	Skipped []string
}
