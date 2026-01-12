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

	ListCreditCards(ctx context.Context, ledgerID, parentAccountID string) ([]CreditCard, error)
	GetCreditCard(ctx context.Context, cardID string) (CreditCard, error)
	CreateCreditCard(ctx context.Context, params CreditCard) (CreditCard, error)
	UpdateCreditCard(ctx context.Context, params CreditCard, updatedAt time.Time) (CreditCard, error)
	DeleteCreditCard(ctx context.Context, ledgerID, cardID string) error
	GetAccount(ctx context.Context, userID, accountID string) (AccountInfo, error)
	CreateAccount(ctx context.Context, ledgerID, name, accountType, nature string, isActive bool) (AccountInfo, error)
	GetAccountNature(ctx context.Context, ledgerID, accountID string) (string, error)
	GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error)
	LedgerExists(ctx context.Context, ledgerID string) (bool, error)
	GetCategoryBudgetInfo(ctx context.Context, ledgerID, categoryID string) (string, bool, error)
	FindCategoryByName(ctx context.Context, ledgerID, name string) (string, error)
	GetCardNetwork(ctx context.Context, code string) (CardNetwork, error)

	CreatePlanWithInstallments(ctx context.Context, plan InstallmentPlan, installments []Installment) (InstallmentPlan, error)
	ListPlans(ctx context.Context, ledgerID, cardID string, status *string) ([]InstallmentPlan, error)
	GetPlan(ctx context.Context, ledgerID, cardID, planID string) (InstallmentPlan, error)
	UpdatePlanStatus(ctx context.Context, ledgerID, planID, status string, updatedAt time.Time) error
	HasPostedInstallments(ctx context.Context, ledgerID, planID string) (bool, error)

	ListInstallments(ctx context.Context, ledgerID, cardID string, month *time.Time, status *string) ([]Installment, error)
	UpdateInstallmentStatus(ctx context.Context, ledgerID, installmentID, status string, updatedAt time.Time) (Installment, error)
	ListInstallmentsForPosting(ctx context.Context, ledgerID, cardID string, month time.Time) ([]Installment, error)
	GetPlanCategory(ctx context.Context, ledgerID, planID string) (string, error)
	MarkInstallmentPosted(ctx context.Context, ledgerID, installmentID, transactionID string, updatedAt time.Time) error

	GetStatement(ctx context.Context, ledgerID, cardID, statementID string) (Statement, error)
	GetStatementByMonth(ctx context.Context, ledgerID, cardID string, month time.Time) (Statement, error)
	CreateStatement(ctx context.Context, statement Statement) (Statement, error)
	UpdateStatementTotals(ctx context.Context, ledgerID, cardID, statementID string, totalCharges, totalPayments int64, status string, updatedAt time.Time) (Statement, error)
	ListStatements(ctx context.Context, ledgerID, cardID string, month *time.Time) ([]Statement, error)
	SetStatementPayment(ctx context.Context, ledgerID, cardID, statementID, paymentTransactionID string, totalPayments int64, status string, updatedAt time.Time) (Statement, error)
	MarkInstallmentsPaid(ctx context.Context, ledgerID, cardID string, month time.Time, statementID string, updatedAt time.Time) error
	SumStatementCharges(ctx context.Context, ledgerID, cardID string, month time.Time) (int64, error)

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

func (s *Service) ListCreditCards(ctx context.Context, userID, accountID string) ([]CreditCard, error) {
	account, err := s.requireAccountAccess(ctx, userID, accountID, "viewer")
	if err != nil {
		return nil, err
	}
	if account.Nature == "liability" {
		return nil, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"account_id": "invalid"})
	}
	return s.repo.ListCreditCards(ctx, account.LedgerID, account.ID)
}

func (s *Service) GetCreditCard(ctx context.Context, userID, cardID string) (CreditCard, error) {
	card, err := s.requireCardAccess(ctx, userID, cardID, "viewer")
	if err != nil {
		return CreditCard{}, err
	}
	return card, nil
}

