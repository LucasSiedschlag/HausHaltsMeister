package postgres

import (
	"context"
	"fmt"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/sqlc"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/installment"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InstallmentRepository struct {
	q *sqlc.Queries
}

func NewInstallmentRepository(db *pgxpool.Pool) *InstallmentRepository {
	return &InstallmentRepository{
		q: sqlc.New(db),
	}
}

func (r *InstallmentRepository) CreatePlan(ctx context.Context, plan *installment.InstallmentPlan) (*installment.InstallmentPlan, error) {
	pgDate := pgtype.Date{Time: plan.StartMonth, Valid: true}
	var total, instAmount pgtype.Numeric
	total.Scan(fmt.Sprintf("%.2f", plan.TotalAmount))
	instAmount.Scan(fmt.Sprintf("%.2f", plan.InstallmentAmount))
	pmID := pgtype.Int4{Int32: plan.PaymentMethodID, Valid: true}

	row, err := r.q.CreateInstallmentPlan(ctx, sqlc.CreateInstallmentPlanParams{
		Description:            plan.Description,
		TotalAmount:            total,
		InstallmentCount:       plan.InstallmentCount,
		InstallmentAmount:      instAmount,
		StartDate:              pgDate,
		PaymentMethodID:        pmID,
		StartsOnCurrentInvoice: true, // Default for now
	})
	if err != nil {
		return nil, err
	}

	tVal, _ := row.TotalAmount.Float64Value()
	iVal, _ := row.InstallmentAmount.Float64Value()
	methodID := int32(0)
	if row.PaymentMethodID.Valid {
		methodID = row.PaymentMethodID.Int32
	}

	return &installment.InstallmentPlan{
		ID:                row.InstallmentPlanID,
		Description:       row.Description,
		TotalAmount:       tVal.Float64,
		InstallmentCount:  row.InstallmentCount,
		InstallmentAmount: iVal.Float64,
		StartMonth:        row.StartDate.Time,
		PaymentMethodID:   methodID,
	}, nil
}

func (r *InstallmentRepository) CreateExpenseDetail(ctx context.Context, cashFlowID int32, paymentMethodID int32, planID int32, affectsCardInvoice bool) error {
	plID := pgtype.Int4{Int32: planID, Valid: true}
	pmID := pgtype.Int4{Int32: paymentMethodID, Valid: true} // payment_method_id in exp_details is int4

	return r.q.CreateExpenseDetail(ctx, sqlc.CreateExpenseDetailParams{
		CashFlowID:         cashFlowID,
		PaymentMethodID:    pmID,
		InstallmentPlanID:  plID,
		AffectsCardInvoice: affectsCardInvoice,
	})
}
