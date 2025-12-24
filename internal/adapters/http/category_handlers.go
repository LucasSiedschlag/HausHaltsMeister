package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/dto"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/category"
)

type CategoryHandler struct {
	service category.Service
}

func NewCategoryHandler(service category.Service) *CategoryHandler {
	return &CategoryHandler{service: service}
}

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
