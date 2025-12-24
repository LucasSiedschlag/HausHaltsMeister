package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/dto"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/budget"
	"github.com/labstack/echo/v4"
)

type BudgetHandler struct {
	service budget.Service
}

func NewBudgetHandler(service budget.Service) *BudgetHandler {
	return &BudgetHandler{service: service}
}

// GetSummary returns the budget summary for a given month.
// @Summary Visualizar Orçamento (Planned vs Actual)
// @Description Returns the budget summary comparing planned vs actual expenses for the month.
// @Tags Budgets
// @Accept json
// @Produce json
// @Param month path string true "Reference Month (YYYY-MM-DD)" format(date) example(2024-03-01)
// @Success 200 {object} dto.BudgetSummaryResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /budgets/{month}/summary [get]
func (h *BudgetHandler) GetSummary(c echo.Context) error {
	monthStr := c.Param("month") // Expects YYYY-MM-DD

	parsedMonth, err := time.Parse("2006-01-02", monthStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid month format"})
	}

	summary, err := h.service.GetBudgetSummary(c.Request().Context(), parsedMonth)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to get budget summary"})
	}

	items := make([]dto.BudgetItemResponse, len(summary.Items))
	for i, it := range summary.Items {
		items[i] = toBudgetItemResponse(&it)
	}

	return c.JSON(http.StatusOK, dto.BudgetSummaryResponse{
		Month: summary.Month.Format("2006-01-02"),
		Items: items,
	})
}

// SetItem sets a budget item for a specific category and month.
// @Summary Definir Item de Orçamento
// @Description Sets or updates the planned amount for a specific category in a given month.
// @Tags Budgets
// @Accept json
// @Produce json
// @Param month path string true "Reference Month (YYYY-MM-DD)" format(date) example(2024-03-01)
// @Param payload body dto.SetBudgetItemRequest true "Budget Item Payload"
// @Success 200 {object} dto.BudgetItemResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /budgets/{month}/items [post]
func (h *BudgetHandler) SetItem(c echo.Context) error {
	monthStr := c.Param("month")
	parsedMonth, err := time.Parse("2006-01-02", monthStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid month format"})
	}

	var req dto.SetBudgetItemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid payload"})
	}

	updated, err := h.service.SetBudgetItem(c.Request().Context(), parsedMonth, req.CategoryID, req.PlannedAmount)
	if err != nil {
		if err == budget.ErrInvalidCategory {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: fmt.Sprintf("failed to set budget item: %v", err)})
	}

	return c.JSON(http.StatusOK, toBudgetItemResponse(updated))
}

// SetBatch sets budget items for a range of months.
// @Summary Definir Orçamento em Lote
// @Description Sets the planned amount for a specific category across a range of months.
// @Tags Budgets
// @Accept json
// @Produce json
// @Param payload body dto.SetBudgetBatchRequest true "Batch Budget Payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /budgets/batch [post]
func (h *BudgetHandler) SetBatch(c echo.Context) error {
	var req dto.SetBudgetBatchRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid payload"})
	}

	start, err := time.Parse("2006-01-02", req.StartMonth)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid start_month format"})
	}

	end, err := time.Parse("2006-01-02", req.EndMonth)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid end_month format"})
	}

	err = h.service.SetBudgetBatch(c.Request().Context(), start, end, req.CategoryID, req.PlannedAmount)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "success"})
}

func toBudgetItemResponse(it *budget.BudgetItem) dto.BudgetItemResponse {
	return dto.BudgetItemResponse{
		ID:             it.ID,
		BudgetPeriodID: it.BudgetPeriodID,
		CategoryID:     it.CategoryID,
		CategoryName:   it.CategoryName,
		Mode:           it.Mode,
		PlannedAmount:  it.PlannedAmount,
		ActualAmount:   it.ActualAmount,
	}
}

func RegisterBudgetRoutes(e *echo.Echo, h *BudgetHandler) {
	g := e.Group("/budgets")
	// GET /budgets/:month/summary
	g.GET("/:month/summary", h.GetSummary)
	// POST /budgets/:month/items
	g.POST("/:month/items", h.SetItem)
	// POST /budgets/batch
	g.POST("/batch", h.SetBatch)
}
