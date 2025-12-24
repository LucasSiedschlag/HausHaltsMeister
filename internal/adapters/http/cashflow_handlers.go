package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/dto"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/cashflow"
)

type CashFlowHandler struct {
	service cashflow.Service
}

func NewCashFlowHandler(service cashflow.Service) *CashFlowHandler {
	return &CashFlowHandler{service: service}
}

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
