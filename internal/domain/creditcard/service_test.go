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
	account           AccountInfo
	accountNature     string
	categoryDirection string
	categoryRelevant  bool
	hasPosted         bool
	networkErr        error
	cardLookupErr     error
	card              CreditCard
	installments      []Installment
	posted            []string
	sumCharges        int64
	statement         Statement
	createdTxs        int
	createdAccount    AccountInfo
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

func (f *fakeRepo) ListCreditCards(ctx context.Context, ledgerID, parentAccountID string) ([]CreditCard, error) {
	return nil, nil
}

func (f *fakeRepo) GetCreditCard(ctx context.Context, cardID string) (CreditCard, error) {
	if f.cardLookupErr != nil {
		return CreditCard{}, f.cardLookupErr
	}
	if f.card.ID == "" {
		return CreditCard{}, ErrNotFound
	}
	return f.card, nil
}

func (f *fakeRepo) CreateCreditCard(ctx context.Context, params CreditCard) (CreditCard, error) {
	return params, nil
}

func (f *fakeRepo) UpdateCreditCard(ctx context.Context, params CreditCard, updatedAt time.Time) (CreditCard, error) {
	return params, nil
}

func (f *fakeRepo) DeleteCreditCard(ctx context.Context, ledgerID, cardID string) error {
	return nil
}

func (f *fakeRepo) GetAccount(ctx context.Context, userID, accountID string) (AccountInfo, error) {
	if f.account.ID == "" {
		return AccountInfo{}, ErrNotFound
	}
	return f.account, nil
}

func (f *fakeRepo) CreateAccount(ctx context.Context, ledgerID, name, accountType, nature string, isActive bool) (AccountInfo, error) {
	f.createdAccount = AccountInfo{
		ID:       "liab-1",
		LedgerID: ledgerID,
		Type:     accountType,
		Nature:   nature,
		IsActive: isActive,
	}
	return f.createdAccount, nil
}

func (f *fakeRepo) GetAccountNature(ctx context.Context, ledgerID, accountID string) (string, error) {
	return f.accountNature, nil
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
	if f.networkErr != nil {
		return CardNetwork{}, f.networkErr
	}
	return CardNetwork{Code: code}, nil
}

func (f *fakeRepo) CreatePlanWithInstallments(ctx context.Context, plan InstallmentPlan, installments []Installment) (InstallmentPlan, error) {
	if plan.ID == "" {
		plan.ID = "plan-1"
	}
	return plan, nil
}

func (f *fakeRepo) ListPlans(ctx context.Context, ledgerID, cardID string, status *string) ([]InstallmentPlan, error) {
	return nil, nil
}

func (f *fakeRepo) GetPlan(ctx context.Context, ledgerID, cardID, planID string) (InstallmentPlan, error) {
	return InstallmentPlan{ID: planID, LedgerID: ledgerID, CreditCardID: cardID}, nil
}

func (f *fakeRepo) UpdatePlanStatus(ctx context.Context, ledgerID, planID, status string, updatedAt time.Time) error {
	return nil
}

func (f *fakeRepo) HasPostedInstallments(ctx context.Context, ledgerID, planID string) (bool, error) {
	return f.hasPosted, nil
}

func (f *fakeRepo) ListInstallments(ctx context.Context, ledgerID, cardID string, month *time.Time, status *string) ([]Installment, error) {
	return nil, nil
}

func (f *fakeRepo) UpdateInstallmentStatus(ctx context.Context, ledgerID, installmentID, status string, updatedAt time.Time) (Installment, error) {
	return Installment{}, nil
}

func (f *fakeRepo) ListInstallmentsForPosting(ctx context.Context, ledgerID, cardID string, month time.Time) ([]Installment, error) {
	return f.installments, nil
}

func (f *fakeRepo) GetPlanCategory(ctx context.Context, ledgerID, planID string) (string, error) {
	return "cat-1", nil
}

func (f *fakeRepo) MarkInstallmentPosted(ctx context.Context, ledgerID, installmentID, transactionID string, updatedAt time.Time) error {
	f.posted = append(f.posted, installmentID)
	return nil
}

func (f *fakeRepo) GetStatement(ctx context.Context, ledgerID, cardID, statementID string) (Statement, error) {
	return f.statement, nil
}

func (f *fakeRepo) GetStatementByMonth(ctx context.Context, ledgerID, cardID string, month time.Time) (Statement, error) {
	if f.statement.ID == "" {
		return Statement{}, ErrNotFound
	}
	return f.statement, nil
}

func (f *fakeRepo) CreateStatement(ctx context.Context, statement Statement) (Statement, error) {
	return statement, nil
}

func (f *fakeRepo) UpdateStatementTotals(ctx context.Context, ledgerID, cardID, statementID string, totalCharges, totalPayments int64, status string, updatedAt time.Time) (Statement, error) {
	return Statement{ID: statementID, TotalChargesCents: totalCharges, TotalPaymentsCents: totalPayments, Status: status}, nil
}

func (f *fakeRepo) ListStatements(ctx context.Context, ledgerID, cardID string, month *time.Time) ([]Statement, error) {
	return nil, nil
}

func (f *fakeRepo) SetStatementPayment(ctx context.Context, ledgerID, cardID, statementID, paymentTransactionID string, totalPayments int64, status string, updatedAt time.Time) (Statement, error) {
	return Statement{}, nil
}

func (f *fakeRepo) MarkInstallmentsPaid(ctx context.Context, ledgerID, cardID string, month time.Time, statementID string, updatedAt time.Time) error {
	return nil
}

func (f *fakeRepo) SumStatementCharges(ctx context.Context, ledgerID, cardID string, month time.Time) (int64, error) {
	return f.sumCharges, nil
}

