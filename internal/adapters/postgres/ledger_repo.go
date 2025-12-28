package postgres

import (
	"context"
	"fmt"
	"time"

	ledgerSqlc "github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/sqlc-ledger"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/ledger"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LedgerRepository struct {
	q *ledgerSqlc.Queries
}

func NewLedgerRepository(db *pgxpool.Pool) *LedgerRepository {
	return &LedgerRepository{
		q: ledgerSqlc.New(db),
	}
}

func (r *LedgerRepository) CreateAccount(ctx context.Context, account *ledger.Account) (*ledger.Account, error) {
	params := ledgerSqlc.CreateLedgerAccountParams{
		Name:     account.Name,
		Type:     account.Type,
		Currency: account.Currency,
		IsActive: account.IsActive,
	}

	row, err := r.q.CreateLedgerAccount(ctx, params)
	if err != nil {
		return nil, err
	}

	return &ledger.Account{
		ID:        row.AccountID,
		Name:      row.Name,
		Type:      row.Type,
		Currency:  row.Currency,
		IsActive:  row.IsActive,
		CreatedAt: toTime(row.CreatedAt),
		UpdatedAt: toTime(row.UpdatedAt),
	}, nil
}

func (r *LedgerRepository) ListAccounts(ctx context.Context) ([]*ledger.Account, error) {
	rows, err := r.q.ListLedgerAccounts(ctx)
	if err != nil {
		return nil, err
	}

	accounts := make([]*ledger.Account, len(rows))
	for i, row := range rows {
		accounts[i] = &ledger.Account{
			ID:        row.AccountID,
			Name:      row.Name,
			Type:      row.Type,
			Currency:  row.Currency,
			IsActive:  row.IsActive,
			CreatedAt: toTime(row.CreatedAt),
			UpdatedAt: toTime(row.UpdatedAt),
		}
	}

	return accounts, nil
}

func (r *LedgerRepository) CreateTransaction(ctx context.Context, transaction *ledger.Transaction) (*ledger.Transaction, error) {
	pgDate := pgtype.Date{Time: transaction.OccurredAt, Valid: true}
	params := ledgerSqlc.CreateLedgerTransactionParams{
		OccurredAt:  pgDate,
		Description: transaction.Description,
		Reference:   textFromString(transaction.Reference),
		Notes:       textFromString(transaction.Notes),
	}

	row, err := r.q.CreateLedgerTransaction(ctx, params)
	if err != nil {
		return nil, err
	}

	return &ledger.Transaction{
		ID:          row.TransactionID,
		OccurredAt:  row.OccurredAt.Time,
		Description: row.Description,
		Reference:   textToString(row.Reference),
		Notes:       textToString(row.Notes),
		CreatedAt:   toTime(row.CreatedAt),
		UpdatedAt:   toTime(row.UpdatedAt),
	}, nil
}

func (r *LedgerRepository) UpdateTransaction(ctx context.Context, transaction *ledger.Transaction) (*ledger.Transaction, error) {
	pgDate := pgtype.Date{Time: transaction.OccurredAt, Valid: true}
	params := ledgerSqlc.UpdateLedgerTransactionParams{
		TransactionID: transaction.ID,
		OccurredAt:    pgDate,
		Description:   transaction.Description,
		Reference:     textFromString(transaction.Reference),
		Notes:         textFromString(transaction.Notes),
	}

	row, err := r.q.UpdateLedgerTransaction(ctx, params)
	if err != nil {
		return nil, err
	}

	return &ledger.Transaction{
		ID:          row.TransactionID,
		OccurredAt:  row.OccurredAt.Time,
		Description: row.Description,
		Reference:   textToString(row.Reference),
		Notes:       textToString(row.Notes),
		CreatedAt:   toTime(row.CreatedAt),
		UpdatedAt:   toTime(row.UpdatedAt),
	}, nil
}

func (r *LedgerRepository) DeleteTransaction(ctx context.Context, transactionID int32) error {
	return r.q.DeleteLedgerTransaction(ctx, transactionID)
}

