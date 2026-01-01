package creditcard

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres"
	journalrepo "github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/journal"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/creditcard"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/journal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool        *pgxpool.Pool
	journalRepo *journalrepo.Repository
}

func NewRepository(store *postgres.Store) *Repository {
	return &Repository{
		pool:        store.Pool(),
		journalRepo: journalrepo.NewRepository(store),
	}
}

func (r *Repository) ListCardNetworks(ctx context.Context) ([]creditcard.CardNetwork, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT code, display_name, created_at, updated_at
		FROM card_networks
		ORDER BY code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []creditcard.CardNetwork
	for rows.Next() {
		var item creditcard.CardNetwork
		if err := rows.Scan(&item.Code, &item.DisplayName, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) CreateCardNetwork(ctx context.Context, code, displayName string) (creditcard.CardNetwork, error) {
	var item creditcard.CardNetwork
	row := r.pool.QueryRow(ctx, `
		INSERT INTO card_networks (code, display_name)
		VALUES ($1, $2)
		RETURNING code, display_name, created_at, updated_at
	`, code, displayName)
	if err := row.Scan(&item.Code, &item.DisplayName, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return creditcard.CardNetwork{}, err
	}
	return item, nil
}

func (r *Repository) UpdateCardNetwork(ctx context.Context, code, displayName string, updatedAt time.Time) (creditcard.CardNetwork, error) {
	var item creditcard.CardNetwork
	row := r.pool.QueryRow(ctx, `
		UPDATE card_networks
		SET display_name = $2, updated_at = $3
		WHERE code = $1
		RETURNING code, display_name, created_at, updated_at
	`, code, displayName, updatedAt)
	if err := row.Scan(&item.Code, &item.DisplayName, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return creditcard.CardNetwork{}, creditcard.ErrNotFound
		}
		return creditcard.CardNetwork{}, err
	}
	return item, nil
}

func (r *Repository) DeleteCardNetwork(ctx context.Context, code string) error {
	cmd, err := r.pool.Exec(ctx, `DELETE FROM card_networks WHERE code = $1`, code)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return creditcard.ErrNotFound
	}
	return nil
}

func (r *Repository) GetCardNetwork(ctx context.Context, code string) (creditcard.CardNetwork, error) {
	var item creditcard.CardNetwork
	row := r.pool.QueryRow(ctx, `
		SELECT code, display_name, created_at, updated_at
		FROM card_networks
		WHERE code = $1
	`, code)
	if err := row.Scan(&item.Code, &item.DisplayName, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return creditcard.CardNetwork{}, creditcard.ErrNotFound
		}
		return creditcard.CardNetwork{}, err
	}
	return item, nil
}

