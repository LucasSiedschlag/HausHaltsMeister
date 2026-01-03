package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/dto"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/httpx"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/budget"
	"github.com/labstack/echo/v4"
)

type BudgetHandler struct {
	Service BudgetService
}

type BudgetService interface {
	GetPlan(ctx context.Context, userID, ledgerID string) (budget.Plan, error)
	CreatePlan(ctx context.Context, userID, ledgerID, name string) (budget.Plan, error)
	UpdatePlan(ctx context.Context, userID, ledgerID, name string) (budget.Plan, error)
	ListVersions(ctx context.Context, userID, ledgerID string, from, to *time.Time) ([]budget.Version, error)
	GetVersion(ctx context.Context, userID, ledgerID, versionID string) (budget.Version, error)
	CreateVersion(ctx context.Context, userID, ledgerID string, effectiveFrom time.Time, lines []budget.LineInput) (budget.Version, error)
	AddLine(ctx context.Context, userID, ledgerID, versionID string, line budget.LineInput) (budget.Line, error)
	UpdateLine(ctx context.Context, userID, ledgerID, lineID string, percent float64, includeChildren bool) (budget.Line, error)
	DeleteLine(ctx context.Context, userID, ledgerID, lineID string) error
	MonthlySummary(ctx context.Context, userID, ledgerID string, month time.Time) (budget.MonthlySummary, error)
	PeriodSummary(ctx context.Context, userID, ledgerID string, from, to time.Time) (budget.PeriodSummary, error)
}

type planRequest = dto.BudgetPlanRequest
type planResponse = dto.BudgetPlanResponse
type versionRequest = dto.BudgetVersionRequest
type versionLineRequest = dto.BudgetVersionLineRequest
type versionResponse = dto.BudgetVersionResponse
type lineResponse = dto.BudgetLineResponse
type monthlyResponse = dto.BudgetMonthlyResponse
type monthlyLineResponse = dto.BudgetMonthlyLineResponse
type periodResponse = dto.BudgetPeriodResponse
type periodLineResponse = dto.BudgetPeriodLineResponse

func (h *BudgetHandler) Register(base *echo.Group) {
	plan := base.Group("/plan")
	plan.GET("", h.GetPlan)
	plan.POST("", h.CreatePlan)
	plan.PATCH("", h.UpdatePlan)
	plan.DELETE("", h.DeletePlan)

	versions := base.Group("/versions")
	versions.POST("", h.CreateVersion)
	versions.GET("", h.ListVersions)
	versions.GET("/:versionId", h.GetVersion)
	versions.PATCH("/:versionId", h.UpdateVersion)
	versions.DELETE("/:versionId", h.DeleteVersion)

	lines := base.Group("/versions/:versionId/lines")
	lines.POST("", h.AddLine)
	lines.PATCH("/:lineId", h.UpdateLine)
	lines.DELETE("/:lineId", h.DeleteLine)

	base.GET("/monthly", h.Monthly)
	base.GET("/period", h.PeriodSummary)
}

func (h *BudgetHandler) GetPlan(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	if ledgerID == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"ledger_id": "required"})
	}

	plan, err := h.Service.GetPlan(c.Request().Context(), user.ID, ledgerID)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.JSON(http.StatusOK, toPlanResponse(plan))
}

func (h *BudgetHandler) CreatePlan(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	if ledgerID == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"ledger_id": "required"})
	}
	var req planRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	plan, err := h.Service.CreatePlan(c.Request().Context(), user.ID, ledgerID, req.Name)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.JSON(http.StatusCreated, toPlanResponse(plan))
}

func (h *BudgetHandler) UpdatePlan(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	if ledgerID == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"ledger_id": "required"})
	}
	var req planRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	plan, err := h.Service.UpdatePlan(c.Request().Context(), user.ID, ledgerID, req.Name)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.JSON(http.StatusOK, toPlanResponse(plan))
}

