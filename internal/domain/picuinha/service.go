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

func (s *PicuinhaService) AddDiff(ctx context.Context, personID int32, amount float64, kind string, cashFlowID *int32, autoCreateFlow bool) (*Entry, error) {
	if amount <= 0 {
		return nil, ErrAmountRequired
	}
	if kind != "PLUS" && kind != "MINUS" {
		return nil, ErrInvalidKind
	}

	var linkedCfID *int32 = cashFlowID

	// Automatic CashFlow creation if requested and not already linked
	if autoCreateFlow && linkedCfID == nil {
		// Find "Picuinhas" category for the direction
		cfDirection := "OUT" // PLUS (Eu emprestei) -> Dinheiro saindo
		if kind == "MINUS" {
			cfDirection = "IN" // MINUS (Me pagaram) -> Dinheiro entrando
		}

		cats, err := s.catRepo.List(ctx, true) // Active only
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
		linkedCfID = &cf.ID
	}

	entry := &Entry{
		PersonID:   personID,
		Date:       time.Now(),
		Kind:       kind,
		Amount:     amount,
		CashFlowID: linkedCfID,
	}
	return s.repo.AddEntry(ctx, entry)
}

func (s *PicuinhaService) Lend(ctx context.Context, personID int32, amount float64, cashFlowID *int32) (*Entry, error) {
	// Lend -> Eu emprestei -> Pessoa me deve -> PLUS. Auto-create if not linked.
	return s.AddDiff(ctx, personID, amount, "PLUS", cashFlowID, true)
}

func (s *PicuinhaService) Receive(ctx context.Context, personID int32, amount float64, cashFlowID *int32) (*Entry, error) {
	// Receive -> Pessoa me pagou -> Dívida diminui -> MINUS. Auto-create if not linked.
	return s.AddDiff(ctx, personID, amount, "MINUS", cashFlowID, true)
}
