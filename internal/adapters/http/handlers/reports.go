package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/dto"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/httpx"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/reports"
	"github.com/labstack/echo/v4"
)

type ReportsHandler struct {
	Service ReportsService
}

type ReportsService interface {
	Balances(ctx context.Context, userID, ledgerID string, month time.Time) (reports.BalanceReport, error)
	CategorySummary(ctx context.Context, userID, ledgerID string, from, to time.Time) (reports.CategoryReport, error)
	Cashflow(ctx context.Context, userID, ledgerID string, from, to time.Time) (reports.CashflowReport, error)
}

type balanceResponse = dto.BalanceResponse
type balanceItemResponse = dto.BalanceItemResponse
type categorySummaryResponse = dto.CategorySummaryResponse
type categorySummaryItemResponse = dto.CategorySummaryItemResponse
type cashflowResponse = dto.CashflowResponse
type cashflowItemResponse = dto.CashflowItemResponse

func (h *ReportsHandler) Register(g *echo.Group) {
	g.GET("/balances", h.Balances)
	g.GET("/categories", h.Categories)
	g.GET("/cashflow", h.Cashflow)
}

func (h *ReportsHandler) Balances(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	if ledgerID == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"ledger_id": "required"})
	}
	monthValue := strings.TrimSpace(c.QueryParam("month"))
	if monthValue == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"month": "required"})
	}
	month, err := httpx.ParseMonth(monthValue)
	if err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"month": "invalid"})
	}

	report, err := h.Service.Balances(c.Request().Context(), user.ID, ledgerID, month)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	items := make([]balanceItemResponse, 0, len(report.Items))
	for _, item := range report.Items {
		items = append(items, balanceItemResponse{
			AccountID:    item.AccountID,
			AccountName:  item.AccountName,
			AccountType:  item.AccountType,
			BalanceCents: item.BalanceCents,
		})
	}

	return c.JSON(http.StatusOK, balanceResponse{Month: report.Month, Items: items})
}

func (h *ReportsHandler) Categories(c echo.Context) error {
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

	report, err := h.Service.CategorySummary(c.Request().Context(), user.ID, ledgerID, from, to)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	items := make([]categorySummaryItemResponse, 0, len(report.Items))
	for _, item := range report.Items {
		items = append(items, categorySummaryItemResponse{
			CategoryID: item.CategoryID,
			Name:       item.Name,
			Direction:  item.Direction,
			TotalCents: item.TotalCents,
		})
	}

	return c.JSON(http.StatusOK, categorySummaryResponse{From: report.From, To: report.To, Items: items})
}

func (h *ReportsHandler) Cashflow(c echo.Context) error {
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

	report, err := h.Service.Cashflow(c.Request().Context(), user.ID, ledgerID, from, to)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	items := make([]cashflowItemResponse, 0, len(report.Items))
	for _, item := range report.Items {
		items = append(items, cashflowItemResponse{
			Month:         item.Month,
			TotalInCents:  item.TotalInCents,
			TotalOutCents: item.TotalOutCents,
			NetCents:      item.NetCents,
		})
	}

	return c.JSON(http.StatusOK, cashflowResponse{From: report.From, To: report.To, Items: items})
}
