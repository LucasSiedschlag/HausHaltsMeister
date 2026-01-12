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

func (r *Repository) ListCreditCards(ctx context.Context, ledgerID, parentAccountID string) ([]creditcard.CreditCard, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.ledger_id, c.parent_account_id, c.liability_account_id, c.label, c.brand, c.last4, c.cvv, c.holder_name,
			c.active, c.color, c.style, c.closing_day, c.due_day, c.created_at, c.updated_at
		FROM credit_cards c
		WHERE c.ledger_id = $1 AND c.parent_account_id = $2
		ORDER BY c.created_at
	`, ledgerID, parentAccountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []creditcard.CreditCard
	for rows.Next() {
		var card creditcard.CreditCard
		if err := rows.Scan(
			&card.ID,
			&card.LedgerID,
			&card.ParentAccountID,
			&card.LiabilityAccountID,
			&card.Label,
			&card.Brand,
			&card.Last4,
			&card.CVV,
			&card.HolderName,
			&card.Active,
			&card.Color,
			&card.Style,
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

func (r *Repository) GetCreditCard(ctx context.Context, cardID string) (creditcard.CreditCard, error) {
	var card creditcard.CreditCard
	row := r.pool.QueryRow(ctx, `
		SELECT c.id, c.ledger_id, c.parent_account_id, c.liability_account_id, c.label, c.brand, c.last4, c.cvv, c.holder_name,
			c.active, c.color, c.style, c.closing_day, c.due_day, c.created_at, c.updated_at
		FROM credit_cards c
		WHERE c.id = $1
	`, cardID)
	if err := row.Scan(
		&card.ID,
		&card.LedgerID,
		&card.ParentAccountID,
		&card.LiabilityAccountID,
		&card.Label,
		&card.Brand,
		&card.Last4,
		&card.CVV,
		&card.HolderName,
		&card.Active,
		&card.Color,
		&card.Style,
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
		INSERT INTO credit_cards (
			ledger_id,
			parent_account_id,
			liability_account_id,
			label,
			brand,
			last4,
			cvv,
			holder_name,
			active,
			color,
			style,
			closing_day,
			due_day
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, ledger_id, parent_account_id, liability_account_id, label, brand, last4, cvv, holder_name,
			active, color, style, closing_day, due_day, created_at, updated_at
	`, params.LedgerID, params.ParentAccountID, params.LiabilityAccountID, params.Label, params.Brand, params.Last4, params.CVV, params.HolderName, params.Active, params.Color, params.Style, params.ClosingDay, params.DueDay)
	if err := row.Scan(
		&card.ID,
		&card.LedgerID,
		&card.ParentAccountID,
		&card.LiabilityAccountID,
		&card.Label,
		&card.Brand,
		&card.Last4,
		&card.CVV,
		&card.HolderName,
		&card.Active,
		&card.Color,
		&card.Style,
		&card.ClosingDay,
		&card.DueDay,
		&card.CreatedAt,
		&card.UpdatedAt,
	); err != nil {
		return creditcard.CreditCard{}, err
	}
	return card, nil
}

func (r *Repository) UpdateCreditCard(ctx context.Context, params creditcard.CreditCard, updatedAt time.Time) (creditcard.CreditCard, error) {
	var card creditcard.CreditCard
	row := r.pool.QueryRow(ctx, `
		UPDATE credit_cards c
		SET label = $3,
			brand = $4,
			last4 = $5,
			cvv = $6,
			holder_name = $7,
			active = $8,
			color = $9,
			style = $10,
			closing_day = $11,
			due_day = $12,
			updated_at = $13
		WHERE c.ledger_id = $1 AND c.id = $2
		RETURNING c.id, c.ledger_id, c.parent_account_id, c.liability_account_id, c.label, c.brand, c.last4, c.cvv, c.holder_name,
			c.active, c.color, c.style, c.closing_day, c.due_day, c.created_at, c.updated_at
	`, params.LedgerID, params.ID, params.Label, params.Brand, params.Last4, params.CVV, params.HolderName, params.Active, params.Color, params.Style, params.ClosingDay, params.DueDay, updatedAt)
	if err := row.Scan(
		&card.ID,
		&card.LedgerID,
		&card.ParentAccountID,
		&card.LiabilityAccountID,
		&card.Label,
		&card.Brand,
		&card.Last4,
		&card.CVV,
		&card.HolderName,
		&card.Active,
		&card.Color,
		&card.Style,
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

func (r *Repository) DeleteCreditCard(ctx context.Context, ledgerID, cardID string) error {
	cmd, err := r.pool.Exec(ctx, `
		DELETE FROM credit_cards
		WHERE ledger_id = $1 AND id = $2
	`, ledgerID, cardID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return creditcard.ErrNotFound
	}
	return nil
}

func (r *Repository) GetAccount(ctx context.Context, userID, accountID string) (creditcard.AccountInfo, error) {
	var account creditcard.AccountInfo
	row := r.pool.QueryRow(ctx, `
		SELECT a.id, a.ledger_id, a.type, a.nature, a.is_active
		FROM accounts a
		JOIN ledger_members lm ON lm.ledger_id = a.ledger_id AND lm.user_id = $2 AND lm.removed_at IS NULL
		WHERE a.id = $1
	`, accountID, userID)
	if err := row.Scan(&account.ID, &account.LedgerID, &account.Type, &account.Nature, &account.IsActive); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return creditcard.AccountInfo{}, creditcard.ErrNotFound
		}
		return creditcard.AccountInfo{}, err
	}
	return account, nil
}

