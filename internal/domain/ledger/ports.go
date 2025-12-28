package ledger

import (
	"context"
	"time"
)

type Repository interface {
	CreateAccount(ctx context.Context, account *Account) (*Account, error)
	ListAccounts(ctx context.Context) ([]*Account, error)
	CreateTransaction(ctx context.Context, transaction *Transaction) (*Transaction, error)
	UpdateTransaction(ctx context.Context, transaction *Transaction) (*Transaction, error)
	DeleteTransaction(ctx context.Context, transactionID int32) error
	ListTransactionsByMonth(ctx context.Context, month time.Time) ([]*Transaction, error)
	CreatePosting(ctx context.Context, posting *Posting) (*Posting, error)
	UpdatePosting(ctx context.Context, posting *Posting) (*Posting, error)
	ListPostingsByTransaction(ctx context.Context, transactionID int32) ([]*Posting, error)
}

type Service interface {
	CreateAccount(ctx context.Context, name, accountType, currency string, isActive bool) (*Account, error)
	ListAccounts(ctx context.Context) ([]*Account, error)
	CreateTransaction(ctx context.Context, occurredAt time.Time, description, reference, notes string, postings []*Posting) (*Transaction, error)
	UpdateTransaction(ctx context.Context, transactionID int32, occurredAt time.Time, description, reference, notes string, postings []*Posting) (*Transaction, error)
	DeleteTransaction(ctx context.Context, transactionID int32) error
	ListTransactionsByMonth(ctx context.Context, month time.Time) ([]*Transaction, error)
}
