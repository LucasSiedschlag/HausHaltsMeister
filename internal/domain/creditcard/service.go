package creditcard

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/journal"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/ledger"
)

type Repository interface {
	ListCardNetworks(ctx context.Context) ([]CardNetwork, error)
	CreateCardNetwork(ctx context.Context, code, displayName string) (CardNetwork, error)
	UpdateCardNetwork(ctx context.Context, code, displayName string, updatedAt time.Time) (CardNetwork, error)
	DeleteCardNetwork(ctx context.Context, code string) error

	ListCreditCards(ctx context.Context, ledgerID string) ([]CreditCard, error)
	GetCreditCard(ctx context.Context, ledgerID, cardAccountID string) (CreditCard, error)
	CreateCreditCard(ctx context.Context, params CreditCard) (CreditCard, error)
	UpdateCreditCard(ctx context.Context, params CreditCard, updatedAt time.Time) (CreditCard, error)
	DeleteCreditCard(ctx context.Context, ledgerID, cardAccountID string) error
	GetAccountType(ctx context.Context, ledgerID, accountID string) (string, error)
	GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error)
	LedgerExists(ctx context.Context, ledgerID string) (bool, error)
	GetCategoryBudgetInfo(ctx context.Context, ledgerID, categoryID string) (string, bool, error)
	FindCategoryByName(ctx context.Context, ledgerID, name string) (string, error)
	GetCardNetwork(ctx context.Context, code string) (CardNetwork, error)

	CreatePlanWithInstallments(ctx context.Context, plan InstallmentPlan, installments []Installment) (InstallmentPlan, error)
	ListPlans(ctx context.Context, ledgerID, cardAccountID string, status *string) ([]InstallmentPlan, error)
	GetPlan(ctx context.Context, ledgerID, cardAccountID, planID string) (InstallmentPlan, error)
	UpdatePlanStatus(ctx context.Context, ledgerID, planID, status string, updatedAt time.Time) error
	HasPostedInstallments(ctx context.Context, ledgerID, planID string) (bool, error)

	ListInstallments(ctx context.Context, ledgerID, cardAccountID string, month *time.Time, status *string) ([]Installment, error)
	UpdateInstallmentStatus(ctx context.Context, ledgerID, installmentID, status string, updatedAt time.Time) (Installment, error)
	ListInstallmentsForPosting(ctx context.Context, ledgerID, cardAccountID string, month time.Time) ([]Installment, error)
	GetPlanCategory(ctx context.Context, ledgerID, planID string) (string, error)
	MarkInstallmentPosted(ctx context.Context, ledgerID, installmentID, transactionID string, updatedAt time.Time) error

	GetStatement(ctx context.Context, ledgerID, cardAccountID, statementID string) (Statement, error)
	GetStatementByMonth(ctx context.Context, ledgerID, cardAccountID string, month time.Time) (Statement, error)
	CreateStatement(ctx context.Context, statement Statement) (Statement, error)
	UpdateStatementTotals(ctx context.Context, ledgerID, cardAccountID, statementID string, totalCharges, totalPayments int64, status string, updatedAt time.Time) (Statement, error)
	ListStatements(ctx context.Context, ledgerID, cardAccountID string, month *time.Time) ([]Statement, error)
	SetStatementPayment(ctx context.Context, ledgerID, cardAccountID, statementID, paymentTransactionID string, totalPayments int64, status string, updatedAt time.Time) (Statement, error)
	MarkInstallmentsPaid(ctx context.Context, ledgerID, cardAccountID string, month time.Time, statementID string, updatedAt time.Time) error
	SumStatementCharges(ctx context.Context, ledgerID, cardAccountID string, month time.Time) (int64, error)

	CreateTransaction(ctx context.Context, params journal.CreateTransactionParams) (journal.Transaction, error)
}

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now().UTC}
}

func (s *Service) ListCardNetworks(ctx context.Context) ([]CardNetwork, error) {
	return s.repo.ListCardNetworks(ctx)
}

