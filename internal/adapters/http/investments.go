package httpapi

import (
	"context"
	"net/http"
	"strings"
	"time"

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

type investmentRequest struct {
	AmountCents int64   `json:"amount_cents"`
	OccurredAt  string  `json:"occurred_at"`
	Memo        *string `json:"memo"`
}

type investmentSummaryResponse struct {
	From               time.Time `json:"from"`
	To                 time.Time `json:"to"`
	TotalContributions int64     `json:"total_contributions"`
	TotalRedemptions   int64     `json:"total_redemptions"`
	TotalEarnings      int64     `json:"total_earnings"`
	TotalLosses        int64     `json:"total_losses"`
	NetVariation       int64     `json:"net_variation"`
}

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
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	if ledgerID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"ledger_id": "required"})
	}

	fromValue := strings.TrimSpace(c.QueryParam("from"))
	toValue := strings.TrimSpace(c.QueryParam("to"))
	if fromValue == "" || toValue == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"range": "required"})
	}

	from, err := parseDateTime(fromValue)
	if err != nil {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"from": "invalid"})
	}
	to, err := parseDateTime(toValue)
	if err != nil {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"to": "invalid"})
	}

	summary, err := h.Service.Summary(c.Request().Context(), user.ID, ledgerID, from, to)
	if err != nil {
		return WriteAppError(c, err)
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
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	if ledgerID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"ledger_id": "required"})
	}

	var req investmentRequest
	if err := c.Bind(&req); err != nil {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	occurredAt, err := parseDateTime(req.OccurredAt)
	if err != nil {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"occurred_at": "invalid"})
	}

	created, err := fn(c.Request().Context(), user.ID, ledgerID, req.AmountCents, occurredAt, req.Memo)
	if err != nil {
		return WriteAppError(c, err)
	}
	return c.JSON(http.StatusCreated, toTransactionResponse(created))
}
