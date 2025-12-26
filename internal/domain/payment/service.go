package payment

import (
	"context"
	"time"
)

type PaymentService struct {
	repo Repository
}

func NewService(repo Repository) *PaymentService {
	return &PaymentService{repo: repo}
}

func (s *PaymentService) CreatePaymentMethod(ctx context.Context, name, kind, bankName string, creditLimit *float64, closingDay, dueDay *int32) (*PaymentMethod, error) {
	m := &PaymentMethod{
		Name:        name,
		Kind:        kind,
		BankName:    bankName,
		CreditLimit: creditLimit,
		ClosingDay:  closingDay,
		DueDay:      dueDay,
		IsActive:    true,
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, m)
}

func (s *PaymentService) ListPaymentMethods(ctx context.Context) ([]PaymentMethod, error) {
	// List ALL for management purposes. In future we can add ListActiveMethods.
	return s.repo.List(ctx, false)
}

func (s *PaymentService) UpdatePaymentMethod(ctx context.Context, id int32, name, kind, bankName string, creditLimit *float64, closingDay, dueDay *int32, isActive bool) (*PaymentMethod, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrPaymentMethodNotFound
	}

	updated := &PaymentMethod{
		ID:          id,
		Name:        name,
		Kind:        kind,
		BankName:    bankName,
		CreditLimit: creditLimit,
		ClosingDay:  closingDay,
		DueDay:      dueDay,
		IsActive:    isActive,
	}
	if err := updated.Validate(); err != nil {
		return nil, err
	}

	return s.repo.Update(ctx, updated)
}

func (s *PaymentService) DeletePaymentMethod(ctx context.Context, id int32) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrPaymentMethodNotFound
	}

	existing.IsActive = false
	_, err = s.repo.Update(ctx, existing)
	if err != nil {
		return err
	}

	return nil
}

func (s *PaymentService) GetInvoice(ctx context.Context, paymentMethodID int32, month time.Time) (*Invoice, error) {
	entries, err := s.repo.GetInvoiceEntries(ctx, paymentMethodID, month)
	if err != nil {
		return nil, err
	}

	var total float64
	for _, e := range entries {
		total += e.Amount
	}

	totalRemaining, err := s.repo.GetOutstandingAmount(ctx, paymentMethodID, month)
	if err != nil {
		return nil, err
	}

	return &Invoice{
		PaymentMethodID: paymentMethodID,
		Month:           month,
		Total:           total,
		TotalRemaining:  totalRemaining,
		Entries:         entries,
	}, nil
}
