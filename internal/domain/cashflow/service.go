package cashflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/category"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/installment"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/ledger"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/payment"
)

var (
	ErrCategoryNotFound     = errors.New("category not found")
	ErrDirectionMismatch    = errors.New("cash flow direction does not match category direction")
	ErrFixedCategoryOnly    = errors.New("fixed expenses must use category Custos fixos")
	ErrFixedDirectionOnly   = errors.New("fixed expenses must be OUT")
	ErrCashFlowNotFound     = errors.New("cash flow not found")
	ErrReversalNotAllowed   = errors.New("cannot reverse a reversal entry")
	ErrPaymentMethodMissing = errors.New("payment method not found")
)

type CashFlowService struct {
	repo      Repository
	catRepo   category.Repository
	ledgerSvc ledger.Service
	payRepo   payment.Repository
	instSvc   installment.Service
}

func NewService(repo Repository, catRepo category.Repository, ledgerSvc ledger.Service, payRepo payment.Repository, instSvc installment.Service) *CashFlowService {
	return &CashFlowService{
		repo:      repo,
		catRepo:   catRepo,
		ledgerSvc: ledgerSvc,
		payRepo:   payRepo,
		instSvc:   instSvc,
	}
}

func (s *CashFlowService) CreateCashFlow(ctx context.Context, date time.Time, categoryID int32, paymentMethodID int32, direction, title string, amount float64, isFixed bool) (*CashFlow, error) {
	cat, err := s.catRepo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	if cat == nil {
		return nil, ErrCategoryNotFound
	}

	resolvedDirection := direction
	if resolvedDirection == "" {
		resolvedDirection = cat.Direction
	}
	if cat.Direction != resolvedDirection {
		return nil, ErrDirectionMismatch
	}

	if isFixed {
		if resolvedDirection != category.DirectionOut {
			return nil, ErrFixedDirectionOnly
		}
		if cat.Name != "Custos fixos" {
			return nil, ErrFixedCategoryOnly
		}
	}

	paymentMethod, err := s.payRepo.GetByID(ctx, paymentMethodID)
	if err != nil {
		return nil, err
	}
	if paymentMethod == nil {
		return nil, ErrPaymentMethodMissing
	}

	occurredAt := normalizeMonth(date)
	if paymentMethod.Kind == payment.KindCreditCard && resolvedDirection == category.DirectionOut {
		occurredAt = calculateCardDueDate(paymentMethod, date)
	}

	newFlow, err := New(occurredAt, categoryID, paymentMethodID, resolvedDirection, title, amount, isFixed)
	if err != nil {
		return nil, fmt.Errorf("domain validation failed: %w", err)
	}

	tx, err := s.createLedgerTransaction(ctx, paymentMethod, cat, occurredAt, title, amount)
	if err != nil {
		return nil, err
	}

	newFlow.TransactionID = tx.ID
	created, err := s.repo.Create(ctx, newFlow)
	if err != nil {
		return nil, err
	}
	created.CategoryName = cat.Name
	created.PaymentMethodName = paymentMethod.Name
	return created, nil
}

func (s *CashFlowService) UpdateCashFlow(ctx context.Context, id int32, date time.Time, categoryID int32, paymentMethodID int32, direction, title string, amount float64, isFixed bool) (*CashFlow, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrCashFlowNotFound
	}

	cat, err := s.catRepo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	if cat == nil {
		return nil, ErrCategoryNotFound
	}

	resolvedDirection := direction
	if resolvedDirection == "" {
		resolvedDirection = cat.Direction
	}
	if cat.Direction != resolvedDirection {
		return nil, ErrDirectionMismatch
	}

	if isFixed {
		if resolvedDirection != category.DirectionOut {
			return nil, ErrFixedDirectionOnly
		}
		if cat.Name != "Custos fixos" {
			return nil, ErrFixedCategoryOnly
		}
	}

	paymentMethod, err := s.payRepo.GetByID(ctx, paymentMethodID)
	if err != nil {
		return nil, err
	}
	if paymentMethod == nil {
		return nil, ErrPaymentMethodMissing
	}

	occurredAt := normalizeMonth(date)
	if paymentMethod.Kind == payment.KindCreditCard && resolvedDirection == category.DirectionOut {
		occurredAt = calculateCardDueDate(paymentMethod, date)
	}

	transactions, err := s.ledgerSvc.ListTransactionsByMonth(ctx, normalizeMonth(existing.Date))
	if err != nil {
		return nil, err
	}

	var entryPostings []*ledger.Posting
	for _, tx := range transactions {
		if tx.ID == existing.TransactionID {
			entryPostings = tx.Postings
			break
		}
	}
	if len(entryPostings) == 0 {
		return nil, fmt.Errorf("transaction postings not found")
	}

	updatedPostings, err := s.buildUpdatedPostings(ctx, existing.TransactionID, entryPostings, paymentMethod, cat, resolvedDirection, amount)
	if err != nil {
		return nil, err
	}

	if _, err := s.ledgerSvc.UpdateTransaction(ctx, existing.TransactionID, occurredAt, title, "", "", updatedPostings); err != nil {
		return nil, err
	}

	updated := &CashFlow{
		ID:              existing.ID,
		TransactionID:   existing.TransactionID,
		CategoryID:      categoryID,
		PaymentMethodID: paymentMethodID,
		Direction:       resolvedDirection,
		Title:           title,
		Amount:          amount,
		IsFixed:         isFixed,
		Date:              occurredAt,
		ReversalOfEntryID: existing.ReversalOfEntryID,
	}

	updatedFlow, err := s.repo.Update(ctx, updated)
	if err != nil {
		return nil, err
	}
	updatedFlow.CategoryName = cat.Name
	updatedFlow.PaymentMethodName = paymentMethod.Name
	return updatedFlow, nil
}

