package creditcard

import (
	"context"
	"testing"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/journal"
	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	role              string
	accountType       string
	categoryDirection string
	categoryRelevant  bool
	hasPosted         bool
}

func (f *fakeRepo) ListCardNetworks(ctx context.Context) ([]CardNetwork, error) {
	return nil, nil
}

func (f *fakeRepo) CreateCardNetwork(ctx context.Context, code, displayName string) (CardNetwork, error) {
	return CardNetwork{}, nil
}

func (f *fakeRepo) UpdateCardNetwork(ctx context.Context, code, displayName string, updatedAt time.Time) (CardNetwork, error) {
	return CardNetwork{}, nil
}

func (f *fakeRepo) DeleteCardNetwork(ctx context.Context, code string) error {
	return nil
}

func (f *fakeRepo) ListCreditCards(ctx context.Context, ledgerID string) ([]CreditCard, error) {
	return nil, nil
}

func (f *fakeRepo) GetCreditCard(ctx context.Context, ledgerID, cardAccountID string) (CreditCard, error) {
	return CreditCard{}, nil
}

func (f *fakeRepo) CreateCreditCard(ctx context.Context, params CreditCard) (CreditCard, error) {
	return params, nil
}

func (f *fakeRepo) UpdateCreditCard(ctx context.Context, params CreditCard, updatedAt time.Time) (CreditCard, error) {
	return params, nil
}

func (f *fakeRepo) DeleteCreditCard(ctx context.Context, ledgerID, cardAccountID string) error {
	return nil
}

func (f *fakeRepo) GetAccountType(ctx context.Context, ledgerID, accountID string) (string, error) {
	return f.accountType, nil
}

func (f *fakeRepo) GetAccountLedger(ctx context.Context, accountID string) (string, error) {
	return "ledger-1", nil
}

func (f *fakeRepo) GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error) {
	return f.role, nil
}

func (f *fakeRepo) LedgerExists(ctx context.Context, ledgerID string) (bool, error) {
	return true, nil
}

func (f *fakeRepo) GetCategoryBudgetInfo(ctx context.Context, ledgerID, categoryID string) (string, bool, error) {
	return f.categoryDirection, f.categoryRelevant, nil
}

func (f *fakeRepo) FindCategoryByName(ctx context.Context, ledgerID, name string) (string, error) {
	return "cat-1", nil
}

func (f *fakeRepo) GetCardNetwork(ctx context.Context, code string) (CardNetwork, error) {
	return CardNetwork{Code: code}, nil
}

func (f *fakeRepo) CreatePlanWithInstallments(ctx context.Context, plan InstallmentPlan, installments []Installment) (InstallmentPlan, error) {
	if plan.ID == "" {
		plan.ID = "plan-1"
	}
	return plan, nil
}

func (f *fakeRepo) ListPlans(ctx context.Context, ledgerID, cardAccountID string, status *string) ([]InstallmentPlan, error) {
	return nil, nil
}

func (f *fakeRepo) GetPlan(ctx context.Context, ledgerID, cardAccountID, planID string) (InstallmentPlan, error) {
	return InstallmentPlan{ID: planID, LedgerID: ledgerID, CardAccountID: cardAccountID}, nil
}

func (f *fakeRepo) UpdatePlanStatus(ctx context.Context, ledgerID, planID, status string, updatedAt time.Time) error {
	return nil
}

func (f *fakeRepo) HasPostedInstallments(ctx context.Context, ledgerID, planID string) (bool, error) {
	return f.hasPosted, nil
}

func (f *fakeRepo) ListInstallments(ctx context.Context, ledgerID, cardAccountID string, month *time.Time, status *string) ([]Installment, error) {
	return nil, nil
}

func (f *fakeRepo) UpdateInstallmentStatus(ctx context.Context, ledgerID, installmentID, status string, updatedAt time.Time) (Installment, error) {
	return Installment{}, nil
}

func (f *fakeRepo) ListInstallmentsForPosting(ctx context.Context, ledgerID, cardAccountID string, month time.Time) ([]Installment, error) {
	return nil, nil
}

func (f *fakeRepo) GetPlanCategory(ctx context.Context, ledgerID, planID string) (string, error) {
	return "cat-1", nil
}

func (f *fakeRepo) MarkInstallmentPosted(ctx context.Context, ledgerID, installmentID, transactionID string, updatedAt time.Time) error {
	return nil
}

func (f *fakeRepo) GetStatement(ctx context.Context, ledgerID, cardAccountID, statementID string) (Statement, error) {
	return Statement{}, nil
}

func (f *fakeRepo) GetStatementByMonth(ctx context.Context, ledgerID, cardAccountID string, month time.Time) (Statement, error) {
	return Statement{}, nil
}

func (f *fakeRepo) CreateStatement(ctx context.Context, statement Statement) (Statement, error) {
	return statement, nil
}

func (f *fakeRepo) UpdateStatementTotals(ctx context.Context, statementID string, totalCharges, totalPayments int64, status string, updatedAt time.Time) (Statement, error) {
	return Statement{}, nil
}

func (f *fakeRepo) ListStatements(ctx context.Context, ledgerID, cardAccountID string, month *time.Time) ([]Statement, error) {
	return nil, nil
}

func (f *fakeRepo) SetStatementPayment(ctx context.Context, statementID, paymentTransactionID string, totalPayments int64, status string, updatedAt time.Time) (Statement, error) {
	return Statement{}, nil
}

func (f *fakeRepo) MarkInstallmentsPaid(ctx context.Context, ledgerID, cardAccountID string, month time.Time, statementID string, updatedAt time.Time) error {
	return nil
}

func (f *fakeRepo) SumStatementCharges(ctx context.Context, ledgerID, cardAccountID string, month time.Time) (int64, error) {
	return 0, nil
}

func (f *fakeRepo) CreateTransaction(ctx context.Context, params journal.CreateTransactionParams) (journal.Transaction, error) {
	return journal.Transaction{}, nil
}

func TestCancelPlanBlocksPostedInstallments(t *testing.T) {
	repo := &fakeRepo{role: "editor", hasPosted: true}
	service := NewService(repo)

	err := service.CancelPlan(context.Background(), "user-1", "ledger-1", "card-1", "plan-1")
	require.Equal(t, ErrTransactionReferenced, err)
}

func TestCreatePlanRejectsNonBudgetCategory(t *testing.T) {
	repo := &fakeRepo{role: "editor", accountType: "credit_card", categoryDirection: "in", categoryRelevant: true}
	service := NewService(repo)

	firstDue := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := service.CreatePlan(context.Background(), "user-1", "ledger-1", "card-1", InstallmentPlan{
		PurchaseOccurredAt: time.Now().UTC(),
		Description:        "Compra",
		CategoryID:         "cat-1",
		TotalAmountCents:   10000,
		FirstDueMonth:      firstDue,
	}, 2, 5000)

	var appErr *Error
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, "VALIDATION_ERROR", appErr.Code())
}
