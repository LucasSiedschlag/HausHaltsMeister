package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/dto"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/httpx"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/audit"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/ledger"
	"github.com/labstack/echo/v4"
)

type LedgerHandler struct {
	Service LedgerService
	Audit   audit.Recorder
}

type LedgerService interface {
	ListLedgers(ctx context.Context, userID string) ([]ledger.LedgerWithRole, error)
	GetMembership(ctx context.Context, userID, ledgerID string) (string, error)
	CreateLedger(ctx context.Context, userID, name, currencyCode string, createDefaultAccounts, includeInvestment bool) (ledger.Ledger, error)
	GetLedger(ctx context.Context, userID, ledgerID string) (ledger.LedgerWithRole, error)
	UpdateLedger(ctx context.Context, userID, ledgerID, name string) (ledger.Ledger, error)
	DeleteLedger(ctx context.Context, userID, ledgerID string) error
	ListMembers(ctx context.Context, userID, ledgerID string) ([]ledger.Member, error)
	AddMember(ctx context.Context, userID, ledgerID, memberUserID, role string) (ledger.Member, error)
	AddMemberByEmail(ctx context.Context, userID, ledgerID, email, role string) (ledger.Member, error)
	UpdateMember(ctx context.Context, userID, ledgerID, memberUserID, role string) (ledger.Member, error)
	RemoveMember(ctx context.Context, userID, ledgerID, memberUserID string) error
}

type ledgerCreateRequest = dto.LedgerCreateRequest
type ledgerUpdateRequest = dto.LedgerUpdateRequest
type memberRequest = dto.LedgerMemberRequest
type ledgerResponse = dto.LedgerResponse
type memberResponse = dto.LedgerMemberResponse
type ledgerMeResponse = dto.LedgerMeResponse

type LedgerGuards struct {
	Viewer echo.MiddlewareFunc
	Editor echo.MiddlewareFunc
	Owner  echo.MiddlewareFunc
}

func (h *LedgerHandler) Register(g *echo.Group, guards LedgerGuards) {
	g.GET("", h.ListLedgers)
	g.POST("", h.CreateLedger)

	g.GET("/:ledgerId", h.GetLedger, guards.Viewer)
	g.GET("/:ledgerId/me", h.Me, guards.Viewer)
	g.PATCH("/:ledgerId", h.UpdateLedger, guards.Editor)
	g.DELETE("/:ledgerId", h.DeleteLedger, guards.Owner)

	g.GET("/:ledgerId/members", h.ListMembers, guards.Viewer)
	g.POST("/:ledgerId/members", h.AddMember, guards.Owner)
	g.PATCH("/:ledgerId/members/:userId", h.UpdateMember, guards.Owner)
	g.DELETE("/:ledgerId/members/:userId", h.DeleteMember, guards.Owner)
}

func (h *LedgerHandler) ListLedgers(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}

	items, err := h.Service.ListLedgers(c.Request().Context(), user.ID)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	response := make([]ledgerResponse, 0, len(items))
	for _, item := range items {
		response = append(response, ledgerResponse{
			ID:           item.Ledger.ID,
			OwnerUserID:  item.Ledger.OwnerUserID,
			Name:         item.Ledger.Name,
			CurrencyCode: item.Ledger.CurrencyCode,
			Role:         item.Role,
			CreatedAt:    item.Ledger.CreatedAt,
			UpdatedAt:    item.Ledger.UpdatedAt,
		})
	}

	return c.JSON(http.StatusOK, response)
}

func (h *LedgerHandler) Me(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID, err := httpx.RequireUUIDParam(c, "ledgerId")
	if err != nil {
		return err
	}

	role, err := h.Service.GetMembership(c.Request().Context(), user.ID, ledgerID)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	return c.JSON(http.StatusOK, ledgerMeResponse{
		LedgerID: ledgerID,
		Role:     role,
	})
}

func (h *LedgerHandler) CreateLedger(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}

	var req ledgerCreateRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	createDefaultAccounts := true
	if req.CreateDefaultAccounts != nil {
		createDefaultAccounts = *req.CreateDefaultAccounts
	}
	includeInvestment := false
	if req.IncludeInvestment != nil {
		includeInvestment = *req.IncludeInvestment
	}

	created, err := h.Service.CreateLedger(c.Request().Context(), user.ID, req.Name, req.CurrencyCode, createDefaultAccounts, includeInvestment)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	return c.JSON(http.StatusCreated, ledgerResponse{
		ID:           created.ID,
		OwnerUserID:  created.OwnerUserID,
		Name:         created.Name,
		CurrencyCode: created.CurrencyCode,
		CreatedAt:    created.CreatedAt,
		UpdatedAt:    created.UpdatedAt,
	})
}

func (h *LedgerHandler) GetLedger(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}

	ledgerID, err := httpx.RequireUUIDParam(c, "ledgerId")
	if err != nil {
		return err
	}

	item, err := h.Service.GetLedger(c.Request().Context(), user.ID, ledgerID)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	return c.JSON(http.StatusOK, ledgerResponse{
		ID:           item.Ledger.ID,
		OwnerUserID:  item.Ledger.OwnerUserID,
		Name:         item.Ledger.Name,
		CurrencyCode: item.Ledger.CurrencyCode,
		Role:         item.Role,
		CreatedAt:    item.Ledger.CreatedAt,
		UpdatedAt:    item.Ledger.UpdatedAt,
	})
}

