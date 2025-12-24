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

func (s *PaymentService) CreatePaymentMethod(ctx context.Context, name, kind, bankName string, closingDay, dueDay *int32) (*PaymentMethod, error) {
	m := &PaymentMethod{
		Name:       name,
		Kind:       kind,
		BankName:   bankName,
		ClosingDay: closingDay,
		DueDay:     dueDay,
		IsActive:   true,
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

func (s *PaymentService) GetInvoice(ctx context.Context, paymentMethodID int32, month time.Time) (*Invoice, error) {
	entries, err := s.repo.GetInvoiceEntries(ctx, paymentMethodID, month)
	if err != nil {
		return nil, err
	}

	var total float64
	for _, e := range entries {
		total += e.Amount
	}

	return &Invoice{
		PaymentMethodID: paymentMethodID,
		Month:           month,
		Total:           total,
		Entries:         entries,
	}, nil
}
