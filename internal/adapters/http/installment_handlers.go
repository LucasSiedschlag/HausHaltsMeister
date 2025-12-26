package http

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/dto"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/cashflow"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/installment"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/payment"
	"github.com/labstack/echo/v4"
)

type InstallmentHandler struct {
	service installment.Service
}

func NewInstallmentHandler(service installment.Service) *InstallmentHandler {
	return &InstallmentHandler{service: service}
}

// Create registers a new installment purchase.
// @Summary Criar Compra Parcelada
// @Description Creates a new purchase that is split into multiple installments.
// @Tags Cards
// @Accept json
// @Produce json
// @Param payload body dto.CreateInstallmentRequest true "Installment Payload"
// @Success 201 {object} dto.InstallmentPlanResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /installments [post]
func (h *InstallmentHandler) Create(c echo.Context) error {
	var req dto.CreateInstallmentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid payload"})
	}
	if strings.TrimSpace(req.Description) == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "description is required"})
	}
	if req.Count < 1 {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "installment count must be at least 1"})
	}
	if req.CategoryID < 1 {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "category_id is required"})
	}
	if req.PaymentMethodID < 1 {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "payment_method_id is required"})
	}

	pDate, err := time.Parse("2006-01-02", req.PurchaseDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid date format"})
	}

	if req.TotalAmount > 0 && req.InstallmentAmount > 0 {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "provide only one: total_amount or installment_amount"})
	}

	amountMode := strings.ToUpper(strings.TrimSpace(req.AmountMode))
	if amountMode == "" {
		if req.InstallmentAmount > 0 {
			amountMode = "INSTALLMENT"
		} else {
			amountMode = "TOTAL"
		}
	}

	totalAmount := req.TotalAmount
	switch amountMode {
	case "TOTAL":
		if totalAmount <= 0 {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "total_amount must be greater than zero"})
		}
	case "INSTALLMENT":
		if req.InstallmentAmount <= 0 {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "installment_amount must be greater than zero"})
		}
		totalAmount = req.InstallmentAmount * float64(req.Count)
	default:
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "amount_mode must be TOTAL or INSTALLMENT"})
	}

	plan, err := h.service.CreateInstallmentPurchase(
		c.Request().Context(),
		req.Description,
		totalAmount,
		req.Count,
		req.CategoryID,
		req.PaymentMethodID,
		pDate,
	)
	if err != nil {
		if errors.Is(err, installment.ErrInvalidTotalAmount) || errors.Is(err, installment.ErrInvalidCount) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		}
		if errors.Is(err, payment.ErrPaymentMethodNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		}
		if errors.Is(err, cashflow.ErrCategoryNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		}
		if errors.Is(err, cashflow.ErrDirectionMismatch) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		}
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
