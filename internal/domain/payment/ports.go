package payment

import (
	"context"
	"time"
)

type InvoiceEntry struct {
	CashFlowID   int32
	Date         time.Time
	Title        string
	Amount       float64
	CategoryName string
}

type Invoice struct {
	PaymentMethodID int32
	Month           time.Time
	Total           float64
	TotalRemaining  float64
	Entries         []InvoiceEntry
}

type Repository interface {
	Create(ctx context.Context, method *PaymentMethod) (*PaymentMethod, error)
	List(ctx context.Context, activeOnly bool) ([]PaymentMethod, error)
	GetByID(ctx context.Context, id int32) (*PaymentMethod, error)
	Update(ctx context.Context, method *PaymentMethod) (*PaymentMethod, error)
	GetInvoiceEntries(ctx context.Context, paymentMethodID int32, month time.Time) ([]InvoiceEntry, error)
	GetOutstandingAmount(ctx context.Context, paymentMethodID int32, month time.Time) (float64, error)
}

type Service interface {
	CreatePaymentMethod(ctx context.Context, name, kind, bankName string, creditLimit *float64, closingDay, dueDay *int32) (*PaymentMethod, error)
	ListPaymentMethods(ctx context.Context) ([]PaymentMethod, error)
	GetInvoice(ctx context.Context, paymentMethodID int32, month time.Time) (*Invoice, error)
	UpdatePaymentMethod(ctx context.Context, id int32, name, kind, bankName string, creditLimit *float64, closingDay, dueDay *int32, isActive bool) (*PaymentMethod, error)
	DeletePaymentMethod(ctx context.Context, id int32) error
}
