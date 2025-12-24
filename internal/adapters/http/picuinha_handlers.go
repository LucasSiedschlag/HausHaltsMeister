package http

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/seuuser/cashflow/internal/domain/picuinha"
)

type PicuinhaHandler struct {
	service *picuinha.PicuinhaService
}

func NewPicuinhaHandler(service *picuinha.PicuinhaService) *PicuinhaHandler {
	return &PicuinhaHandler{service: service}
}

type CreatePersonRequest struct {
	Name  string `json:"name"`
	Notes string `json:"notes"`
}

func (h *PicuinhaHandler) CreatePerson(c echo.Context) error {
	var req CreatePersonRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}

	person, err := h.service.CreatePerson(c.Request().Context(), req.Name, req.Notes)
	if err != nil {
		if err == picuinha.ErrPersonNameRequired {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("failed to create person: %v", err)})
	}

	return c.JSON(http.StatusCreated, person)
}

func (h *PicuinhaHandler) ListPersons(c echo.Context) error {
	persons, err := h.service.ListPersons(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list persons"})
	}
	return c.JSON(http.StatusOK, persons)
}

type AddEntryRequest struct {
	PersonID       int32   `json:"person_id"`
	Kind           string  `json:"kind"` // PLUS or MINUS
	Amount         float64 `json:"amount"`
	CashFlowID     *int32  `json:"cash_flow_id"` // Optional
	AutoCreateFlow bool    `json:"auto_create_flow"`
}

func (h *PicuinhaHandler) AddEntry(c echo.Context) error {
	var req AddEntryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}

	entry, err := h.service.AddDiff(c.Request().Context(), req.PersonID, req.Amount, req.Kind, req.CashFlowID, req.AutoCreateFlow)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// If CashFlowID was present, we might need a separate call or update AddDiff signature.
	// Service.AddDiff doesn't take CashFlowID currently.
	// Service.Lend / Receive do.
	// Let's just use generic AddDiff for now as per minimal viable.

	return c.JSON(http.StatusCreated, entry)
}

func RegisterPicuinhaRoutes(e *echo.Echo, h *PicuinhaHandler) {
	g := e.Group("/picuinhas")
	g.POST("/persons", h.CreatePerson)
	g.GET("/persons", h.ListPersons)
	g.POST("/entries", h.AddEntry)
}
