package budget

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/ledger"
)

type Repository interface {
	GetPlanByLedger(ctx context.Context, ledgerID string) (Plan, error)
	CreatePlan(ctx context.Context, ledgerID, name string) (Plan, error)
	UpdatePlan(ctx context.Context, ledgerID, name string, updatedAt time.Time) (Plan, error)
	ListVersions(ctx context.Context, ledgerID string, from, to *time.Time) ([]Version, error)
	GetVersion(ctx context.Context, ledgerID, versionID string) (Version, error)
	CreateVersionWithLines(ctx context.Context, ledgerID, planID, userID string, effectiveFrom time.Time, lines []LineInput) (Version, error)
	UpdateLine(ctx context.Context, lineID string, percent float64, includeChildren bool, updatedAt time.Time) (Line, error)
	DeleteLine(ctx context.Context, lineID string) error
	AddLine(ctx context.Context, versionID string, line LineInput) (Line, error)
	GetLinesByVersion(ctx context.Context, versionID string) ([]Line, error)
	GetApplicableVersion(ctx context.Context, ledgerID string, month time.Time) (Version, error)
	GetCategoryInfo(ctx context.Context, ledgerID string, categoryIDs []string) (map[string]CategoryInfo, error)
	GetCategoryDescendants(ctx context.Context, ledgerID, categoryID string) ([]string, error)
	IncomeBaseForMonth(ctx context.Context, ledgerID string, month time.Time) (int64, error)
	SpentActualForMonth(ctx context.Context, ledgerID string, categoryIDs []string, month time.Time) (int64, error)
	OutsideBudgetForMonth(ctx context.Context, ledgerID string, month time.Time) (int64, error)
	GetLedgerRole(ctx context.Context, ledgerID, userID string) (string, error)
	LedgerExists(ctx context.Context, ledgerID string) (bool, error)
}

type CategoryInfo struct {
	Direction        string
	IsBudgetRelevant bool
}

type LineInput struct {
	CategoryID      string
	Percent         float64
	IncludeChildren bool
}

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now().UTC}
}

func (s *Service) GetPlan(ctx context.Context, userID, ledgerID string) (Plan, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "viewer"); err != nil {
		return Plan{}, err
	}
	plan, err := s.repo.GetPlanByLedger(ctx, ledgerID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Plan{}, ErrPlanNotFound
		}
		return Plan{}, err
	}
	return plan, nil
}

func (s *Service) CreatePlan(ctx context.Context, userID, ledgerID, name string) (Plan, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return Plan{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Default"
	}
	return s.repo.CreatePlan(ctx, ledgerID, name)
}

func (s *Service) UpdatePlan(ctx context.Context, userID, ledgerID, name string) (Plan, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return Plan{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return Plan{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"name": "required"})
	}
	return s.repo.UpdatePlan(ctx, ledgerID, name, s.now())
}

func (s *Service) ListVersions(ctx context.Context, userID, ledgerID string, from, to *time.Time) ([]Version, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "viewer"); err != nil {
		return nil, err
	}
	return s.repo.ListVersions(ctx, ledgerID, from, to)
}

func (s *Service) GetVersion(ctx context.Context, userID, ledgerID, versionID string) (Version, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "viewer"); err != nil {
		return Version{}, err
	}
	version, err := s.repo.GetVersion(ctx, ledgerID, versionID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Version{}, ErrVersionNotFound
		}
		return Version{}, err
	}
	lines, err := s.repo.GetLinesByVersion(ctx, version.ID)
	if err != nil {
		return Version{}, err
	}
	version.Lines = lines
	return version, nil
}

func (s *Service) CreateVersion(ctx context.Context, userID, ledgerID string, effectiveFrom time.Time, lines []LineInput) (Version, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return Version{}, err
	}
	if !isMonthStart(effectiveFrom) {
		return Version{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"effective_from_month": "invalid"})
	}
	if len(lines) == 0 {
		return Version{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"lines": "required"})
	}

	plan, err := s.repo.GetPlanByLedger(ctx, ledgerID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			plan, err = s.repo.CreatePlan(ctx, ledgerID, "Default")
		}
	}
	if err != nil {
		return Version{}, err
	}

	validated, err := s.validateLines(ctx, ledgerID, lines)
	if err != nil {
		return Version{}, err
	}

	return s.repo.CreateVersionWithLines(ctx, ledgerID, plan.ID, userID, effectiveFrom, validated)
}

