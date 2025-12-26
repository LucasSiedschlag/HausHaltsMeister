package picuinha

import (
	"context"
)

type Repository interface {
	CreatePerson(ctx context.Context, name, notes string) (*Person, error)
	ListPersons(ctx context.Context) ([]Person, error)
	GetPerson(ctx context.Context, id int32) (*Person, error)
	UpdatePerson(ctx context.Context, id int32, name, notes string) (*Person, error)
	DeletePerson(ctx context.Context, id int32) error
	CountEntriesByPerson(ctx context.Context, personID int32) (int64, error)
	CountCasesByPerson(ctx context.Context, personID int32) (int64, error)

	AddEntry(ctx context.Context, entry *Entry) (*Entry, error)
	ListEntries(ctx context.Context) ([]Entry, error)
	ListEntriesByPerson(ctx context.Context, personID int32) ([]Entry, error)
	GetEntry(ctx context.Context, id int32) (*Entry, error)
	UpdateEntry(ctx context.Context, entry *Entry) (*Entry, error)
	DeleteEntry(ctx context.Context, id int32) error
	GetBalance(ctx context.Context, personID int32) (float64, error)

	CreateCase(ctx context.Context, picCase *Case) (*Case, error)
	UpdateCase(ctx context.Context, picCase *Case) (*Case, error)
	DeleteCase(ctx context.Context, id int32) error
	GetCase(ctx context.Context, id int32) (*Case, error)
	ListCasesByPerson(ctx context.Context, personID int32) ([]CaseSummary, error)

	CreateInstallment(ctx context.Context, installment *CaseInstallment) (*CaseInstallment, error)
	UpdateInstallment(ctx context.Context, installment *CaseInstallment) (*CaseInstallment, error)
	GetInstallment(ctx context.Context, id int32) (*CaseInstallment, error)
	ListInstallmentsByCase(ctx context.Context, caseID int32) ([]CaseInstallment, error)
}

type Service interface {
	CreatePerson(ctx context.Context, name, notes string) (*Person, error)
	ListPersons(ctx context.Context) ([]Person, error)
	UpdatePerson(ctx context.Context, id int32, name, notes string) (*Person, error)
	DeletePerson(ctx context.Context, id int32) error

	AddDiff(ctx context.Context, personID int32, amount float64, kind string, cashFlowID *int32, paymentMethodID *int32, cardOwner string, autoCreateFlow bool) (*Entry, error)
	ListEntries(ctx context.Context, personID *int32) ([]Entry, error)
	UpdateEntry(ctx context.Context, id int32, personID int32, amount float64, kind string, paymentMethodID *int32, cardOwner string, autoCreateFlow bool) (*Entry, error)
	DeleteEntry(ctx context.Context, id int32) error
	// Higher level operations might be better:
	Lend(ctx context.Context, personID int32, amount float64, cashFlowID *int32) (*Entry, error)
	Receive(ctx context.Context, personID int32, amount float64, cashFlowID *int32) (*Entry, error)

	CreateCase(ctx context.Context, req CreateCaseRequest) (*CaseSummary, error)
	ListCasesByPerson(ctx context.Context, personID int32) ([]CaseSummary, error)
	UpdateCase(ctx context.Context, id int32, req CreateCaseRequest) (*CaseSummary, error)
	DeleteCase(ctx context.Context, id int32) error
	ListInstallmentsByCase(ctx context.Context, caseID int32) ([]CaseInstallment, error)
	UpdateInstallment(ctx context.Context, id int32, req UpdateInstallmentRequest) (*CaseInstallment, error)
}