func (s *Service) CreateCardNetwork(ctx context.Context, code, displayName string) (CardNetwork, error) {
	code = strings.TrimSpace(strings.ToLower(code))
	displayName = strings.TrimSpace(displayName)
	if code == "" || displayName == "" {
		return CardNetwork{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"code": "required"})
	}
	return s.repo.CreateCardNetwork(ctx, code, displayName)
}

func (s *Service) UpdateCardNetwork(ctx context.Context, code, displayName string) (CardNetwork, error) {
	code = strings.TrimSpace(strings.ToLower(code))
	displayName = strings.TrimSpace(displayName)
	if code == "" || displayName == "" {
		return CardNetwork{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"code": "required"})
	}
	return s.repo.UpdateCardNetwork(ctx, code, displayName, s.now())
}

func (s *Service) DeleteCardNetwork(ctx context.Context, code string) error {
	code = strings.TrimSpace(strings.ToLower(code))
	if code == "" {
		return NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"code": "required"})
	}
	return s.repo.DeleteCardNetwork(ctx, code)
}

func (s *Service) ListCreditCards(ctx context.Context, userID, ledgerID string) ([]CreditCard, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "viewer"); err != nil {
		return nil, err
	}
	return s.repo.ListCreditCards(ctx, ledgerID)
}

func (s *Service) GetCreditCard(ctx context.Context, userID, ledgerID, cardAccountID string) (CreditCard, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "viewer"); err != nil {
		return CreditCard{}, err
	}
	card, err := s.repo.GetCreditCard(ctx, ledgerID, cardAccountID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return CreditCard{}, ErrCardNotFound
		}
		return CreditCard{}, err
	}
	return card, nil
}

func (s *Service) CreateCreditCard(ctx context.Context, userID, ledgerID string, card CreditCard) (CreditCard, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return CreditCard{}, err
	}

	if card.AccountID == "" {
		return CreditCard{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"account_id": "required"})
	}
	accountType, err := s.repo.GetAccountType(ctx, ledgerID, card.AccountID)
	if err != nil {
		return CreditCard{}, ErrCardNotFound
	}
	if accountType != "credit_card" {
		return CreditCard{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"account_id": "type"})
	}

	if card.Network == "" {
		card.Network = "other"
	}
	if _, err := s.repo.GetCardNetwork(ctx, card.Network); err != nil {
		return CreditCard{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"network": "invalid"})
	}
	if card.ClosingDay < 1 || card.DueDay < 1 {
		return CreditCard{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"closing_day": "invalid"})
	}
	card.LedgerID = ledgerID

	return s.repo.CreateCreditCard(ctx, card)
}

func (s *Service) UpdateCreditCard(ctx context.Context, userID, ledgerID, cardAccountID string, card CreditCard) (CreditCard, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return CreditCard{}, err
	}
	card.AccountID = cardAccountID
	card.LedgerID = ledgerID

	if card.Network == "" {
		card.Network = "other"
	}
	if _, err := s.repo.GetCardNetwork(ctx, card.Network); err != nil {
		return CreditCard{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"network": "invalid"})
	}
	updated, err := s.repo.UpdateCreditCard(ctx, card, s.now())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return CreditCard{}, ErrCardNotFound
		}
		return CreditCard{}, err
	}
	return updated, nil
}

func (s *Service) DeleteCreditCard(ctx context.Context, userID, ledgerID, cardAccountID string) error {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return err
	}
	if err := s.repo.DeleteCreditCard(ctx, ledgerID, cardAccountID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrCardNotFound
		}
		return err
	}
	return nil
}

