package http

import (
	"errors"
	"fmt"
	"net/http"
	"time"

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
	monthParam := c.QueryParam("month")

	var (
		list []*category.Category
		err  error
	)
	if monthParam != "" {
		parsedMonth, parseErr := time.Parse("2006-01-02", monthParam)
		if parseErr != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid month format"})
		}
		list, err = h.service.ListCategoriesByMonth(c.Request().Context(), activeOnly, parsedMonth)
	} else {
		list, err = h.service.ListCategories(c.Request().Context(), activeOnly)
	}
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

	effectiveMonth := c.QueryParam("effective_month")
	month := time.Now().UTC()
	month = time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	if effectiveMonth != "" {
		parsedMonth, parseErr := time.Parse("2006-01-02", effectiveMonth)
		if parseErr != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid month format"})
		}
		month = parsedMonth
	}

	err := h.service.DeactivateCategory(c.Request().Context(), id, month)
	if err != nil {
		if errors.Is(err, category.ErrCategoryNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to deactivate category"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "deactivated"})
}

// Update updates a category.
// @Summary Atualizar Categoria
// @Description Updates a category by ID.
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param payload body dto.UpdateCategoryRequest true "Category Payload"
// @Success 200 {object} dto.CategoryResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /categories/{id} [put]
func (h *CategoryHandler) Update(c echo.Context) error {
	idStr := c.Param("id")
	var id int32
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id format"})
	}

	var req dto.UpdateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid payload"})
	}
	if req.Name == nil || req.Direction == nil || req.IsBudgetRelevant == nil || req.IsActive == nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "missing required fields"})
	}

	updated, err := h.service.UpdateCategory(c.Request().Context(), id, *req.Name, *req.Direction, *req.IsBudgetRelevant, *req.IsActive)
	if err != nil {
		if errors.Is(err, category.ErrInvalidDirection) || errors.Is(err, category.ErrEmptyName) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		}
		if errors.Is(err, category.ErrCategoryNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
	}

	return c.JSON(http.StatusOK, toCategoryResponse(updated))
}

func RegisterCategoryRoutes(e *echo.Echo, h *CategoryHandler) {
	g := e.Group("/categories")
	g.POST("", h.Create)
	g.GET("", h.List)
	g.PUT("/:id", h.Update)
	g.PATCH("/:id/deactivate", h.Deactivate)
}

func toCategoryResponse(c *category.Category) dto.CategoryResponse {
	var inactiveFromMonth string
	if c.InactiveFromMonth != nil {
		inactiveFromMonth = c.InactiveFromMonth.Format("2006-01-02")
	}
	return dto.CategoryResponse{
		ID:               c.ID,
		Name:             c.Name,
		Direction:        c.Direction,
		IsBudgetRelevant: c.IsBudgetRelevant,
		IsActive:         c.IsActive,
		InactiveFromMonth: inactiveFromMonth,
	}
}