func (r *Repository) ListCreditCards(ctx context.Context, ledgerID string) ([]creditcard.CreditCard, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.account_id, a.ledger_id, c.issuer_name, c.network, c.nickname, c.last4, c.credit_limit_cents, c.closing_day, c.due_day, c.created_at, c.updated_at
		FROM credit_cards c
		JOIN accounts a ON a.id = c.account_id
		WHERE a.ledger_id = $1
		ORDER BY c.created_at
	`, ledgerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []creditcard.CreditCard
	for rows.Next() {
		var card creditcard.CreditCard
		if err := rows.Scan(
			&card.AccountID,
			&card.LedgerID,
			&card.IssuerName,
			&card.Network,
			&card.Nickname,
			&card.Last4,
			&card.CreditLimitCents,
			&card.ClosingDay,
			&card.DueDay,
			&card.CreatedAt,
			&card.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, card)
	}
	return items, nil
}

func (r *Repository) GetCreditCard(ctx context.Context, ledgerID, cardAccountID string) (creditcard.CreditCard, error) {
	var card creditcard.CreditCard
	row := r.pool.QueryRow(ctx, `
		SELECT c.account_id, a.ledger_id, c.issuer_name, c.network, c.nickname, c.last4, c.credit_limit_cents, c.closing_day, c.due_day, c.created_at, c.updated_at
		FROM credit_cards c
		JOIN accounts a ON a.id = c.account_id
		WHERE a.ledger_id = $1 AND c.account_id = $2
	`, ledgerID, cardAccountID)
	if err := row.Scan(
		&card.AccountID,
		&card.LedgerID,
		&card.IssuerName,
		&card.Network,
		&card.Nickname,
		&card.Last4,
		&card.CreditLimitCents,
		&card.ClosingDay,
		&card.DueDay,
		&card.CreatedAt,
		&card.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return creditcard.CreditCard{}, creditcard.ErrNotFound
		}
		return creditcard.CreditCard{}, err
	}
	return card, nil
}

func (r *Repository) CreateTransaction(ctx context.Context, params journal.CreateTransactionParams) (journal.Transaction, error) {
	return r.journalRepo.CreateTransaction(ctx, params)
}

func (r *Repository) CreateCreditCard(ctx context.Context, params creditcard.CreditCard) (creditcard.CreditCard, error) {
	var card creditcard.CreditCard
	row := r.pool.QueryRow(ctx, `
		INSERT INTO credit_cards (account_id, issuer_name, network, nickname, last4, credit_limit_cents, closing_day, due_day)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING account_id, issuer_name, network, nickname, last4, credit_limit_cents, closing_day, due_day, created_at, updated_at
	`, params.AccountID, params.IssuerName, params.Network, params.Nickname, params.Last4, params.CreditLimitCents, params.ClosingDay, params.DueDay)
	if err := row.Scan(
		&card.AccountID,
		&card.IssuerName,
		&card.Network,
		&card.Nickname,
		&card.Last4,
		&card.CreditLimitCents,
		&card.ClosingDay,
		&card.DueDay,
		&card.CreatedAt,
		&card.UpdatedAt,
	); err != nil {
		return creditcard.CreditCard{}, err
	}
	card.LedgerID = params.LedgerID
	return card, nil
}

func (r *Repository) UpdateCreditCard(ctx context.Context, params creditcard.CreditCard, updatedAt time.Time) (creditcard.CreditCard, error) {
	var card creditcard.CreditCard
	row := r.pool.QueryRow(ctx, `
		UPDATE credit_cards
		SET issuer_name = $2, network = $3, nickname = $4, last4 = $5, credit_limit_cents = $6, closing_day = $7, due_day = $8, updated_at = $9
		WHERE account_id = $1
		RETURNING account_id, issuer_name, network, nickname, last4, credit_limit_cents, closing_day, due_day, created_at, updated_at
	`, params.AccountID, params.IssuerName, params.Network, params.Nickname, params.Last4, params.CreditLimitCents, params.ClosingDay, params.DueDay, updatedAt)
	if err := row.Scan(
		&card.AccountID,
		&card.IssuerName,
		&card.Network,
		&card.Nickname,
		&card.Last4,
		&card.CreditLimitCents,
		&card.ClosingDay,
		&card.DueDay,
		&card.CreatedAt,
		&card.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return creditcard.CreditCard{}, creditcard.ErrNotFound
		}
		return creditcard.CreditCard{}, err
	}
	card.LedgerID = params.LedgerID
	return card, nil
}

func (r *Repository) DeleteCreditCard(ctx context.Context, ledgerID, cardAccountID string) error {
	cmd, err := r.pool.Exec(ctx, `DELETE FROM credit_cards WHERE account_id = $1`, cardAccountID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return creditcard.ErrNotFound
	}
	return nil
}

func (r *Repository) GetAccountType(ctx context.Context, ledgerID, accountID string) (string, error) {
	var accountType string
	row := r.pool.QueryRow(ctx, `
		SELECT type
		FROM accounts
		WHERE ledger_id = $1 AND id = $2
	`, ledgerID, accountID)
	if err := row.Scan(&accountType); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", creditcard.ErrNotFound
		}
		return "", err
	}
	return accountType, nil
}

func (r *Repository) GetAccountLedger(ctx context.Context, accountID string) (string, error) {
	var ledgerID string
	row := r.pool.QueryRow(ctx, `SELECT ledger_id FROM accounts WHERE id = $1`, accountID)
	if err := row.Scan(&ledgerID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", creditcard.ErrNotFound
		}
		return "", err
	}
	return ledgerID, nil
}

func (r *Repository) GetCategoryBudgetInfo(ctx context.Context, ledgerID, categoryID string) (string, bool, error) {
	var direction string
	var relevant bool
	row := r.pool.QueryRow(ctx, `
		SELECT direction, is_budget_relevant
		FROM categories
		WHERE ledger_id = $1 AND id = $2
	`, ledgerID, categoryID)
	if err := row.Scan(&direction, &relevant); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", false, creditcard.ErrNotFound
		}
		return "", false, err
	}
	return direction, relevant, nil
}

func (r *Repository) FindCategoryByName(ctx context.Context, ledgerID, name string) (string, error) {
	var id string
	row := r.pool.QueryRow(ctx, `
		SELECT id
		FROM categories
		WHERE ledger_id = $1 AND name = $2
		LIMIT 1
	`, ledgerID, name)
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", creditcard.ErrNotFound
		}
		return "", err
	}
	return id, nil
}

func (r *Repository) CreatePlanWithInstallments(ctx context.Context, plan creditcard.InstallmentPlan, installments []creditcard.Installment) (creditcard.InstallmentPlan, error) {
	var created creditcard.InstallmentPlan
	err := postgres.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, `
			INSERT INTO installment_plans (ledger_id, card_account_id, purchase_occurred_at, merchant, description, category_id, total_amount_cents, installments_count, installment_amount_cents, first_due_month, status, created_by_user_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			RETURNING id, ledger_id, card_account_id, purchase_occurred_at, merchant, description, category_id, total_amount_cents, installments_count, installment_amount_cents, first_due_month, status, created_by_user_id, created_at, updated_at
		`, plan.LedgerID, plan.CardAccountID, plan.PurchaseOccurredAt, plan.Merchant, plan.Description, plan.CategoryID, plan.TotalAmountCents, plan.InstallmentsCount, plan.InstallmentAmountCents, plan.FirstDueMonth, plan.Status, plan.CreatedByUserID)
		if err := row.Scan(
			&created.ID,
			&created.LedgerID,
			&created.CardAccountID,
			&created.PurchaseOccurredAt,
			&created.Merchant,
			&created.Description,
			&created.CategoryID,
			&created.TotalAmountCents,
			&created.InstallmentsCount,
			&created.InstallmentAmountCents,
			&created.FirstDueMonth,
			&created.Status,
			&created.CreatedByUserID,
			&created.CreatedAt,
			&created.UpdatedAt,
		); err != nil {
			return err
		}

		for _, inst := range installments {
			_, err := tx.Exec(ctx, `
				INSERT INTO installments (ledger_id, plan_id, installment_no, due_month, amount_cents, status)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, plan.LedgerID, created.ID, inst.InstallmentNo, inst.DueMonth, inst.AmountCents, inst.Status)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return creditcard.InstallmentPlan{}, err
	}
	return created, nil
}

