package journal

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/journal"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

type dbQuerier interface {
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
}

func NewRepository(store *postgres.Store) *Repository {
	return &Repository{pool: store.Pool()}
}

func (r *Repository) CreateTransaction(ctx context.Context, params journal.CreateTransactionParams) (journal.Transaction, error) {
	var created journal.Transaction
	entries := make([]journal.Entry, 0, len(params.Entries))

	if params.Idempotency != nil && params.Idempotency.Key != "" {
		record, err := r.getIdempotencyKey(ctx, r.pool, params.LedgerID, params.Idempotency.Key, journal.IdempotencyOperationTransactionCreate)
		if err == nil {
			return r.handleIdempotencyHit(ctx, r.pool, record, params.Idempotency.RequestHash)
		}
		if err != nil && !errors.Is(err, journal.ErrNotFound) {
			return journal.Transaction{}, err
		}
	}

	err := postgres.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		transactionID := ""
		if params.Idempotency != nil && params.Idempotency.Key != "" {
			transactionID = uuid.NewString()
			_, err := tx.Exec(ctx, `
				INSERT INTO idempotency_keys (ledger_id, key, resource_type, resource_id, request_hash, expires_at)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, params.LedgerID, params.Idempotency.Key, journal.IdempotencyOperationTransactionCreate, transactionID, params.Idempotency.RequestHash, params.Idempotency.ExpiresAt)
			if err != nil {
				if postgres.IsUniqueViolation(err) {
					record, err := r.getIdempotencyKey(ctx, tx, params.LedgerID, params.Idempotency.Key, journal.IdempotencyOperationTransactionCreate)
					if err != nil {
						return err
					}
					item, err := r.handleIdempotencyHit(ctx, tx, record, params.Idempotency.RequestHash)
					if err != nil {
						return err
					}
					created = item
					return nil
				}
				return err
			}
		}

		if created.ID != "" {
			return nil
		}

		var row pgx.Row
		if transactionID != "" {
			row = tx.QueryRow(ctx, `
				INSERT INTO transactions (id, ledger_id, occurred_at, description, notes, created_by_user_id, credit_card_id, investment_action)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				RETURNING id, ledger_id, occurred_at, description, notes, created_by_user_id, credit_card_id, investment_action, created_at, updated_at
			`, transactionID, params.LedgerID, params.OccurredAt, params.Description, params.Notes, params.CreatedByUserID, params.CreditCardID, params.InvestmentAction)
		} else {
			row = tx.QueryRow(ctx, `
				INSERT INTO transactions (ledger_id, occurred_at, description, notes, created_by_user_id, credit_card_id, investment_action)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				RETURNING id, ledger_id, occurred_at, description, notes, created_by_user_id, credit_card_id, investment_action, created_at, updated_at
			`, params.LedgerID, params.OccurredAt, params.Description, params.Notes, params.CreatedByUserID, params.CreditCardID, params.InvestmentAction)
		}
		if err := scanTransaction(row, &created); err != nil {
			return err
		}

		for _, entry := range params.Entries {
			var inserted journal.Entry
			row = tx.QueryRow(ctx, `
				INSERT INTO entries (ledger_id, transaction_id, account_id, category_id, kind, amount_cents, memo)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				RETURNING id, ledger_id, transaction_id, account_id, category_id, kind, amount_cents, memo, created_at, updated_at
			`, params.LedgerID, created.ID, entry.AccountID, entry.CategoryID, entry.Kind, entry.AmountCents, entry.Memo)
			if err := scanEntry(row, &inserted); err != nil {
				return err
			}
			entries = append(entries, inserted)
		}
		return nil
	})
	if err != nil {
		return journal.Transaction{}, err
	}
	if len(created.Entries) == 0 {
		created.Entries = entries
	}
	return created, nil
}

func (r *Repository) getIdempotencyKey(ctx context.Context, q dbQuerier, ledgerID, key, operation string) (journal.IdempotencyRecord, error) {
	row := q.QueryRow(ctx, `
		SELECT ledger_id, key, resource_type, request_hash, resource_id, expires_at
		FROM idempotency_keys
		WHERE ledger_id = $1 AND key = $2 AND resource_type = $3
	`, ledgerID, key, operation)

	var record journal.IdempotencyRecord
	if err := row.Scan(&record.LedgerID, &record.Key, &record.Operation, &record.RequestHash, &record.ResourceID, &record.ExpiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return journal.IdempotencyRecord{}, journal.ErrNotFound
		}
		return journal.IdempotencyRecord{}, err
	}
	return record, nil
}

func (r *Repository) handleIdempotencyHit(ctx context.Context, q dbQuerier, record journal.IdempotencyRecord, requestHash string) (journal.Transaction, error) {
	if time.Now().UTC().After(record.ExpiresAt) {
		return journal.Transaction{}, journal.ErrIdempotencyExpired
	}
	if record.RequestHash != requestHash {
		return journal.Transaction{}, journal.ErrIdempotencyConflict
	}
	item, err := r.getTransactionWithEntries(ctx, q, record.LedgerID, record.ResourceID)
	if errors.Is(err, journal.ErrNotFound) {
		return journal.Transaction{}, journal.ErrTransactionNotFound
	}
	return item, err
}

func (r *Repository) ListTransactions(ctx context.Context, params journal.ListTransactionsParams) (journal.ListResult, error) {
	ids, cursor, err := r.listTransactionIDs(ctx, params)
	if err != nil {
		return journal.ListResult{}, err
	}
	if len(ids) == 0 {
		return journal.ListResult{Items: []journal.Transaction{}}, nil
	}

	transactions, err := r.getTransactionsWithEntries(ctx, ids)
	if err != nil {
		return journal.ListResult{}, err
	}

	return journal.ListResult{Items: transactions, NextCursor: cursor}, nil
}

func (r *Repository) listTransactionIDs(ctx context.Context, params journal.ListTransactionsParams) ([]string, *journal.Cursor, error) {
	query := `
		SELECT t.id, t.occurred_at
		FROM transactions t
		WHERE t.ledger_id = $1
	`
	args := []interface{}{params.LedgerID}
	idx := 2

	if params.From != nil {
		query += " AND t.occurred_at >= $" + strconv.Itoa(idx)
		args = append(args, *params.From)
		idx++
	}
	if params.To != nil {
		query += " AND t.occurred_at <= $" + strconv.Itoa(idx)
		args = append(args, *params.To)
		idx++
	}
	if params.AccountID != nil {
		query += " AND EXISTS (SELECT 1 FROM entries e WHERE e.transaction_id = t.id AND e.account_id = $" + strconv.Itoa(idx) + ")"
		args = append(args, *params.AccountID)
		idx++
	}
	if params.CategoryID != nil {
		query += " AND EXISTS (SELECT 1 FROM entries e WHERE e.transaction_id = t.id AND e.category_id = $" + strconv.Itoa(idx) + ")"
		args = append(args, *params.CategoryID)
		idx++
	}
	if params.Query != nil {
		query += " AND (t.description ILIKE $" + strconv.Itoa(idx) + " OR t.notes ILIKE $" + strconv.Itoa(idx) + ")"
		args = append(args, "%"+*params.Query+"%")
		idx++
	}
	if params.Cursor != nil {
		query += " AND (t.occurred_at < $" + strconv.Itoa(idx) + " OR (t.occurred_at = $" + strconv.Itoa(idx+1) + " AND t.id < $" + strconv.Itoa(idx+1) + "))"
		args = append(args, params.Cursor.OccurredAt, params.Cursor.ID)
		idx += 2
	}

	query += " ORDER BY t.occurred_at DESC, t.id DESC LIMIT $" + strconv.Itoa(idx)
	args = append(args, params.Limit)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	ids := []string{}
	var lastCursor *journal.Cursor
	for rows.Next() {
		var id string
		var occurredAt time.Time
		if err := rows.Scan(&id, &occurredAt); err != nil {
			return nil, nil, err
		}
		ids = append(ids, id)
		lastCursor = &journal.Cursor{OccurredAt: occurredAt, ID: id}
	}

	return ids, lastCursor, nil
}

func (r *Repository) getTransactionsWithEntries(ctx context.Context, ids []string) ([]journal.Transaction, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT t.id, t.ledger_id, t.occurred_at, t.description, t.notes, t.created_by_user_id, t.credit_card_id, t.investment_action, t.created_at, t.updated_at,
			e.id, e.ledger_id, e.transaction_id, e.account_id, e.category_id, e.kind, e.amount_cents, e.memo, e.created_at, e.updated_at
		FROM transactions t
		JOIN entries e ON e.transaction_id = t.id
		WHERE t.id = ANY($1)
		ORDER BY t.occurred_at DESC, t.id DESC, e.created_at ASC
	`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []journal.Transaction{}
	index := map[string]int{}
	for rows.Next() {
		var tx journal.Transaction
		var entry journal.Entry
		if err := rows.Scan(
			&tx.ID,
			&tx.LedgerID,
			&tx.OccurredAt,
			&tx.Description,
			&tx.Notes,
			&tx.CreatedByUserID,
			&tx.CreditCardID,
			&tx.InvestmentAction,
			&tx.CreatedAt,
			&tx.UpdatedAt,
			&entry.ID,
			&entry.LedgerID,
			&entry.TransactionID,
			&entry.AccountID,
			&entry.CategoryID,
			&entry.Kind,
			&entry.AmountCents,
			&entry.Memo,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		); err != nil {
			return nil, err
		}

		idx, ok := index[tx.ID]
		if !ok {
			tx.Entries = []journal.Entry{entry}
			items = append(items, tx)
			index[tx.ID] = len(items) - 1
			continue
		}
		items[idx].Entries = append(items[idx].Entries, entry)
	}

	return items, nil
}

func (r *Repository) GetTransaction(ctx context.Context, ledgerID, transactionID string) (journal.Transaction, error) {
	return r.getTransactionWithEntries(ctx, r.pool, ledgerID, transactionID)
}

func (r *Repository) getTransactionWithEntries(ctx context.Context, q dbQuerier, ledgerID, transactionID string) (journal.Transaction, error) {
	rows, err := q.Query(ctx, `
		SELECT t.id, t.ledger_id, t.occurred_at, t.description, t.notes, t.created_by_user_id, t.credit_card_id, t.investment_action, t.created_at, t.updated_at,
			e.id, e.ledger_id, e.transaction_id, e.account_id, e.category_id, e.kind, e.amount_cents, e.memo, e.created_at, e.updated_at
		FROM transactions t
		JOIN entries e ON e.transaction_id = t.id
		WHERE t.ledger_id = $1 AND t.id = $2
		ORDER BY e.created_at ASC
	`, ledgerID, transactionID)
	if err != nil {
		return journal.Transaction{}, err
	}
	defer rows.Close()

	var result *journal.Transaction
	for rows.Next() {
		var tx journal.Transaction
		var entry journal.Entry
		if err := rows.Scan(
			&tx.ID,
			&tx.LedgerID,
			&tx.OccurredAt,
			&tx.Description,
			&tx.Notes,
			&tx.CreatedByUserID,
			&tx.CreditCardID,
			&tx.InvestmentAction,
			&tx.CreatedAt,
			&tx.UpdatedAt,
			&entry.ID,
			&entry.LedgerID,
			&entry.TransactionID,
			&entry.AccountID,
			&entry.CategoryID,
			&entry.Kind,
			&entry.AmountCents,
			&entry.Memo,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		); err != nil {
			return journal.Transaction{}, err
		}
		if result == nil {
			result = &tx
			result.Entries = []journal.Entry{entry}
			continue
		}
		result.Entries = append(result.Entries, entry)
	}

	if result == nil {
		return journal.Transaction{}, journal.ErrNotFound
	}
	return *result, nil
}

func (r *Repository) UpdateTransaction(ctx context.Context, params journal.UpdateTransactionParams) (journal.Transaction, error) {
	var updated journal.Transaction
	err := postgres.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		setParts := []string{}
		args := []interface{}{params.LedgerID, params.TransactionID}
		idx := 3

		if params.OccurredAt != nil {
			setParts = append(setParts, "occurred_at = $"+strconv.Itoa(idx))
			args = append(args, *params.OccurredAt)
			idx++
		}
		if params.Description != nil {
			setParts = append(setParts, "description = $"+strconv.Itoa(idx))
			args = append(args, *params.Description)
			idx++
		}
		if params.Notes != nil {
			setParts = append(setParts, "notes = $"+strconv.Itoa(idx))
			args = append(args, *params.Notes)
			idx++
		}

		setParts = append(setParts, "updated_at = $"+strconv.Itoa(idx))
		args = append(args, params.UpdatedAt)
		idx++

		query := `
			UPDATE transactions
			SET ` + strings.Join(setParts, ", ") + `
			WHERE ledger_id = $1 AND id = $2
			RETURNING id, ledger_id, occurred_at, description, notes, created_by_user_id, credit_card_id, investment_action, created_at, updated_at
		`

		row := tx.QueryRow(ctx, query, args...)
		if err := scanTransaction(row, &updated); err != nil {
			return err
		}

		if params.Entries != nil {
			if _, err := tx.Exec(ctx, `DELETE FROM entries WHERE ledger_id = $1 AND transaction_id = $2`, params.LedgerID, params.TransactionID); err != nil {
				return err
			}
			for _, entry := range *params.Entries {
				if _, err := tx.Exec(ctx, `
					INSERT INTO entries (transaction_id, ledger_id, account_id, category_id, kind, amount_cents, memo)
					VALUES ($1, $2, $3, $4, $5, $6, $7)
				`, params.TransactionID, params.LedgerID, entry.AccountID, entry.CategoryID, entry.Kind, entry.AmountCents, entry.Memo); err != nil {
					return err
				}
			}
		}

		entries, err := getEntriesByTransactionTx(ctx, tx, updated.ID)
		if err != nil {
			return err
		}
		updated.Entries = entries
		return nil
	})
	if err != nil {
		return journal.Transaction{}, err
	}
	return updated, nil
}

