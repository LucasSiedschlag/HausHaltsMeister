package cashflow

import "time"

type Direction string

const (
	DirectionIn  Direction = "IN"
	DirectionOut Direction = "OUT"
)

type CashFlow struct {
	ID         int64
	Date       time.Time
	CategoryID int64
	Direction  Direction
	Title      string
	Amount     float64
}

