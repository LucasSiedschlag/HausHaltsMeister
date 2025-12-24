package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/seuuser/cashflow/internal/adapters/http/dto"
	"github.com/seuuser/cashflow/internal/domain/installment"
)

type InstallmentHandler struct {
	service installment.Service
}

func NewInstallmentHandler(service installment.Service) *InstallmentHandler {
	return &InstallmentHandler{service: service}
}

func (h *InstallmentHandler) Create(c echo.Context) error {
	var req dto.CreateInstallmentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid payload"})
	}

	pDate, err := time.Parse("2006-01-02", req.PurchaseDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid date format"})
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
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: fmt.Sprintf("failed to create installment plan: %v", err)})
	}

	return c.JSON(http.StatusCreated, toInstallmentPlanResponse(plan))
}

func RegisterInstallmentRoutes(e *echo.Echo, h *InstallmentHandler) {
	g := e.Group("/installments")
	g.POST("", h.Create)
}

func toInstallmentPlanResponse(p *installment.InstallmentPlan) dto.InstallmentPlanResponse {
	return dto.InstallmentPlanResponse{
		ID:                p.ID,
		Description:       p.Description,
		TotalAmount:       p.TotalAmount,
		InstallmentCount:  p.InstallmentCount,
		InstallmentAmount: p.InstallmentAmount,
		StartMonth:        p.StartMonth.Format("2006-01-02"),
		PaymentMethodID:   p.PaymentMethodID,
	}
}