func (s *CashFlowService) DeleteCashFlow(ctx context.Context, id int32) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrCashFlowNotFound
	}
	return s.ledgerSvc.DeleteTransaction(ctx, existing.TransactionID)
}

func (s *CashFlowService) ReverseCashFlow(ctx context.Context, id int32) (*CashFlow, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrCashFlowNotFound
	}
	if existing.ReversalOfEntryID != nil {
		return nil, ErrReversalNotAllowed
	}

	transactions, err := s.ledgerSvc.ListTransactionsByMonth(ctx, normalizeMonth(existing.Date))
	if err != nil {
		return nil, err
	}

	var entryPostings []*ledger.Posting
	for _, tx := range transactions {
		if tx.ID == existing.TransactionID {
			entryPostings = tx.Postings
			break
		}
	}
	if len(entryPostings) == 0 {
		return nil, fmt.Errorf("transaction postings not found")
	}

	reversedPostings := make([]*ledger.Posting, len(entryPostings))
	for i, posting := range entryPostings {
		side := ledger.PostingSideDebit
		if posting.Side == ledger.PostingSideDebit {
			side = ledger.PostingSideCredit
		}
		reversedPostings[i] = &ledger.Posting{
			AccountID:  posting.AccountID,
			CategoryID: posting.CategoryID,
			PartyID:    posting.PartyID,
			Side:       side,
			Amount:     posting.Amount,
			Memo:       "Estorno",
		}
	}

	reversalTitle := fmt.Sprintf("Estorno: %s", existing.Title)
	tx, err := s.ledgerSvc.CreateTransaction(ctx, existing.Date, reversalTitle, "", "", reversedPostings)
	if err != nil {
		return nil, err
	}

	reversal := &CashFlow{
		TransactionID:     tx.ID,
		CategoryID:        existing.CategoryID,
		PaymentMethodID:   existing.PaymentMethodID,
		Direction:         existing.Direction,
		Title:             reversalTitle,
		Amount:            existing.Amount,
		IsFixed:           false,
		Date:              existing.Date,
		ReversalOfEntryID: &existing.ID,
	}

	created, err := s.repo.Create(ctx, reversal)
	if err != nil {
		return nil, err
	}
	created.CategoryName = existing.CategoryName
	created.PaymentMethodName = existing.PaymentMethodName
	return created, nil
}

func (s *CashFlowService) ListCashFlows(ctx context.Context, month time.Time, direction *string, isFixed *bool) ([]*CashFlow, error) {
	return s.repo.ListByMonth(ctx, month, direction, isFixed)
}

func (s *CashFlowService) CopyFixedExpenses(ctx context.Context, fromMonth, toMonth time.Time) (int, error) {
	sourceFlows, err := s.repo.ListByMonth(ctx, fromMonth, nil, boolPtr(true))
	if err != nil {
		return 0, fmt.Errorf("failed to list source month expenses: %w", err)
	}

	count := 0
	for _, flow := range sourceFlows {
		if flow.ReversalOfEntryID != nil {
			continue
		}

		targetDate := normalizeMonth(toMonth)
		_, err := s.CreateCashFlow(
			ctx,
			targetDate,
			flow.CategoryID,
			flow.PaymentMethodID,
			flow.Direction,
			flow.Title,
			flow.Amount,
			true,
		)
		if err != nil {
			return count, fmt.Errorf("failed to copy flow %d: %w", flow.ID, err)
		}
		count++
	}
	return count, nil
}

func (s *CashFlowService) GetMonthlySummary(ctx context.Context, month time.Time) (*MonthlySummary, error) {
	return s.repo.GetMonthlySummary(ctx, month)
}

