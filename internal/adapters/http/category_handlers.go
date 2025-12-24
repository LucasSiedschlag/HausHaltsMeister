package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/seuuser/cashflow/internal/domain/category"
)

type CategoryHandler struct {
	service category.Service
}

func NewCategoryHandler(service category.Service) *CategoryHandler {
	return &CategoryHandler{service: service}
}

type CreateCategoryRequest struct {
	Name      string `json:"name"`
	Direction string `json:"direction"`
}

func (h *CategoryHandler) Create(c echo.Context) error {
	var req CreateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}

	cat, err := h.service.CreateCategory(c.Request().Context(), req.Name, req.Direction)
	if err != nil {
		// Map domain errors to status codes
		if err == category.ErrInvalidDirection || err == category.ErrEmptyName {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return c.JSON(http.StatusCreated, cat)
}

func (h *CategoryHandler) List(c echo.Context) error {
	activeOnly := c.QueryParam("active") == "true"

	list, err := h.service.ListCategories(c.Request().Context(), activeOnly)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list categories"})
	}

	return c.JSON(http.StatusOK, list)
}

func RegisterCategoryRoutes(e *echo.Echo, h *CategoryHandler) {
	g := e.Group("/categories")
	g.POST("", h.Create)
	g.GET("", h.List)
}