func (h *LedgerHandler) UpdateLedger(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}

	ledgerID, err := httpx.RequireUUIDParam(c, "ledgerId")
	if err != nil {
		return err
	}

	var req ledgerUpdateRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	updated, err := h.Service.UpdateLedger(c.Request().Context(), user.ID, ledgerID, req.Name)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	return c.JSON(http.StatusOK, ledgerResponse{
		ID:           updated.ID,
		OwnerUserID:  updated.OwnerUserID,
		Name:         updated.Name,
		CurrencyCode: updated.CurrencyCode,
		CreatedAt:    updated.CreatedAt,
		UpdatedAt:    updated.UpdatedAt,
	})
}

func (h *LedgerHandler) DeleteLedger(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}

	ledgerID, err := httpx.RequireUUIDParam(c, "ledgerId")
	if err != nil {
		return err
	}

	if err := h.Service.DeleteLedger(c.Request().Context(), user.ID, ledgerID); err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *LedgerHandler) ListMembers(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}

	ledgerID, err := httpx.RequireUUIDParam(c, "ledgerId")
	if err != nil {
		return err
	}

	members, err := h.Service.ListMembers(c.Request().Context(), user.ID, ledgerID)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	response := make([]memberResponse, 0, len(members))
	for _, member := range members {
		response = append(response, memberResponse{
			LedgerID:    member.LedgerID,
			UserID:      member.UserID,
			Role:        member.Role,
			DisplayName: member.DisplayName,
			Email:       member.Email,
			AvatarURL:   member.AvatarURL,
			CreatedAt:   member.CreatedAt,
			UpdatedAt:   member.UpdatedAt,
		})
	}
	return c.JSON(http.StatusOK, response)
}

func (h *LedgerHandler) AddMember(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}

	ledgerID, err := httpx.RequireUUIDParam(c, "ledgerId")
	if err != nil {
		return err
	}

	var req memberRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	memberID := strings.TrimSpace(req.UserID)
	memberEmail := strings.TrimSpace(req.Email)
	if memberID == "" && memberEmail == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{
			"user_id": "required",
		})
	}
	if memberID != "" && memberEmail != "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{
			"user_id": "conflict",
			"email":   "conflict",
		})
	}
	if memberID != "" && !httpx.IsUUID(memberID) {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"user_id": "invalid"})
	}

	var (
		member    ledger.Member
		memberErr error
	)
	if memberEmail != "" {
		member, memberErr = h.Service.AddMemberByEmail(c.Request().Context(), user.ID, ledgerID, memberEmail, req.Role)
	} else {
		member, memberErr = h.Service.AddMember(c.Request().Context(), user.ID, ledgerID, memberID, req.Role)
	}
	if memberErr != nil {
		return httpx.WriteAppError(c, memberErr)
	}

	entityID := member.UserID
	recordAudit(c, h.Audit, audit.Event{
		LedgerID:  ledgerID,
		UserID:    user.ID,
		Action:    "ledger.member.add",
		EntityID:  &entityID,
		IP:        c.RealIP(),
		UserAgent: c.Request().UserAgent(),
	})

	return c.JSON(http.StatusCreated, memberResponse{
		LedgerID:    member.LedgerID,
		UserID:      member.UserID,
		Role:        member.Role,
		DisplayName: member.DisplayName,
		Email:       member.Email,
		AvatarURL:   member.AvatarURL,
		CreatedAt:   member.CreatedAt,
		UpdatedAt:   member.UpdatedAt,
	})
}

func (h *LedgerHandler) UpdateMember(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}

	ledgerID, err := httpx.RequireUUIDParam(c, "ledgerId")
	if err != nil {
		return err
	}
	memberUserID, err := httpx.RequireUUIDParam(c, "userId")
	if err != nil {
		return err
	}

	var req memberRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	member, err := h.Service.UpdateMember(c.Request().Context(), user.ID, ledgerID, memberUserID, req.Role)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	recordAudit(c, h.Audit, audit.Event{
		LedgerID:  ledgerID,
		UserID:    user.ID,
		Action:    "ledger.member.role_change",
		EntityID:  &memberUserID,
		IP:        c.RealIP(),
		UserAgent: c.Request().UserAgent(),
	})

	return c.JSON(http.StatusOK, memberResponse{
		LedgerID:    member.LedgerID,
		UserID:      member.UserID,
		Role:        member.Role,
		DisplayName: member.DisplayName,
		Email:       member.Email,
		AvatarURL:   member.AvatarURL,
		CreatedAt:   member.CreatedAt,
		UpdatedAt:   member.UpdatedAt,
	})
}

func (h *LedgerHandler) DeleteMember(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}

	ledgerID, err := httpx.RequireUUIDParam(c, "ledgerId")
	if err != nil {
		return err
	}
	memberUserID, err := httpx.RequireUUIDParam(c, "userId")
	if err != nil {
		return err
	}

	if err := h.Service.RemoveMember(c.Request().Context(), user.ID, ledgerID, memberUserID); err != nil {
		return httpx.WriteAppError(c, err)
	}

	recordAudit(c, h.Audit, audit.Event{
		LedgerID:  ledgerID,
		UserID:    user.ID,
		Action:    "ledger.member.remove",
		EntityID:  &memberUserID,
		IP:        c.RealIP(),
		UserAgent: c.Request().UserAgent(),
	})

	return c.NoContent(http.StatusNoContent)
}
