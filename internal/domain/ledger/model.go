package ledger

import (
	"errors"
	"math"
	"strings"
	"time"
)

const (
	AccountTypeAsset     = "ASSET"
	AccountTypeLiability = "LIABILITY"
	AccountTypeEquity    = "EQUITY"
	AccountTypeIncome    = "INCOME"
	AccountTypeExpense   = "EXPENSE"

	PostingSideDebit  = "DEBIT"
	PostingSideCredit = "CREDIT"
)

var (
	ErrInvalidAccountName    = errors.New("account name is required")
	ErrInvalidAccountType    = errors.New("invalid account type")
	ErrInvalidCurrency       = errors.New("invalid currency")
	ErrInvalidTransaction    = errors.New("transaction date is required")
	ErrEmptyDescription      = errors.New("description is required")
	ErrNoPostings            = errors.New("at least two postings are required")
	ErrInvalidPostingSide    = errors.New("posting side must be DEBIT or CREDIT")
	ErrInvalidPostingAmount  = errors.New("posting amount must be greater than zero")
	ErrInvalidPostingAccount = errors.New("posting account_id is required")
	ErrUnbalancedTransaction = errors.New("transaction is not balanced")
)

type Account struct {
	ID        int32
	Name      string
	Type      string
	Currency  string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Transaction struct {
	ID          int32
	OccurredAt  time.Time
	Description string
	Reference   string
	Notes       string
	Postings    []*Posting
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Posting struct {
	ID            int32
	TransactionID int32
	AccountID     int32
	CategoryID    *int32
	PartyID       *int32
	Side          string
	Amount        float64
	Memo          string
	CreatedAt     time.Time
}

func NewAccount(name, accountType, currency string, isActive bool) (*Account, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidAccountName
	}

	accountType = strings.ToUpper(strings.TrimSpace(accountType))
	if !isValidAccountType(accountType) {
		return nil, ErrInvalidAccountType
	}

	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		currency = "BRL"
	}
	if len(currency) != 3 {
		return nil, ErrInvalidCurrency
	}

	return &Account{
		Name:     name,
		Type:     accountType,
		Currency: currency,
		IsActive: isActive,
	}, nil
}

func NewTransaction(occurredAt time.Time, description, reference, notes string, postings []*Posting) (*Transaction, error) {
	if occurredAt.IsZero() {
		return nil, ErrInvalidTransaction
	}

	description = strings.TrimSpace(description)
	if description == "" {
		return nil, ErrEmptyDescription
	}

	if len(postings) < 2 {
		return nil, ErrNoPostings
	}

	normalized := make([]*Posting, len(postings))
	var debitTotal float64
	var creditTotal float64

	for i, posting := range postings {
		if posting == nil {
			return nil, ErrInvalidPostingAccount
		}
		if posting.AccountID <= 0 {
			return nil, ErrInvalidPostingAccount
		}
		if posting.Amount <= 0 {
			return nil, ErrInvalidPostingAmount
		}

		side := strings.ToUpper(strings.TrimSpace(posting.Side))
		if side != PostingSideDebit && side != PostingSideCredit {
			return nil, ErrInvalidPostingSide
		}

		if side == PostingSideDebit {
			debitTotal += posting.Amount
		} else {
			creditTotal += posting.Amount
		}

		cp := *posting
		cp.Side = side
		cp.Memo = strings.TrimSpace(cp.Memo)
		normalized[i] = &cp
	}

	if round2(debitTotal) != round2(creditTotal) {
		return nil, ErrUnbalancedTransaction
	}

	return &Transaction{
		OccurredAt:  occurredAt,
		Description: description,
		Reference:   strings.TrimSpace(reference),
		Notes:       strings.TrimSpace(notes),
		Postings:    normalized,
	}, nil
}

func isValidAccountType(accountType string) bool {
	switch accountType {
	case AccountTypeAsset, AccountTypeLiability, AccountTypeEquity, AccountTypeIncome, AccountTypeExpense:
		return true
	default:
		return false
	}
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}
