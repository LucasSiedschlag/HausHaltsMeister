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
	if req.PaymentMethodID <= 0 {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "payment_method_id is required"})
	}

	parsedDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid date format, use YYYY-MM-DD"})
	}

	created, err := h.service.CreateCashFlow(
		c.Request().Context(),
		parsedDate,
		req.CategoryID,
		req.PaymentMethodID,
		req.Direction,
		req.Title,
		req.Amount,
		req.IsFixed,
	)
	if err != nil {
		if err == cashflow.ErrDirectionMismatch || err == cashflow.ErrCategoryNotFound || err == cashflow.ErrFixedCategoryOnly || err == cashflow.ErrFixedDirectionOnly {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		}
		if err == cashflow.ErrPaymentMethodMissing {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
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

	var direction *string
	if dir := c.QueryParam("direction"); dir != "" {
		direction = &dir
	}

	var isFixed *bool
	if fixed := c.QueryParam("is_fixed"); fixed != "" {
		value := fixed == "true" || fixed == "1"
		isFixed = &value
	}

	list, err := h.service.ListCashFlows(c.Request().Context(), parsedMonth, direction, isFixed)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to list cash flows"})
	}

	resp := make([]dto.CashFlowResponse, len(list))
	for i, cf := range list {
		resp[i] = toCashFlowResponse(cf)
	}

	return c.JSON(http.StatusOK, resp)
}

// Update updates an existing cash flow entry.
// @Summary Atualizar Lançamento
// @Description Updates a cash flow entry (income or expense).
// @Tags CashFlows
// @Accept json
// @Produce json
// @Param id path int true "CashFlow ID"
// @Param payload body dto.CreateCashFlowRequest true "CashFlow Payload"
// @Success 200 {object} dto.CashFlowResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /cashflows/{id} [put]
func (h *CashFlowHandler) Update(c echo.Context) error {
	idStr := c.Param("id")
	var id int32
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id format"})
	}

	var req dto.CreateCashFlowRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid payload"})
	}
	if req.PaymentMethodID <= 0 {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "payment_method_id is required"})
	}

	parsedDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid date format, use YYYY-MM-DD"})
	}

	updated, err := h.service.UpdateCashFlow(
		c.Request().Context(),
		id,
		parsedDate,
		req.CategoryID,
		req.PaymentMethodID,
		req.Direction,
		req.Title,
		req.Amount,
		req.IsFixed,
	)
	if err != nil {
		if err == cashflow.ErrCashFlowNotFound {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		}
		if err == cashflow.ErrDirectionMismatch || err == cashflow.ErrCategoryNotFound || err == cashflow.ErrFixedCategoryOnly || err == cashflow.ErrFixedDirectionOnly {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		}
		if err == cashflow.ErrPaymentMethodMissing {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to update cash flow"})
	}

	return c.JSON(http.StatusOK, toCashFlowResponse(updated))
}

// Delete removes a cash flow entry.
// @Summary Excluir Lançamento
// @Description Deletes a cash flow entry.
// @Tags CashFlows
// @Accept json
// @Produce json
// @Param id path int true "CashFlow ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} dto.ErrorResponse
// @Router /cashflows/{id} [delete]
func (h *CashFlowHandler) Delete(c echo.Context) error {
	idStr := c.Param("id")
	var id int32
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id format"})
	}

	if err := h.service.DeleteCashFlow(c.Request().Context(), id); err != nil {
		if err == cashflow.ErrCashFlowNotFound {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to delete cash flow"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "deleted"})
}

// Reverse creates a reversal entry for a cash flow entry.
// @Summary Estornar Lançamento
// @Description Creates a reversal transaction for a cash flow entry.
// @Tags CashFlows
// @Accept json
// @Produce json
// @Param id path int true "CashFlow ID"
// @Success 201 {object} dto.CashFlowResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /cashflows/{id}/reverse [post]
func (h *CashFlowHandler) Reverse(c echo.Context) error {
	idStr := c.Param("id")
	var id int32
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id format"})
	}

	reversed, err := h.service.ReverseCashFlow(c.Request().Context(), id)
	if err != nil {
		if err == cashflow.ErrCashFlowNotFound {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		}
		if err == cashflow.ErrReversalNotAllowed {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to reverse cash flow"})
	}

	return c.JSON(http.StatusCreated, toCashFlowResponse(reversed))
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
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
	g.POST("/:id/reverse", h.Reverse)
	g.POST("/copy-fixed", h.CopyFixed)
	g.GET("/summary", h.MonthlySummary)
	g.GET("/category-summary", h.CategorySummary)
}

func toCashFlowResponse(cf *cashflow.CashFlow) dto.CashFlowResponse {
	return dto.CashFlowResponse{
		ID:                cf.ID,
		Date:              cf.Date.Format("2006-01-02"),
		CategoryID:        cf.CategoryID,
		CategoryName:      cf.CategoryName,
		PaymentMethodID:   cf.PaymentMethodID,
		PaymentMethodName: cf.PaymentMethodName,
		Direction:         cf.Direction,
		Title:             cf.Title,
		Amount:            cf.Amount,
		IsFixed:           cf.IsFixed,
		ReversalOfEntryID: cf.ReversalOfEntryID,
	}
}