func (s *Service) CreatePlan(ctx context.Context, userID, ledgerID, cardAccountID string, input InstallmentPlan, installmentsCount int, installmentAmount int64) (InstallmentPlan, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return InstallmentPlan{}, err
	}
	if input.Description == "" {
		return InstallmentPlan{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"description": "required"})
	}
	if installmentsCount < 1 || installmentAmount <= 0 || input.TotalAmountCents <= 0 {
		return InstallmentPlan{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"installments_count": "invalid"})
	}
	if !isMonthStart(input.FirstDueMonth) {
		return InstallmentPlan{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"first_due_month": "invalid"})
	}

	accountType, err := s.repo.GetAccountType(ctx, ledgerID, cardAccountID)
	if err != nil || accountType != "credit_card" {
		return InstallmentPlan{}, ErrCardNotFound
	}

	direction, relevant, err := s.repo.GetCategoryBudgetInfo(ctx, ledgerID, input.CategoryID)
	if err != nil {
		return InstallmentPlan{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"category_id": "invalid"})
	}
	if direction != "out" || !relevant {
		return InstallmentPlan{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"category_id": "not_eligible"})
	}

	plan := InstallmentPlan{
		LedgerID:               ledgerID,
		CardAccountID:          cardAccountID,
		PurchaseOccurredAt:     input.PurchaseOccurredAt,
		Merchant:               input.Merchant,
		Description:            input.Description,
		CategoryID:             input.CategoryID,
		TotalAmountCents:       input.TotalAmountCents,
		InstallmentsCount:      installmentsCount,
		InstallmentAmountCents: installmentAmount,
		FirstDueMonth:          input.FirstDueMonth,
		Status:                 "active",
		CreatedByUserID:        userID,
	}

	installments, err := buildInstallments(plan)
	if err != nil {
		return InstallmentPlan{}, err
	}

	return s.repo.CreatePlanWithInstallments(ctx, plan, installments)
}

func (s *Service) ListPlans(ctx context.Context, userID, ledgerID, cardAccountID string, status *string) ([]InstallmentPlan, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "viewer"); err != nil {
		return nil, err
	}
	return s.repo.ListPlans(ctx, ledgerID, cardAccountID, status)
}

func (s *Service) GetPlan(ctx context.Context, userID, ledgerID, cardAccountID, planID string) (InstallmentPlan, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "viewer"); err != nil {
		return InstallmentPlan{}, err
	}
	plan, err := s.repo.GetPlan(ctx, ledgerID, cardAccountID, planID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return InstallmentPlan{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"plan_id": "not_found"})
		}
		return InstallmentPlan{}, err
	}
	return plan, nil
}

func (s *Service) CancelPlan(ctx context.Context, userID, ledgerID, cardAccountID, planID string) error {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return err
	}
	_, err := s.repo.GetPlan(ctx, ledgerID, cardAccountID, planID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"plan_id": "not_found"})
		}
		return err
	}
	hasPosted, err := s.repo.HasPostedInstallments(ctx, ledgerID, planID)
	if err != nil {
		return err
	}
	if hasPosted {
		return ErrTransactionReferenced
	}
	return s.repo.UpdatePlanStatus(ctx, ledgerID, planID, "cancelled", s.now())
}

func (s *Service) ListInstallments(ctx context.Context, userID, ledgerID, cardAccountID string, month *time.Time, status *string) ([]Installment, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "viewer"); err != nil {
		return nil, err
	}
	return s.repo.ListInstallments(ctx, ledgerID, cardAccountID, month, status)
}

func (s *Service) UpdateInstallment(ctx context.Context, userID, ledgerID, installmentID, status string) (Installment, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return Installment{}, err
	}
	if status != "skipped" {
		return Installment{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"status": "invalid"})
	}
	return s.repo.UpdateInstallmentStatus(ctx, ledgerID, installmentID, status, s.now())
}

func (s *Service) PostMonth(ctx context.Context, userID, ledgerID, cardAccountID string, month time.Time) (PostingResult, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return PostingResult{}, err
	}
	if !isMonthStart(month) {
		return PostingResult{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"month": "invalid"})
	}

	installments, err := s.repo.ListInstallmentsForPosting(ctx, ledgerID, cardAccountID, month)
	if err != nil {
		return PostingResult{}, err
	}
	if len(installments) == 0 {
		return PostingResult{PostedCount: 0}, nil
	}

	transactionIDs := []string{}
	for _, inst := range installments {
		categoryID, err := s.repo.GetPlanCategory(ctx, ledgerID, inst.PlanID)
		if err != nil {
			return PostingResult{}, err
		}
		entry := journal.EntryInput{
			AccountID:   cardAccountID,
			CategoryID:  &categoryID,
			Kind:        "normal",
			AmountCents: inst.AmountCents,
		}
		created, err := s.repo.CreateTransaction(ctx, journal.CreateTransactionParams{
			LedgerID:        ledgerID,
			OccurredAt:      month,
			Description:     "Parcela",
			CreatedByUserID: userID,
			Entries:         []journal.EntryInput{entry},
		})
		if err != nil {
			return PostingResult{}, err
		}

		if err := s.repo.MarkInstallmentPosted(ctx, ledgerID, inst.ID, created.ID, s.now()); err != nil {
			return PostingResult{}, err
		}
		transactionIDs = append(transactionIDs, created.ID)
	}

	return PostingResult{PostedCount: len(installments), Transactions: transactionIDs}, nil
}

