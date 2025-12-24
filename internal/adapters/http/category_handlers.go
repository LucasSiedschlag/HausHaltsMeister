package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/dto"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/category"
	"github.com/labstack/echo/v4"
)

type CategoryHandler struct {
	service category.Service
}

func NewCategoryHandler(service category.Service) *CategoryHandler {
	return &CategoryHandler{service: service}
}

// Create creates a new category.
// @Summary Create Categoria
// @Description Creates a new category for financial flows.
// @Tags Categories
// @Accept json
// @Produce json
// @Param payload body dto.CreateCategoryRequest true "Category Payload"
// @Success 201 {object} dto.CategoryResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /categories [post]
func (h *CategoryHandler) Create(c echo.Context) error {
	var req dto.CreateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid payload"})
	}

	cat, err := h.service.CreateCategory(c.Request().Context(), req.Name, req.Direction, req.IsBudgetRelevant)
	if err != nil {
		if errors.Is(err, category.ErrInvalidDirection) || errors.Is(err, category.ErrEmptyName) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
	}

	return c.JSON(http.StatusCreated, toCategoryResponse(cat))
}

// List returns a list of categories.
// @Summary Listar Categorias
// @Description Returns a list of all categories, optionally filtered by active status.
// @Tags Categories
// @Accept json
// @Produce json
// @Param active query boolean false "Filter only active categories"
// @Success 200 {array} dto.CategoryResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /categories [get]
func (h *CategoryHandler) List(c echo.Context) error {
	activeOnly := c.QueryParam("active") == "true"

	list, err := h.service.ListCategories(c.Request().Context(), activeOnly)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to list categories"})
	}

	resp := make([]dto.CategoryResponse, len(list))
	for i, cat := range list {
		resp[i] = toCategoryResponse(cat)
	}

	return c.JSON(http.StatusOK, resp)
}

// Deactivate disables a category.
// @Summary Desativar Categoria
// @Description Deactivates a specific category by ID.
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Router /categories/{id}/deactivate [patch]
func (h *CategoryHandler) Deactivate(c echo.Context) error {
	idStr := c.Param("id")
	var id int32
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id format"})
	}

	err := h.service.DeactivateCategory(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to deactivate category"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "deactivated"})
}

func RegisterCategoryRoutes(e *echo.Echo, h *CategoryHandler) {
	g := e.Group("/categories")
	g.POST("", h.Create)
	g.GET("", h.List)
	g.PATCH("/:id/deactivate", h.Deactivate)
}

func toCategoryResponse(c *category.Category) dto.CategoryResponse {
	return dto.CategoryResponse{
		ID:               c.ID,
		Name:             c.Name,
		Direction:        c.Direction,
		IsBudgetRelevant: c.IsBudgetRelevant,
		IsActive:         c.IsActive,
	}
}
