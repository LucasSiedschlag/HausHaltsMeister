package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/dto"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/picuinha"
)

type PicuinhaHandler struct {
	service *picuinha.PicuinhaService
}

func NewPicuinhaHandler(service *picuinha.PicuinhaService) *PicuinhaHandler {
	return &PicuinhaHandler{service: service}
}

func (h *PicuinhaHandler) CreatePerson(c echo.Context) error {
	var req dto.CreatePersonRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid payload"})
	}

	person, err := h.service.CreatePerson(c.Request().Context(), req.Name, req.Notes)
	if err != nil {
		if err == picuinha.ErrPersonNameRequired {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: fmt.Sprintf("failed to create person: %v", err)})
	}

	return c.JSON(http.StatusCreated, toPersonResponse(person))
}

func (h *PicuinhaHandler) ListPersons(c echo.Context) error {
	persons, err := h.service.ListPersons(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to list persons"})
	}

	resp := make([]dto.PersonResponse, len(persons))
	for i, p := range persons {
		resp[i] = toPersonResponse(&p)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *PicuinhaHandler) AddEntry(c echo.Context) error {
	var req dto.AddEntryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid payload"})
	}

	entry, err := h.service.AddDiff(c.Request().Context(), req.PersonID, req.Amount, req.Kind, req.CashFlowID, req.AutoCreateFlow)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusCreated, toPicuinhaEntryResponse(entry))
}

func RegisterPicuinhaRoutes(e *echo.Echo, h *PicuinhaHandler) {
	g := e.Group("/picuinhas")
	g.POST("/persons", h.CreatePerson)
	g.GET("/persons", h.ListPersons)
	g.POST("/entries", h.AddEntry)
}

func toPersonResponse(p *picuinha.Person) dto.PersonResponse {
	return dto.PersonResponse{
		ID:      p.ID,
		Name:    p.Name,
		Notes:   p.Notes,
		Balance: p.Balance,
	}
}

func toPicuinhaEntryResponse(e *picuinha.Entry) dto.PicuinhaEntryResponse {
	return dto.PicuinhaEntryResponse{
		ID:         e.ID,
		PersonID:   e.PersonID,
		Amount:     e.Amount,
		Kind:       e.Kind,
		CashFlowID: e.CashFlowID,
		CreatedAt:  e.Date.Format(time.RFC3339),
	}
}