func (s *Service) CreateCreditCard(ctx context.Context, userID, accountID string, card CreditCard) (CreditCard, error) {
	account, err := s.requireAccountAccess(ctx, userID, accountID, "editor")
	if err != nil {
		return CreditCard{}, err
	}
	if account.Nature == "liability" {
		return CreditCard{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"account_id": "invalid"})
	}

	card.ParentAccountID = account.ID
	card.LedgerID = account.LedgerID

	if card.Brand == "" {
		card.Brand = "other"
	}
	if _, err := s.repo.GetCardNetwork(ctx, card.Brand); err != nil {
		return CreditCard{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"brand": "invalid"})
	}
	if card.ClosingDay < 1 || card.DueDay < 1 {
		return CreditCard{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"closing_day": "invalid"})
	}

	liabilityName := buildLiabilityAccountName(card)
	liabilityAccount, err := s.repo.CreateAccount(ctx, card.LedgerID, liabilityName, "current", "liability", true)
	if err != nil {
		return CreditCard{}, err
	}
	card.LiabilityAccountID = liabilityAccount.ID

	created, err := s.repo.CreateCreditCard(ctx, card)
	if err != nil {
		return CreditCard{}, err
	}
	return created, nil
}

func (s *Service) UpdateCreditCard(ctx context.Context, userID, cardID string, card CreditCard) (CreditCard, error) {
	current, err := s.requireCardAccess(ctx, userID, cardID, "editor")
	if err != nil {
		return CreditCard{}, err
	}
	card.ID = cardID
	card.LedgerID = current.LedgerID
	card.ParentAccountID = current.ParentAccountID
	card.LiabilityAccountID = current.LiabilityAccountID

	if card.Brand == "" {
		card.Brand = "other"
	}
	if _, err := s.repo.GetCardNetwork(ctx, card.Brand); err != nil {
		return CreditCard{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"brand": "invalid"})
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

func (s *Service) DeleteCreditCard(ctx context.Context, userID, cardID string) error {
	card, err := s.requireCardAccess(ctx, userID, cardID, "editor")
	if err != nil {
		return err
	}
	if err := s.repo.DeleteCreditCard(ctx, card.LedgerID, card.ID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrCardNotFound
		}
		return err
	}
	return nil
}

func (s *Service) CreatePlan(ctx context.Context, userID, cardID string, input InstallmentPlan, installmentsCount int, installmentAmount int64) (InstallmentPlan, error) {
	card, err := s.requireCardAccess(ctx, userID, cardID, "editor")
	if err != nil {
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

	direction, relevant, err := s.repo.GetCategoryBudgetInfo(ctx, card.LedgerID, input.CategoryID)
	if err != nil {
		return InstallmentPlan{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"category_id": "invalid"})
	}
	if direction != "out" || !relevant {
		return InstallmentPlan{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"category_id": "not_eligible"})
	}

	plan := InstallmentPlan{
		LedgerID:               card.LedgerID,
		CreditCardID:           card.ID,
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

func (s *Service) ListPlans(ctx context.Context, userID, cardID string, status *string) ([]InstallmentPlan, error) {
	card, err := s.requireCardAccess(ctx, userID, cardID, "viewer")
	if err != nil {
		return nil, err
	}
	return s.repo.ListPlans(ctx, card.LedgerID, card.ID, status)
}

func (s *Service) GetPlan(ctx context.Context, userID, cardID, planID string) (InstallmentPlan, error) {
	card, err := s.requireCardAccess(ctx, userID, cardID, "viewer")
	if err != nil {
		return InstallmentPlan{}, err
	}
	plan, err := s.repo.GetPlan(ctx, card.LedgerID, card.ID, planID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return InstallmentPlan{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"plan_id": "not_found"})
		}
		return InstallmentPlan{}, err
	}
	return plan, nil
}

