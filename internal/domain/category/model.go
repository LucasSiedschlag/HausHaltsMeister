package category

import (
	"errors"
	"time"
)

// Direction types
const (
	DirectionIn  = "IN"
	DirectionOut = "OUT"
)

var (
	ErrInvalidDirection = errors.New("invalid direction: must be IN or OUT")
	ErrEmptyName        = errors.New("category name cannot be empty")
	ErrCategoryNotFound = errors.New("category not found")
)

type Category struct {
	ID               int32
	Name             string
	Direction        string
	IsBudgetRelevant bool
	IsActive         bool
	InactiveFromMonth *time.Time
}

func New(name, direction string, isBudgetRelevant bool) (*Category, error) {
	if name == "" {
		return nil, ErrEmptyName
	}
	if direction != DirectionIn && direction != DirectionOut {
		return nil, ErrInvalidDirection
	}

	return &Category{
		Name:             name,
		Direction:        direction,
		IsBudgetRelevant: isBudgetRelevant,
		IsActive:         true,
	}, nil
}
