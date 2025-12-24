package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/dto"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/cashflow"
	"github.com/labstack/echo/v4"
)

type CashFlowHandler struct {
	service cashflow.Service
}

func NewCashFlowHandler(service cashflow.Service) *CashFlowHandler {
	return &CashFlowHandler{service: service}
}

// Create creates a new cash flow entry.
// @Summary Criar Lançamento
// @Description Creates a new cash flow (income or expense).
// @Tags CashFlows
// @Accept json
// @Produce json
// @Param payload body dto.CreateCashFlowRequest true "CashFlow Payload"
// @Success 201 {object} dto.CashFlowResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /cashflows [post]
func (h *CashFlowHandler) Create(c echo.Context) error {
	var req dto.CreateCashFlowRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid payload"})
	}

	parsedDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid date format, use YYYY-MM-DD"})
	}

	created, err := h.service.CreateCashFlow(
		c.Request().Context(),
		parsedDate,
		req.CategoryID,
		req.Direction,
		req.Title,
		req.Amount,
		req.IsFixed,
	)
	if err != nil {
		if err == cashflow.ErrDirectionMismatch || err == cashflow.ErrCategoryNotFound {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: fmt.Sprintf("failed to create cash flow: %v", err)})
	}

	return c.JSON(http.StatusCreated, toCashFlowResponse(created))
}

// ListByMonth returns a list of cash flows for a given month.
// @Summary Listar Fluxos (Extrato)
// @Description Returns a list of cash flows for the specified month.
// @Tags CashFlows
// @Accept json
// @Produce json
// @Param month query string true "Reference Month (YYYY-MM-DD)" format(date) example(2024-03-01)
// @Success 200 {array} dto.CashFlowResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /cashflows [get]
func (h *CashFlowHandler) ListByMonth(c echo.Context) error {
	monthStr := c.QueryParam("month")
	if monthStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "month parameter is required (YYYY-MM-DD)"})
	}

	parsedMonth, err := time.Parse("2006-01-02", monthStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid month format, use YYYY-MM-DD"})
	}

	list, err := h.service.ListCashFlows(c.Request().Context(), parsedMonth)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to list cash flows"})
	}

	resp := make([]dto.CashFlowResponse, len(list))
	for i, cf := range list {
		resp[i] = toCashFlowResponse(cf)
	}

	return c.JSON(http.StatusOK, resp)
}

// MonthlySummary returns the financial summary for a given month.
// @Summary Resumo Mensal
// @Description Returns the total income, expense, and balance for the specified month.
// @Tags CashFlows
// @Accept json
// @Produce json
// @Param month query string true "Reference Month (YYYY-MM-DD)" format(date) example(2024-03-01)
// @Success 200 {object} dto.MonthlySummaryResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /cashflows/summary [get]
func (h *CashFlowHandler) MonthlySummary(c echo.Context) error {
	monthStr := c.QueryParam("month")
	if monthStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "month is required"})
	}
	parsedMonth, err := time.Parse("2006-01-02", monthStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid month format"})
	}

	summary, err := h.service.GetMonthlySummary(c.Request().Context(), parsedMonth)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to get summary"})
	}

	return c.JSON(http.StatusOK, dto.MonthlySummaryResponse{
		TotalIncome:  summary.TotalIncome,
		TotalExpense: summary.TotalExpense,
		Balance:      summary.Balance,
	})
}

// CategorySummary returns the financial summary grouped by category for a given month.
// @Summary Resumo por Categoria
// @Description Returns a list of expenses/incomes grouped by category for the specified month.
// @Tags CashFlows
// @Accept json
// @Produce json
// @Param month query string true "Reference Month (YYYY-MM-DD)" format(date) example(2024-03-01)
// @Success 200 {array} dto.CategorySummaryResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /cashflows/category-summary [get]
func (h *CashFlowHandler) CategorySummary(c echo.Context) error {
	monthStr := c.QueryParam("month")
	if monthStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "month is required"})
	}
	parsedMonth, err := time.Parse("2006-01-02", monthStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid month format"})
	}

	summary, err := h.service.GetCategorySummary(c.Request().Context(), parsedMonth)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to get category summary"})
	}

	resp := make([]dto.CategorySummaryResponse, len(summary))
	for i, s := range summary {
		resp[i] = dto.CategorySummaryResponse{
			CategoryName: s.CategoryName,
			Direction:    s.Direction,
			TotalAmount:  s.TotalAmount,
		}
	}

	return c.JSON(http.StatusOK, resp)
}

// CopyFixed copies fixed expenses from one month to another.
// @Summary Copiar Gastos Fixos
// @Description Copies fixed expenses from a source month to a target month.
// @Tags CashFlows
// @Accept json
// @Produce json
// @Param payload body dto.CopyFixedRequest true "Copy Fixed Param"
// @Success 200 {object} map[string]int
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /cashflows/copy-fixed [post]
func (h *CashFlowHandler) CopyFixed(c echo.Context) error {
	var req dto.CopyFixedRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid payload"})
	}

	from, err := time.Parse("2006-01-02", req.FromMonth)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid from_month format"})
	}
	to, err := time.Parse("2006-01-02", req.ToMonth)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid to_month format"})
	}

	count, err := h.service.CopyFixedExpenses(c.Request().Context(), from, to)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: fmt.Sprintf("failed to copy expenses: %v", err)})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"copied_count": count})
}

func RegisterCashFlowRoutes(e *echo.Echo, h *CashFlowHandler) {
	g := e.Group("/cashflows")
	g.POST("", h.Create)
	g.GET("", h.ListByMonth)
	g.POST("/copy-fixed", h.CopyFixed)
	g.GET("/summary", h.MonthlySummary)
	g.GET("/category-summary", h.CategorySummary)
}

func toCashFlowResponse(cf *cashflow.CashFlow) dto.CashFlowResponse {
	return dto.CashFlowResponse{
		ID:         cf.ID,
		Date:       cf.Date.Format("2006-01-02"),
		CategoryID: cf.CategoryID,
		Direction:  cf.Direction,
		Title:      cf.Title,
		Amount:     cf.Amount,
		IsFixed:    cf.IsFixed,
	}
}
