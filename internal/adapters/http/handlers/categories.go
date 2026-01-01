package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/dto"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/httpx"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/categories"
	"github.com/labstack/echo/v4"
)

type CategoriesHandler struct {
	Service CategoriesService
}

type CategoriesService interface {
	ListCategories(ctx context.Context, userID, ledgerID string, direction *string, active *bool) ([]categories.Category, error)
	GetCategory(ctx context.Context, userID, ledgerID, categoryID string) (categories.Category, error)
	CreateCategory(ctx context.Context, userID, ledgerID string, input categories.CreateCategoryParams) (categories.Category, error)
	UpdateCategory(ctx context.Context, userID, ledgerID, categoryID string, input categories.UpdateCategoryParams) (categories.Category, error)
	DeleteCategory(ctx context.Context, userID, ledgerID, categoryID string) error
}

type categoryRequest = dto.CategoryRequest
type categoryResponse = dto.CategoryResponse

func (h *CategoriesHandler) Register(g *echo.Group) {
	g.GET("", h.List)
	g.POST("", h.Create)
	g.GET("/:categoryId", h.Get)
	g.PATCH("/:categoryId", h.Update)
	g.DELETE("/:categoryId", h.Delete)
}

func (h *CategoriesHandler) List(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	if ledgerID == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"ledger_id": "required"})
	}

	var direction *string
	if value := strings.TrimSpace(c.QueryParam("direction")); value != "" {
		direction = &value
	}
	var active *bool
	if value := strings.TrimSpace(c.QueryParam("active")); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"active": "invalid"})
		}
		active = &parsed
	}

	items, err := h.Service.ListCategories(c.Request().Context(), user.ID, ledgerID, direction, active)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}

	response := make([]categoryResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toCategoryResponse(item))
	}
	return c.JSON(http.StatusOK, response)
}

func (h *CategoriesHandler) Get(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	categoryID := c.Param("categoryId")
	if ledgerID == "" || categoryID == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"category_id": "required"})
	}

	item, err := h.Service.GetCategory(c.Request().Context(), user.ID, ledgerID, categoryID)
	if err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.JSON(http.StatusOK, toCategoryResponse(item))
}

func (h *CategoriesHandler) Create(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	if ledgerID == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"ledger_id": "required"})
	}

	var req categoryRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	isBudgetBase := false
	if req.IsBudgetBase != nil {
		isBudgetBase = *req.IsBudgetBase
	}
	isBudgetRelevant := true
	if req.IsBudgetRelevant != nil {
		isBudgetRelevant = *req.IsBudgetRelevant
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	created, err := h.Service.CreateCategory(c.Request().Context(), user.ID, ledgerID, categories.CreateCategoryParams{
		LedgerID:         ledgerID,
		ParentID:         req.ParentID,
		Name:             req.Name,
		Direction:        req.Direction,
		IsBudgetBase:     isBudgetBase,
		IsBudgetRelevant: isBudgetRelevant,
		IsActive:         isActive,
	})
	if err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.JSON(http.StatusCreated, toCategoryResponse(created))
}

func (h *CategoriesHandler) Update(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	categoryID := c.Param("categoryId")
	if ledgerID == "" || categoryID == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"category_id": "required"})
	}

	var req categoryRequest
	if err := c.Bind(&req); err != nil {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Payload invalido", nil)
	}

	isBudgetBase := false
	if req.IsBudgetBase != nil {
		isBudgetBase = *req.IsBudgetBase
	}
	isBudgetRelevant := true
	if req.IsBudgetRelevant != nil {
		isBudgetRelevant = *req.IsBudgetRelevant
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	updated, err := h.Service.UpdateCategory(c.Request().Context(), user.ID, ledgerID, categoryID, categories.UpdateCategoryParams{
		LedgerID:         ledgerID,
		CategoryID:       categoryID,
		ParentID:         req.ParentID,
		Name:             req.Name,
		IsBudgetBase:     isBudgetBase,
		IsBudgetRelevant: isBudgetRelevant,
		IsActive:         isActive,
	})
	if err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.JSON(http.StatusOK, toCategoryResponse(updated))
}

func (h *CategoriesHandler) Delete(c echo.Context) error {
	user, ok := httpx.GetUser(c)
	if !ok {
		return httpx.WriteError(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Credenciais invalidas", nil)
	}
	ledgerID := c.Param("ledgerId")
	categoryID := c.Param("categoryId")
	if ledgerID == "" || categoryID == "" {
		return httpx.WriteError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Parametros invalidos", map[string]string{"category_id": "required"})
	}

	if err := h.Service.DeleteCategory(c.Request().Context(), user.ID, ledgerID, categoryID); err != nil {
		return httpx.WriteAppError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func toCategoryResponse(item categories.Category) categoryResponse {
	return categoryResponse{
		ID:               item.ID,
		LedgerID:         item.LedgerID,
		ParentID:         item.ParentID,
		Name:             item.Name,
		Direction:        item.Direction,
		IsBudgetBase:     item.IsBudgetBase,
		IsBudgetRelevant: item.IsBudgetRelevant,
		IsActive:         item.IsActive,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}
}