func (r *Repository) ListPlans(ctx context.Context, ledgerID, cardAccountID string, status *string) ([]creditcard.InstallmentPlan, error) {
	query := `
		SELECT id, ledger_id, card_account_id, purchase_occurred_at, merchant, description, category_id, total_amount_cents, installments_count, installment_amount_cents, first_due_month, status, created_by_user_id, created_at, updated_at
		FROM installment_plans
		WHERE ledger_id = $1 AND card_account_id = $2
	`
	args := []interface{}{ledgerID, cardAccountID}
	idx := 3
	if status != nil {
		query += " AND status = $" + strconv.Itoa(idx)
		args = append(args, *status)
		idx++
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []creditcard.InstallmentPlan
	for rows.Next() {
		var plan creditcard.InstallmentPlan
		if err := rows.Scan(
			&plan.ID,
			&plan.LedgerID,
			&plan.CardAccountID,
			&plan.PurchaseOccurredAt,
			&plan.Merchant,
			&plan.Description,
			&plan.CategoryID,
			&plan.TotalAmountCents,
			&plan.InstallmentsCount,
			&plan.InstallmentAmountCents,
			&plan.FirstDueMonth,
			&plan.Status,
			&plan.CreatedByUserID,
			&plan.CreatedAt,
			&plan.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, plan)
	}
	return items, nil
}

func (r *Repository) GetPlan(ctx context.Context, ledgerID, cardAccountID, planID string) (creditcard.InstallmentPlan, error) {
	var plan creditcard.InstallmentPlan
	row := r.pool.QueryRow(ctx, `
		SELECT id, ledger_id, card_account_id, purchase_occurred_at, merchant, description, category_id, total_amount_cents, installments_count, installment_amount_cents, first_due_month, status, created_by_user_id, created_at, updated_at
		FROM installment_plans
		WHERE ledger_id = $1 AND card_account_id = $2 AND id = $3
	`, ledgerID, cardAccountID, planID)
	if err := row.Scan(
		&plan.ID,
		&plan.LedgerID,
		&plan.CardAccountID,
		&plan.PurchaseOccurredAt,
		&plan.Merchant,
		&plan.Description,
		&plan.CategoryID,
		&plan.TotalAmountCents,
		&plan.InstallmentsCount,
		&plan.InstallmentAmountCents,
		&plan.FirstDueMonth,
		&plan.Status,
		&plan.CreatedByUserID,
		&plan.CreatedAt,
		&plan.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return creditcard.InstallmentPlan{}, creditcard.ErrNotFound
		}
		return creditcard.InstallmentPlan{}, err
	}
	return plan, nil
}

