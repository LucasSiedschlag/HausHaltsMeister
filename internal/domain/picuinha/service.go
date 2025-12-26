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
		return nil, ErrEntryNotFound
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
