package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	ledgerSqlc "github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/sqlc-ledger"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/ledger"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/payment"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepository struct {
	q *ledgerSqlc.Queries
}

func NewPaymentRepository(db *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{
		q: ledgerSqlc.New(db),
	}
}

func (r *PaymentRepository) Create(ctx context.Context, m *payment.PaymentMethod) (*payment.PaymentMethod, error) {
	accountID := m.AccountID
	if accountID == 0 {
		accountType := accountTypeForPaymentKind(m.Kind)
		account, err := r.q.CreateLedgerAccount(ctx, ledgerSqlc.CreateLedgerAccountParams{
			Name:     m.Name,
			Type:     accountType,
			Currency: "BRL",
			IsActive: m.IsActive,
		})
		if err != nil {
			return nil, err
		}
		accountID = account.AccountID
	}

	bank := pgtype.Text{String: m.BankName, Valid: m.BankName != ""}
	limit := pgtype.Numeric{Valid: false}
	if m.CreditLimit != nil {
		limit.Scan(fmt.Sprintf("%.2f", *m.CreditLimit))
	}
	cDay := pgtype.Int4{Valid: false}
	if m.ClosingDay != nil {
		cDay = pgtype.Int4{Int32: *m.ClosingDay, Valid: true}
	}
	dDay := pgtype.Int4{Valid: false}
	if m.DueDay != nil {
		dDay = pgtype.Int4{Int32: *m.DueDay, Valid: true}
	}

	row, err := r.q.CreatePaymentMethod(ctx, ledgerSqlc.CreatePaymentMethodParams{
		AccountID:   accountID,
		Name:        m.Name,
		Kind:        m.Kind,
		BankName:    bank,
		CreditLimit: limit,
		ClosingDay:  cDay,
		DueDay:      dDay,
		IsActive:    m.IsActive,
	})
	if err != nil {
		return nil, err
	}

	return mapPaymentMethodRow(row)
}

func (r *PaymentRepository) Update(ctx context.Context, m *payment.PaymentMethod) (*payment.PaymentMethod, error) {
	if m.AccountID == 0 {
		return nil, fmt.Errorf("account_id is required for payment method update")
	}

	accountType := accountTypeForPaymentKind(m.Kind)
	if _, err := r.q.UpdateLedgerAccount(ctx, ledgerSqlc.UpdateLedgerAccountParams{
		AccountID: m.AccountID,
		Name:      m.Name,
		Type:      accountType,
		IsActive:  m.IsActive,
	}); err != nil {
		return nil, err
	}

	bank := pgtype.Text{String: m.BankName, Valid: m.BankName != ""}
	limit := pgtype.Numeric{Valid: false}
	if m.CreditLimit != nil {
		limit.Scan(fmt.Sprintf("%.2f", *m.CreditLimit))
	}
	cDay := pgtype.Int4{Valid: false}
	if m.ClosingDay != nil {
		cDay = pgtype.Int4{Int32: *m.ClosingDay, Valid: true}
	}
	dDay := pgtype.Int4{Valid: false}
	if m.DueDay != nil {
		dDay = pgtype.Int4{Int32: *m.DueDay, Valid: true}
	}

	row, err := r.q.UpdatePaymentMethod(ctx, ledgerSqlc.UpdatePaymentMethodParams{
		PaymentMethodID: m.ID,
		AccountID:       m.AccountID,
		Name:            m.Name,
		Kind:            m.Kind,
		BankName:        bank,
		CreditLimit:     limit,
		ClosingDay:      cDay,
		DueDay:          dDay,
		IsActive:        m.IsActive,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, payment.ErrPaymentMethodNotFound
		}
		return nil, err
	}

	return mapPaymentMethodRow(row)
}

func (r *PaymentRepository) List(ctx context.Context, activeOnly bool) ([]payment.PaymentMethod, error) {
	filter := pgtype.Bool{Bool: true, Valid: activeOnly}
	if !activeOnly {
		filter = pgtype.Bool{Valid: false}
	}

	rows, err := r.q.ListPaymentMethods(ctx, filter)
	if err != nil {
		return nil, err
	}

	methods := make([]payment.PaymentMethod, len(rows))
	for i, row := range rows {
		method, err := mapPaymentMethodRow(row)
		if err != nil {
			return nil, err
		}
		methods[i] = *method
	}
	return methods, nil
}

func (r *PaymentRepository) GetByID(ctx context.Context, id int32) (*payment.PaymentMethod, error) {
	row, err := r.q.GetPaymentMethod(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return mapPaymentMethodRow(row)
}

func (r *PaymentRepository) GetInvoiceEntries(ctx context.Context, paymentMethodID int32, month time.Time) ([]payment.InvoiceEntry, error) {
	pgDate := pgtype.Date{Time: month, Valid: true}
	rows, err := r.q.GetInvoiceEntries(ctx, ledgerSqlc.GetInvoiceEntriesParams{
		PaymentMethodID: paymentMethodID,
		Column2:         pgDate,
	})
	if err != nil {
		return nil, err
	}

	entries := make([]payment.InvoiceEntry, len(rows))
	for i, row := range rows {
		title := row.Title
		amt := float64(row.Amount)
		entries[i] = payment.InvoiceEntry{
			CashFlowID:   row.TransactionID,
			Date:         row.OccurredAt.Time,
			Title:        title,
			Amount:       amt,
			CategoryName: row.CategoryName,
		}
	}
	return entries, nil
}

func (r *PaymentRepository) GetOutstandingAmount(ctx context.Context, paymentMethodID int32, month time.Time) (float64, error) {
	pgDate := pgtype.Date{Time: month, Valid: true}
	return r.q.GetOutstandingAmount(ctx, ledgerSqlc.GetOutstandingAmountParams{
		PaymentMethodID: paymentMethodID,
		Column2:         pgDate,
	})
}

func mapPaymentMethodRow(row ledgerSqlc.PaymentMethod) (*payment.PaymentMethod, error) {
	var closing, due *int32
	var creditLimit *float64
	if row.CreditLimit.Valid {
		limitVal, _ := row.CreditLimit.Float64Value()
		value := limitVal.Float64
		creditLimit = &value
	}
	if row.ClosingDay.Valid {
		closing = &row.ClosingDay.Int32
	}
	if row.DueDay.Valid {
		due = &row.DueDay.Int32
	}

	return &payment.PaymentMethod{
		ID:          row.PaymentMethodID,
		AccountID:   row.AccountID,
		Name:        row.Name,
		Kind:        row.Kind,
		BankName:    row.BankName.String,
		CreditLimit: creditLimit,
		ClosingDay:  closing,
		DueDay:      due,
		IsActive:    row.IsActive,
	}, nil
}

func accountTypeForPaymentKind(kind string) string {
	switch kind {
	case payment.KindCreditCard:
		return ledger.AccountTypeLiability
	case payment.KindDebitCard, payment.KindCash, payment.KindPix, payment.KindBankSlip:
		return ledger.AccountTypeAsset
	default:
		return ledger.AccountTypeAsset
	}
}
