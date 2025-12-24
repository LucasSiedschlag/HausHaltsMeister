package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/seuuser/cashflow/internal/adapters/http/dto"
	"github.com/seuuser/cashflow/internal/domain/payment"
)

type PaymentHandler struct {
	service payment.Service
}

func NewPaymentHandler(service payment.Service) *PaymentHandler {
	return &PaymentHandler{service: service}
}

func (h *PaymentHandler) Create(c echo.Context) error {
	var req dto.CreatePaymentMethodRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid payload"})
	}

	created, err := h.service.CreatePaymentMethod(
		c.Request().Context(),
		req.Name, req.Kind, req.BankName, req.ClosingDay, req.DueDay,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: fmt.Sprintf("failed to create payment method: %v", err)})
	}

	return c.JSON(http.StatusCreated, toPaymentMethodResponse(created))
}

func (h *PaymentHandler) List(c echo.Context) error {
	list, err := h.service.ListPaymentMethods(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to list payment methods"})
	}

	resp := make([]dto.PaymentMethodResponse, len(list))
	for i, m := range list {
		resp[i] = toPaymentMethodResponse(&m)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *PaymentHandler) GetInvoice(c echo.Context) error {
	idStr := c.Param("id")
	var id int32
	fmt.Sscanf(idStr, "%d", &id)

	monthStr := c.QueryParam("month")
	if monthStr == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "month parameter is required (YYYY-MM-DD)"})
	}

	parsedMonth, err := time.Parse("2006-01-02", monthStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid month format, use YYYY-MM-DD"})
	}

	invoice, err := h.service.GetInvoice(c.Request().Context(), id, parsedMonth)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to get invoice"})
	}

	entries := make([]dto.InvoiceEntryResponse, len(invoice.Entries))
	for i, e := range invoice.Entries {
		entries[i] = dto.InvoiceEntryResponse{
			CashFlowID:   e.CashFlowID,
			Date:         e.Date.Format("2006-01-02"),
			Title:        e.Title,
			Amount:       e.Amount,
			CategoryName: e.CategoryName,
		}
	}

	return c.JSON(http.StatusOK, dto.InvoiceResponse{
		PaymentMethodID: invoice.PaymentMethodID,
		Month:           invoice.Month.Format("2006-01-02"),
		Total:           invoice.Total,
		Entries:         entries,
	})
}

func RegisterPaymentRoutes(e *echo.Echo, h *PaymentHandler) {
	g := e.Group("/payment-methods")
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:id/invoice", h.GetInvoice)
}

func toPaymentMethodResponse(m *payment.PaymentMethod) dto.PaymentMethodResponse {
	return dto.PaymentMethodResponse{
		ID:         m.ID,
		Name:       m.Name,
		Kind:       m.Kind,
		BankName:   m.BankName,
		ClosingDay: m.ClosingDay,
		DueDay:     m.DueDay,
		IsActive:   m.IsActive,
	}
}
