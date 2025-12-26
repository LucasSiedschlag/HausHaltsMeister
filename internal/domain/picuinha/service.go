package picuinha

import (
	"context"
	"fmt"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/cashflow"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/category"
)

type PicuinhaService struct {
	repo      Repository
	cfService *cashflow.CashFlowService
	catRepo   category.Repository
}

func NewService(repo Repository, cfService *cashflow.CashFlowService, catRepo category.Repository) *PicuinhaService {
	return &PicuinhaService{
		repo:      repo,
		cfService: cfService,
		catRepo:   catRepo,
	}
}

func (s *PicuinhaService) CreatePerson(ctx context.Context, name, notes string) (*Person, error) {
	if name == "" {
		return nil, ErrPersonNameRequired
	}
	return s.repo.CreatePerson(ctx, name, notes)
}

func (s *PicuinhaService) ListPersons(ctx context.Context) ([]Person, error) {
	people, err := s.repo.ListPersons(ctx)
	if err != nil {
		return nil, err
	}

	// Enrich with balance
	for i := range people {
		bal, err := s.repo.GetBalance(ctx, people[i].ID)
		if err == nil {
			people[i].Balance = bal
		}
	}
	return people, nil
}

func (s *PicuinhaService) UpdatePerson(ctx context.Context, id int32, name, notes string) (*Person, error) {
	if name == "" {
		return nil, ErrPersonNameRequired
	}

	person, err := s.repo.UpdatePerson(ctx, id, name, notes)
	if err != nil {
		return nil, err
	}
	if person == nil {
		return nil, ErrPersonNotFound
	}

	bal, err := s.repo.GetBalance(ctx, id)
	if err == nil {
		person.Balance = bal
	}

	return person, nil
}

func (s *PicuinhaService) DeletePerson(ctx context.Context, id int32) error {
	person, err := s.repo.GetPerson(ctx, id)
	if err != nil {
		return err
	}
	if person == nil {
		return ErrPersonNotFound
	}

	count, err := s.repo.CountEntriesByPerson(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrPersonHasEntries
	}
	caseCount, err := s.repo.CountCasesByPerson(ctx, id)
	if err != nil {
		return err
	}
	if caseCount > 0 {
		return ErrPersonHasEntries
	}

	return s.repo.DeletePerson(ctx, id)
}

func (s *PicuinhaService) AddDiff(ctx context.Context, personID int32, amount float64, kind string, cashFlowID *int32, paymentMethodID *int32, cardOwner string, autoCreateFlow bool) (*Entry, error) {
	if amount <= 0 {
		return nil, ErrAmountRequired
	}
	if kind != "PLUS" && kind != "MINUS" {
		return nil, ErrInvalidKind
	}

	normalizedOwner, err := normalizeCardOwner(cardOwner)
	if err != nil {
		return nil, err
	}

	linkedCfID, err := s.ensureCashFlowID(ctx, personID, amount, kind, cashFlowID, autoCreateFlow, normalizedOwner)
	if err != nil {
		return nil, err
	}

	entry := &Entry{
		PersonID:        personID,
		Date:            time.Now(),
		Kind:            kind,
		Amount:          amount,
		CashFlowID:      linkedCfID,
		PaymentMethodID: paymentMethodID,
		CardOwner:       normalizedOwner,
	}
	return s.repo.AddEntry(ctx, entry)
}

func (s *PicuinhaService) ListEntries(ctx context.Context, personID *int32) ([]Entry, error) {
	if personID != nil {
		person, err := s.repo.GetPerson(ctx, *personID)
		if err != nil {
			return nil, err
		}
		if person == nil {
			return nil, ErrPersonNotFound
		}
		return s.repo.ListEntriesByPerson(ctx, *personID)
	}

	return s.repo.ListEntries(ctx)
}

