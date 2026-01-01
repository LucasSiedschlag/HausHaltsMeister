package httpapi

import (
	"context"
	"net/http"
	"strings"
	"time"

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

type balanceResponse struct {
	Month time.Time             `json:"month"`
	Items []balanceItemResponse `json:"items"`
}

type balanceItemResponse struct {
	AccountID    string `json:"account_id"`
	AccountName  string `json:"account_name"`
	AccountType  string `json:"account_type"`
	BalanceCents int64  `json:"balance_cents"`
}

type categorySummaryResponse struct {
	From  time.Time                     `json:"from"`
	To    time.Time                     `json:"to"`
	Items []categorySummaryItemResponse `json:"items"`
}

type categorySummaryItemResponse struct {
	CategoryID string `json:"category_id"`
	Name       string `json:"name"`
	Direction  string `json:"direction"`
	TotalCents int64  `json:"total_cents"`
}

type cashflowResponse struct {
	From  time.Time              `json:"from"`
	To    time.Time              `json:"to"`
	Items []cashflowItemResponse `json:"items"`
}

type cashflowItemResponse struct {
	Month         time.Time `json:"month"`
	TotalInCents  int64     `json:"total_in_cents"`
	TotalOutCents int64     `json:"total_out_cents"`
	NetCents      int64     `json:"net_cents"`
}

func (h *ReportsHandler) Register(g *echo.Group) {
	g.GET("/balances", h.Balances)
	g.GET("/categories", h.Categories)
	g.GET("/cashflow", h.Cashflow)
}

func (h *ReportsHandler) Balances(c echo.Context) error {
	user, ok := GetUser(c)
	if !ok {
		return WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	if ledgerID == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"ledger_id": "required"})
	}
	monthValue := strings.TrimSpace(c.QueryParam("month"))
	if monthValue == "" {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"month": "required"})
	}
	month, err := parseMonth(monthValue)
	if err != nil {
		return WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"month": "invalid"})
	}

	report, err := h.Service.Balances(c.Request().Context(), user.ID, ledgerID, month)
	if err != nil {
		return WriteAppError(c, err)
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

	report, err := h.Service.CategorySummary(c.Request().Context(), user.ID, ledgerID, from, to)
	if err != nil {
		return WriteAppError(c, err)
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

	report, err := h.Service.Cashflow(c.Request().Context(), user.ID, ledgerID, from, to)
	if err != nil {
		return WriteAppError(c, err)
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
