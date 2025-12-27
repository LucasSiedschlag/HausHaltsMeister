package installment

import (
	"context"
	"fmt"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/category"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/ledger"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/payment"
)

type InstallmentService struct {
	repo          Repository
	catRepo       category.Repository
	ledgerService ledger.Service
	payRepo       payment.Repository
}

func NewService(repo Repository, catRepo category.Repository, ledgerService ledger.Service, payRepo payment.Repository) *InstallmentService {
	return &InstallmentService{
		repo:          repo,
		catRepo:       catRepo,
		ledgerService: ledgerService,
		payRepo:       payRepo,
	}
}

func (s *InstallmentService) CreateInstallmentPurchase(ctx context.Context, description string, totalAmount float64, count int32, categoryID int32, paymentMethodID int32, purchaseDate time.Time) (*InstallmentPlan, error) {
	pm, err := s.payRepo.GetByID(ctx, paymentMethodID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment method: %w", err)
	}
	if pm == nil {
		return nil, payment.ErrPaymentMethodNotFound
	}
	if pm.AccountID == 0 {
		return nil, fmt.Errorf("payment method account is not configured")
	}

	cat, err := s.catRepo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	if cat == nil {
		return nil, category.ErrCategoryNotFound
	}
	if cat.Direction != category.DirectionOut {
		return nil, ErrInvalidCategory
	}

	firstDueDate := calculateFirstDueDate(pm, purchaseDate)

	plan, err := NewPlan(description, totalAmount, count, firstDueDate, paymentMethodID, categoryID)
	if err != nil {
		return nil, err
	}

	createdPlan, err := s.repo.CreatePlan(ctx, plan)
	if err != nil {
		return nil, err
	}

	expenseAccount, err := s.ensureExpenseAccount(ctx)
	if err != nil {
		return nil, err
	}

	currentDueDate := firstDueDate
	for i := 0; i < int(count); i++ {
		title := fmt.Sprintf("%s (%d/%d)", description, i+1, count)

		tx, err := s.ledgerService.CreateTransaction(ctx, currentDueDate, title, "", "", []*ledger.Posting{
			{
				AccountID:  expenseAccount.ID,
				CategoryID: &categoryID,
				Side:       ledger.PostingSideDebit,
				Amount:     createdPlan.InstallmentAmount,
			},
			{
				AccountID: pm.AccountID,
				Side:      ledger.PostingSideCredit,
				Amount:    createdPlan.InstallmentAmount,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create posting for installment %d: %w", i+1, err)
		}

		item := &InstallmentPlanItem{
			PlanID:        createdPlan.ID,
			Sequence:      int32(i + 1),
			DueDate:       currentDueDate,
			Amount:        createdPlan.InstallmentAmount,
			TransactionID: &tx.ID,
		}
		if _, err := s.repo.CreatePlanItem(ctx, item); err != nil {
			return nil, fmt.Errorf("failed to persist installment %d: %w", i+1, err)
		}

		currentDueDate = currentDueDate.AddDate(0, 1, 0)
	}

	return createdPlan, nil
}

func (s *InstallmentService) ensureExpenseAccount(ctx context.Context) (*ledger.Account, error) {
	accounts, err := s.ledgerService.ListAccounts(ctx)
	if err != nil {
		return nil, err
	}
	for _, account := range accounts {
		if account.Type == ledger.AccountTypeExpense && account.Name == "Despesa" {
			return account, nil
		}
	}

	return s.ledgerService.CreateAccount(ctx, "Despesa", ledger.AccountTypeExpense, "BRL", true)
}

func calculateFirstDueDate(pm *payment.PaymentMethod, purchaseDate time.Time) time.Time {
	if pm.Kind != payment.KindCreditCard || pm.ClosingDay == nil || pm.DueDay == nil {
		return purchaseDate
	}

	cDay := *pm.ClosingDay
	dDay := *pm.DueDay

	closingDate := time.Date(purchaseDate.Year(), purchaseDate.Month(), int(cDay), 0, 0, 0, 0, purchaseDate.Location())
	if purchaseDate.Day() >= int(cDay) {
		closingDate = closingDate.AddDate(0, 1, 0)
	}

	dueMonth := closingDate
	if int(dDay) <= int(cDay) {
		dueMonth = dueMonth.AddDate(0, 1, 0)
	}

	return time.Date(dueMonth.Year(), dueMonth.Month(), int(dDay), 0, 0, 0, 0, dueMonth.Location())
}