func (s *Service) AddLine(ctx context.Context, userID, ledgerID, versionID string, line LineInput) (Line, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return Line{}, err
	}
	validated, err := s.validateLines(ctx, ledgerID, []LineInput{line})
	if err != nil {
		return Line{}, err
	}
	return s.repo.AddLine(ctx, versionID, validated[0])
}

func (s *Service) UpdateLine(ctx context.Context, userID, ledgerID, lineID string, percent float64, includeChildren bool) (Line, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return Line{}, err
	}
	if percent <= 0 {
		return Line{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"percent": "invalid"})
	}
	line, err := s.repo.UpdateLine(ctx, lineID, percent, includeChildren, s.now())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Line{}, ErrLineNotFound
		}
		return Line{}, err
	}
	return line, nil
}

func (s *Service) DeleteLine(ctx context.Context, userID, ledgerID, lineID string) error {
	if err := s.requireRole(ctx, ledgerID, userID, "editor"); err != nil {
		return err
	}
	if err := s.repo.DeleteLine(ctx, lineID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrLineNotFound
		}
		return err
	}
	return nil
}

func (s *Service) MonthlySummary(ctx context.Context, userID, ledgerID string, month time.Time) (MonthlySummary, error) {
	if err := s.requireRole(ctx, ledgerID, userID, "viewer"); err != nil {
		return MonthlySummary{}, err
	}
	if !isMonthStart(month) {
		return MonthlySummary{}, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"month": "invalid"})
	}

	version, err := s.repo.GetApplicableVersion(ctx, ledgerID, month)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return MonthlySummary{Month: month}, nil
		}
		return MonthlySummary{}, err
	}
	lines, err := s.repo.GetLinesByVersion(ctx, version.ID)
	if err != nil {
		return MonthlySummary{}, err
	}
	version.Lines = lines

	incomeBase, err := s.repo.IncomeBaseForMonth(ctx, ledgerID, month)
	if err != nil {
		return MonthlySummary{}, err
	}

	items := make([]MonthlyLine, 0, len(lines))
	for _, line := range lines {
		categoryIDs := []string{line.CategoryID}
		if line.IncludeChildren {
			desc, err := s.repo.GetCategoryDescendants(ctx, ledgerID, line.CategoryID)
			if err != nil {
				return MonthlySummary{}, err
			}
			categoryIDs = uniqueStrings(append(categoryIDs, desc...))
		}

		spent, err := s.repo.SpentActualForMonth(ctx, ledgerID, categoryIDs, month)
		if err != nil {
			return MonthlySummary{}, err
		}

		limit := int64(float64(incomeBase) * (line.Percent / 100.0))
		delta := spent - limit
		usage := 0.0
		if limit > 0 {
			usage = float64(spent) / float64(limit)
		}
		items = append(items, MonthlyLine{
			CategoryID:       line.CategoryID,
			Percent:          line.Percent,
			IncludeChildren:  line.IncludeChildren,
			BudgetLimitCents: limit,
			SpentActualCents: spent,
			DeltaCents:       delta,
			UsagePct:         usage,
		})
	}

	outside, err := s.repo.OutsideBudgetForMonth(ctx, ledgerID, month)
	if err != nil {
		return MonthlySummary{}, err
	}

	return MonthlySummary{
		Month:              month,
		Version:            &version,
		IncomeBaseCents:    incomeBase,
		Lines:              items,
		OutsideBudgetCents: outside,
	}, nil
}

func (s *Service) validateLines(ctx context.Context, ledgerID string, lines []LineInput) ([]LineInput, error) {
	categoryIDs := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line.CategoryID) == "" {
			return nil, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"category_id": "required"})
		}
		if line.Percent <= 0 {
			return nil, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"percent": "invalid"})
		}
		categoryIDs = append(categoryIDs, line.CategoryID)
	}

	info, err := s.repo.GetCategoryInfo(ctx, ledgerID, uniqueStrings(categoryIDs))
	if err != nil {
		return nil, err
	}
	for _, line := range lines {
		data, ok := info[line.CategoryID]
		if !ok {
			return nil, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"category_id": "invalid"})
		}
		if data.Direction != "out" || !data.IsBudgetRelevant {
			return nil, NewError("VALIDATION_ERROR", "Validacao falhou", map[string]string{"category_id": "not_eligible"})
		}
	}
	return lines, nil
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

func isMonthStart(value time.Time) bool {
	return value.Day() == 1 && value.Equal(time.Date(value.Year(), value.Month(), 1, 0, 0, 0, 0, value.Location()))
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