func (s *Service) CancelPlan(ctx context.Context, userID, cardID, planID string) error {
	card, err := s.requireCardAccess(ctx, userID, cardID, "editor")
	if err != nil {
		return err
	}
	_, err = s.repo.GetPlan(ctx, card.LedgerID, card.ID, planID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"plan_id": "not_found"})
		}
		return err
	}
	hasPosted, err := s.repo.HasPostedInstallments(ctx, card.LedgerID, planID)
	if err != nil {
		return err
	}
	if hasPosted {
		return ErrTransactionReferenced
	}
	return s.repo.UpdatePlanStatus(ctx, card.LedgerID, planID, "cancelled", s.now())
}

func (s *Service) ListInstallments(ctx context.Context, userID, cardID string, month *time.Time, status *string) ([]Installment, error) {
	card, err := s.requireCardAccess(ctx, userID, cardID, "viewer")
	if err != nil {
		return nil, err
	}
	return s.repo.ListInstallments(ctx, card.LedgerID, card.ID, month, status)
}

func (s *Service) UpdateInstallment(ctx context.Context, userID, cardID, installmentID, status string) (Installment, error) {
	card, err := s.requireCardAccess(ctx, userID, cardID, "editor")
	if err != nil {
		return Installment{}, err
	}
	if status != "skipped" {
		return Installment{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"status": "invalid"})
	}
	return s.repo.UpdateInstallmentStatus(ctx, card.LedgerID, installmentID, status, s.now())
}

func (s *Service) PostMonth(ctx context.Context, userID, cardID string, month time.Time) (PostingResult, error) {
	card, err := s.requireCardAccess(ctx, userID, cardID, "editor")
	if err != nil {
		return PostingResult{}, err
	}
	if !isMonthStart(month) {
		return PostingResult{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"month": "invalid"})
	}

	installments, err := s.repo.ListInstallmentsForPosting(ctx, card.LedgerID, card.ID, month)
	if err != nil {
		return PostingResult{}, err
	}
	if len(installments) == 0 {
		return PostingResult{PostedCount: 0}, nil
	}

	transactionIDs := []string{}
	for _, inst := range installments {
		categoryID, err := s.repo.GetPlanCategory(ctx, card.LedgerID, inst.PlanID)
		if err != nil {
			return PostingResult{}, err
		}
		entry := journal.EntryInput{
			AccountID:   card.LiabilityAccountID,
			CategoryID:  &categoryID,
			Kind:        "normal",
			AmountCents: inst.AmountCents,
		}
		created, err := s.repo.CreateTransaction(ctx, journal.CreateTransactionParams{
			LedgerID:        card.LedgerID,
			OccurredAt:      month,
			Description:     "Parcela",
			CreatedByUserID: userID,
			CreditCardID:    &card.ID,
			Entries:         []journal.EntryInput{entry},
		})
		if err != nil {
			return PostingResult{}, err
		}

		if err := s.repo.MarkInstallmentPosted(ctx, card.LedgerID, inst.ID, created.ID, s.now()); err != nil {
			return PostingResult{}, err
		}
		transactionIDs = append(transactionIDs, created.ID)
	}

	return PostingResult{PostedCount: len(installments), Transactions: transactionIDs}, nil
}

func (s *Service) ListStatements(ctx context.Context, userID, cardID string, month *time.Time) ([]Statement, error) {
	card, err := s.requireCardAccess(ctx, userID, cardID, "viewer")
	if err != nil {
		return nil, err
	}
	return s.repo.ListStatements(ctx, card.LedgerID, card.ID, month)
}

func (s *Service) GetStatement(ctx context.Context, userID, cardID, statementID string) (Statement, error) {
	card, err := s.requireCardAccess(ctx, userID, cardID, "viewer")
	if err != nil {
		return Statement{}, err
	}
	statement, err := s.repo.GetStatement(ctx, card.LedgerID, card.ID, statementID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Statement{}, ErrStatementNotFound
		}
		return Statement{}, err
	}
	return statement, nil
}