func (r *Repository) DeleteTransaction(ctx context.Context, ledgerID, transactionID string) error {
	return postgres.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `DELETE FROM entries WHERE ledger_id = $1 AND transaction_id = $2`, ledgerID, transactionID)
		if err != nil {
			return err
		}
		cmd, err := tx.Exec(ctx, `DELETE FROM transactions WHERE ledger_id = $1 AND id = $2`, ledgerID, transactionID)
		if err != nil {
			return err
		}
		if cmd.RowsAffected() == 0 {
			return journal.ErrNotFound
		}
		return nil
	})
}

func (r *Repository) HasTransactionReferences(ctx context.Context, ledgerID, transactionID string) (bool, error) {
	var exists bool
	row := r.pool.QueryRow(ctx, `
		SELECT (
			EXISTS (SELECT 1 FROM installments WHERE ledger_id = $1 AND posted_transaction_id = $2)
			OR EXISTS (SELECT 1 FROM credit_card_statements WHERE ledger_id = $1 AND payment_transaction_id = $2)
		)
	`, ledgerID, transactionID)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *Repository) GetAccountsByIDs(ctx context.Context, ledgerID string, ids []string) (map[string]bool, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id FROM accounts
		WHERE ledger_id = $1 AND id = ANY($2)
	`, ledgerID, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result[id] = true
	}
	return result, nil
}

func (r *Repository) GetCategoriesByIDs(ctx context.Context, ledgerID string, ids []string) (map[string]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, direction FROM categories
		WHERE ledger_id = $1 AND id = ANY($2)
	`, ledgerID, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[string]string{}
	for rows.Next() {
		var id string
		var direction string
		if err := rows.Scan(&id, &direction); err != nil {
			return nil, err
		}
		result[id] = direction
	}
	return result, nil
}

