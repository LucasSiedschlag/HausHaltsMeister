package http

import (
	"fmt"
	"math"
	"net/http"
	"strings"
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
		Month:       summary.Month.Format("2006-01-02"),
		TotalIncome: summary.TotalIncome,
		Items:       items,
	})
}

// SetItem sets a budget item for a specific category and month.
// @Summary Definir Item de Orçamento
// @Description Sets or updates the budget item (percentual or absoluto) for a specific category in a given month.
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

	mode, plannedAmount, targetPercent, err := normalizeBudgetInput(req.Mode, req.PlannedAmount, req.TargetPercent)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
	}

	updated, err := h.service.SetBudgetItem(c.Request().Context(), parsedMonth, req.CategoryID, mode, plannedAmount, targetPercent)
	if err != nil {
		if err == budget.ErrInvalidCategory {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		}
		if err == budget.ErrInvalidAmount || err == budget.ErrInvalidPercent || err == budget.ErrInvalidMode {
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

	mode, plannedAmount, targetPercent, err := normalizeBudgetInput(req.Mode, req.PlannedAmount, req.TargetPercent)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
	}

	err = h.service.SetBudgetBatch(c.Request().Context(), start, end, req.CategoryID, mode, plannedAmount, targetPercent)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "success"})
}

// BulkUpdateItems updates all budget items for a given month.
// @Summary Atualizar Orçamento em Lote
// @Description Updates all budget items for a month and validates 100% distribution.
// @Tags Budgets
// @Accept json
// @Produce json
// @Param month path string true "Reference Month (YYYY-MM-DD)" format(date) example(2024-03-01)
// @Param payload body dto.BulkBudgetItemsRequest true "Bulk Budget Items Payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /budgets/{month}/items [put]
func (h *BudgetHandler) BulkUpdateItems(c echo.Context) error {
	monthStr := c.Param("month")
	parsedMonth, err := time.Parse("2006-01-02", monthStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid month format"})
	}

	var req dto.BulkBudgetItemsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid payload"})
	}
	if len(req.Items) == 0 {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "items are required"})
	}

	total := 0.0
	seen := make(map[int32]struct{}, len(req.Items))
	for _, item := range req.Items {
		if _, exists := seen[item.CategoryID]; exists {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "duplicate category_id"})
		}
		seen[item.CategoryID] = struct{}{}
		total += item.TargetPercent
	}
	if math.Abs(total-100.0) > 0.001 {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "total percent must be 100"})
	}

	for _, item := range req.Items {
		_, err := h.service.SetBudgetItem(c.Request().Context(), parsedMonth, item.CategoryID, budget.ModePercentOfIncome, 0, item.TargetPercent)
		if err != nil {
			if err == budget.ErrInvalidCategory || err == budget.ErrInvalidPercent || err == budget.ErrInvalidMode {
				return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
			}
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to set budget items"})
		}
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "success"})
}

// UpdateItem updates a budget item.
// @Summary Atualizar Item de Orçamento
// @Description Updates the budget item (percentual ou absoluto) for a specific budget item.
// @Tags Budgets
// @Accept json
// @Produce json
// @Param id path int true "Budget Item ID"
// @Param payload body dto.UpdateBudgetItemRequest true "Budget Item Payload"
// @Success 200 {object} dto.BudgetItemResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /budgets/items/{id} [put]
func (h *BudgetHandler) UpdateItem(c echo.Context) error {
	idStr := c.Param("id")
	var id int32
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id format"})
	}

	var req dto.UpdateBudgetItemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid payload"})
	}
	mode, plannedAmount, targetPercent, err := normalizeBudgetInput(req.Mode, req.PlannedAmount, req.TargetPercent)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
	}

	updated, err := h.service.UpdateBudgetItem(c.Request().Context(), id, mode, plannedAmount, targetPercent)
	if err != nil {
		if err == budget.ErrInvalidAmount {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		}
		if err == budget.ErrInvalidPercent || err == budget.ErrInvalidMode {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		}
		if err == budget.ErrBudgetItemNotFound {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to update budget item"})
	}

	return c.JSON(http.StatusOK, toBudgetItemResponse(updated))
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
		TargetPercent:  it.TargetPercent,
	}
}

func normalizeBudgetInput(mode string, plannedAmount *float64, targetPercent *float64) (string, float64, float64, error) {
	normalizedMode := strings.ToUpper(strings.TrimSpace(mode))
	if targetPercent != nil {
		normalizedMode = budget.ModePercentOfIncome
	} else if normalizedMode == "" {
		if plannedAmount != nil {
			normalizedMode = budget.ModeAbsolute
		} else {
			normalizedMode = budget.ModePercentOfIncome
		}
	}

	if normalizedMode == "PERCENT" {
		normalizedMode = budget.ModePercentOfIncome
	}

	switch normalizedMode {
	case budget.ModePercentOfIncome:
		if targetPercent == nil {
			return "", 0, 0, fmt.Errorf("target_percent is required for percent mode")
		}
		return normalizedMode, 0, *targetPercent, nil
	case budget.ModeAbsolute:
		if plannedAmount == nil {
			return "", 0, 0, fmt.Errorf("planned_amount is required for absolute mode")
		}
		return normalizedMode, *plannedAmount, 0, nil
	default:
		return "", 0, 0, budget.ErrInvalidMode
	}
}

func RegisterBudgetRoutes(e *echo.Echo, h *BudgetHandler) {
	g := e.Group("/budgets")
	// GET /budgets/:month/summary
	g.GET("/:month/summary", h.GetSummary)
	// POST /budgets/:month/items
	g.POST("/:month/items", h.SetItem)
	// PUT /budgets/:month/items
	g.PUT("/:month/items", h.BulkUpdateItems)
	// PUT /budgets/items/:id
	g.PUT("/items/:id", h.UpdateItem)
	// POST /budgets/batch
	g.POST("/batch", h.SetBatch)
}
