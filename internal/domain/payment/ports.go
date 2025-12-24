package payment

import "context"

type Repository interface {
	Create(ctx context.Context, method *PaymentMethod) (*PaymentMethod, error)
	List(ctx context.Context, activeOnly bool) ([]PaymentMethod, error)
	GetByID(ctx context.Context, id int32) (*PaymentMethod, error)
}

type Service interface {
	CreatePaymentMethod(ctx context.Context, name, kind, bankName string, closingDay, dueDay *int32) (*PaymentMethod, error)
	ListPaymentMethods(ctx context.Context) ([]PaymentMethod, error)
}
