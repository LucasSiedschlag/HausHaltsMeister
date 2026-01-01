package journal

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/ledger"
)

type Repository interface {
	CreateTransaction(ctx context.Context, params CreateTransactionParams) (Transaction, error)
	ListTransactions(ctx context.Context, params ListTransactionsParams) (ListResult, error)
	GetTransaction(ctx context.Context, ledgerID, transactionID string) (Transaction, error)
	UpdateTransaction(ctx context.Context, params UpdateTransactionParams) (Transaction, error)
	DeleteTransaction(ctx context.Context, ledgerID, transactionID string) error
	HasTransactionReferences(ctx context.Context, ledgerID, transactionID string) (bool, error)
	GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error)
	LedgerExists(ctx context.Context, ledgerID string) (bool, error)
	GetAccountsByIDs(ctx context.Context, ledgerID string, ids []string) (map[string]bool, error)
	GetCategoriesByIDs(ctx context.Context, ledgerID string, ids []string) (map[string]string, error)
}

type EntryInput struct {
	AccountID   string
	CategoryID  *string
	Kind        string
	AmountCents int64
	Memo        *string
}

type CreateTransactionParams struct {
	LedgerID        string
	OccurredAt      time.Time
	Description     string
	Notes           *string
	CreatedByUserID string
	Entries         []EntryInput
}

type UpdateTransactionParams struct {
	LedgerID      string
	TransactionID string
	OccurredAt    *time.Time
	Description   *string
	Notes         *string
	UpdatedAt     time.Time
}

type ListTransactionsParams struct {
	LedgerID   string
	From       *time.Time
	To         *time.Time
	AccountID  *string
	CategoryID *string
	Query      *string
	Limit      int
	Cursor     *Cursor
}

type Service struct {
	repo         Repository
	defaultLimit int
	maxLimit     int
	now          func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, defaultLimit: 50, maxLimit: 200, now: time.Now().UTC}
}

func (s *Service) CreateTransaction(ctx context.Context, userID, ledgerID string, params CreateTransactionParams) (Transaction, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return Transaction{}, err
	}

	if strings.TrimSpace(params.Description) == "" {
		return Transaction{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"description": "required"})
	}
	if params.OccurredAt.IsZero() {
		return Transaction{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"occurred_at": "required"})
	}
	if len(params.Entries) == 0 {
		return Transaction{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"entries": "required"})
	}

	validatedEntries, err := s.validateEntries(ctx, ledgerID, params.Entries)
	if err != nil {
		return Transaction{}, err
	}

	params.LedgerID = ledgerID
	params.CreatedByUserID = userID
	params.Entries = validatedEntries

	created, err := s.repo.CreateTransaction(ctx, params)
	if err != nil {
		return Transaction{}, err
	}
	return created, nil
}

func (s *Service) ListTransactions(ctx context.Context, userID, ledgerID string, params ListTransactionsParams) (ListResult, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "viewer"); err != nil {
		return ListResult{}, err
	}

	limit := params.Limit
	if limit <= 0 {
		limit = s.defaultLimit
	}
	if limit > s.maxLimit {
		limit = s.maxLimit
	}

	params.LedgerID = ledgerID
	params.Limit = limit

	return s.repo.ListTransactions(ctx, params)
}

func (s *Service) GetTransaction(ctx context.Context, userID, ledgerID, transactionID string) (Transaction, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "viewer"); err != nil {
		return Transaction{}, err
	}
	item, err := s.repo.GetTransaction(ctx, ledgerID, transactionID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Transaction{}, ErrTransactionNotFound
		}
		return Transaction{}, s.mapAccessError(ctx, ledgerID, err)
	}
	return item, nil
}

func (s *Service) UpdateTransaction(ctx context.Context, userID, ledgerID, transactionID string, params UpdateTransactionParams) (Transaction, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return Transaction{}, err
	}
	if params.OccurredAt == nil && params.Description == nil && params.Notes == nil {
		return Transaction{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"fields": "empty"})
	}

	referenced, err := s.repo.HasTransactionReferences(ctx, ledgerID, transactionID)
	if err != nil {
		return Transaction{}, err
	}
	if referenced {
		return Transaction{}, ErrTransactionReferenced
	}

	params.LedgerID = ledgerID
	params.TransactionID = transactionID
	params.UpdatedAt = s.now()

	updated, err := s.repo.UpdateTransaction(ctx, params)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Transaction{}, ErrTransactionNotFound
		}
		return Transaction{}, err
	}
	return updated, nil
}

