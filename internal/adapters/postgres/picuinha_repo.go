package postgres

import (
	"context"
	"database/sql"

	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/sqlc"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/picuinha"
)

type PicuinhaRepository struct {
	q *sqlc.Queries
}

func NewPicuinhaRepository(db *pgxpool.Pool) *PicuinhaRepository {
	return &PicuinhaRepository{
		q: sqlc.New(db),
	}
}

func (r *PicuinhaRepository) CreatePerson(ctx context.Context, name, notes string) (*picuinha.Person, error) {
	n := pgtype.Text{String: notes, Valid: notes != ""}
	row, err := r.q.CreatePerson(ctx, sqlc.CreatePersonParams{Name: name, Notes: n})
	if err != nil {
		return nil, err
	}
	return &picuinha.Person{
		ID:    row.PersonID,
		Name:  row.Name,
		Notes: row.Notes.String,
	}, nil
}

func (r *PicuinhaRepository) ListPersons(ctx context.Context) ([]picuinha.Person, error) {
	rows, err := r.q.ListPersons(ctx)
	if err != nil {
		return nil, err
	}
	persons := make([]picuinha.Person, len(rows))
	for i, row := range rows {
		persons[i] = picuinha.Person{
			ID:    row.PersonID,
			Name:  row.Name,
			Notes: row.Notes.String,
		}
	}
	return persons, nil
}

func (r *PicuinhaRepository) GetPerson(ctx context.Context, id int32) (*picuinha.Person, error) {
	row, err := r.q.GetPerson(ctx, id)
	if err != nil {
		return nil, err
	}
	return &picuinha.Person{
		ID:    row.PersonID,
		Name:  row.Name,
		Notes: row.Notes.String,
	}, nil
}

func (r *PicuinhaRepository) AddEntry(ctx context.Context, entry *picuinha.Entry) (*picuinha.Entry, error) {
	pgDate := pgtype.Date{Time: entry.Date, Valid: true}
	var am pgtype.Numeric
	am.Scan(fmt.Sprintf("%.2f", entry.Amount))

	cfID := pgtype.Int4{Valid: false}
	if entry.CashFlowID != nil {
		cfID = pgtype.Int4{Int32: *entry.CashFlowID, Valid: true}
	}

	row, err := r.q.CreatePicuinhaEntry(ctx, sqlc.CreatePicuinhaEntryParams{
		PersonID:   entry.PersonID,
		Date:       pgDate,
		Kind:       entry.Kind,
		Amount:     am,
		CashFlowID: cfID,
	})
	if err != nil {
		return nil, err
	}

	val, _ := row.Amount.Float64Value()
	var retCfID *int32
	if row.CashFlowID.Valid {
		retCfID = &row.CashFlowID.Int32
	}

	return &picuinha.Entry{
		ID:         row.PicuinhaEntryID,
		PersonID:   row.PersonID,
		Date:       row.Date.Time,
		Kind:       row.Kind,
		Amount:     val.Float64,
		CashFlowID: retCfID,
	}, nil
}

func (r *PicuinhaRepository) ListEntries(ctx context.Context, personID int32) ([]picuinha.Entry, error) {
	rows, err := r.q.ListEntriesByPerson(ctx, personID)
	if err != nil {
		return nil, err
	}
	entries := make([]picuinha.Entry, len(rows))
	for i, row := range rows {
		val, _ := row.Amount.Float64Value()
		var retCfID *int32
		if row.CashFlowID.Valid {
			retCfID = &row.CashFlowID.Int32
		}
		entries[i] = picuinha.Entry{
			ID:         row.PicuinhaEntryID,
			PersonID:   row.PersonID,
			Date:       row.Date.Time,
			Kind:       row.Kind,
			Amount:     val.Float64,
			CashFlowID: retCfID,
		}
	}
	return entries, nil
}

func (r *PicuinhaRepository) GetBalance(ctx context.Context, personID int32) (float64, error) {
	bal, err := r.q.GetPersonBalance(ctx, personID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	val, _ := bal.Float64Value()
	return val.Float64, nil
}
