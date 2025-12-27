package ledger

import (
	"context"
	"fmt"
	"time"
)

type LedgerService struct {
	repo Repository
}

func NewService(repo Repository) *LedgerService {
	return &LedgerService{repo: repo}
}

func (s *LedgerService) CreateAccount(ctx context.Context, name, accountType, currency string, isActive bool) (*Account, error) {
	account, err := NewAccount(name, accountType, currency, isActive)
	if err != nil {
		return nil, err
	}
	return s.repo.CreateAccount(ctx, account)
}

func (s *LedgerService) ListAccounts(ctx context.Context) ([]*Account, error) {
	return s.repo.ListAccounts(ctx)
}

func (s *LedgerService) CreateTransaction(ctx context.Context, occurredAt time.Time, description, reference, notes string, postings []*Posting) (*Transaction, error) {
	newTransaction, err := NewTransaction(occurredAt, description, reference, notes, postings)
	if err != nil {
		return nil, err
	}

	createdTx, err := s.repo.CreateTransaction(ctx, newTransaction)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	createdPostings := make([]*Posting, len(newTransaction.Postings))
	for i, posting := range newTransaction.Postings {
		posting.TransactionID = createdTx.ID
		createdPosting, err := s.repo.CreatePosting(ctx, posting)
		if err != nil {
			return nil, fmt.Errorf("failed to create posting: %w", err)
		}
		createdPostings[i] = createdPosting
	}

	createdTx.Postings = createdPostings
	return createdTx, nil
}

func (s *LedgerService) ListTransactionsByMonth(ctx context.Context, month time.Time) ([]*Transaction, error) {
	transactions, err := s.repo.ListTransactionsByMonth(ctx, month)
	if err != nil {
		return nil, err
	}

	for _, tx := range transactions {
		postings, err := s.repo.ListPostingsByTransaction(ctx, tx.ID)
		if err != nil {
			return nil, err
		}
		tx.Postings = postings
	}

	return transactions, nil
}