func (s *Service) DeleteTransaction(ctx context.Context, userID, ledgerID, transactionID string) error {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return err
	}

	referenced, err := s.repo.HasTransactionReferences(ctx, ledgerID, transactionID)
	if err != nil {
		return err
	}
	if referenced {
		return ErrTransactionReferenced
	}

	if err := s.repo.DeleteTransaction(ctx, ledgerID, transactionID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrTransactionNotFound
		}
		return err
	}
	return nil
}

func (s *Service) validateEntries(ctx context.Context, ledgerID string, entries []EntryInput) ([]EntryInput, error) {
	accountIDs := make([]string, 0, len(entries))
	categoryIDs := make([]string, 0, len(entries))

	requiresCategories := false
	hasTransfer := false
	for i, entry := range entries {
		if strings.TrimSpace(entry.AccountID) == "" {
			return nil, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"entries": "account_id"})
		}
		kind := strings.TrimSpace(entry.Kind)
		if kind == "" {
			kind = "normal"
		}
		entries[i].Kind = kind
		if !isValidKind(kind) {
			return nil, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"entries": "kind"})
		}
		if entry.AmountCents <= 0 {
			return nil, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"entries": "amount_cents"})
		}

		if kind == "transfer" {
			requiresCategories = true
			hasTransfer = true
		}

		accountIDs = append(accountIDs, entry.AccountID)
		if entry.CategoryID != nil && strings.TrimSpace(*entry.CategoryID) != "" {
			categoryIDs = append(categoryIDs, *entry.CategoryID)
		} else if kind == "transfer" {
			return nil, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"entries": "category_id"})
		}
	}

	if hasTransfer {
		for _, entry := range entries {
			if entry.Kind != "transfer" {
				return nil, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"entries": "mixed_kinds"})
			}
		}
	}

	accountMap, err := s.repo.GetAccountsByIDs(ctx, ledgerID, uniqueStrings(accountIDs))
	if err != nil {
		return nil, err
	}
	for _, id := range accountIDs {
		if !accountMap[id] {
			return nil, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"entries": "account_id"})
		}
	}

	if len(categoryIDs) > 0 {
		categoryMap, err := s.repo.GetCategoriesByIDs(ctx, ledgerID, uniqueStrings(categoryIDs))
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if entry.CategoryID == nil {
				continue
			}
			categoryID := strings.TrimSpace(*entry.CategoryID)
			if categoryID == "" {
				continue
			}
			if _, ok := categoryMap[categoryID]; !ok {
				return nil, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"entries": "category_id"})
			}
		}

		if hasTransfer || requiresCategories {
			if err := validateTransferBalance(entries, categoryMap); err != nil {
				return nil, err
			}
		}
	}

	return entries, nil
}

func validateTransferBalance(entries []EntryInput, categories map[string]string) error {
	var inTotal int64
	var outTotal int64
	var inCount int
	var outCount int

	for _, entry := range entries {
		if entry.Kind != "transfer" {
			continue
		}
		if entry.CategoryID == nil {
			return NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"entries": "category_id"})
		}
		direction := categories[*entry.CategoryID]
		switch direction {
		case "in":
			inTotal += entry.AmountCents
			inCount++
		case "out":
			outTotal += entry.AmountCents
			outCount++
		default:
			return NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"entries": "direction"})
		}
	}

	if inCount == 0 || outCount == 0 {
		return ErrTransferNotBalanced
	}
	if inTotal != outTotal {
		return ErrTransferNotBalanced
	}
	return nil
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

func isValidKind(kind string) bool {
	switch kind {
	case "normal", "transfer", "adjust":
		return true
	default:
		return false
	}
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

func uniqueStrings(items []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}
