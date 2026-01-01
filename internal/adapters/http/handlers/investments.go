package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/dto"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/httpx"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/investments"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/journal"
	"github.com/labstack/echo/v4"
)

type InvestmentsHandler struct {
	Service InvestmentsService
}

type InvestmentsService interface {
	Contribution(ctx context.Context, userID, ledgerID string, amount int64, occurredAt time.Time, memo *string) (journal.Transaction, error)
	Redemption(ctx context.Context, userID, ledgerID string, amount int64, occurredAt time.Time, memo *string) (journal.Transaction, error)
	Earnings(ctx context.Context, userID, ledgerID string, amount int64, occurredAt time.Time, memo *string) (journal.Transaction, error)
	Summary(ctx context.Context, userID, ledgerID string, from, to time.Time) (investments.Summary, error)
}

type investmentRequest = dto.InvestmentRequest
type investmentSummaryResponse = dto.InvestmentSummaryResponse

func (h *InvestmentsHandler) Register(g *echo.Group) {
	g.POST("/contributions", h.Contribution)
	g.POST("/redemptions", h.Redemption)
	g.POST("/earnings", h.Earnings)
	g.GET("/summary", h.Summary)
}

func (h *InvestmentsHandler) Contribution(c echo.Context) error {
	return h.handleEntry(c, h.Service.Contribution)
}

func (h *InvestmentsHandler) Redemption(c echo.Context) error {
	return h.handleEntry(c, h.Service.Redemption)
}

func (h *InvestmentsHandler) Earnings(c echo.Context) error {
	return h.handleEntry(c, h.Service.Earnings)
}

func (h *InvestmentsHandler) Summary(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	if ledgerID == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"ledger_id": "required"})
	}

	fromValue := strings.TrimSpace(c.QueryParam("from"))
	toValue := strings.TrimSpace(c.QueryParam("to"))
	if fromValue == "" || toValue == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"range": "required"})
	}

	from, err := httpx.ParseDateTime(fromValue)
	if err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"from": "invalid"})
	}
	to, err := httpx.ParseDateTime(toValue)
	if err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"to": "invalid"})
	}

	summary, err := h.Service.Summary(c.Request().Context(), user.ID, ledgerID, from, to)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	return c.JSON(http.StatusOK, investmentSummaryResponse{
		From:               summary.From,
		To:                 summary.To,
		TotalContributions: summary.TotalContributions,
		TotalRedemptions:   summary.TotalRedemptions,
		TotalEarnings:      summary.TotalEarnings,
		TotalLosses:        summary.TotalLosses,
		NetVariation:       summary.NetVariation,
	})
}

func (h *InvestmentsHandler) handleEntry(c echo.Context, fn func(context.Context, string, string, int64, time.Time, *string) (journal.Transaction, error)) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	if ledgerID == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"ledger_id": "required"})
	}

	var req investmentRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	occurredAt, err := httpx.ParseDateTime(req.OccurredAt)
	if err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"occurred_at": "invalid"})
	}

	created, err := fn(c.Request().Context(), user.ID, ledgerID, req.AmountCents, occurredAt, req.Memo)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.JSON(http.StatusCreated, toTransactionResponse(created))
}
