package investments

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/journal"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/ledger"
)

type Repository interface {
	GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error)
	LedgerExists(ctx context.Context, ledgerID string) (bool, error)
	FindAccountByType(ctx context.Context, ledgerID, accountType string) (string, error)
	FindCategoryByName(ctx context.Context, ledgerID, name string) (string, error)
	CreateTransaction(ctx context.Context, params journal.CreateTransactionParams) (journal.Transaction, error)
	SumByInvestmentAction(ctx context.Context, ledgerID, action string, from, to time.Time) (int64, error)
}

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now().UTC}
}

func (s *Service) Contribution(ctx context.Context, userID, ledgerID string, amount int64, occurredAt time.Time, memo *string) (journal.Transaction, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return journal.Transaction{}, err
	}
	if amount <= 0 {
		return journal.Transaction{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"amount_cents": "invalid"})
	}
	if occurredAt.IsZero() {
		return journal.Transaction{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"occurred_at": "required"})
	}

	cashAccount, investAccount, cats, err := s.loadDefaults(ctx, ledgerID)
	if err != nil {
		return journal.Transaction{}, err
	}

	entries := []journal.EntryInput{
		{AccountID: cashAccount, CategoryID: &cats.InvestmentsOut, Kind: "transfer", AmountCents: amount, Memo: memo},
		{AccountID: investAccount, CategoryID: &cats.InvestmentsIn, Kind: "transfer", AmountCents: amount, Memo: memo},
	}

	return s.repo.CreateTransaction(ctx, journal.CreateTransactionParams{
		LedgerID:         ledgerID,
		OccurredAt:       occurredAt,
		Description:      "Aporte investimentos",
		Notes:            memo,
		CreatedByUserID:  userID,
		InvestmentAction: strPtr("contribution"),
		Entries:          entries,
	})
}

func (s *Service) Redemption(ctx context.Context, userID, ledgerID string, amount int64, occurredAt time.Time, memo *string) (journal.Transaction, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return journal.Transaction{}, err
	}
	if amount <= 0 {
		return journal.Transaction{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"amount_cents": "invalid"})
	}
	if occurredAt.IsZero() {
		return journal.Transaction{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"occurred_at": "required"})
	}

	cashAccount, investAccount, cats, err := s.loadDefaults(ctx, ledgerID)
	if err != nil {
		return journal.Transaction{}, err
	}

	entries := []journal.EntryInput{
		{AccountID: investAccount, CategoryID: &cats.InvestmentsOut, Kind: "transfer", AmountCents: amount, Memo: memo},
		{AccountID: cashAccount, CategoryID: &cats.InvestmentsIn, Kind: "transfer", AmountCents: amount, Memo: memo},
	}

	return s.repo.CreateTransaction(ctx, journal.CreateTransactionParams{
		LedgerID:         ledgerID,
		OccurredAt:       occurredAt,
		Description:      "Resgate investimentos",
		Notes:            memo,
		CreatedByUserID:  userID,
		InvestmentAction: strPtr("redemption"),
		Entries:          entries,
	})
}

func (s *Service) Earnings(ctx context.Context, userID, ledgerID string, amount int64, occurredAt time.Time, memo *string) (journal.Transaction, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return journal.Transaction{}, err
	}
	if amount <= 0 {
		return journal.Transaction{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"amount_cents": "invalid"})
	}
	if occurredAt.IsZero() {
		return journal.Transaction{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"occurred_at": "required"})
	}

	_, investAccount, cats, err := s.loadDefaults(ctx, ledgerID)
	if err != nil {
		return journal.Transaction{}, err
	}

	entries := []journal.EntryInput{
		{AccountID: investAccount, CategoryID: &cats.InvestmentsIn, Kind: "adjust", AmountCents: amount, Memo: memo},
	}

	return s.repo.CreateTransaction(ctx, journal.CreateTransactionParams{
		LedgerID:         ledgerID,
		OccurredAt:       occurredAt,
		Description:      "Rendimento investimentos",
		Notes:            memo,
		CreatedByUserID:  userID,
		InvestmentAction: strPtr("earnings"),
		Entries:          entries,
	})
}

func (s *Service) Summary(ctx context.Context, userID, ledgerID string, from, to time.Time) (Summary, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "viewer"); err != nil {
		return Summary{}, err
	}
	if from.IsZero() || to.IsZero() {
		return Summary{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"range": "required"})
	}
	if to.Before(from) {
		return Summary{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"range": "invalid"})
	}

	_, _, _, err := s.loadDefaults(ctx, ledgerID)
	if err != nil {
		return Summary{}, err
	}

	contrib, err := s.repo.SumByInvestmentAction(ctx, ledgerID, "contribution", from, to)
	if err != nil {
		return Summary{}, err
	}
	redempt, err := s.repo.SumByInvestmentAction(ctx, ledgerID, "redemption", from, to)
	if err != nil {
		return Summary{}, err
	}
	earnings, err := s.repo.SumByInvestmentAction(ctx, ledgerID, "earnings", from, to)
	if err != nil {
		return Summary{}, err
	}

	losses, err := s.repo.SumByInvestmentAction(ctx, ledgerID, "loss", from, to)
	if err != nil {
		return Summary{}, err
	}

	net := (earnings - losses) + (contrib - redempt)
	return Summary{
		From:               from,
		To:                 to,
		TotalContributions: contrib,
		TotalRedemptions:   redempt,
		TotalEarnings:      earnings,
		TotalLosses:        losses,
		NetVariation:       net,
	}, nil
}

func (s *Service) loadDefaults(ctx context.Context, ledgerID string) (string, string, FlowCategoryIDs, error) {
	cashAccount, err := s.repo.FindAccountByType(ctx, ledgerID, "wallet")
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			cashAccount, err = s.repo.FindAccountByType(ctx, ledgerID, "current")
		}
		if err != nil {
			return "", "", FlowCategoryIDs{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"account": "wallet"})
		}
	}
	investAccount, err := s.repo.FindAccountByType(ctx, ledgerID, "investment")
	if err != nil {
		return "", "", FlowCategoryIDs{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"account": "investment"})
	}

	inCategoryID, err := s.repo.FindCategoryByName(ctx, ledgerID, "Investimentos (Entrada)")
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", "", FlowCategoryIDs{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"category": "Investimentos (Entrada)"})
		}
		return "", "", FlowCategoryIDs{}, err
	}

	outCategoryID, err := s.repo.FindCategoryByName(ctx, ledgerID, "Investimentos (Saída)")
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", "", FlowCategoryIDs{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"category": "Investimentos (Saída)"})
		}
		return "", "", FlowCategoryIDs{}, err
	}

	return cashAccount, investAccount, FlowCategoryIDs{
		InvestmentsIn:  inCategoryID,
		InvestmentsOut: outCategoryID,
	}, nil
}

func strPtr(value string) *string {
	return &value
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

func normalizeName(name string) string {
	return strings.TrimSpace(name)
}
