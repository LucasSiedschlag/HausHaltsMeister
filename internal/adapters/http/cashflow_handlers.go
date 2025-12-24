package http

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/seuuser/cashflow/internal/domain/cashflow"
)

type CashFlowHandler struct {
	service *cashflow.Service
}

func RegisterCashFlowRoutes(e *echo.Echo, svc *cashflow.Service) {
	h := &CashFlowHandler{service: svc}
	e.POST("/cashflows", h.CreateCashFlow)
	e.GET("/cashflows", h.ListCashFlows)
}

type createCashFlowRequest struct {
	Date       string  `json:"date"`        // ex.: "2025-12-01"
	CategoryID int64   `json:"category_id"` // ex.: categoria Ganho/Investimento
	Direction  string  `json:"direction"`   // "IN" ou "OUT"
	Title      string  `json:"title"`
	Amount     float64 `json:"amount"`
}

func (h *CashFlowHandler) CreateCashFlow(c echo.Context) error {
	var req createCashFlowRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
	}

	d, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid date"})
	}

	in := cashflow.CreateCashFlowInput{
		Date:       d,
		CategoryID: req.CategoryID,
		Direction:  cashflow.Direction(req.Direction),
		Title:      req.Title,
		Amount:     req.Amount,
	}

	cf, err := h.service.CreateCashFlow(c.Request().Context(), in)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, cf)
}

func (h *CashFlowHandler) ListCashFlows(c echo.Context) error {
	monthStr := c.QueryParam("month") // ex.: "2025-12-01"
	if monthStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "month query param required"})
	}

	month, err := time.Parse("2006-01-02", monthStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid month"})
	}

	list, err := h.service.ListCashFlowsByMonth(c.Request().Context(), month)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, list)
}