func (h *BudgetHandler) DeletePlan(c echo.Context) error {
	return httpx.WriteError(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Funcionalidade nao disponivel", nil)
}

func (h *BudgetHandler) CreateVersion(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	if ledgerID == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"ledger_id": "required"})
	}

	var req versionRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}
	month, err := httpx.ParseMonth(req.EffectiveFromMonth)
	if err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"effective_from_month": "invalid"})
	}

	lines := make([]budget.LineInput, 0, len(req.Lines))
	for _, line := range req.Lines {
		lines = append(lines, budget.LineInput{
			CategoryID:      line.CategoryID,
			Percent:         line.Percent,
			IncludeChildren: line.IncludeChildren,
		})
	}

	version, err := h.Service.CreateVersion(c.Request().Context(), user.ID, ledgerID, month, lines)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.JSON(http.StatusCreated, toVersionResponse(version, ledgerID))
}

func (h *BudgetHandler) ListVersions(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	if ledgerID == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"ledger_id": "required"})
	}

	var from *time.Time
	if value := strings.TrimSpace(c.QueryParam("from")); value != "" {
		parsed, err := httpx.ParseMonth(value)
		if err != nil {
			return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"from": "invalid"})
		}
		from = &parsed
	}
	var to *time.Time
	if value := strings.TrimSpace(c.QueryParam("to")); value != "" {
		parsed, err := httpx.ParseMonth(value)
		if err != nil {
			return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"to": "invalid"})
		}
		to = &parsed
	}

	versions, err := h.Service.ListVersions(c.Request().Context(), user.ID, ledgerID, from, to)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	response := make([]versionResponse, 0, len(versions))
	for _, version := range versions {
		response = append(response, toVersionResponse(version, ledgerID))
	}
	return c.JSON(http.StatusOK, response)
}

func (h *BudgetHandler) GetVersion(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	versionID := c.Param("versionId")
	if ledgerID == "" || versionID == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"version_id": "required"})
	}

	version, err := h.Service.GetVersion(c.Request().Context(), user.ID, ledgerID, versionID)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.JSON(http.StatusOK, toVersionResponse(version, ledgerID))
}

func (h *BudgetHandler) UpdateVersion(c echo.Context) error {
	return httpx.WriteError(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Funcionalidade nao disponivel", nil)
}

func (h *BudgetHandler) DeleteVersion(c echo.Context) error {
	return httpx.WriteError(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Funcionalidade nao disponivel", nil)
}

func (h *BudgetHandler) AddLine(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	versionID := c.Param("versionId")
	if ledgerID == "" || versionID == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"version_id": "required"})
	}

	var req versionLineRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	line, err := h.Service.AddLine(c.Request().Context(), user.ID, ledgerID, versionID, budget.LineInput{
		CategoryID:      req.CategoryID,
		Percent:         req.Percent,
		IncludeChildren: req.IncludeChildren,
	})
	if err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.JSON(http.StatusCreated, toLineResponse(line, ledgerID))
}

func (h *BudgetHandler) UpdateLine(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	lineID := c.Param("lineId")
	if ledgerID == "" || lineID == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"line_id": "required"})
	}

	var req versionLineRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	line, err := h.Service.UpdateLine(c.Request().Context(), user.ID, ledgerID, lineID, req.Percent, req.IncludeChildren)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.JSON(http.StatusOK, toLineResponse(line, ledgerID))
}

func (h *BudgetHandler) DeleteLine(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	lineID := c.Param("lineId")
	if ledgerID == "" || lineID == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"line_id": "required"})
	}

	if err := h.Service.DeleteLine(c.Request().Context(), user.ID, ledgerID, lineID); err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *BudgetHandler) Monthly(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	if ledgerID == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"ledger_id": "required"})
	}

	monthValue := strings.TrimSpace(c.QueryParam("month"))
	if monthValue == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"month": "required"})
	}
	month, err := httpx.ParseMonth(monthValue)
	if err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"month": "invalid"})
	}

	summary, err := h.Service.MonthlySummary(c.Request().Context(), user.ID, ledgerID, month)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	return c.JSON(http.StatusOK, toMonthlyResponse(summary, ledgerID))
}

