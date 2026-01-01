package budget

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	role         string
	categoryInfo map[string]CategoryInfo
}

func (f *fakeRepo) GetPlanByLedger(ctx context.Context, ledgerID string) (Plan, error) {
	return Plan{}, ErrNotFound
}

func (f *fakeRepo) CreatePlan(ctx context.Context, ledgerID, name string) (Plan, error) {
	return Plan{ID: "plan-1", LedgerID: ledgerID, Name: name}, nil
}

func (f *fakeRepo) UpdatePlan(ctx context.Context, ledgerID, name string, updatedAt time.Time) (Plan, error) {
	return Plan{}, nil
}

func (f *fakeRepo) ListVersions(ctx context.Context, ledgerID string, from, to *time.Time) ([]Version, error) {
	return nil, nil
}

func (f *fakeRepo) GetVersion(ctx context.Context, ledgerID, versionID string) (Version, error) {
	return Version{}, nil
}

func (f *fakeRepo) CreateVersionWithLines(ctx context.Context, ledgerID, planID, userID string, effectiveFrom time.Time, lines []LineInput) (Version, error) {
	return Version{}, nil
}

func (f *fakeRepo) UpdateLine(ctx context.Context, lineID string, percent float64, includeChildren bool, updatedAt time.Time) (Line, error) {
	return Line{}, nil
}

func (f *fakeRepo) DeleteLine(ctx context.Context, lineID string) error {
	return nil
}

func (f *fakeRepo) AddLine(ctx context.Context, versionID string, line LineInput) (Line, error) {
	return Line{}, nil
}

func (f *fakeRepo) GetLinesByVersion(ctx context.Context, versionID string) ([]Line, error) {
	return nil, nil
}

func (f *fakeRepo) GetApplicableVersion(ctx context.Context, ledgerID string, month time.Time) (Version, error) {
	return Version{}, ErrNotFound
}

func (f *fakeRepo) GetCategoryInfo(ctx context.Context, ledgerID string, categoryIDs []string) (map[string]CategoryInfo, error) {
	return f.categoryInfo, nil
}

func (f *fakeRepo) GetCategoryDescendants(ctx context.Context, ledgerID, categoryID string) ([]string, error) {
	return nil, nil
}

func (f *fakeRepo) IncomeBaseForMonth(ctx context.Context, ledgerID string, month time.Time) (int64, error) {
	return 0, nil
}

func (f *fakeRepo) SpentActualForMonth(ctx context.Context, ledgerID string, categoryIDs []string, month time.Time) (int64, error) {
	return 0, nil
}

func (f *fakeRepo) OutsideBudgetForMonth(ctx context.Context, ledgerID string, month time.Time) (int64, error) {
	return 0, nil
}

func (f *fakeRepo) GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error) {
	return f.role, nil
}

func (f *fakeRepo) LedgerExists(ctx context.Context, ledgerID string) (bool, error) {
	return true, nil
}

func TestCreateVersionValidatesCategory(t *testing.T) {
	repo := &fakeRepo{
		role:         "editor",
		categoryInfo: map[string]CategoryInfo{"cat-1": {Direction: "in", IsBudgetRelevant: true}},
	}
	service := NewService(repo)

	_, err := service.CreateVersion(context.Background(), "user-1", "ledger-1", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), []LineInput{{CategoryID: "cat-1", Percent: 10}})
	require.Error(t, err)
}

func TestMonthlySummaryRequiresMonthStart(t *testing.T) {
	repo := &fakeRepo{role: "viewer"}
	service := NewService(repo)

	_, err := service.MonthlySummary(context.Background(), "user-1", "ledger-1", time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC))
	require.Error(t, err)
}
