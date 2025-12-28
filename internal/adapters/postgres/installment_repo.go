package postgres

import (
	"context"

	ledgerSqlc "github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/sqlc-ledger"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/installment"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InstallmentRepository struct {
	q *ledgerSqlc.Queries
}

func NewInstallmentRepository(db *pgxpool.Pool) *InstallmentRepository {
	return &InstallmentRepository{
		q: ledgerSqlc.New(db),
	}
}

func (r *InstallmentRepository) CreatePlan(ctx context.Context, plan *installment.InstallmentPlan) (*installment.InstallmentPlan, error) {
	pgDate := pgtype.Date{Time: plan.StartDate, Valid: true}
	total := numericFromValue(plan.TotalAmount)
	instAmount := numericFromValue(plan.InstallmentAmount)
	count := pgtype.Int4{Int32: plan.InstallmentCount, Valid: plan.InstallmentCount > 0}
	pmID := pgtype.Int4{Int32: plan.PaymentMethodID, Valid: plan.PaymentMethodID > 0}
	accountID := pgtype.Int4{Valid: false}
	categoryID := pgtype.Int4{Int32: plan.CategoryID, Valid: plan.CategoryID > 0}

	row, err := r.q.CreateInstallmentPlan(ctx, ledgerSqlc.CreateInstallmentPlanParams{
		Description:            plan.Description,
		PlanType:               plan.PlanType,
		TotalAmount:            total,
		InstallmentCount:       count,
		InstallmentAmount:      instAmount,
		StartDate:              pgDate,
		PaymentMethodID:        pmID,
		AccountID:              accountID,
		CategoryID:             categoryID,
		StartsOnCurrentInvoice: true,
	})
	if err != nil {
		return nil, err
	}

	return mapInstallmentPlanRow(row)
}

func (r *InstallmentRepository) CreatePlanItem(ctx context.Context, item *installment.InstallmentPlanItem) (*installment.InstallmentPlanItem, error) {
	pgDate := pgtype.Date{Time: item.DueDate, Valid: true}
	amount := numericFromValue(item.Amount)
	txID := pgtype.Int4{Valid: false}
	if item.TransactionID != nil {
		txID = pgtype.Int4{Int32: *item.TransactionID, Valid: true}
	}

	row, err := r.q.CreateInstallmentPlanItem(ctx, ledgerSqlc.CreateInstallmentPlanItemParams{
		InstallmentPlanID: item.PlanID,
		Sequence:          item.Sequence,
		DueDate:           pgDate,
		Amount:            amount,
		Status:            "POSTED",
		TransactionID:     txID,
		ExtraAmount:       numericFromValue(0),
		IsPaid:            false,
		PaidAt:            pgtype.Timestamptz{Valid: false},
	})
	if err != nil {
		return nil, err
	}

	return &installment.InstallmentPlanItem{
		ID:            row.InstallmentPlanItemID,
		PlanID:        row.InstallmentPlanID,
		Sequence:      row.Sequence,
		DueDate:       row.DueDate.Time,
		Amount:        numericToValue(row.Amount),
		TransactionID: int4ToPtr(row.TransactionID),
	}, nil
}

func (r *InstallmentRepository) ListPlanItemsByPlan(ctx context.Context, planID int32) ([]*installment.InstallmentPlanItem, error) {
	rows, err := r.q.ListInstallmentPlanItemsByPlan(ctx, planID)
	if err != nil {
		return nil, err
	}

	items := make([]*installment.InstallmentPlanItem, len(rows))
	for i, row := range rows {
		items[i] = &installment.InstallmentPlanItem{
			ID:            row.InstallmentPlanItemID,
			PlanID:        row.InstallmentPlanID,
			Sequence:      row.Sequence,
			DueDate:       row.DueDate.Time,
			Amount:        numericToValue(row.Amount),
			TransactionID: int4ToPtr(row.TransactionID),
		}
	}
	return items, nil
}

func mapInstallmentPlanRow(row ledgerSqlc.InstallmentPlan) (*installment.InstallmentPlan, error) {
	total, _ := row.TotalAmount.Float64Value()
	instAmount, _ := row.InstallmentAmount.Float64Value()
	methodID := int32(0)
	if row.PaymentMethodID.Valid {
		methodID = row.PaymentMethodID.Int32
	}
	categoryID := int32(0)
	if row.CategoryID.Valid {
		categoryID = row.CategoryID.Int32
	}
	count := int32(0)
	if row.InstallmentCount.Valid {
		count = row.InstallmentCount.Int32
	}

	return &installment.InstallmentPlan{
		ID:                row.InstallmentPlanID,
		Description:       row.Description,
		PlanType:          row.PlanType,
		TotalAmount:       total.Float64,
		InstallmentCount:  count,
		InstallmentAmount: instAmount.Float64,
		StartDate:         row.StartDate.Time,
		PaymentMethodID:   methodID,
		CategoryID:        categoryID,
	}, nil
}