func (h *BudgetHandler) PeriodSummary(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	if ledgerID == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"ledger_id": "required"})
	}

	fromValue := strings.TrimSpace(c.QueryParam("from"))
	toValue := strings.TrimSpace(c.QueryParam("to"))
	if fromValue == "" || toValue == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"range": "required"})
	}
	from, err := httpx.ParseMonth(fromValue)
	if err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"from": "invalid"})
	}
	to, err := httpx.ParseMonth(toValue)
	if err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"to": "invalid"})
	}

	summary, err := h.Service.PeriodSummary(c.Request().Context(), user.ID, ledgerID, from, to)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	response := periodResponse{
		LedgerID:           ledgerID,
		From:               summary.From,
		To:                 summary.To,
		OutsideBudgetCents: summary.OutsideBudgetCents,
		TotalBudgetedCents: summary.TotalBudgetedCents,
		TotalSpentCents:    summary.TotalSpentCents,
		TotalDeltaCents:    summary.TotalDeltaCents,
		Months:             []monthlyResponse{},
		Categories:         []periodLineResponse{},
	}

	for _, month := range summary.Months {
		response.Months = append(response.Months, toMonthlyResponse(month, ledgerID))
	}
	for _, line := range summary.Categories {
		response.Categories = append(response.Categories, periodLineResponse{
			CategoryID:       line.CategoryID,
			BudgetLimitCents: line.BudgetLimitCents,
			SpentActualCents: line.SpentActualCents,
			DeltaCents:       line.DeltaCents,
			UsagePct:         line.UsagePct,
		})
	}

	return c.JSON(http.StatusOK, response)
}

func toPlanResponse(plan budget.Plan) planResponse {
	return planResponse{
		ID:        plan.ID,
		LedgerID:  plan.LedgerID,
		Name:      plan.Name,
		CreatedAt: plan.CreatedAt,
		UpdatedAt: plan.UpdatedAt,
	}
}

func toVersionResponse(version budget.Version, ledgerID string) versionResponse {
	resp := versionResponse{
		ID:                 version.ID,
		LedgerID:           ledgerID,
		PlanID:             version.PlanID,
		EffectiveFromMonth: version.EffectiveFromMonth,
		CreatedByUserID:    version.CreatedByUserID,
		CreatedAt:          version.CreatedAt,
		UpdatedAt:          version.UpdatedAt,
	}

	if len(version.Lines) > 0 {
		lines := make([]lineResponse, 0, len(version.Lines))
		for _, line := range version.Lines {
			lines = append(lines, toLineResponse(line, ledgerID))
		}
		resp.Lines = lines
	}
	return resp
}

func toLineResponse(line budget.Line, ledgerID string) lineResponse {
	return lineResponse{
		ID:              line.ID,
		LedgerID:        ledgerID,
		VersionID:       line.VersionID,
		CategoryID:      line.CategoryID,
		Percent:         line.Percent,
		IncludeChildren: line.IncludeChildren,
		CreatedAt:       line.CreatedAt,
		UpdatedAt:       line.UpdatedAt,
	}
}

func toMonthlyResponse(summary budget.MonthlySummary, ledgerID string) monthlyResponse {
	response := monthlyResponse{
		LedgerID:           ledgerID,
		Month:              summary.Month,
		IncomeBaseCents:    summary.IncomeBaseCents,
		OutsideBudgetCents: summary.OutsideBudgetCents,
		Lines:              []monthlyLineResponse{},
	}

	if summary.Version != nil {
		version := toVersionResponse(*summary.Version, ledgerID)
		response.Version = &version
	}

	for _, line := range summary.Lines {
		response.Lines = append(response.Lines, monthlyLineResponse{
			CategoryID:       line.CategoryID,
			Percent:          line.Percent,
			IncludeChildren:  line.IncludeChildren,
			BudgetLimitCents: line.BudgetLimitCents,
			SpentActualCents: line.SpentActualCents,
			DeltaCents:       line.DeltaCents,
			UsagePct:         line.UsagePct,
		})
	}

	return response
}
