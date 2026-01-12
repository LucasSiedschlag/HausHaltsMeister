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
	GetCategoryDirection(ctx context.Context, ledgerID, categoryID string) (string, error)
	CreateTransaction(ctx context.Context, params journal.CreateTransactionParams) (journal.Transaction, error)
	SumByCategory(ctx context.Context, ledgerID, categoryID string, from, to time.Time) (int64, error)
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
		{AccountID: cashAccount, CategoryID: &cats.ContributionOut, Kind: "transfer", AmountCents: amount, Memo: memo},
		{AccountID: investAccount, CategoryID: &cats.ContributionIn, Kind: "transfer", AmountCents: amount, Memo: memo},
	}

	return s.repo.CreateTransaction(ctx, journal.CreateTransactionParams{
		LedgerID:        ledgerID,
		OccurredAt:      occurredAt,
		Description:     "Aporte investimentos",
		Notes:           memo,
		CreatedByUserID: userID,
		Entries:         entries,
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
		{AccountID: investAccount, CategoryID: &cats.RedemptionOut, Kind: "transfer", AmountCents: amount, Memo: memo},
		{AccountID: cashAccount, CategoryID: &cats.RedemptionIn, Kind: "transfer", AmountCents: amount, Memo: memo},
	}

	return s.repo.CreateTransaction(ctx, journal.CreateTransactionParams{
		LedgerID:        ledgerID,
		OccurredAt:      occurredAt,
		Description:     "Resgate investimentos",
		Notes:           memo,
		CreatedByUserID: userID,
		Entries:         entries,
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
		{AccountID: investAccount, CategoryID: &cats.EarningsIn, Kind: "adjust", AmountCents: amount, Memo: memo},
	}

	return s.repo.CreateTransaction(ctx, journal.CreateTransactionParams{
		LedgerID:        ledgerID,
		OccurredAt:      occurredAt,
		Description:     "Rendimento investimentos",
		Notes:           memo,
		CreatedByUserID: userID,
		Entries:         entries,
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

	_, _, cats, err := s.loadDefaults(ctx, ledgerID)
	if err != nil {
		return Summary{}, err
	}

	contrib, err := s.repo.SumByCategory(ctx, ledgerID, cats.ContributionOut, from, to)
	if err != nil {
		return Summary{}, err
	}
	redempt, err := s.repo.SumByCategory(ctx, ledgerID, cats.RedemptionOut, from, to)
	if err != nil {
		return Summary{}, err
	}
	earnings, err := s.repo.SumByCategory(ctx, ledgerID, cats.EarningsIn, from, to)
	if err != nil {
		return Summary{}, err
	}

	losses := int64(0)
	if cats.LossesOut != nil {
		losses, err = s.repo.SumByCategory(ctx, ledgerID, *cats.LossesOut, from, to)
		if err != nil {
			return Summary{}, err
		}
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

	categories := map[string]string{
		"Aportes Investimentos":           "",
		"Entrada Investimentos (Aporte)":  "",
		"Resgate Investimentos":           "",
		"Entrada Resgate (Investimentos)": "",
		"Rendimentos":                     "",
		"Perdas":                          "",
	}

	for name := range categories {
		id, err := s.repo.FindCategoryByName(ctx, ledgerID, name)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				if name == "Perdas" {
					continue
				}
				return "", "", FlowCategoryIDs{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"category": name})
			}
			return "", "", FlowCategoryIDs{}, err
		}
		categories[name] = id
	}

	result := FlowCategoryIDs{
		ContributionOut: categories["Aportes Investimentos"],
		ContributionIn:  categories["Entrada Investimentos (Aporte)"],
		RedemptionOut:   categories["Resgate Investimentos"],
		RedemptionIn:    categories["Entrada Resgate (Investimentos)"],
		EarningsIn:      categories["Rendimentos"],
	}
	if categories["Perdas"] != "" {
		lossesID := categories["Perdas"]
		result.LossesOut = &lossesID
	}

	return cashAccount, investAccount, result, nil
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
