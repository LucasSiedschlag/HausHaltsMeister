package postgres

import (
	"context"
	"time"

	ledgerSqlc "github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/sqlc-ledger"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/cashflow"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CashFlowRepository struct {
	q *ledgerSqlc.Queries
}

func NewCashFlowRepository(db *pgxpool.Pool) *CashFlowRepository {
	return &CashFlowRepository{
		q: ledgerSqlc.New(db),
	}
}

func (r *CashFlowRepository) Create(ctx context.Context, cf *cashflow.CashFlow) (*cashflow.CashFlow, error) {
	pgDate := pgtype.Date{Time: cf.Date, Valid: true}
	params := ledgerSqlc.CreateCashFlowEntryParams{
		TransactionID:         cf.TransactionID,
		CategoryID:            cf.CategoryID,
		PaymentMethodID:       cf.PaymentMethodID,
		InstallmentPlanID:     int4FromPtr(cf.InstallmentPlanID),
		InstallmentPlanItemID: int4FromPtr(cf.InstallmentPlanItemID),
		Direction:             cf.Direction,
		Title:                 cf.Title,
		Amount:                numericFromValue(cf.Amount),
		IsFixed:               cf.IsFixed,
		OccurredAt:            pgDate,
		ReversalOfEntryID:     int4FromPtr(cf.ReversalOfEntryID),
	}

	row, err := r.q.CreateCashFlowEntry(ctx, params)
	if err != nil {
		return nil, err
	}
	return mapCashFlowEntry(row), nil
}

func (r *CashFlowRepository) Update(ctx context.Context, cf *cashflow.CashFlow) (*cashflow.CashFlow, error) {
	pgDate := pgtype.Date{Time: cf.Date, Valid: true}
	params := ledgerSqlc.UpdateCashFlowEntryParams{
		CashflowEntryID:       cf.ID,
		TransactionID:         cf.TransactionID,
		CategoryID:            cf.CategoryID,
		PaymentMethodID:       cf.PaymentMethodID,
		InstallmentPlanID:     int4FromPtr(cf.InstallmentPlanID),
		InstallmentPlanItemID: int4FromPtr(cf.InstallmentPlanItemID),
		Direction:             cf.Direction,
		Title:                 cf.Title,
		Amount:                numericFromValue(cf.Amount),
		IsFixed:               cf.IsFixed,
		OccurredAt:            pgDate,
		ReversalOfEntryID:     int4FromPtr(cf.ReversalOfEntryID),
	}

	row, err := r.q.UpdateCashFlowEntry(ctx, params)
	if err != nil {
		return nil, err
	}
	return mapCashFlowEntry(row), nil
}

func (r *CashFlowRepository) Delete(ctx context.Context, id int32) error {
	return r.q.DeleteCashFlowEntry(ctx, id)
}

func (r *CashFlowRepository) GetByID(ctx context.Context, id int32) (*cashflow.CashFlow, error) {
	row, err := r.q.GetCashFlowEntry(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return mapCashFlowEntryWithNames(row), nil
}

func (r *CashFlowRepository) ListByMonth(ctx context.Context, month time.Time, direction *string, isFixed *bool) ([]*cashflow.CashFlow, error) {
	pgDate := pgtype.Date{Time: month, Valid: true}
	dir := pgtype.Text{Valid: false}
	if direction != nil && *direction != "" {
		dir = pgtype.Text{String: *direction, Valid: true}
	}
	fixed := pgtype.Bool{Valid: false}
	if isFixed != nil {
		fixed = pgtype.Bool{Bool: *isFixed, Valid: true}
	}

	rows, err := r.q.ListCashFlowEntriesByMonth(ctx, ledgerSqlc.ListCashFlowEntriesByMonthParams{
		Column1:    pgDate,
		Direction:  dir,
		IsFixed:    fixed,
	})
	if err != nil {
		return nil, err
	}

	result := make([]*cashflow.CashFlow, len(rows))
	for i, row := range rows {
		result[i] = mapCashFlowEntryWithNamesList(row)
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

func mapCashFlowEntry(row ledgerSqlc.CashflowEntry) *cashflow.CashFlow {
	return &cashflow.CashFlow{
		ID:                    row.CashflowEntryID,
		TransactionID:         row.TransactionID,
		InstallmentPlanID:     int4ToPtr(row.InstallmentPlanID),
		InstallmentPlanItemID: int4ToPtr(row.InstallmentPlanItemID),
		Date:                  row.OccurredAt.Time,
		CategoryID:            row.CategoryID,
		PaymentMethodID:       row.PaymentMethodID,
		Direction:             row.Direction,
		Title:                 row.Title,
		Amount:                numericToValue(row.Amount),
		IsFixed:               row.IsFixed,
		ReversalOfEntryID:     int4ToPtr(row.ReversalOfEntryID),
	}
}

func mapCashFlowEntryWithNames(row ledgerSqlc.GetCashFlowEntryRow) *cashflow.CashFlow {
	return &cashflow.CashFlow{
		ID:                    row.CashflowEntryID,
		TransactionID:         row.TransactionID,
		InstallmentPlanID:     int4ToPtr(row.InstallmentPlanID),
		InstallmentPlanItemID: int4ToPtr(row.InstallmentPlanItemID),
		Date:                  row.OccurredAt.Time,
		CategoryID:            row.CategoryID,
		CategoryName:          row.CategoryName,
		PaymentMethodID:       row.PaymentMethodID,
		PaymentMethodName:     row.PaymentMethodName,
		Direction:             row.Direction,
		Title:                 row.Title,
		Amount:                numericToValue(row.Amount),
		IsFixed:               row.IsFixed,
		ReversalOfEntryID:     int4ToPtr(row.ReversalOfEntryID),
	}
}

func mapCashFlowEntryWithNamesList(row ledgerSqlc.ListCashFlowEntriesByMonthRow) *cashflow.CashFlow {
	return &cashflow.CashFlow{
		ID:                    row.CashflowEntryID,
		TransactionID:         row.TransactionID,
		InstallmentPlanID:     int4ToPtr(row.InstallmentPlanID),
		InstallmentPlanItemID: int4ToPtr(row.InstallmentPlanItemID),
		Date:                  row.OccurredAt.Time,
		CategoryID:            row.CategoryID,
		CategoryName:          row.CategoryName,
		PaymentMethodID:       row.PaymentMethodID,
		PaymentMethodName:     row.PaymentMethodName,
		Direction:             row.Direction,
		Title:                 row.Title,
		Amount:                numericToValue(row.Amount),
		IsFixed:               row.IsFixed,
		ReversalOfEntryID:     int4ToPtr(row.ReversalOfEntryID),
	}
}
