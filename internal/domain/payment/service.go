package payment

import (
	"context"
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
	// List active only or all? Let's default to all for now or active?
	// The port said "List(activeOnly bool)" in Repo, but Service interface says "ListPaymentMethods".
	// Let's return all for config, or default active.
	// Actually for management UI we usually want all. Dropdowns want active.
	// Let's just return ALL for now.
	return s.repo.List(ctx, false) // false -> Valid: false -> List ALL
}
