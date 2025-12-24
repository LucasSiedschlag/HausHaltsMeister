package picuinha

import (
	"context"
	"time"
)

type PicuinhaService struct {
	repo Repository
}

func NewService(repo Repository) *PicuinhaService {
	return &PicuinhaService{repo: repo}
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

	// Enrich with balance (N+1 problem usually, but for a single user/few people it is fine)
	for i := range people {
		bal, err := s.repo.GetBalance(ctx, people[i].ID)
		if err == nil {
			people[i].Balance = bal
		}
	}
	return people, nil
}

func (s *PicuinhaService) AddDiff(ctx context.Context, personID int32, amount float64, kind string) (*Entry, error) {
	if amount <= 0 {
		return nil, ErrAmountRequired
	}
	if kind != "PLUS" && kind != "MINUS" {
		return nil, ErrInvalidKind // Using string convention for now
	}

	entry := &Entry{
		PersonID: personID,
		Date:     time.Now(),
		Kind:     kind,
		Amount:   amount,
	}
	return s.repo.AddEntry(ctx, entry)
}

func (s *PicuinhaService) Lend(ctx context.Context, personID int32, amount float64, cashFlowID *int32) (*Entry, error) {
	// Lend -> Eu emprestei -> Pessoa me deve -> PLUS
	entry := &Entry{
		PersonID:   personID,
		Date:       time.Now(),
		Kind:       "PLUS",
		Amount:     amount,
		CashFlowID: cashFlowID,
	}
	return s.repo.AddEntry(ctx, entry)
}

func (s *PicuinhaService) Receive(ctx context.Context, personID int32, amount float64, cashFlowID *int32) (*Entry, error) {
	// Receive -> Pessoa me pagou -> Dívida diminui -> MINUS
	entry := &Entry{
		PersonID:   personID,
		Date:       time.Now(),
		Kind:       "MINUS",
		Amount:     amount,
		CashFlowID: cashFlowID,
	}
	return s.repo.AddEntry(ctx, entry)
}