func (s *Service) ListStatements(ctx context.Context, userID, ledgerID, cardAccountID string, month *time.Time) ([]Statement, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "viewer"); err != nil {
		return nil, err
	}
	return s.repo.ListStatements(ctx, ledgerID, cardAccountID, month)
}

func (s *Service) GetStatement(ctx context.Context, userID, ledgerID, cardAccountID, statementID string) (Statement, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "viewer"); err != nil {
		return Statement{}, err
	}
	statement, err := s.repo.GetStatement(ctx, ledgerID, cardAccountID, statementID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Statement{}, ErrStatementNotFound
		}
		return Statement{}, err
	}
	return statement, nil
}

func (s *Service) CloseStatement(ctx context.Context, userID, ledgerID, cardAccountID string, month time.Time) (Statement, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return Statement{}, err
	}
	if !isMonthStart(month) {
		return Statement{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"month": "invalid"})
	}

	card, err := s.repo.GetCreditCard(ctx, ledgerID, cardAccountID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Statement{}, ErrCardNotFound
		}
		return Statement{}, err
	}

	statement, err := s.repo.GetStatementByMonth(ctx, ledgerID, cardAccountID, month)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			statement = Statement{
				LedgerID:       ledgerID,
				CardAccountID:  cardAccountID,
				StatementMonth: month,
				ClosingDate:    adjustDate(month, card.ClosingDay),
				DueDate:        adjustDate(month, card.DueDay),
				Status:         "open",
			}
			statement, err = s.repo.CreateStatement(ctx, statement)
		}
	}
	if err != nil {
		return Statement{}, err
	}

	charges, err := s.repo.SumStatementCharges(ctx, ledgerID, cardAccountID, month)
	if err != nil {
		return Statement{}, err
	}

	updated, err := s.repo.UpdateStatementTotals(ctx, statement.LedgerID, statement.CardAccountID, statement.ID, charges, statement.TotalPaymentsCents, "closed", s.now())
	if err != nil {
		return Statement{}, err
	}
	return updated, nil
}

func (s *Service) PayStatement(ctx context.Context, userID, ledgerID, cardAccountID, statementID, cashAccountID string, payAmount int64, paymentDate time.Time) (Statement, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return Statement{}, err
	}
	if payAmount <= 0 {
		return Statement{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"pay_amount_cents": "invalid"})
	}

	statement, err := s.repo.GetStatement(ctx, ledgerID, cardAccountID, statementID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Statement{}, ErrStatementNotFound
		}
		return Statement{}, err
	}
	if statement.Status == "paid" {
		return Statement{}, ErrStatementAlreadyPaid
	}

	if payAmount+statement.TotalPaymentsCents > statement.TotalChargesCents {
		return Statement{}, ErrPaymentExceedsTotal
	}

	accountType, err := s.repo.GetAccountType(ctx, ledgerID, cashAccountID)
	if err != nil || accountType != "cash" {
		return Statement{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"cash_account_id": "invalid"})
	}

	outCategory, inCategory, err := s.loadPaymentCategories(ctx, ledgerID)
	if err != nil {
		return Statement{}, err
	}

	entries := []journal.EntryInput{
		{AccountID: cashAccountID, CategoryID: &outCategory, Kind: "transfer", AmountCents: payAmount},
		{AccountID: cardAccountID, CategoryID: &inCategory, Kind: "transfer", AmountCents: payAmount},
	}

	paymentTx, err := s.repo.CreateTransaction(ctx, journal.CreateTransactionParams{
		LedgerID:        ledgerID,
		OccurredAt:      paymentDate,
		Description:     "Pagamento fatura",
		CreatedByUserID: userID,
		Entries:         entries,
	})
	if err != nil {
		return Statement{}, err
	}

	newTotalPayments := statement.TotalPaymentsCents + payAmount
	status := statement.Status
	if newTotalPayments >= statement.TotalChargesCents {
		status = "paid"
	}

	updated, err := s.repo.SetStatementPayment(ctx, statement.LedgerID, statement.CardAccountID, statement.ID, paymentTx.ID, newTotalPayments, status, s.now())
	if err != nil {
		return Statement{}, err
	}
	if status == "paid" {
		if err := s.repo.MarkInstallmentsPaid(ctx, ledgerID, cardAccountID, statement.StatementMonth, statement.ID, s.now()); err != nil {
			return Statement{}, err
		}
	}

	return updated, nil
}