func (s *Service) CloseStatement(ctx context.Context, userID, cardID string, month time.Time) (Statement, error) {
	card, err := s.requireCardAccess(ctx, userID, cardID, "editor")
	if err != nil {
		return Statement{}, err
	}
	if !isMonthStart(month) {
		return Statement{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"month": "invalid"})
	}

	statement, err := s.repo.GetStatementByMonth(ctx, card.LedgerID, card.ID, month)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			statement = Statement{
				LedgerID:       card.LedgerID,
				CreditCardID:   card.ID,
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

	charges, err := s.repo.SumStatementCharges(ctx, card.LedgerID, card.ID, month)
	if err != nil {
		return Statement{}, err
	}

	updated, err := s.repo.UpdateStatementTotals(ctx, statement.LedgerID, statement.CreditCardID, statement.ID, charges, statement.TotalPaymentsCents, "closed", s.now())
	if err != nil {
		return Statement{}, err
	}
	return updated, nil
}

func (s *Service) PayStatement(ctx context.Context, userID, cardID, statementID, payingAccountID string, payAmount int64, paymentDate time.Time) (Statement, error) {
	card, err := s.requireCardAccess(ctx, userID, cardID, "editor")
	if err != nil {
		return Statement{}, err
	}
	if payAmount <= 0 {
		return Statement{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"pay_amount_cents": "invalid"})
	}

	statement, err := s.repo.GetStatement(ctx, card.LedgerID, card.ID, statementID)
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

	nature, err := s.repo.GetAccountNature(ctx, card.LedgerID, payingAccountID)
	if err != nil || nature != "asset" {
		return Statement{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"paying_account_id": "invalid"})
	}

	outCategory, inCategory, err := s.loadPaymentCategories(ctx, card.LedgerID)
	if err != nil {
		return Statement{}, err
	}

	entries := []journal.EntryInput{
		{AccountID: payingAccountID, CategoryID: &outCategory, Kind: "transfer", AmountCents: payAmount},
		{AccountID: card.LiabilityAccountID, CategoryID: &inCategory, Kind: "transfer", AmountCents: payAmount},
	}

	paymentTx, err := s.repo.CreateTransaction(ctx, journal.CreateTransactionParams{
		LedgerID:        card.LedgerID,
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

	updated, err := s.repo.SetStatementPayment(ctx, statement.LedgerID, statement.CreditCardID, statement.ID, paymentTx.ID, newTotalPayments, status, s.now())
	if err != nil {
		return Statement{}, err
	}
	if status == "paid" {
		if err := s.repo.MarkInstallmentsPaid(ctx, card.LedgerID, card.ID, statement.StatementMonth, statement.ID, s.now()); err != nil {
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

func (s *Service) requireAccountAccess(ctx context.Context, userID, accountID, minRole string) (AccountInfo, error) {
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return AccountInfo{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"account_id": "required"})
	}
	account, err := s.repo.GetAccount(ctx, userID, accountID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return AccountInfo{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"account_id": "invalid"})
		}
		return AccountInfo{}, err
	}
	if err := s.requireRole(ctx, account.LedgerID, userID, minRole); err != nil {
		return AccountInfo{}, err
	}
	if !account.IsActive {
		return AccountInfo{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"account_id": "inactive"})
	}
	return account, nil
}

func (s *Service) requireCardAccess(ctx context.Context, userID, cardID, minRole string) (CreditCard, error) {
	cardID = strings.TrimSpace(cardID)
	if cardID == "" {
		return CreditCard{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"card_id": "required"})
	}
	card, err := s.repo.GetCreditCard(ctx, cardID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return CreditCard{}, ErrCardNotFound
		}
		return CreditCard{}, err
	}
	if err := s.requireRole(ctx, card.LedgerID, userID, minRole); err != nil {
		return CreditCard{}, err
	}
	return card, nil
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

func buildLiabilityAccountName(card CreditCard) string {
	label := strings.TrimSpace(ptrString(card.Label))
	last4 := strings.TrimSpace(ptrString(card.Last4))
	brand := strings.TrimSpace(card.Brand)

	parts := []string{"Cartao"}
	if label != "" {
		parts = append(parts, label)
	}
	if last4 != "" {
		parts = append(parts, last4)
	} else if label == "" && brand != "" {
		parts = append(parts, strings.ToUpper(brand))
	}

	return strings.Join(parts, " ")
}

func ptrString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
