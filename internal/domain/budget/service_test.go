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
	version      Version
	lines        []Line
	incomeBase   int64
	spentActual  map[string]int64
	outside      int64
	createdLines []LineInput
	incomeByMonth map[string]int64
	spentByMonth  map[string]map[string]int64
	outsideByMonth map[string]int64
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
	f.createdLines = lines
	return Version{ID: "ver-1", PlanID: planID, EffectiveFromMonth: effectiveFrom}, nil
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
	return f.lines, nil
}

func (f *fakeRepo) GetApplicableVersion(ctx context.Context, ledgerID string, month time.Time) (Version, error) {
	if f.version.ID == "" {
		return Version{}, ErrNotFound
	}
	return f.version, nil
}

func (f *fakeRepo) GetCategoryInfo(ctx context.Context, ledgerID string, categoryIDs []string) (map[string]CategoryInfo, error) {
	return f.categoryInfo, nil
}

func (f *fakeRepo) GetCategoryDescendants(ctx context.Context, ledgerID, categoryID string) ([]string, error) {
	return nil, nil
}

func (f *fakeRepo) IncomeBaseForMonth(ctx context.Context, ledgerID string, month time.Time) (int64, error) {
	if f.incomeByMonth != nil {
		if value, ok := f.incomeByMonth[month.Format("2006-01-02")]; ok {
			return value, nil
		}
	}
	return f.incomeBase, nil
}

func (f *fakeRepo) SpentActualForMonth(ctx context.Context, ledgerID string, categoryIDs []string, month time.Time) (int64, error) {
	var total int64
	if f.spentByMonth != nil {
		if values, ok := f.spentByMonth[month.Format("2006-01-02")]; ok {
			for _, id := range categoryIDs {
				total += values[id]
			}
			return total, nil
		}
	}
	for _, id := range categoryIDs {
		total += f.spentActual[id]
	}
	return total, nil
}

func (f *fakeRepo) OutsideBudgetForMonth(ctx context.Context, ledgerID string, month time.Time) (int64, error) {
	if f.outsideByMonth != nil {
		if value, ok := f.outsideByMonth[month.Format("2006-01-02")]; ok {
			return value, nil
		}
	}
	return f.outside, nil
}

func (f *fakeRepo) GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error) {
	return f.role, nil
}

func (f *fakeRepo) LedgerExists(ctx context.Context, ledgerID string) (bool, error) {
	return true, nil
}

func TestCreatePlanRequiresEditor(t *testing.T) {
	repo := &fakeRepo{role: "viewer"}
	service := NewService(repo)

	_, err := service.CreatePlan(context.Background(), "user-1", "ledger-1", "Pessoal")
	require.Equal(t, ErrAccessDenied, err)
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

func TestMonthlySummaryCalculatesTotals(t *testing.T) {
	month := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	repo := &fakeRepo{
		role:       "viewer",
		version:    Version{ID: "ver-1", PlanID: "plan-1", EffectiveFromMonth: month},
		lines:      []Line{{CategoryID: "cat-1", Percent: 50, IncludeChildren: false}},
		incomeBase: 10000,
		spentActual: map[string]int64{
			"cat-1": 6000,
		},
		outside: 200,
	}
	service := NewService(repo)

	summary, err := service.MonthlySummary(context.Background(), "user-1", "ledger-1", month)
	require.NoError(t, err)
	require.Equal(t, int64(10000), summary.IncomeBaseCents)
	require.Equal(t, int64(200), summary.OutsideBudgetCents)
	require.Len(t, summary.Lines, 1)
	require.Equal(t, int64(5000), summary.Lines[0].BudgetLimitCents)
	require.Equal(t, int64(6000), summary.Lines[0].SpentActualCents)
	require.Equal(t, int64(1000), summary.Lines[0].DeltaCents)
	require.Equal(t, 1.2, summary.Lines[0].UsagePct)
}

func TestCreateVersionSuccess(t *testing.T) {
	repo := &fakeRepo{
		role:         "editor",
		categoryInfo: map[string]CategoryInfo{"cat-1": {Direction: "out", IsBudgetRelevant: true}},
	}
	service := NewService(repo)

	month := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	_, err := service.CreateVersion(context.Background(), "user-1", "ledger-1", month, []LineInput{{CategoryID: "cat-1", Percent: 10}})
	require.NoError(t, err)
	require.Len(t, repo.createdLines, 1)
}

func TestMonthlySummaryNoVersionReturnsEmpty(t *testing.T) {
	repo := &fakeRepo{role: "viewer"}
	service := NewService(repo)

	month := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	summary, err := service.MonthlySummary(context.Background(), "user-1", "ledger-1", month)
	require.NoError(t, err)
	require.Empty(t, summary.Lines)
}

func TestPeriodSummaryAggregatesMonths(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	repo := &fakeRepo{
		role:    "viewer",
		version: Version{ID: "ver-1", PlanID: "plan-1", EffectiveFromMonth: from},
		lines:   []Line{{CategoryID: "cat-1", Percent: 10, IncludeChildren: false}},
		incomeByMonth: map[string]int64{
			"2026-01-01": 10000,
			"2026-02-01": 20000,
		},
		spentByMonth: map[string]map[string]int64{
			"2026-01-01": {"cat-1": 800},
			"2026-02-01": {"cat-1": 1500},
		},
		outsideByMonth: map[string]int64{
			"2026-01-01": 100,
			"2026-02-01": 200,
		},
	}
	service := NewService(repo)

	period, err := service.PeriodSummary(context.Background(), "user-1", "ledger-1", from, to)
	require.NoError(t, err)
	require.Len(t, period.Months, 2)
	require.Equal(t, int64(300), period.OutsideBudgetCents)
	require.Len(t, period.Categories, 1)
	require.Equal(t, int64(3000), period.TotalBudgetedCents)
	require.Equal(t, int64(2300), period.TotalSpentCents)
	require.Equal(t, int64(-700), period.TotalDeltaCents)
}