func (s *Service) loadPaymentCategories(ctx context.Context, ledgerID string) (string, string, error) {
	outID, err := s.repo.FindCategoryByName(ctx, ledgerID, "Pagamento Fatura Cartao")
	if err != nil {
		return "", "", NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"category": "Pagamento Fatura Cartao"})
	}
	inID, err := s.repo.FindCategoryByName(ctx, ledgerID, "Entrada Pagamento Cartao")
	if err != nil {
		return "", "", NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"category": "Entrada Pagamento Cartao"})
	}
	return outID, inID, nil
}

func (s *Service) requireRole(ctx context.Context, ledgerID, userID, minRole string) error {
	role, err := s.repo.GetLedgerRole(ctx, ledgerID, userID)
	if err != nil {
		return s.mapAccessError(ctx, ledgerID, err)
	}
	if roleRank(role) < roleRank(minRole) {
		return ErrAccessDenied
	}
	return nil
}

func (s *Service) mapAccessError(ctx context.Context, ledgerID string, err error) error {
	if errors.Is(err, ErrNotFound) || errors.Is(err, ledger.ErrNotFound) {
		exists, checkErr := s.repo.LedgerExists(ctx, ledgerID)
		if checkErr == nil && exists {
			return ErrAccessDenied
		}
		return ErrLedgerNotFound
	}
	return err
}

func roleRank(role string) int {
	switch role {
	case "owner":
		return 3
	case "editor":
		return 2
	case "viewer":
		return 1
	default:
		return 0
	}
}

func buildInstallments(plan InstallmentPlan) ([]Installment, error) {
	if plan.InstallmentsCount < 1 {
		return nil, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"installments_count": "invalid"})
	}

	installments := make([]Installment, 0, plan.InstallmentsCount)
	remaining := plan.TotalAmountCents
	for i := 1; i <= plan.InstallmentsCount; i++ {
		amount := plan.InstallmentAmountCents
		if i == plan.InstallmentsCount {
			amount = remaining
		}
		remaining -= amount
		installments = append(installments, Installment{
			LedgerID:      plan.LedgerID,
			PlanID:        plan.ID,
			InstallmentNo: i,
			DueMonth:      plan.FirstDueMonth.AddDate(0, i-1, 0),
			AmountCents:   amount,
			Status:        "scheduled",
		})
	}
	if remaining != 0 {
		return nil, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"total_amount_cents": "mismatch"})
	}
	return installments, nil
}

func isMonthStart(value time.Time) bool {
	return value.Day() == 1 && value.Equal(time.Date(value.Year(), value.Month(), 1, 0, 0, 0, 0, value.Location()))
}

func adjustDate(month time.Time, day int) time.Time {
	if day < 1 {
		day = 1
	}
	year := month.Year()
	mon := month.Month()
	last := time.Date(year, mon+1, 0, 0, 0, 0, 0, month.Location()).Day()
	if day > last {
		day = last
	}
	return time.Date(year, mon, day, 0, 0, 0, 0, month.Location())
}