func (s *CashFlowService) GetCategorySummary(ctx context.Context, month time.Time) ([]CategorySummary, error) {
	return s.repo.GetCategorySummary(ctx, month)
}

func (s *CashFlowService) createLedgerTransaction(ctx context.Context, pm *payment.PaymentMethod, cat *category.Category, occurredAt time.Time, title string, amount float64) (*ledger.Transaction, error) {
	accountType := ledger.AccountTypeExpense
	var sideForCategory string
	var sideForPayment string

	if cat.Direction == category.DirectionIn {
		accountType = ledger.AccountTypeIncome
		sideForCategory = ledger.PostingSideCredit
		sideForPayment = ledger.PostingSideDebit
	} else {
		sideForCategory = ledger.PostingSideDebit
		sideForPayment = ledger.PostingSideCredit
	}

	categoryAccount, err := s.ensureAccount(ctx, accountType)
	if err != nil {
		return nil, err
	}

	return s.ledgerSvc.CreateTransaction(ctx, occurredAt, title, "", "", []*ledger.Posting{
		{
			AccountID:  categoryAccount.ID,
			CategoryID: &cat.ID,
			Side:       sideForCategory,
			Amount:     amount,
		},
		{
			AccountID: pm.AccountID,
			Side:      sideForPayment,
			Amount:    amount,
		},
	})
}

func (s *CashFlowService) buildUpdatedPostings(ctx context.Context, transactionID int32, existing []*ledger.Posting, pm *payment.PaymentMethod, cat *category.Category, direction string, amount float64) ([]*ledger.Posting, error) {
	if len(existing) < 2 {
		return nil, fmt.Errorf("transaction postings are incomplete")
	}

	var categoryPosting *ledger.Posting
	var paymentPosting *ledger.Posting
	for _, posting := range existing {
		if posting.CategoryID != nil {
			categoryPosting = posting
		} else {
			paymentPosting = posting
		}
	}
	if categoryPosting == nil || paymentPosting == nil {
		return nil, fmt.Errorf("transaction postings are incomplete")
	}

	accountType := ledger.AccountTypeExpense
	categorySide := ledger.PostingSideDebit
	paymentSide := ledger.PostingSideCredit
	if direction == category.DirectionIn {
		accountType = ledger.AccountTypeIncome
		categorySide = ledger.PostingSideCredit
		paymentSide = ledger.PostingSideDebit
	}

	categoryAccount, err := s.ensureAccount(ctx, accountType)
	if err != nil {
		return nil, err
	}

	categoryPosting.ID = categoryPosting.ID
	categoryPosting.TransactionID = transactionID
	categoryPosting.AccountID = categoryAccount.ID
	categoryPosting.CategoryID = &cat.ID
	categoryPosting.Side = categorySide
	categoryPosting.Amount = amount

	paymentPosting.ID = paymentPosting.ID
	paymentPosting.TransactionID = transactionID
	paymentPosting.AccountID = pm.AccountID
	paymentPosting.CategoryID = nil
	paymentPosting.Side = paymentSide
	paymentPosting.Amount = amount

	return []*ledger.Posting{categoryPosting, paymentPosting}, nil
}

func (s *CashFlowService) ensureAccount(ctx context.Context, accountType string) (*ledger.Account, error) {
	accounts, err := s.ledgerSvc.ListAccounts(ctx)
	if err != nil {
		return nil, err
	}
	name := "Despesas"
	if accountType == ledger.AccountTypeIncome {
		name = "Receitas"
	}
	for _, account := range accounts {
		if account.Type == accountType && account.Name == name {
			return account, nil
		}
	}
	return s.ledgerSvc.CreateAccount(ctx, name, accountType, "BRL", true)
}

func normalizeMonth(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
}

func calculateCardDueDate(pm *payment.PaymentMethod, purchaseDate time.Time) time.Time {
	if pm.Kind != payment.KindCreditCard || pm.ClosingDay == nil || pm.DueDay == nil {
		return normalizeMonth(purchaseDate)
	}

	closing := time.Date(purchaseDate.Year(), purchaseDate.Month(), int(*pm.ClosingDay), 0, 0, 0, 0, purchaseDate.Location())
	if purchaseDate.Day() >= int(*pm.ClosingDay) {
		closing = closing.AddDate(0, 1, 0)
	}

	dueMonth := closing
	if int(*pm.DueDay) <= int(*pm.ClosingDay) {
		dueMonth = dueMonth.AddDate(0, 1, 0)
	}

	return time.Date(dueMonth.Year(), dueMonth.Month(), int(*pm.DueDay), 0, 0, 0, 0, purchaseDate.Location())
}

func boolPtr(value bool) *bool {
	return &value
}
