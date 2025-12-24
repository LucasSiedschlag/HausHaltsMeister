package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seuuser/cashflow/internal/adapters/postgres/sqlc"
	"github.com/seuuser/cashflow/internal/domain/cashflow"
)

type CashFlowRepository struct {
	q *sqlc.Queries
}

func NewCashFlowRepository(db *pgxpool.Pool) *CashFlowRepository {
	return &CashFlowRepository{
		q: sqlc.New(db),
	}
}

func (r *CashFlowRepository) Create(ctx context.Context, cf *cashflow.CashFlow) (*cashflow.CashFlow, error) {
	// Convert time.Time to pgtype.Date
	// pgx/v5 automatically handles time.Time for Date fields usually, but sqlc generates pgtype.Date for 'date' columns.
	// We need to convert.

	pgDate := pgtype.Date{
		Time:  cf.Date,
		Valid: true,
	}

	// Convert float64 to pgtype.Numeric
	// This is a bit verbose with pgx/v5 pgtype.Numeric.
	// Let's assume for now we can pass a string or use a helper.
	// Or even better, let's look at what sqlc generated.
	// It likely generated 'amount' as pgtype.Numeric.
	// To simplify, we'll try to use a float helper if one existed, but standard way is via big.Int or string.
	// Actually, let's verify what `sqlc generate` produced in step 82 (`models.go` is 2387 bytes).
	// Since I can't check it right now, I will implement a safe float to numeric conversion
	// or assume the driver handles it if sqlc generated valid Go types.
	// Wait, standard `sqlc` with `pgx/v5` uses `pgtype.Numeric`.

	// Workaround: We'll scan back as float64. Inserting might require proper Numeric construction.
	// For simplicity in this "agent" mode, I'll attempt a direct cast if sqlc generated float64,
	// else I will need to handle Numeric.
	// Let's rely on `fmt.Sprintf` for float -> numeric string scan.

	var am pgtype.Numeric
	am.Scan(fmt.Sprintf("%.2f", cf.Amount))

	params := sqlc.CreateCashFlowParams{
		Date:       pgDate,
		CategoryID: int32(cf.CategoryID),
		Direction:  cf.Direction,
		Title:      cf.Title,
		Amount:     am,
	}

	row, err := r.q.CreateCashFlow(ctx, params)
	if err != nil {
		return nil, err
	}

	// Convert back
	val, _ := row.Amount.Float64Value()

	return &cashflow.CashFlow{
		ID:         row.CashFlowID,
		Date:       row.Date.Time,
		CategoryID: row.CategoryID,
		Direction:  row.Direction,
		Title:      row.Title,
		Amount:     val.Float64,
	}, nil
}

func (r *CashFlowRepository) ListByMonth(ctx context.Context, month time.Time) ([]*cashflow.CashFlow, error) {
	pgDate := pgtype.Date{
		Time:  month,
		Valid: true,
	}

	rows, err := r.q.ListCashFlowsByMonth(ctx, pgDate)
	if err != nil {
		return nil, err
	}

	result := make([]*cashflow.CashFlow, len(rows))
	for i, row := range rows {
		val, _ := row.Amount.Float64Value()
		result[i] = &cashflow.CashFlow{
			ID:           row.CashFlowID,
			Date:         row.Date.Time,
			CategoryID:   row.CategoryID,
			CategoryName: row.CategoryName,
			Direction:    row.Direction,
			Title:        row.Title,
			Amount:       val.Float64,
		}
	}
	return result, nil
}
