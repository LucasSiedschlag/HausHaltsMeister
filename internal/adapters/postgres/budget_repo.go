package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/sqlc"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/budget"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BudgetRepository struct {
	q *sqlc.Queries
}

func NewBudgetRepository(db *pgxpool.Pool) *BudgetRepository {
	return &BudgetRepository{
		q: sqlc.New(db),
	}
}

func (r *BudgetRepository) GetPeriodByMonth(ctx context.Context, month time.Time) (*budget.BudgetPeriod, error) {
	pgDate := pgtype.Date{Time: month, Valid: true}

	row, err := r.q.GetBudgetPeriodByMonth(ctx, pgDate)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &budget.BudgetPeriod{
		ID:           row.BudgetPeriodID,
		Month:        row.Month.Time,
		AnalysisMode: row.AnalysisMode.String, // Assuming sqlc generates sql.NullString or similar, need to check
		IsClosed:     row.IsClosed,
	}, nil
}

func (r *BudgetRepository) GetLatestPeriodWithItemsBefore(ctx context.Context, month time.Time) (*budget.BudgetPeriod, error) {
	pgDate := pgtype.Date{Time: month, Valid: true}

	row, err := r.q.GetLatestBudgetPeriodWithItemsBefore(ctx, pgDate)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &budget.BudgetPeriod{
		ID:           row.BudgetPeriodID,
		Month:        row.Month.Time,
		AnalysisMode: row.AnalysisMode.String,
		IsClosed:     row.IsClosed,
	}, nil
}

func (r *BudgetRepository) CreatePeriod(ctx context.Context, period *budget.BudgetPeriod) (*budget.BudgetPeriod, error) {
	pgDate := pgtype.Date{Time: period.Month, Valid: true}

	// Handle AnalysisMode (string -> nullstring or just string depending on schema)
	// Schema says varchar(20), nullable. sqlc likely generates pgtype.Text or sql.NullString.
	// Let's assume pgtype.Text for pgx/v5
	mode := pgtype.Text{String: period.AnalysisMode, Valid: period.AnalysisMode != ""}

	params := sqlc.CreateBudgetPeriodParams{
		Month:        pgDate,
		AnalysisMode: mode,
		IsClosed:     period.IsClosed,
	}

	row, err := r.q.CreateBudgetPeriod(ctx, params)
	if err != nil {
		return nil, err
	}

	return &budget.BudgetPeriod{
		ID:           row.BudgetPeriodID,
		Month:        row.Month.Time,
		AnalysisMode: row.AnalysisMode.String,
		IsClosed:     row.IsClosed,
	}, nil
}

func (r *BudgetRepository) UpsertItem(ctx context.Context, item *budget.BudgetItem) (*budget.BudgetItem, error) {
	// Numeric conversion
	var planned pgtype.Numeric
	planned.Scan(fmt.Sprintf("%.2f", item.PlannedAmount))

	var target pgtype.Numeric
	target.Scan(fmt.Sprintf("%.2f", item.TargetPercent))

	// Mode handling
	// item.Mode is string, DB column is varchar(30) NOT NULL.

	// Notes handling (nullable text)
	notes := pgtype.Text{String: item.Notes, Valid: item.Notes != ""}

	params := sqlc.UpsertBudgetItemParams{
		BudgetPeriodID: item.BudgetPeriodID,
		CategoryID:     item.CategoryID,
		Mode:           item.Mode,
		PlannedAmount:  planned, // Nullable in DB? No, in DB it is. Wait, schema: planned_amount decimal(14,2) (nullable).
		TargetPercent:  target,
		Notes:          notes,
	}

	row, err := r.q.UpsertBudgetItem(ctx, params)
	if err != nil {
		return nil, err
	}

	pVal, _ := row.PlannedAmount.Float64Value()
	tVal, _ := row.TargetPercent.Float64Value()

	return &budget.BudgetItem{
		ID:             row.BudgetItemID,
		BudgetPeriodID: row.BudgetPeriodID,
		CategoryID:     row.CategoryID,
		Mode:           row.Mode,
		PlannedAmount:  pVal.Float64,
		TargetPercent:  tVal.Float64,
		Notes:          row.Notes.String,
	}, nil
}

func (r *BudgetRepository) GetItemsByPeriod(ctx context.Context, periodID int32) ([]budget.BudgetItem, error) {
	rows, err := r.q.GetBudgetItemsByPeriod(ctx, periodID)
	if err != nil {
		return nil, err
	}

	items := make([]budget.BudgetItem, len(rows))
	for i, row := range rows {
		pVal, _ := row.PlannedAmount.Float64Value()
		tVal, _ := row.TargetPercent.Float64Value()

		items[i] = budget.BudgetItem{
			ID:             row.BudgetItemID,
			BudgetPeriodID: row.BudgetPeriodID,
			CategoryID:     row.CategoryID,
			CategoryName:   row.CategoryName,
			Mode:           row.Mode,
			PlannedAmount:  pVal.Float64,
			TargetPercent:  tVal.Float64,
			Notes:          row.Notes.String,
		}
	}
	return items, nil
}

func (r *BudgetRepository) GetItemByID(ctx context.Context, id int32) (*budget.BudgetItem, error) {
	row, err := r.q.GetBudgetItemByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	pVal, _ := row.PlannedAmount.Float64Value()
	tVal, _ := row.TargetPercent.Float64Value()

	return &budget.BudgetItem{
		ID:             row.BudgetItemID,
		BudgetPeriodID: row.BudgetPeriodID,
		CategoryID:     row.CategoryID,
		CategoryName:   row.CategoryName,
		Mode:           row.Mode,
		PlannedAmount:  pVal.Float64,
		TargetPercent:  tVal.Float64,
		Notes:          row.Notes.String,
	}, nil
}

func (r *BudgetRepository) UpdateItem(ctx context.Context, item *budget.BudgetItem) (*budget.BudgetItem, error) {
	var planned pgtype.Numeric
	planned.Scan(fmt.Sprintf("%.2f", item.PlannedAmount))

	var target pgtype.Numeric
	target.Scan(fmt.Sprintf("%.2f", item.TargetPercent))

	notes := pgtype.Text{String: item.Notes, Valid: item.Notes != ""}

	params := sqlc.UpdateBudgetItemParams{
		BudgetItemID:  item.ID,
		Mode:          item.Mode,
		PlannedAmount: planned,
		TargetPercent: target,
		Notes:         notes,
	}

	row, err := r.q.UpdateBudgetItem(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, budget.ErrBudgetItemNotFound
		}
		return nil, err
	}

	pVal, _ := row.PlannedAmount.Float64Value()
	tVal, _ := row.TargetPercent.Float64Value()

	return &budget.BudgetItem{
		ID:             row.BudgetItemID,
		BudgetPeriodID: row.BudgetPeriodID,
		CategoryID:     row.CategoryID,
		CategoryName:   row.CategoryName,
		Mode:           row.Mode,
		PlannedAmount:  pVal.Float64,
		TargetPercent:  tVal.Float64,
		Notes:          row.Notes.String,
	}, nil
}
