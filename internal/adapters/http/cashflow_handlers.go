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
}
