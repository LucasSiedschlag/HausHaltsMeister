package category

import (
	"errors"
)

// Direction types
const (
	DirectionIn  = "IN"
	DirectionOut = "OUT"
)

var (
	ErrInvalidDirection = errors.New("invalid direction: must be IN or OUT")
	ErrEmptyName        = errors.New("category name cannot be empty")
)

type Category struct {
	ID               int32
	Name             string
	Direction        string
	IsBudgetRelevant bool
	IsActive         bool
}

func New(name, direction string) (*Category, error) {
	if name == "" {
		return nil, ErrEmptyName
	}
	if direction != DirectionIn && direction != DirectionOut {
		return nil, ErrInvalidDirection
	}

	return &Category{
		Name:             name,
		Direction:        direction,
		IsBudgetRelevant: true,
		IsActive:         true,
	}, nil
}