func (f *fakeRepo) CreateTransaction(ctx context.Context, params journal.CreateTransactionParams) (journal.Transaction, error) {
	f.createdTxs++
	return journal.Transaction{ID: "tx-1"}, nil
}

func TestCreateCreditCardRequiresEditor(t *testing.T) {
	repo := &fakeRepo{
		role: "viewer",
		account: AccountInfo{
			ID:       "acc-1",
			LedgerID: "ledger-1",
			Nature:   "asset",
			IsActive: true,
		},
	}
	service := NewService(repo)

	_, err := service.CreateCreditCard(context.Background(), "user-1", "acc-1", CreditCard{ClosingDay: 10, DueDay: 20})
	require.Equal(t, ErrAccessDenied, err)
}

func TestCreateCreditCardRejectsLiabilityAccount(t *testing.T) {
	repo := &fakeRepo{
		role: "editor",
		account: AccountInfo{
			ID:       "acc-1",
			LedgerID: "ledger-1",
			Nature:   "liability",
			IsActive: true,
		},
	}
	service := NewService(repo)

	_, err := service.CreateCreditCard(context.Background(), "user-1", "acc-1", CreditCard{ClosingDay: 10, DueDay: 20})
	require.Error(t, err)
	require.Equal(t, "VALIDATION_ERROR", err.(*Error).Code())
}

func TestCreateCreditCardSuccess(t *testing.T) {
	repo := &fakeRepo{
		role: "editor",
		account: AccountInfo{
			ID:       "acc-1",
			LedgerID: "ledger-1",
			Nature:   "asset",
			IsActive: true,
		},
	}
	service := NewService(repo)

	card, err := service.CreateCreditCard(context.Background(), "user-1", "acc-1", CreditCard{
		Brand:      "visa",
		ClosingDay: 10,
		DueDay:     20,
	})
	require.NoError(t, err)
	require.Equal(t, "acc-1", card.ParentAccountID)
	require.Equal(t, "ledger-1", card.LedgerID)
	require.Equal(t, "liab-1", card.LiabilityAccountID)
	require.Equal(t, "current", repo.createdAccount.Type)
	require.Equal(t, "liability", repo.createdAccount.Nature)
}

func TestCreatePlanRejectsNonBudgetCategory(t *testing.T) {
	repo := &fakeRepo{
		role:              "editor",
		categoryDirection: "in",
		categoryRelevant:  true,
		card: CreditCard{
			ID:                 "card-1",
			LedgerID:           "ledger-1",
			LiabilityAccountID: "liab-1",
		},
	}
	service := NewService(repo)

	firstDue := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := service.CreatePlan(context.Background(), "user-1", "card-1", InstallmentPlan{
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

func TestCreatePlanSuccess(t *testing.T) {
	repo := &fakeRepo{
		role:              "editor",
		categoryDirection: "out",
		categoryRelevant:  true,
		card: CreditCard{
			ID:                 "card-1",
			LedgerID:           "ledger-1",
			LiabilityAccountID: "liab-1",
		},
	}
	service := NewService(repo)

	firstDue := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	plan, err := service.CreatePlan(context.Background(), "user-1", "card-1", InstallmentPlan{
		PurchaseOccurredAt: time.Now().UTC(),
		Description:        "Compra",
		CategoryID:         "cat-1",
		TotalAmountCents:   10000,
		FirstDueMonth:      firstDue,
	}, 2, 5000)
	require.NoError(t, err)
	require.Equal(t, "card-1", plan.CreditCardID)
}

func TestPostMonthCreatesTransactions(t *testing.T) {
	repo := &fakeRepo{
		role: "editor",
		card: CreditCard{
			ID:                 "card-1",
			LedgerID:           "ledger-1",
			LiabilityAccountID: "liab-1",
		},
		installments: []Installment{{ID: "inst-1", PlanID: "plan-1", AmountCents: 1000}},
	}
	service := NewService(repo)

	month := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	result, err := service.PostMonth(context.Background(), "user-1", "card-1", month)
	require.NoError(t, err)
	require.Equal(t, 1, result.PostedCount)
	require.Equal(t, 1, repo.createdTxs)
	require.Equal(t, []string{"inst-1"}, repo.posted)
}

func TestCloseStatementCreatesAndCloses(t *testing.T) {
	repo := &fakeRepo{
		role:       "editor",
		sumCharges: 12000,
		card: CreditCard{
			ID:                 "card-1",
			LedgerID:           "ledger-1",
			LiabilityAccountID: "liab-1",
		},
	}
	service := NewService(repo)

	month := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	statement, err := service.CloseStatement(context.Background(), "user-1", "card-1", month)
	require.NoError(t, err)
	require.Equal(t, "closed", statement.Status)
	require.Equal(t, int64(12000), statement.TotalChargesCents)
}

func TestPayStatementSuccess(t *testing.T) {
	repo := &fakeRepo{
		role:          "editor",
		accountNature: "asset",
		card: CreditCard{
			ID:                 "card-1",
			LedgerID:           "ledger-1",
			LiabilityAccountID: "liab-1",
		},
		statement: Statement{
			ID:                "stmt-1",
			LedgerID:          "ledger-1",
			CreditCardID:      "card-1",
			StatementMonth:    time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			TotalChargesCents: 10000,
			Status:            "closed",
		},
	}
	service := NewService(repo)

	_, err := service.PayStatement(context.Background(), "user-1", "card-1", "stmt-1", "acc-1", 10000, time.Now().UTC())
	require.NoError(t, err)
	require.Equal(t, 1, repo.createdTxs)
}
