package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/seuuser/cashflow/internal/domain/budget"
)

type BudgetHandler struct {
	service budget.Service
}

func NewBudgetHandler(service budget.Service) *BudgetHandler {
	return &BudgetHandler{service: service}
}

type SetBudgetItemRequest struct {
	CategoryID    int32   `json:"category_id"`
	PlannedAmount float64 `json:"planned_amount"`
}

func (h *BudgetHandler) GetSummary(c echo.Context) error {
	monthStr := c.Param("month") // Expects YYYY-MM-DD (e.g., 2023-12-01)

	parsedMonth, err := time.Parse("2006-01-02", monthStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid month format"})
	}

	summary, err := h.service.GetBudgetSummary(c.Request().Context(), parsedMonth)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get budget summary"})
	}

	return c.JSON(http.StatusOK, summary)
}

func (h *BudgetHandler) SetItem(c echo.Context) error {
	monthStr := c.Param("month")
	parsedMonth, err := time.Parse("2006-01-02", monthStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid month format"})
	}

	var req SetBudgetItemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}

	updated, err := h.service.SetBudgetItem(c.Request().Context(), parsedMonth, req.CategoryID, req.PlannedAmount)
	if err != nil {
		if err == budget.ErrInvalidCategory {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("failed to set budget item: %v", err)})
	}

	return c.JSON(http.StatusOK, updated)
}

func RegisterBudgetRoutes(e *echo.Echo, h *BudgetHandler) {
	g := e.Group("/budgets")
	// GET /budgets/:month/summary
	g.GET("/:month/summary", h.GetSummary)
	// POST /budgets/:month/items
	g.POST("/:month/items", h.SetItem)
}
