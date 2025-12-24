package installment

import (
	"context"
	"fmt"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/cashflow"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/payment"
)

type InstallmentService struct {
	repo      Repository
	cfService *cashflow.CashFlowService
	payRepo   payment.Repository
}

func NewService(repo Repository, cfService *cashflow.CashFlowService, payRepo payment.Repository) *InstallmentService {
	return &InstallmentService{
		repo:      repo,
		cfService: cfService,
		payRepo:   payRepo,
	}
}

func (s *InstallmentService) CreateInstallmentPurchase(ctx context.Context, description string, totalAmount float64, count int32, categoryID int32, paymentMethodID int32, purchaseDate time.Time) (*InstallmentPlan, error) {
	// 1. Get Payment Method
	pm, err := s.payRepo.GetByID(ctx, paymentMethodID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment method: %w", err)
	}
	if pm == nil {
		return nil, fmt.Errorf("payment method not found")
	}

	// 2. Initial Plan Object (without ID)
	plan, err := NewPlan(description, totalAmount, count, purchaseDate, paymentMethodID)
	if err != nil {
		return nil, err
	}

	// 3. Persist Plan
	createdPlan, err := s.repo.CreatePlan(ctx, plan)
	if err != nil {
		return nil, err
	}

	// 4. Generate CashFlows
	installmentAmount := createdPlan.InstallmentAmount

	// Date Logic
	// If Credit Card:
	// Start Month depends on ClosingDay.
	// We need to find the due date of the invoice where 'purchaseDate' falls.
	// Simplification:
	// If purchaseDate < ClosingDay (of purchase month) -> Invoice closes this month -> Due date is this month's DueDay (or next month if DueDay < ClosingDay logic, which is rare but possible).
	// Usually Closing Day Y < Due Day Y+N.
	// Let's assume standard: Closing Day 1, Due Day 7.
	// Purchase Jan 20. Closing Feb 1. Due Feb 7.
	// Logic:
	// Find next Closing Date after Purchase Date.
	// The Due Date associated with that Closing Date is the first payment.

	// Algorithm:
	// baseDate = purchaseDate
	// closingDay = *pm.ClosingDay
	// dueDay = *pm.DueDay

	// Create a date for Closing Day in the purchase month.
	// closingDateThisMonth = Date(purchaseYear, purchaseMonth, closingDay)
	// If purchaseDate < closingDateThisMonth:
	//    This purchase belongs to invoice closing on closingDateThisMonth.
	//    The due date is in the same month as closing date (usually).
	// Else:
	//    Belongs to next month's closing.

	// Is this robust?
	// If Closing Day = 5. Purchase = 4. Belongs to invoice closing on 5th.
	// If Closing Day = 5. Purchase = 6. Belongs to invoice closing on 5th of next month.

	// Refined Logic:
	// First Due Date Calculation:
	// If NOT Credit Card (e.g. Loan, specific agreement), start month is StartMonth?
	// If Credit Card:
	//   closingDay := *pm.ClosingDay
	//   dueDay := *pm.DueDay
	//   effectiveClosingDate = Date(purchaseDate.Year, purchaseDate.Month, closingDay)
	//   if purchaseDate.Day() >= closingDay {
	//      effectiveClosingDate = effectiveClosingDate.AddDate(0, 1, 0) // Next month
	//   }
	//   // Now effectiveClosingDate is when the invoice closes.
	//   // Due date is usually related to closing date.
	//   // Often Due Date is in the SAME month as Closing Date, or Next Month?
	//   // Nubank: Closes 7 days before due date.
	//   // Lets assume Due Date is in the same month as Closing Date? Or simply use the DueDay of that month?
	//   // If Closing is 1st, Due is 7th.
	//   // EffectiveClosing: Feb 1. Due: Feb 7.
	//   // If Closing is 25th. Due is 5th (Next Month).
	//   // This is getting complex.
	//   // AGENTIC SIMPLIFICATION:
	//   // The user just wants "1/N" generated.
	//   // Let's set the FIRST DUE DATE to:
	//   //   If purchase.Day < closing.Day: This month's Due Day (if ThisMonth.DueDay > purchase.Day ?? No)
	//   //   Actually, let's just project the Closing Date forward.
	//   //   Limit: Payment happens on DueDay.
	//   //   We need to find the FIRST DueDay that is > PurchaseDate + (Closing-Due Gap).
	//   //   Let's stick to: "Purchase enters the 'Open' invoice".
	//   //   Open Invoice closes on Next occurrence of Closing Day.
	//   //   Its payment is on the corresponding Due Day.

	currentDueDate := purchaseDate
	if pm.Kind == "CREDIT_CARD" && pm.ClosingDay != nil && pm.DueDay != nil {
		cDay := *pm.ClosingDay
		dDay := *pm.DueDay

		// Find next closing date
		// Set to current month closing day
		closingDate := time.Date(purchaseDate.Year(), purchaseDate.Month(), int(cDay), 0, 0, 0, 0, purchaseDate.Location())

		// If closing date is in past or today (assuming closes at end of day, purchase during day),
		// if purchase >= closingDate -> it's next invoice.
		// Wait, if closing is 1st. Purchase 20th.
		// closingDate (this month) = 1st. Purchase > Closing.
		// Next Closing = Next Month 1st.

		if purchaseDate.Day() >= int(cDay) {
			closingDate = closingDate.AddDate(0, 1, 0)
		}

		// Now we have the Closing Date of the invoice.
		// The Due Date is usually X days after closing.
		// Or strictly on Due Day of ... same month? Or next?
		// If Closing 1st, Due 7th -> Same Month.
		// If Closing 25th, Due 5th -> Next Month.

		// We can infer match based on Days.
		// If DueDay > ClosingDay -> Same Month.
		// If DueDay <= ClosingDay -> Next Month.

		dueMonth := closingDate
		if int(dDay) <= int(cDay) {
			dueMonth = dueMonth.AddDate(0, 1, 0)
		}

		currentDueDate = time.Date(dueMonth.Year(), dueMonth.Month(), int(dDay), 0, 0, 0, 0, dueMonth.Location())
	} else {
		// Not Credit Card, just monthly starting next month?
		// Or start immediately?
		// Let's start immediately (today/purchase date) for 1st installment?
		// Or 1 month from now?
		// Assuming Loan: 1st payment 1 month from now?
		// Let's assume currentDueDate = purchaseDate (immediate) or purchaseDate + 1 month?
		// Better: Input said "StartMonth". We used purchaseDate as StartMonth.
		// Let's use it as first due date if not card.
		currentDueDate = purchaseDate
	}

	for i := 0; i < int(count); i++ {
		title := fmt.Sprintf("%s (%d/%d)", description, i+1, count)

		// Create CashFlow
		// Direction OUT implied for purchases
		cf, err := s.cfService.CreateCashFlow(ctx, currentDueDate, categoryID, "OUT", title, installmentAmount, false)
		if err != nil {
			return nil, fmt.Errorf("failed to create installment %d: %w", i+1, err)
		}

		// Link ExpenseDetail
		affectsCard := (pm.Kind == "CREDIT_CARD")
		err = s.repo.CreateExpenseDetail(ctx, cf.ID, paymentMethodID, createdPlan.ID, affectsCard)
		if err != nil {
			return nil, fmt.Errorf("failed to link installment %d: %w", i+1, err)
		}

		// Advance Date (1 Month)
		currentDueDate = currentDueDate.AddDate(0, 1, 0)
	}

	return createdPlan, nil
}
