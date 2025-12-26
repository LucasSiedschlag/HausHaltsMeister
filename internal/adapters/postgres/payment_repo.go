package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/sqlc"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/payment"
)

type PaymentRepository struct {
	q *sqlc.Queries
}

func NewPaymentRepository(db *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{
		q: sqlc.New(db),
	}
}

func (r *PaymentRepository) Create(ctx context.Context, m *payment.PaymentMethod) (*payment.PaymentMethod, error) {
	// Nullable handling
	bank := pgtype.Text{String: m.BankName, Valid: m.BankName != ""}
	cDay := pgtype.Int4{Valid: false}
	if m.ClosingDay != nil {
		cDay = pgtype.Int4{Int32: *m.ClosingDay, Valid: true}
	}
	dDay := pgtype.Int4{Valid: false}
	if m.DueDay != nil {
		dDay = pgtype.Int4{Int32: *m.DueDay, Valid: true}
	}

	row, err := r.q.CreatePaymentMethod(ctx, sqlc.CreatePaymentMethodParams{
		Name:       m.Name,
		Kind:       m.Kind,
		BankName:   bank,
		ClosingDay: cDay,
		DueDay:     dDay,
		IsActive:   m.IsActive,
	})
	if err != nil {
		return nil, err
	}

	var closing, due *int32
	if row.ClosingDay.Valid {
		closing = &row.ClosingDay.Int32
	}
	if row.DueDay.Valid {
		due = &row.DueDay.Int32
	}

	return &payment.PaymentMethod{
		ID:         row.PaymentMethodID,
		Name:       row.Name,
		Kind:       row.Kind,
		BankName:   row.BankName.String,
		ClosingDay: closing,
		DueDay:     due,
		IsActive:   row.IsActive,
	}, nil
}

func (r *PaymentRepository) Update(ctx context.Context, m *payment.PaymentMethod) (*payment.PaymentMethod, error) {
	bank := pgtype.Text{String: m.BankName, Valid: m.BankName != ""}
	cDay := pgtype.Int4{Valid: false}
	if m.ClosingDay != nil {
		cDay = pgtype.Int4{Int32: *m.ClosingDay, Valid: true}
	}
	dDay := pgtype.Int4{Valid: false}
	if m.DueDay != nil {
		dDay = pgtype.Int4{Int32: *m.DueDay, Valid: true}
	}

	row, err := r.q.UpdatePaymentMethod(ctx, sqlc.UpdatePaymentMethodParams{
		PaymentMethodID: m.ID,
		Name:            m.Name,
		Kind:            m.Kind,
		BankName:        bank,
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

	var closing, due *int32
	if row.ClosingDay.Valid {
		closing = &row.ClosingDay.Int32
	}
	if row.DueDay.Valid {
		due = &row.DueDay.Int32
	}

	return &payment.PaymentMethod{
		ID:         row.PaymentMethodID,
		Name:       row.Name,
		Kind:       row.Kind,
		BankName:   row.BankName.String,
		ClosingDay: closing,
		DueDay:     due,
		IsActive:   row.IsActive,
	}, nil
}

func (r *PaymentRepository) List(ctx context.Context, activeOnly bool) ([]payment.PaymentMethod, error) {
	// Filter handling
	// Query: WHERE ($1::boolean IS NULL OR is_active = $1)
	// If activeOnly is true, we pass true. If we want all, we pass NULL?
	// But bool cannot be NULL in Go.
	// We need sqlc params to accept pgtype.Bool/Int or handle it.
	// Let's check generated code signature.
	// Likely: func (q *Queries) ListPaymentMethods(ctx context.Context, dollar_1 pgtype.Bool)

	filter := pgtype.Bool{Bool: true, Valid: activeOnly}
	if !activeOnly {
		// If we want ALL, we want Param to be NULL.
		filter = pgtype.Bool{Valid: false}
		// Wait, user might want INACTIVE only.
		// My interface says `activeOnly`. I assume implementation: if activeOnly=true, return active. If false, return ALL.
	}

	rows, err := r.q.ListPaymentMethods(ctx, filter)
	if err != nil {
		return nil, err
	}

	methods := make([]payment.PaymentMethod, len(rows))
	for i, row := range rows {
		var closing, due *int32
		if row.ClosingDay.Valid {
			closing = &row.ClosingDay.Int32
		}
		if row.DueDay.Valid {
			due = &row.DueDay.Int32
		}
		methods[i] = payment.PaymentMethod{
			ID:         row.PaymentMethodID,
			Name:       row.Name,
			Kind:       row.Kind,
			BankName:   row.BankName.String,
			ClosingDay: closing,
			DueDay:     due,
			IsActive:   row.IsActive,
		}
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
	var closing, due *int32
	if row.ClosingDay.Valid {
		closing = &row.ClosingDay.Int32
	}
	if row.DueDay.Valid {
		due = &row.DueDay.Int32
	}
	return &payment.PaymentMethod{
		ID:         row.PaymentMethodID,
		Name:       row.Name,
		Kind:       row.Kind,
		BankName:   row.BankName.String,
		ClosingDay: closing,
		DueDay:     due,
		IsActive:   row.IsActive,
	}, nil
}

func (r *PaymentRepository) GetInvoiceEntries(ctx context.Context, paymentMethodID int32, month time.Time) ([]payment.InvoiceEntry, error) {
	pgDate := pgtype.Date{Time: month, Valid: true}
	rows, err := r.q.GetInvoiceEntries(ctx, sqlc.GetInvoiceEntriesParams{
		PaymentMethodID: pgtype.Int4{Int32: paymentMethodID, Valid: true},
		Column2:         pgDate,
	})
	if err != nil {
		return nil, err
	}

	entries := make([]payment.InvoiceEntry, len(rows))
	for i, row := range rows {
		amt, _ := row.Amount.Float64Value()
		entries[i] = payment.InvoiceEntry{
			CashFlowID:   row.CashFlowID,
			Date:         row.Date.Time,
			Title:        row.Title,
			Amount:       amt.Float64,
			CategoryName: row.CategoryName,
		}
	}
	return entries, nil
}