func (r *Repository) CreateAccount(ctx context.Context, ledgerID, name, accountType, nature string, isActive bool) (creditcard.AccountInfo, error) {
	var account creditcard.AccountInfo
	row := r.pool.QueryRow(ctx, `
		INSERT INTO accounts (ledger_id, name, type, nature, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, ledger_id, type, nature, is_active
	`, ledgerID, name, accountType, nature, isActive)
	if err := row.Scan(&account.ID, &account.LedgerID, &account.Type, &account.Nature, &account.IsActive); err != nil {
		return creditcard.AccountInfo{}, err
	}
	return account, nil
}

func (r *Repository) GetAccountNature(ctx context.Context, ledgerID, accountID string) (string, error) {
	var nature string
	row := r.pool.QueryRow(ctx, `
		SELECT nature
		FROM accounts
		WHERE ledger_id = $1 AND id = $2
	`, ledgerID, accountID)
	if err := row.Scan(&nature); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", creditcard.ErrNotFound
		}
		return "", err
	}
	return nature, nil
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
			INSERT INTO installment_plans (ledger_id, credit_card_id, purchase_occurred_at, merchant, description, category_id, total_amount_cents, installments_count, installment_amount_cents, first_due_month, status, created_by_user_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			RETURNING id, ledger_id, credit_card_id, purchase_occurred_at, merchant, description, category_id, total_amount_cents, installments_count, installment_amount_cents, first_due_month, status, created_by_user_id, created_at, updated_at
		`, plan.LedgerID, plan.CreditCardID, plan.PurchaseOccurredAt, plan.Merchant, plan.Description, plan.CategoryID, plan.TotalAmountCents, plan.InstallmentsCount, plan.InstallmentAmountCents, plan.FirstDueMonth, plan.Status, plan.CreatedByUserID)
		if err := row.Scan(
			&created.ID,
			&created.LedgerID,
			&created.CreditCardID,
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

func (r *Repository) ListPlans(ctx context.Context, ledgerID, cardID string, status *string) ([]creditcard.InstallmentPlan, error) {
	query := `
		SELECT id, ledger_id, credit_card_id, purchase_occurred_at, merchant, description, category_id, total_amount_cents, installments_count, installment_amount_cents, first_due_month, status, created_by_user_id, created_at, updated_at
		FROM installment_plans
		WHERE ledger_id = $1 AND credit_card_id = $2
	`
	args := []interface{}{ledgerID, cardID}
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
			&plan.CreditCardID,
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

func (r *Repository) GetPlan(ctx context.Context, ledgerID, cardID, planID string) (creditcard.InstallmentPlan, error) {
	var plan creditcard.InstallmentPlan
	row := r.pool.QueryRow(ctx, `
		SELECT id, ledger_id, credit_card_id, purchase_occurred_at, merchant, description, category_id, total_amount_cents, installments_count, installment_amount_cents, first_due_month, status, created_by_user_id, created_at, updated_at
		FROM installment_plans
		WHERE ledger_id = $1 AND credit_card_id = $2 AND id = $3
	`, ledgerID, cardID, planID)
	if err := row.Scan(
		&plan.ID,
		&plan.LedgerID,
		&plan.CreditCardID,
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

func (r *Repository) ListInstallments(ctx context.Context, ledgerID, cardID string, month *time.Time, status *string) ([]creditcard.Installment, error) {
	query := `
		SELECT i.id, i.ledger_id, i.plan_id, i.installment_no, i.due_month, i.amount_cents, i.status, i.posted_transaction_id, i.paid_statement_id, i.created_at, i.updated_at
		FROM installments i
		JOIN installment_plans p ON p.id = i.plan_id
		WHERE i.ledger_id = $1 AND p.credit_card_id = $2
	`
	args := []interface{}{ledgerID, cardID}
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

func (r *Repository) ListInstallmentsForPosting(ctx context.Context, ledgerID, cardID string, month time.Time) ([]creditcard.Installment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT i.id, i.ledger_id, i.plan_id, i.installment_no, i.due_month, i.amount_cents, i.status, i.posted_transaction_id, i.paid_statement_id, i.created_at, i.updated_at
		FROM installments i
		JOIN installment_plans p ON p.id = i.plan_id
		WHERE i.ledger_id = $1
			AND p.credit_card_id = $2
			AND i.due_month = $3
			AND i.status = 'scheduled'
	`, ledgerID, cardID, month)
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

