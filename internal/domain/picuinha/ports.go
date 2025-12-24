package picuinha

import (
	"context"
)

type Repository interface {
	CreatePerson(ctx context.Context, name, notes string) (*Person, error)
	ListPersons(ctx context.Context) ([]Person, error)
	GetPerson(ctx context.Context, id int32) (*Person, error)

	AddEntry(ctx context.Context, entry *Entry) (*Entry, error)
	ListEntries(ctx context.Context, personID int32) ([]Entry, error)
	GetBalance(ctx context.Context, personID int32) (float64, error)
}

type Service interface {
	CreatePerson(ctx context.Context, name, notes string) (*Person, error)
	ListPersons(ctx context.Context) ([]Person, error)

	AddDiff(ctx context.Context, personID int32, amount float64, description string) (*Entry, error) // Generic "Add or Subtract"
	// Higher level operations might be better:
	Lend(ctx context.Context, personID int32, amount float64, cashFlowID *int32) (*Entry, error)
	Receive(ctx context.Context, personID int32, amount float64, cashFlowID *int32) (*Entry, error)
}