func (s *PicuinhaService) UpdateEntry(ctx context.Context, id int32, personID int32, amount float64, kind string, paymentMethodID *int32, cardOwner string, autoCreateFlow bool) (*Entry, error) {
	if amount <= 0 {
		return nil, ErrAmountRequired
	}
	if kind != "PLUS" && kind != "MINUS" {
		return nil, ErrInvalidKind
	}

	normalizedOwner, err := normalizeCardOwner(cardOwner)
	if err != nil {
		return nil, err
	}

	existing, err := s.repo.GetEntry(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrCaseNotFound
	}

	person, err := s.repo.GetPerson(ctx, personID)
	if err != nil {
		return nil, err
	}
	if person == nil {
		return nil, ErrPersonNotFound
	}

	linkedCfID, err := s.ensureCashFlowID(ctx, personID, amount, kind, existing.CashFlowID, autoCreateFlow, normalizedOwner)
	if err != nil {
		return nil, err
	}

	updated := &Entry{
		ID:              existing.ID,
		PersonID:        personID,
		Date:            existing.Date,
		Kind:            kind,
		Amount:          amount,
		CashFlowID:      linkedCfID,
		PaymentMethodID: paymentMethodID,
		CardOwner:       normalizedOwner,
	}

	saved, err := s.repo.UpdateEntry(ctx, updated)
	if err != nil {
		return nil, err
	}
	if saved == nil {
		return nil, ErrEntryNotFound
	}
	return saved, nil
}

func (s *PicuinhaService) DeleteEntry(ctx context.Context, id int32) error {
	entry, err := s.repo.GetEntry(ctx, id)
	if err != nil {
		return err
	}
	if entry == nil {
		return ErrEntryNotFound
	}

	return s.repo.DeleteEntry(ctx, id)
}

func (s *PicuinhaService) CreateCase(ctx context.Context, req CreateCaseRequest) (*CaseSummary, error) {
	if req.Title == "" {
		return nil, ErrCaseTitleRequired
	}
	if !isValidCaseType(req.CaseType) {
		return nil, ErrCaseTypeInvalid
	}
	if req.StartDate.IsZero() {
		return nil, ErrStartDateRequired
	}
	if req.CaseType == CaseTypeCardInstall && req.PaymentMethodID == nil {
		return nil, ErrPaymentMethodRequired
	}
	if req.InterestRate != nil && req.InterestRateUnit != "" && req.InterestRateUnit != InterestRateMonthly && req.InterestRateUnit != InterestRateAnnual {
		return nil, ErrInterestRateUnit
	}
	if req.CaseType == CaseTypeRecurring && req.RecurrenceIntervalMonths != nil && *req.RecurrenceIntervalMonths <= 0 {
		return nil, ErrRecurrenceInterval
	}
	if req.StartDate.IsZero() {
		return nil, ErrStartDateRequired
	}
	if req.CaseType == CaseTypeCardInstall && req.PaymentMethodID == nil {
		return nil, ErrPaymentMethodRequired
	}
	if req.InterestRate != nil && req.InterestRateUnit != "" && req.InterestRateUnit != InterestRateMonthly && req.InterestRateUnit != InterestRateAnnual {
		return nil, ErrInterestRateUnit
	}
	if req.CaseType == CaseTypeRecurring && req.RecurrenceIntervalMonths != nil && *req.RecurrenceIntervalMonths <= 0 {
		return nil, ErrRecurrenceInterval
	}

	person, err := s.repo.GetPerson(ctx, req.PersonID)
	if err != nil {
		return nil, err
	}
	if person == nil {
		return nil, ErrPersonNotFound
	}

	totalAmount, installmentAmount, installmentCount, recurrenceInterval, err := normalizeCaseAmounts(req)
	if err != nil {
		return nil, err
	}

	picCase := &Case{
		PersonID:                 req.PersonID,
		Title:                    req.Title,
		CaseType:                 req.CaseType,
		TotalAmount:              totalAmount,
		InstallmentCount:         installmentCount,
		InstallmentAmount:        installmentAmount,
		StartDate:                req.StartDate,
		PaymentMethodID:          req.PaymentMethodID,
		InstallmentPlanID:        req.InstallmentPlanID,
		CategoryID:               req.CategoryID,
		InterestRate:             req.InterestRate,
		InterestRateUnit:         req.InterestRateUnit,
		RecurrenceIntervalMonths: recurrenceInterval,
	}

	created, err := s.repo.CreateCase(ctx, picCase)
	if err != nil {
		return nil, err
	}

	count := int32(1)
	if installmentCount != nil {
		count = *installmentCount
	}
	interval := int32(1)
	if recurrenceInterval != nil && *recurrenceInterval > 0 {
		interval = *recurrenceInterval
	}

	amount := float64(0)
	if installmentAmount != nil {
		amount = *installmentAmount
	}

	for i := int32(1); i <= count; i++ {
		dueDate := created.StartDate.AddDate(0, int(interval*(i-1)), 0)
		_, err := s.repo.CreateInstallment(ctx, &CaseInstallment{
			CaseID:            created.ID,
			InstallmentNumber: i,
			DueDate:           dueDate,
			Amount:            amount,
			ExtraAmount:       0,
			IsPaid:            false,
		})
		if err != nil {
			return nil, err
		}
	}

	return &CaseSummary{
		Case:              *created,
		InstallmentsTotal: count,
		InstallmentsPaid:  0,
		AmountPaid:        0,
		AmountRemaining:   float64(count) * amount,
		Status:            caseStatus(req.CaseType, 0, count),
	}, nil
}

