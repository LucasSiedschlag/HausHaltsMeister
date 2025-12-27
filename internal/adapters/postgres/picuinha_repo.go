package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/postgres/sqlc"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/picuinha"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &picuinha.Person{
		ID:    row.PersonID,
		Name:  row.Name,
		Notes: row.Notes.String,
	}, nil
}

func (r *PicuinhaRepository) UpdatePerson(ctx context.Context, id int32, name, notes string) (*picuinha.Person, error) {
	n := pgtype.Text{String: notes, Valid: notes != ""}
	row, err := r.q.UpdatePerson(ctx, sqlc.UpdatePersonParams{
		PersonID: id,
		Name:     name,
		Notes:    n,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &picuinha.Person{
		ID:    row.PersonID,
		Name:  row.Name,
		Notes: row.Notes.String,
	}, nil
}

func (r *PicuinhaRepository) DeletePerson(ctx context.Context, id int32) error {
	return r.q.DeletePerson(ctx, id)
}

func (r *PicuinhaRepository) CountCasesByPerson(ctx context.Context, personID int32) (int64, error) {
	return r.q.CountCasesByPerson(ctx, int4FromPtr(&personID))
}

func (r *PicuinhaRepository) GetBalance(ctx context.Context, personID int32) (float64, error) {
	bal, err := r.q.GetPersonBalance(ctx, int4FromPtr(&personID))
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	val, _ := bal.Float64Value()
	return val.Float64, nil
}

func (r *PicuinhaRepository) CreateCase(ctx context.Context, picCase *picuinha.Case) (*picuinha.Case, error) {
	count := int32(1)
	if picCase.InstallmentCount != nil {
		count = *picCase.InstallmentCount
	}
	row, err := r.q.CreatePicuinhaCase(ctx, sqlc.CreatePicuinhaCaseParams{
		PersonID:                 int4FromPtr(&picCase.PersonID),
		Description:              picCase.Title,
		PlanType:                 picCase.CaseType,
		TotalAmount:              numericFromPtr(picCase.TotalAmount),
		InstallmentCount:         count,
		InstallmentAmount:        numericFromPtr(picCase.InstallmentAmount),
		StartDate:                pgtype.Date{Time: picCase.StartDate, Valid: true},
		PaymentMethodID:          int4FromPtr(picCase.PaymentMethodID),
		CategoryID:               int4FromPtr(picCase.CategoryID),
		InterestRate:             numericFromPtr(picCase.InterestRate),
		InterestRateUnit:         textFromString(picCase.InterestRateUnit),
		RecurrenceIntervalMonths: int4FromPtr(picCase.RecurrenceIntervalMonths),
	})
	if err != nil {
		return nil, err
	}

	return mapCaseRow(row), nil
}

func (r *PicuinhaRepository) UpdateCase(ctx context.Context, picCase *picuinha.Case) (*picuinha.Case, error) {
	count := int32(1)
	if picCase.InstallmentCount != nil {
		count = *picCase.InstallmentCount
	}
	row, err := r.q.UpdatePicuinhaCase(ctx, sqlc.UpdatePicuinhaCaseParams{
		InstallmentPlanID:        picCase.ID,
		PersonID:                 int4FromPtr(&picCase.PersonID),
		Description:              picCase.Title,
		PlanType:                 picCase.CaseType,
		TotalAmount:              numericFromPtr(picCase.TotalAmount),
		InstallmentCount:         count,
		InstallmentAmount:        numericFromPtr(picCase.InstallmentAmount),
		StartDate:                pgtype.Date{Time: picCase.StartDate, Valid: true},
		PaymentMethodID:          int4FromPtr(picCase.PaymentMethodID),
		CategoryID:               int4FromPtr(picCase.CategoryID),
		InterestRate:             numericFromPtr(picCase.InterestRate),
		InterestRateUnit:         textFromString(picCase.InterestRateUnit),
		RecurrenceIntervalMonths: int4FromPtr(picCase.RecurrenceIntervalMonths),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return mapCaseRow(row), nil
}

func (r *PicuinhaRepository) DeleteCase(ctx context.Context, id int32) error {
	return r.q.DeletePicuinhaCase(ctx, id)
}

func (r *PicuinhaRepository) GetCase(ctx context.Context, id int32) (*picuinha.Case, error) {
	row, err := r.q.GetPicuinhaCase(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return mapCaseRow(row), nil
}

func (r *PicuinhaRepository) ListCasesByPerson(ctx context.Context, personID int32) ([]picuinha.CaseSummary, error) {
	rows, err := r.q.ListPicuinhaCasesByPerson(ctx, int4FromPtr(&personID))
	if err != nil {
		return nil, err
	}
	results := make([]picuinha.CaseSummary, len(rows))
	for i, row := range rows {
		c := mapCaseSummaryRow(row)
		results[i] = c
	}
	return results, nil
}

func (r *PicuinhaRepository) CreateInstallment(ctx context.Context, installment *picuinha.CaseInstallment) (*picuinha.CaseInstallment, error) {
	row, err := r.q.CreatePicuinhaCaseInstallment(ctx, sqlc.CreatePicuinhaCaseInstallmentParams{
		InstallmentPlanID: installment.CaseID,
		InstallmentNumber: installment.InstallmentNumber,
		DueDate:           pgtype.Date{Time: installment.DueDate, Valid: true},
		Amount:            numericFromValue(installment.Amount),
		ExtraAmount:       numericFromValue(installment.ExtraAmount),
		IsPaid:            installment.IsPaid,
		PaidAt:            timestampFromPtr(installment.PaidAt),
	})
	if err != nil {
		return nil, err
	}
	return mapCaseInstallmentRow(row), nil
}

func (r *PicuinhaRepository) UpdateInstallment(ctx context.Context, installment *picuinha.CaseInstallment) (*picuinha.CaseInstallment, error) {
	row, err := r.q.UpdatePicuinhaCaseInstallment(ctx, sqlc.UpdatePicuinhaCaseInstallmentParams{
		InstallmentPlanItemID: installment.ID,
		Amount:                numericFromValue(installment.Amount),
		ExtraAmount:           numericFromValue(installment.ExtraAmount),
		IsPaid:                installment.IsPaid,
		PaidAt:                timestampFromPtr(installment.PaidAt),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return mapCaseInstallmentRow(row), nil
}

func (r *PicuinhaRepository) GetInstallment(ctx context.Context, id int32) (*picuinha.CaseInstallment, error) {
	row, err := r.q.GetPicuinhaCaseInstallment(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return mapCaseInstallmentRow(row), nil
}

func (r *PicuinhaRepository) ListInstallmentsByCase(ctx context.Context, caseID int32) ([]picuinha.CaseInstallment, error) {
	rows, err := r.q.ListPicuinhaCaseInstallments(ctx, caseID)
	if err != nil {
		return nil, err
	}
	installments := make([]picuinha.CaseInstallment, len(rows))
	for i, row := range rows {
		installments[i] = *mapCaseInstallmentRow(row)
	}
	return installments, nil
}

func mapCaseRow(row sqlc.InstallmentPlan) *picuinha.Case {
	personID := int32(0)
	if row.PersonID.Valid {
		personID = row.PersonID.Int32
	}
	var planID *int32
	if row.PlanType == picuinha.CaseTypeCardInstall {
		id := row.InstallmentPlanID
		planID = &id
	}
	installmentCount := row.InstallmentCount
	installmentCountPtr := &installmentCount
	return &picuinha.Case{
		ID:                       row.InstallmentPlanID,
		PersonID:                 personID,
		Title:                    row.Description,
		CaseType:                 row.PlanType,
		TotalAmount:              numericToPtr(row.TotalAmount),
		InstallmentCount:         installmentCountPtr,
		InstallmentAmount:        numericToPtr(row.InstallmentAmount),
		StartDate:                row.StartDate.Time,
		PaymentMethodID:          int4ToPtr(row.PaymentMethodID),
		InstallmentPlanID:        planID,
		CategoryID:               int4ToPtr(row.CategoryID),
		InterestRate:             numericToPtr(row.InterestRate),
		InterestRateUnit:         row.InterestRateUnit.String,
		RecurrenceIntervalMonths: int4ToPtr(row.RecurrenceIntervalMonths),
		CreatedAt:                row.CreatedAt.Time,
	}
}

func mapCaseSummaryRow(row sqlc.ListPicuinhaCasesByPersonRow) picuinha.CaseSummary {
	var planID *int32
	if row.PlanType == picuinha.CaseTypeCardInstall {
		id := row.InstallmentPlanID
		planID = &id
	}
	personID := int32(0)
	if row.PersonID.Valid {
		personID = row.PersonID.Int32
	}
	installmentCount := row.InstallmentCount
	caseData := picuinha.Case{
		ID:                       row.InstallmentPlanID,
		PersonID:                 personID,
		Title:                    row.Description,
		CaseType:                 row.PlanType,
		TotalAmount:              numericToPtr(row.TotalAmount),
		InstallmentCount:         &installmentCount,
		InstallmentAmount:        numericToPtr(row.InstallmentAmount),
		StartDate:                row.StartDate.Time,
		PaymentMethodID:          int4ToPtr(row.PaymentMethodID),
		InstallmentPlanID:        planID,
		CategoryID:               int4ToPtr(row.CategoryID),
		InterestRate:             numericToPtr(row.InterestRate),
		InterestRateUnit:         row.InterestRateUnit.String,
		RecurrenceIntervalMonths: int4ToPtr(row.RecurrenceIntervalMonths),
		CreatedAt:                row.CreatedAt.Time,
	}
	paid := int32(row.InstallmentsPaid)
	total := int32(row.InstallmentsTotal)
	status := picuinha.StatusOpen
	if row.PlanType == picuinha.CaseTypeRecurring {
		status = picuinha.StatusRecurringActive
	} else if total > 0 && paid >= total {
		status = picuinha.StatusPaid
	}
	return picuinha.CaseSummary{
		Case:              caseData,
		InstallmentsTotal: total,
		InstallmentsPaid:  paid,
		AmountPaid:        numericToValue(row.AmountPaid),
		AmountRemaining:   numericToValue(row.AmountRemaining),
		Status:            status,
	}
}

func mapCaseInstallmentRow(row sqlc.InstallmentPlanItem) *picuinha.CaseInstallment {
	var paidAt *time.Time
	if row.PaidAt.Valid {
		paidAt = &row.PaidAt.Time
	}
	return &picuinha.CaseInstallment{
		ID:                row.InstallmentPlanItemID,
		CaseID:            row.InstallmentPlanID,
		InstallmentNumber: row.InstallmentNumber,
		DueDate:           row.DueDate.Time,
		Amount:            numericToValue(row.Amount),
		ExtraAmount:       numericToValue(row.ExtraAmount),
		IsPaid:            row.IsPaid,
		PaidAt:            paidAt,
	}
}

func numericFromPtr(value *float64) pgtype.Numeric {
	if value == nil {
		return pgtype.Numeric{Valid: false}
	}
	return numericFromValue(*value)
}

func numericFromValue(value float64) pgtype.Numeric {
	var out pgtype.Numeric
	out.Scan(fmt.Sprintf("%.2f", value))
	return out
}

func numericToPtr(value pgtype.Numeric) *float64 {
	if !value.Valid {
		return nil
	}
	val, _ := value.Float64Value()
	return &val.Float64
}

func numericToValue(value pgtype.Numeric) float64 {
	val, _ := value.Float64Value()
	return val.Float64
}

func int4FromPtr(value *int32) pgtype.Int4 {
	if value == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: *value, Valid: true}
}

func int4ToPtr(value pgtype.Int4) *int32 {
	if !value.Valid {
		return nil
	}
	return &value.Int32
}

func textFromString(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: value, Valid: true}
}

func timestampFromPtr(value *time.Time) pgtype.Timestamp {
	if value == nil {
		return pgtype.Timestamp{Valid: false}
	}
	return pgtype.Timestamp{Time: *value, Valid: true}
}
