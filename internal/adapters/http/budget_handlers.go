package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/dto"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/budget"
)

type BudgetHandler struct {
	service budget.Service
}

func NewBudgetHandler(service budget.Service) *BudgetHandler {
	return &BudgetHandler{service: service}
}

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