func (s *PicuinhaService) ListCasesByPerson(ctx context.Context, personID int32) ([]CaseSummary, error) {
	person, err := s.repo.GetPerson(ctx, personID)
	if err != nil {
		return nil, err
	}
	if person == nil {
		return nil, ErrPersonNotFound
	}
	return s.repo.ListCasesByPerson(ctx, personID)
}

func (s *PicuinhaService) UpdateCase(ctx context.Context, id int32, req CreateCaseRequest) (*CaseSummary, error) {
	if req.Title == "" {
		return nil, ErrCaseTitleRequired
	}
	if !isValidCaseType(req.CaseType) {
		return nil, ErrCaseTypeInvalid
	}

	existing, err := s.repo.GetCase(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrCaseNotFound
	}

	totalAmount, installmentAmount, installmentCount, recurrenceInterval, err := normalizeCaseAmounts(req)
	if err != nil {
		return nil, err
	}

	existing.Title = req.Title
	existing.CaseType = req.CaseType
	existing.TotalAmount = totalAmount
	existing.InstallmentAmount = installmentAmount
	existing.InstallmentCount = installmentCount
	existing.StartDate = req.StartDate
	existing.PaymentMethodID = req.PaymentMethodID
	existing.InstallmentPlanID = req.InstallmentPlanID
	existing.CategoryID = req.CategoryID
	existing.InterestRate = req.InterestRate
	existing.InterestRateUnit = req.InterestRateUnit
	existing.RecurrenceIntervalMonths = recurrenceInterval

	updated, err := s.repo.UpdateCase(ctx, existing)
	if err != nil {
		return nil, err
	}

	summaries, err := s.repo.ListCasesByPerson(ctx, updated.PersonID)
	if err != nil {
		return nil, err
	}
	for _, summary := range summaries {
		if summary.ID == updated.ID {
			return &summary, nil
		}
	}

	return &CaseSummary{Case: *updated}, nil
}

func (s *PicuinhaService) DeleteCase(ctx context.Context, id int32) error {
	existing, err := s.repo.GetCase(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrCaseNotFound
	}
	return s.repo.DeleteCase(ctx, id)
}

func (s *PicuinhaService) ListInstallmentsByCase(ctx context.Context, caseID int32) ([]CaseInstallment, error) {
	existing, err := s.repo.GetCase(ctx, caseID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrCaseNotFound
	}
	return s.repo.ListInstallmentsByCase(ctx, caseID)
}

func (s *PicuinhaService) UpdateInstallment(ctx context.Context, id int32, req UpdateInstallmentRequest) (*CaseInstallment, error) {
	existing, err := s.repo.GetInstallment(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrInstallmentNotFound
	}

	updated := &CaseInstallment{
		ID:                existing.ID,
		CaseID:            existing.CaseID,
		InstallmentNumber: existing.InstallmentNumber,
		DueDate:           existing.DueDate,
		Amount:            existing.Amount,
		ExtraAmount:       req.ExtraAmount,
		IsPaid:            req.IsPaid,
	}
	if req.IsPaid {
		now := time.Now()
		updated.PaidAt = &now
	} else {
		updated.PaidAt = nil
	}

	return s.repo.UpdateInstallment(ctx, updated)
}

