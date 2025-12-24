package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/seuuser/cashflow/internal/domain/cashflow"
)

type CashFlowHandler struct {
	service cashflow.Service
}

func NewCashFlowHandler(service cashflow.Service) *CashFlowHandler {
	return &CashFlowHandler{service: service}
}

type CreateCashFlowRequest struct {
	Date       string  `json:"date"` // YYYY-MM-DD
	CategoryID int32   `json:"category_id"`
	Direction  string  `json:"direction"`
	Title      string  `json:"title"`
	Amount     float64 `json:"amount"`
	IsFixed    bool    `json:"is_fixed"`
}

func (h *CashFlowHandler) Create(c echo.Context) error {
	var req CreateCashFlowRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}

	parsedDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid date format, use YYYY-MM-DD"})
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
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("failed to create cash flow: %v", err)})
	}

	return c.JSON(http.StatusCreated, created)
}

func (h *CashFlowHandler) ListByMonth(c echo.Context) error {
	monthStr := c.QueryParam("month")
	if monthStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "month parameter is required (YYYY-MM-DD)"})
	}

	parsedMonth, err := time.Parse("2006-01-02", monthStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid month format, use YYYY-MM-DD"})
	}

	list, err := h.service.ListCashFlows(c.Request().Context(), parsedMonth)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list cash flows"})
	}

	return c.JSON(http.StatusOK, list)
}

func RegisterCashFlowRoutes(e *echo.Echo, h *CashFlowHandler) {
	g := e.Group("/cashflows")
	g.POST("", h.Create)
	g.GET("", h.ListByMonth)
	g.POST("/copy-fixed", h.CopyFixed)
	g.GET("/summary", h.MonthlySummary)
	g.GET("/category-summary", h.CategorySummary)
}

func (h *CashFlowHandler) MonthlySummary(c echo.Context) error {
	monthStr := c.QueryParam("month")
	if monthStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "month is required"})
	}
	parsedMonth, err := time.Parse("2006-01-02", monthStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid month format"})
	}

	summary, err := h.service.GetMonthlySummary(c.Request().Context(), parsedMonth)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get summary"})
	}
	return c.JSON(http.StatusOK, summary)
}

func (h *CashFlowHandler) CategorySummary(c echo.Context) error {
	monthStr := c.QueryParam("month")
	if monthStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "month is required"})
	}
	parsedMonth, err := time.Parse("2006-01-02", monthStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid month format"})
	}

	summary, err := h.service.GetCategorySummary(c.Request().Context(), parsedMonth)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get category summary"})
	}
	return c.JSON(http.StatusOK, summary)
}

func (h *CashFlowHandler) CopyFixed(c echo.Context) error {
	type Request struct {
		FromMonth string `json:"from_month"`
		ToMonth   string `json:"to_month"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}

	from, err := time.Parse("2006-01-02", req.FromMonth)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid from_month format"})
	}
	to, err := time.Parse("2006-01-02", req.ToMonth)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid to_month format"})
	}

	count, err := h.service.CopyFixedExpenses(c.Request().Context(), from, to)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("failed to copy expenses: %v", err)})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"copied_count": count})
}