func (r *Repository) getEntriesByTransaction(ctx context.Context, transactionID string) ([]journal.Entry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, ledger_id, transaction_id, account_id, category_id, kind, amount_cents, memo, created_at, updated_at
		FROM entries
		WHERE transaction_id = $1
		ORDER BY created_at ASC
	`, transactionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []journal.Entry{}
	for rows.Next() {
		var entry journal.Entry
		if err := rows.Scan(
			&entry.ID,
			&entry.LedgerID,
			&entry.TransactionID,
			&entry.AccountID,
			&entry.CategoryID,
			&entry.Kind,
			&entry.AmountCents,
			&entry.Memo,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func getEntriesByTransactionTx(ctx context.Context, tx pgx.Tx, transactionID string) ([]journal.Entry, error) {
	rows, err := tx.Query(ctx, `
		SELECT id, ledger_id, transaction_id, account_id, category_id, kind, amount_cents, memo, created_at, updated_at
		FROM entries
		WHERE transaction_id = $1
		ORDER BY created_at ASC
	`, transactionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []journal.Entry{}
	for rows.Next() {
		var entry journal.Entry
		if err := rows.Scan(
			&entry.ID,
			&entry.LedgerID,
			&entry.TransactionID,
			&entry.AccountID,
			&entry.CategoryID,
			&entry.Kind,
			&entry.AmountCents,
			&entry.Memo,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func scanTransaction(row pgx.Row, tx *journal.Transaction) error {
	if err := row.Scan(
		&tx.ID,
		&tx.LedgerID,
		&tx.OccurredAt,
		&tx.Description,
		&tx.Notes,
		&tx.CreatedByUserID,
		&tx.CreditCardID,
		&tx.InvestmentAction,
		&tx.CreatedAt,
		&tx.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return journal.ErrNotFound
		}
		return err
	}
	return nil
}

func (r *Repository) GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error) {
	return postgres.GetLedgerRole(ctx, r.pool, ledgerID, userID)
}

func (r *Repository) LedgerExists(ctx context.Context, ledgerID string) (bool, error) {
	return postgres.LedgerExists(ctx, r.pool, ledgerID)
}

func scanEntry(row pgx.Row, entry *journal.Entry) error {
	if err := row.Scan(
		&entry.ID,
		&entry.LedgerID,
		&entry.TransactionID,
		&entry.AccountID,
		&entry.CategoryID,
		&entry.Kind,
		&entry.AmountCents,
		&entry.Memo,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	); err != nil {
		return err
	}
	return nil
}