func (s *PicuinhaService) ensureCashFlowID(ctx context.Context, personID int32, amount float64, kind string, cashFlowID *int32, autoCreateFlow bool, cardOwner string) (*int32, error) {
	if !autoCreateFlow || cashFlowID != nil {
		return cashFlowID, nil
	}
	if cardOwner == CardOwnerThird {
		return nil, ErrCardOwnerUnsupported
	}
	if s.cfService == nil || s.catRepo == nil {
		return nil, fmt.Errorf("cash flow service unavailable")
	}

	cfDirection := "OUT"
	if kind == "MINUS" {
		cfDirection = "IN"
	}

	cats, err := s.catRepo.List(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}

	var foundID int32
	for _, c := range cats {
		if c.Name == "Picuinhas" && c.Direction == cfDirection {
			foundID = c.ID
			break
		}
	}

	if foundID == 0 {
		return nil, fmt.Errorf("category 'Picuinhas' (%s) not found in system", cfDirection)
	}

	person, _ := s.repo.GetPerson(ctx, personID)
	name := "Dívida"
	if person != nil {
		name = person.Name
	}
	title := fmt.Sprintf("Picuinha: %s", name)
	if kind == "MINUS" {
		title = fmt.Sprintf("Recebimento: %s", name)
	}

	cf, err := s.cfService.CreateCashFlow(ctx, time.Now(), foundID, cfDirection, title, amount, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create automatic cash flow: %w", err)
	}
	return &cf.ID, nil
}

func (s *PicuinhaService) Lend(ctx context.Context, personID int32, amount float64, cashFlowID *int32) (*Entry, error) {
	// Lend -> Eu emprestei -> Pessoa me deve -> PLUS. Auto-create if not linked.
	return s.AddDiff(ctx, personID, amount, "PLUS", cashFlowID, nil, CardOwnerSelf, true)
}

func (s *PicuinhaService) Receive(ctx context.Context, personID int32, amount float64, cashFlowID *int32) (*Entry, error) {
	// Receive -> Pessoa me pagou -> Dívida diminui -> MINUS. Auto-create if not linked.
	return s.AddDiff(ctx, personID, amount, "MINUS", cashFlowID, nil, CardOwnerSelf, true)
}

func normalizeCardOwner(cardOwner string) (string, error) {
	if cardOwner == "" {
		return CardOwnerSelf, nil
	}
	if cardOwner != CardOwnerSelf && cardOwner != CardOwnerThird {
		return "", ErrInvalidCardOwner
	}
	return cardOwner, nil
}

func isValidCaseType(caseType string) bool {
	switch caseType {
	case CaseTypeOneOff, CaseTypeInstallment, CaseTypeRecurring, CaseTypeCardInstall:
		return true
	default:
		return false
	}
}

func normalizeCaseAmounts(req CreateCaseRequest) (*float64, *float64, *int32, *int32, error) {
	switch req.CaseType {
	case CaseTypeOneOff:
		if req.TotalAmount <= 0 && req.InstallmentAmount <= 0 {
			return nil, nil, nil, nil, ErrAmountRequired
		}
		amount := req.TotalAmount
		if amount <= 0 {
			amount = req.InstallmentAmount
		}
		total := amount
		installmentAmount := amount
		installmentCount := int32(1)
		return &total, &installmentAmount, &installmentCount, nil, nil
	case CaseTypeInstallment, CaseTypeCardInstall:
		if req.InstallmentCount <= 0 {
			return nil, nil, nil, nil, ErrInstallmentCount
		}
		if req.TotalAmount <= 0 && req.InstallmentAmount <= 0 {
			return nil, nil, nil, nil, ErrAmountRequired
		}
		total := req.TotalAmount
		installmentAmount := req.InstallmentAmount
		if total <= 0 {
			total = installmentAmount * float64(req.InstallmentCount)
		}
		if installmentAmount <= 0 {
			installmentAmount = total / float64(req.InstallmentCount)
		}
		count := req.InstallmentCount
		return &total, &installmentAmount, &count, nil, nil
	case CaseTypeRecurring:
		if req.InstallmentAmount <= 0 && req.TotalAmount <= 0 {
			return nil, nil, nil, nil, ErrAmountRequired
		}
		installmentAmount := req.InstallmentAmount
		if installmentAmount <= 0 {
			installmentAmount = req.TotalAmount
		}
		count := int32(24)
		interval := int32(1)
		if req.RecurrenceIntervalMonths != nil && *req.RecurrenceIntervalMonths > 0 {
			interval = *req.RecurrenceIntervalMonths
		}
		return nil, &installmentAmount, &count, &interval, nil
	default:
		return nil, nil, nil, nil, ErrCaseTypeInvalid
	}
}

func caseStatus(caseType string, paid, total int32) string {
	if caseType == CaseTypeRecurring {
		return StatusRecurringActive
	}
	if total > 0 && paid >= total {
		return StatusPaid
	}
	return StatusOpen
}
