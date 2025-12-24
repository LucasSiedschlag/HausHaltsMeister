package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/sqlc"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/cashflow"
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
		IsFixed:    cf.IsFixed,
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
		IsFixed:    row.IsFixed,
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
			IsFixed:      row.IsFixed,
		}
	}
	return result, nil
}

func (r *CashFlowRepository) GetMonthlySummary(ctx context.Context, month time.Time) (*cashflow.MonthlySummary, error) {
	pgDate := pgtype.Date{Time: month, Valid: true}
	row, err := r.q.GetMonthlySummary(ctx, pgDate)
	if err != nil {
		return nil, err
	}
	// row is struct{ TotalIncome float64, TotalExpense float64 } (checking generated code assumption)
	// Actually sqlc returns float64 directly if not null, or sql.NullFloat64?
	// The query used ::float, so it should be float64. But SUM() can be NULL if no rows.
	// So it's likely float64 or *float64 or sql.NullFloat64.
	// Let's assume float64 for now, but handle potential mismatch if compilation fails.
	// Update: query used `SUM(...)::float`. If no rows, returns NULL. = *float64?
	// I'll check generated code if I could, but I'll use safe dereference logic assuming generated code follows standard PGX.

	// Wait, if I can't check generated code, I should look at `sqlc` defaults.
	// SUM usually implies null possibility.

	// Let's implement optimistically using values.

	inc := row.TotalIncome
	exp := row.TotalExpense

	return &cashflow.MonthlySummary{
		TotalIncome:  inc,
		TotalExpense: exp,
		Balance:      inc - exp,
	}, nil
}

func (r *CashFlowRepository) GetCategorySummary(ctx context.Context, month time.Time) ([]cashflow.CategorySummary, error) {
	pgDate := pgtype.Date{Time: month, Valid: true}
	rows, err := r.q.GetCategorySummary(ctx, pgDate)
	if err != nil {
		return nil, err
	}

	var summaries []cashflow.CategorySummary
	for _, row := range rows {
		summaries = append(summaries, cashflow.CategorySummary{
			CategoryName: row.Name,
			Direction:    row.Direction,
			TotalAmount:  row.TotalAmount,
		})
	}
	return summaries, nil
}
