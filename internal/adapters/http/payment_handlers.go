package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/seuuser/cashflow/internal/domain/payment"
)

type PaymentHandler struct {
	service payment.Service
}

func NewPaymentHandler(service payment.Service) *PaymentHandler {
	return &PaymentHandler{service: service}
}

type CreatePaymentMethodRequest struct {
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	BankName   string `json:"bank_name"`
	ClosingDay *int32 `json:"closing_day"`
	DueDay     *int32 `json:"due_day"`
}

func (h *PaymentHandler) Create(c echo.Context) error {
	var req CreatePaymentMethodRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}

	created, err := h.service.CreatePaymentMethod(
		c.Request().Context(),
		req.Name, req.Kind, req.BankName, req.ClosingDay, req.DueDay,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("failed to create payment method: %v", err)})
	}

	return c.JSON(http.StatusCreated, created)
}

func (h *PaymentHandler) List(c echo.Context) error {
	list, err := h.service.ListPaymentMethods(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list payment methods"})
	}
	return c.JSON(http.StatusOK, list)
}

func (h *PaymentHandler) GetInvoice(c echo.Context) error {
	idStr := c.Param("id")
	var id int32
	fmt.Sscanf(idStr, "%d", &id)

	monthStr := c.QueryParam("month")
	if monthStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "month parameter is required (YYYY-MM-DD)"})
	}

	parsedMonth, err := time.Parse("2006-01-02", monthStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid month format, use YYYY-MM-DD"})
	}

	invoice, err := h.service.GetInvoice(c.Request().Context(), id, parsedMonth)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get invoice"})
	}

	return c.JSON(http.StatusOK, invoice)
}

func RegisterPaymentRoutes(e *echo.Echo, h *PaymentHandler) {
	g := e.Group("/payment-methods")
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:id/invoice", h.GetInvoice)
}