func (r *LedgerRepository) ListTransactionsByMonth(ctx context.Context, month time.Time) ([]*ledger.Transaction, error) {
	pgDate := pgtype.Date{Time: month, Valid: true}
	rows, err := r.q.ListLedgerTransactionsByMonth(ctx, pgDate)
	if err != nil {
		return nil, err
	}

	transactions := make([]*ledger.Transaction, len(rows))
	for i, row := range rows {
		transactions[i] = &ledger.Transaction{
			ID:          row.TransactionID,
			OccurredAt:  row.OccurredAt.Time,
			Description: row.Description,
			Reference:   textToString(row.Reference),
			Notes:       textToString(row.Notes),
			CreatedAt:   toTime(row.CreatedAt),
			UpdatedAt:   toTime(row.UpdatedAt),
		}
	}

	return transactions, nil
}

func (r *LedgerRepository) CreatePosting(ctx context.Context, posting *ledger.Posting) (*ledger.Posting, error) {
	amount, err := numericFromFloat(posting.Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	params := ledgerSqlc.CreateLedgerPostingParams{
		TransactionID: posting.TransactionID,
		AccountID:     posting.AccountID,
		CategoryID:    int4FromPtr(posting.CategoryID),
		PartyID:       int4FromPtr(posting.PartyID),
		Side:          posting.Side,
		Amount:        amount,
		Memo:          textFromString(posting.Memo),
	}

	row, err := r.q.CreateLedgerPosting(ctx, params)
	if err != nil {
		return nil, err
	}

	amountValue, _ := row.Amount.Float64Value()

	return &ledger.Posting{
		ID:            row.PostingID,
		TransactionID: row.TransactionID,
		AccountID:     row.AccountID,
		CategoryID:    int4ToPtr(row.CategoryID),
		PartyID:       int4ToPtr(row.PartyID),
		Side:          row.Side,
		Amount:        amountValue.Float64,
		Memo:          textToString(row.Memo),
		CreatedAt:     toTime(row.CreatedAt),
	}, nil
}

func (r *LedgerRepository) UpdatePosting(ctx context.Context, posting *ledger.Posting) (*ledger.Posting, error) {
	amount, err := numericFromFloat(posting.Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	params := ledgerSqlc.UpdateLedgerPostingParams{
		PostingID:   posting.ID,
		AccountID:   posting.AccountID,
		CategoryID:  int4FromPtr(posting.CategoryID),
		PartyID:     int4FromPtr(posting.PartyID),
		Side:        posting.Side,
		Amount:      amount,
		Memo:        textFromString(posting.Memo),
	}

	row, err := r.q.UpdateLedgerPosting(ctx, params)
	if err != nil {
		return nil, err
	}

	amountValue, _ := row.Amount.Float64Value()
	return &ledger.Posting{
		ID:            row.PostingID,
		TransactionID: row.TransactionID,
		AccountID:     row.AccountID,
		CategoryID:    int4ToPtr(row.CategoryID),
		PartyID:       int4ToPtr(row.PartyID),
		Side:          row.Side,
		Amount:        amountValue.Float64,
		Memo:          textToString(row.Memo),
		CreatedAt:     toTime(row.CreatedAt),
	}, nil
}

func (r *LedgerRepository) ListPostingsByTransaction(ctx context.Context, transactionID int32) ([]*ledger.Posting, error) {
	rows, err := r.q.ListLedgerPostingsByTransaction(ctx, transactionID)
	if err != nil {
		return nil, err
	}

	postings := make([]*ledger.Posting, len(rows))
	for i, row := range rows {
		amountValue, _ := row.Amount.Float64Value()
		postings[i] = &ledger.Posting{
			ID:            row.PostingID,
			TransactionID: row.TransactionID,
			AccountID:     row.AccountID,
			CategoryID:    int4ToPtr(row.CategoryID),
			PartyID:       int4ToPtr(row.PartyID),
			Side:          row.Side,
			Amount:        amountValue.Float64,
			Memo:          textToString(row.Memo),
			CreatedAt:     toTime(row.CreatedAt),
		}
	}

	return postings, nil
}

func textToString(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func numericFromFloat(value float64) (pgtype.Numeric, error) {
	var numeric pgtype.Numeric
	if err := numeric.Scan(fmt.Sprintf("%.2f", value)); err != nil {
		return pgtype.Numeric{}, err
	}
	return numeric, nil
}

func toTime(value pgtype.Timestamptz) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time
}
