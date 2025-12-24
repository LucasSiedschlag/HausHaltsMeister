package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/seuuser/cashflow/internal/domain/installment"
)

type InstallmentHandler struct {
	service installment.Service
}

func NewInstallmentHandler(service installment.Service) *InstallmentHandler {
	return &InstallmentHandler{service: service}
}

type CreateInstallmentRequest struct {
	Description     string  `json:"description"`
	TotalAmount     float64 `json:"total_amount"`
	Count           int32   `json:"count"`
	CategoryID      int32   `json:"category_id"`
	PaymentMethodID int32   `json:"payment_method_id"`
	PurchaseDate    string  `json:"purchase_date"` // YYYY-MM-DD
}

func (h *InstallmentHandler) Create(c echo.Context) error {
	var req CreateInstallmentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}

	pDate, err := time.Parse("2006-01-02", req.PurchaseDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid date format"})
	}

	plan, err := h.service.CreateInstallmentPurchase(
		c.Request().Context(),
		req.Description,
		req.TotalAmount,
		req.Count,
		req.CategoryID,
		req.PaymentMethodID,
		pDate,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("failed to create installment plan: %v", err)})
	}

	return c.JSON(http.StatusCreated, plan)
}

func RegisterInstallmentRoutes(e *echo.Echo, h *InstallmentHandler) {
	g := e.Group("/installments")
	g.POST("", h.Create)
}