func (r *Repository) UpdatePlanStatus(ctx context.Context, ledgerID, planID, status string, updatedAt time.Time) error {
	cmd, err := r.pool.Exec(ctx, `
		UPDATE installment_plans
		SET status = $2, updated_at = $3
		WHERE ledger_id = $1 AND id = $4
	`, ledgerID, status, updatedAt, planID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return creditcard.ErrNotFound
	}
	return nil
}

func (r *Repository) ListInstallments(ctx context.Context, ledgerID, cardAccountID string, month *time.Time, status *string) ([]creditcard.Installment, error) {
	query := `
		SELECT i.id, i.ledger_id, i.plan_id, i.installment_no, i.due_month, i.amount_cents, i.status, i.posted_transaction_id, i.paid_statement_id, i.created_at, i.updated_at
		FROM installments i
		JOIN installment_plans p ON p.id = i.plan_id
		WHERE i.ledger_id = $1 AND p.card_account_id = $2
	`
	args := []interface{}{ledgerID, cardAccountID}
	idx := 3
	if month != nil {
		query += " AND i.due_month = $" + strconv.Itoa(idx)
		args = append(args, *month)
		idx++
	}
	if status != nil {
		query += " AND i.status = $" + strconv.Itoa(idx)
		args = append(args, *status)
		idx++
	}
	query += " ORDER BY i.due_month, i.installment_no"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []creditcard.Installment
	for rows.Next() {
		var item creditcard.Installment
		if err := rows.Scan(
			&item.ID,
			&item.LedgerID,
			&item.PlanID,
			&item.InstallmentNo,
			&item.DueMonth,
			&item.AmountCents,
			&item.Status,
			&item.PostedTransactionID,
			&item.PaidStatementID,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) UpdateInstallmentStatus(ctx context.Context, ledgerID, installmentID, status string, updatedAt time.Time) (creditcard.Installment, error) {
	var item creditcard.Installment
	row := r.pool.QueryRow(ctx, `
		UPDATE installments
		SET status = $2, updated_at = $3
		WHERE ledger_id = $1 AND id = $4
		RETURNING id, ledger_id, plan_id, installment_no, due_month, amount_cents, status, posted_transaction_id, paid_statement_id, created_at, updated_at
	`, ledgerID, status, updatedAt, installmentID)
	if err := row.Scan(
		&item.ID,
		&item.LedgerID,
		&item.PlanID,
		&item.InstallmentNo,
		&item.DueMonth,
		&item.AmountCents,
		&item.Status,
		&item.PostedTransactionID,
		&item.PaidStatementID,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return creditcard.Installment{}, creditcard.ErrNotFound
		}
		return creditcard.Installment{}, err
	}
	return item, nil
}

func (r *Repository) ListInstallmentsForPosting(ctx context.Context, ledgerID, cardAccountID string, month time.Time) ([]creditcard.Installment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT i.id, i.ledger_id, i.plan_id, i.installment_no, i.due_month, i.amount_cents, i.status, i.posted_transaction_id, i.paid_statement_id, i.created_at, i.updated_at
		FROM installments i
		JOIN installment_plans p ON p.id = i.plan_id
		WHERE i.ledger_id = $1
			AND p.card_account_id = $2
			AND i.due_month = $3
			AND i.status = 'scheduled'
	`, ledgerID, cardAccountID, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []creditcard.Installment
	for rows.Next() {
		var item creditcard.Installment
		if err := rows.Scan(
			&item.ID,
			&item.LedgerID,
			&item.PlanID,
			&item.InstallmentNo,
			&item.DueMonth,
			&item.AmountCents,
			&item.Status,
			&item.PostedTransactionID,
			&item.PaidStatementID,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) GetPlanCategory(ctx context.Context, ledgerID, planID string) (string, error) {
	var categoryID string
	row := r.pool.QueryRow(ctx, `
		SELECT category_id FROM installment_plans
		WHERE ledger_id = $1 AND id = $2
	`, ledgerID, planID)
	if err := row.Scan(&categoryID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", creditcard.ErrNotFound
		}
		return "", err
	}
	return categoryID, nil
}

func (r *Repository) MarkInstallmentPosted(ctx context.Context, ledgerID, installmentID, transactionID string, updatedAt time.Time) error {
	cmd, err := r.pool.Exec(ctx, `
		UPDATE installments
		SET status = 'posted', posted_transaction_id = $2, updated_at = $3
		WHERE ledger_id = $1 AND id = $4 AND status = 'scheduled'
	`, ledgerID, transactionID, updatedAt, installmentID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return creditcard.ErrInstallmentAlreadyPosted
	}
	return nil
}

func (r *Repository) GetStatement(ctx context.Context, ledgerID, cardAccountID, statementID string) (creditcard.Statement, error) {
	var statement creditcard.Statement
	row := r.pool.QueryRow(ctx, `
		SELECT id, ledger_id, card_account_id, statement_month, closing_date, due_date, total_charges_cents, total_payments_cents, status, payment_transaction_id, created_at, updated_at
		FROM credit_card_statements
		WHERE ledger_id = $1 AND card_account_id = $2 AND id = $3
	`, ledgerID, cardAccountID, statementID)
	if err := row.Scan(
		&statement.ID,
		&statement.LedgerID,
		&statement.CardAccountID,
		&statement.StatementMonth,
		&statement.ClosingDate,
		&statement.DueDate,
		&statement.TotalChargesCents,
		&statement.TotalPaymentsCents,
		&statement.Status,
		&statement.PaymentTransactionID,
		&statement.CreatedAt,
		&statement.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return creditcard.Statement{}, creditcard.ErrNotFound
		}
		return creditcard.Statement{}, err
	}
	return statement, nil
}

func (r *Repository) GetStatementByMonth(ctx context.Context, ledgerID, cardAccountID string, month time.Time) (creditcard.Statement, error) {
	var statement creditcard.Statement
	row := r.pool.QueryRow(ctx, `
		SELECT id, ledger_id, card_account_id, statement_month, closing_date, due_date, total_charges_cents, total_payments_cents, status, payment_transaction_id, created_at, updated_at
		FROM credit_card_statements
		WHERE ledger_id = $1 AND card_account_id = $2 AND statement_month = $3
	`, ledgerID, cardAccountID, month)
	if err := row.Scan(
		&statement.ID,
		&statement.LedgerID,
		&statement.CardAccountID,
		&statement.StatementMonth,
		&statement.ClosingDate,
		&statement.DueDate,
		&statement.TotalChargesCents,
		&statement.TotalPaymentsCents,
		&statement.Status,
		&statement.PaymentTransactionID,
		&statement.CreatedAt,
		&statement.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return creditcard.Statement{}, creditcard.ErrNotFound
		}
		return creditcard.Statement{}, err
	}
	return statement, nil
}

func (r *Repository) CreateStatement(ctx context.Context, statement creditcard.Statement) (creditcard.Statement, error) {
	var created creditcard.Statement
	row := r.pool.QueryRow(ctx, `
		INSERT INTO credit_card_statements (ledger_id, card_account_id, statement_month, closing_date, due_date, total_charges_cents, total_payments_cents, status, payment_transaction_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, ledger_id, card_account_id, statement_month, closing_date, due_date, total_charges_cents, total_payments_cents, status, payment_transaction_id, created_at, updated_at
	`, statement.LedgerID, statement.CardAccountID, statement.StatementMonth, statement.ClosingDate, statement.DueDate, statement.TotalChargesCents, statement.TotalPaymentsCents, statement.Status, statement.PaymentTransactionID)
	if err := row.Scan(
		&created.ID,
		&created.LedgerID,
		&created.CardAccountID,
		&created.StatementMonth,
		&created.ClosingDate,
		&created.DueDate,
		&created.TotalChargesCents,
		&created.TotalPaymentsCents,
		&created.Status,
		&created.PaymentTransactionID,
		&created.CreatedAt,
		&created.UpdatedAt,
	); err != nil {
		return creditcard.Statement{}, err
	}
	return created, nil
}

func (r *Repository) UpdateStatementTotals(ctx context.Context, statementID string, totalCharges, totalPayments int64, status string, updatedAt time.Time) (creditcard.Statement, error) {
	var updated creditcard.Statement
	row := r.pool.QueryRow(ctx, `
		UPDATE credit_card_statements
		SET total_charges_cents = $2, total_payments_cents = $3, status = $4, updated_at = $5
		WHERE id = $1
		RETURNING id, ledger_id, card_account_id, statement_month, closing_date, due_date, total_charges_cents, total_payments_cents, status, payment_transaction_id, created_at, updated_at
	`, statementID, totalCharges, totalPayments, status, updatedAt)
	if err := row.Scan(
		&updated.ID,
		&updated.LedgerID,
		&updated.CardAccountID,
		&updated.StatementMonth,
		&updated.ClosingDate,
		&updated.DueDate,
		&updated.TotalChargesCents,
		&updated.TotalPaymentsCents,
		&updated.Status,
		&updated.PaymentTransactionID,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return creditcard.Statement{}, creditcard.ErrNotFound
		}
		return creditcard.Statement{}, err
	}
	return updated, nil
}

func (r *Repository) ListStatements(ctx context.Context, ledgerID, cardAccountID string, month *time.Time) ([]creditcard.Statement, error) {
	query := `
		SELECT id, ledger_id, card_account_id, statement_month, closing_date, due_date, total_charges_cents, total_payments_cents, status, payment_transaction_id, created_at, updated_at
		FROM credit_card_statements
		WHERE ledger_id = $1 AND card_account_id = $2
	`
	args := []interface{}{ledgerID, cardAccountID}
	idx := 3
	if month != nil {
		query += " AND statement_month = $" + strconv.Itoa(idx)
		args = append(args, *month)
		idx++
	}
	query += " ORDER BY statement_month DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []creditcard.Statement
	for rows.Next() {
		var statement creditcard.Statement
		if err := rows.Scan(
			&statement.ID,
			&statement.LedgerID,
			&statement.CardAccountID,
			&statement.StatementMonth,
			&statement.ClosingDate,
			&statement.DueDate,
			&statement.TotalChargesCents,
			&statement.TotalPaymentsCents,
			&statement.Status,
			&statement.PaymentTransactionID,
			&statement.CreatedAt,
			&statement.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, statement)
	}
	return items, nil
}

func (r *Repository) SetStatementPayment(ctx context.Context, statementID, paymentTransactionID string, totalPayments int64, status string, updatedAt time.Time) (creditcard.Statement, error) {
	var updated creditcard.Statement
	row := r.pool.QueryRow(ctx, `
		UPDATE credit_card_statements
		SET payment_transaction_id = $2, total_payments_cents = $3, status = $4, updated_at = $5
		WHERE id = $1
		RETURNING id, ledger_id, card_account_id, statement_month, closing_date, due_date, total_charges_cents, total_payments_cents, status, payment_transaction_id, created_at, updated_at
	`, statementID, paymentTransactionID, totalPayments, status, updatedAt)
	if err := row.Scan(
		&updated.ID,
		&updated.LedgerID,
		&updated.CardAccountID,
		&updated.StatementMonth,
		&updated.ClosingDate,
		&updated.DueDate,
		&updated.TotalChargesCents,
		&updated.TotalPaymentsCents,
		&updated.Status,
		&updated.PaymentTransactionID,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return creditcard.Statement{}, creditcard.ErrNotFound
		}
		return creditcard.Statement{}, err
	}
	return updated, nil
}

func (r *Repository) MarkInstallmentsPaid(ctx context.Context, ledgerID, cardAccountID string, month time.Time, statementID string, updatedAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE installments i
		SET status = 'paid', paid_statement_id = $4, updated_at = $5
		FROM installment_plans p
		WHERE i.plan_id = p.id
			AND i.ledger_id = $1
			AND p.card_account_id = $2
			AND i.due_month = $3
			AND i.status = 'posted'
	`, ledgerID, cardAccountID, month, statementID, updatedAt)
	return err
}

func (r *Repository) SumStatementCharges(ctx context.Context, ledgerID, cardAccountID string, month time.Time) (int64, error) {
	var total int64
	row := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(i.amount_cents), 0)
		FROM installments i
		JOIN installment_plans p ON p.id = i.plan_id
		WHERE i.ledger_id = $1
			AND p.card_account_id = $2
			AND i.due_month = $3
			AND i.status IN ('posted', 'paid')
	`, ledgerID, cardAccountID, month)
	if err := row.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error) {
	return postgres.GetLedgerRole(ctx, r.pool, ledgerID, userID)
}

func (r *Repository) LedgerExists(ctx context.Context, ledgerID string) (bool, error) {
	return postgres.LedgerExists(ctx, r.pool, ledgerID)
}

func (r *Repository) HasPostedInstallments(ctx context.Context, ledgerID, planID string) (bool, error) {
	var exists bool
	row := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM installments
			WHERE ledger_id = $1 AND plan_id = $2 AND status IN ('posted', 'paid')
		)
	`, ledgerID, planID)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}