func (r *Repository) GetStatement(ctx context.Context, ledgerID, cardID, statementID string) (creditcard.Statement, error) {
	var statement creditcard.Statement
	row := r.pool.QueryRow(ctx, `
		SELECT id, ledger_id, credit_card_id, statement_month, closing_date, due_date, total_charges_cents, total_payments_cents, status, payment_transaction_id, created_at, updated_at
		FROM credit_card_statements
		WHERE ledger_id = $1 AND credit_card_id = $2 AND id = $3
	`, ledgerID, cardID, statementID)
	if err := row.Scan(
		&statement.ID,
		&statement.LedgerID,
		&statement.CreditCardID,
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

func (r *Repository) GetStatementByMonth(ctx context.Context, ledgerID, cardID string, month time.Time) (creditcard.Statement, error) {
	var statement creditcard.Statement
	row := r.pool.QueryRow(ctx, `
		SELECT id, ledger_id, credit_card_id, statement_month, closing_date, due_date, total_charges_cents, total_payments_cents, status, payment_transaction_id, created_at, updated_at
		FROM credit_card_statements
		WHERE ledger_id = $1 AND credit_card_id = $2 AND statement_month = $3
	`, ledgerID, cardID, month)
	if err := row.Scan(
		&statement.ID,
		&statement.LedgerID,
		&statement.CreditCardID,
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
		INSERT INTO credit_card_statements (ledger_id, credit_card_id, statement_month, closing_date, due_date, total_charges_cents, total_payments_cents, status, payment_transaction_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, ledger_id, credit_card_id, statement_month, closing_date, due_date, total_charges_cents, total_payments_cents, status, payment_transaction_id, created_at, updated_at
	`, statement.LedgerID, statement.CreditCardID, statement.StatementMonth, statement.ClosingDate, statement.DueDate, statement.TotalChargesCents, statement.TotalPaymentsCents, statement.Status, statement.PaymentTransactionID)
	if err := row.Scan(
		&created.ID,
		&created.LedgerID,
		&created.CreditCardID,
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

func (r *Repository) UpdateStatementTotals(ctx context.Context, ledgerID, cardID, statementID string, totalCharges, totalPayments int64, status string, updatedAt time.Time) (creditcard.Statement, error) {
	var updated creditcard.Statement
	row := r.pool.QueryRow(ctx, `
		UPDATE credit_card_statements
		SET total_charges_cents = $4, total_payments_cents = $5, status = $6, updated_at = $7
		WHERE ledger_id = $1 AND credit_card_id = $2 AND id = $3
		RETURNING id, ledger_id, credit_card_id, statement_month, closing_date, due_date, total_charges_cents, total_payments_cents, status, payment_transaction_id, created_at, updated_at
	`, ledgerID, cardID, statementID, totalCharges, totalPayments, status, updatedAt)
	if err := row.Scan(
		&updated.ID,
		&updated.LedgerID,
		&updated.CreditCardID,
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

func (r *Repository) ListStatements(ctx context.Context, ledgerID, cardID string, month *time.Time) ([]creditcard.Statement, error) {
	query := `
		SELECT id, ledger_id, credit_card_id, statement_month, closing_date, due_date, total_charges_cents, total_payments_cents, status, payment_transaction_id, created_at, updated_at
		FROM credit_card_statements
		WHERE ledger_id = $1 AND credit_card_id = $2
	`
	args := []interface{}{ledgerID, cardID}
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
			&statement.CreditCardID,
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

func (r *Repository) SetStatementPayment(ctx context.Context, ledgerID, cardID, statementID, paymentTransactionID string, totalPayments int64, status string, updatedAt time.Time) (creditcard.Statement, error) {
	var updated creditcard.Statement
	row := r.pool.QueryRow(ctx, `
		UPDATE credit_card_statements
		SET payment_transaction_id = $4, total_payments_cents = $5, status = $6, updated_at = $7
		WHERE ledger_id = $1 AND credit_card_id = $2 AND id = $3
		RETURNING id, ledger_id, credit_card_id, statement_month, closing_date, due_date, total_charges_cents, total_payments_cents, status, payment_transaction_id, created_at, updated_at
	`, ledgerID, cardID, statementID, paymentTransactionID, totalPayments, status, updatedAt)
	if err := row.Scan(
		&updated.ID,
		&updated.LedgerID,
		&updated.CreditCardID,
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

func (r *Repository) MarkInstallmentsPaid(ctx context.Context, ledgerID, cardID string, month time.Time, statementID string, updatedAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE installments i
		SET status = 'paid', paid_statement_id = $4, updated_at = $5
		FROM installment_plans p
		WHERE i.plan_id = p.id
			AND i.ledger_id = $1
			AND p.credit_card_id = $2
			AND i.due_month = $3
			AND i.status = 'posted'
	`, ledgerID, cardID, month, statementID, updatedAt)
	return err
}

func (r *Repository) SumStatementCharges(ctx context.Context, ledgerID, cardID string, month time.Time) (int64, error) {
	var total int64
	row := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(i.amount_cents), 0)
		FROM installments i
		JOIN installment_plans p ON p.id = i.plan_id
		WHERE i.ledger_id = $1
			AND p.credit_card_id = $2
			AND i.due_month = $3
			AND i.status IN ('posted', 'paid')
	`, ledgerID, cardID, month)
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
