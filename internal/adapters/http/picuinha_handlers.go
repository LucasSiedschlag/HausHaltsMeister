package http

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/adapters/http/dto"
	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/picuinha"
	"github.com/labstack/echo/v4"
)

type PicuinhaHandler struct {
	service *picuinha.PicuinhaService
}

func NewPicuinhaHandler(service *picuinha.PicuinhaService) *PicuinhaHandler {
	return &PicuinhaHandler{service: service}
}

// CreatePerson registers a new person in the picuinha module.
// @Summary Cadastrar Pessoa
// @Description Registers a new person for sharing expenses.
// @Tags Picuinhas
// @Accept json
// @Produce json
// @Param payload body dto.CreatePersonRequest true "Person Payload"
// @Success 201 {object} dto.PersonResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /picuinhas/persons [post]
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

// UpdatePerson updates a person in the picuinha module.
// @Summary Atualizar Pessoa
// @Description Updates a person in picuinhas.
// @Tags Picuinhas
// @Accept json
// @Produce json
// @Param id path int true "Person ID"
// @Param payload body dto.UpdatePersonRequest true "Person Payload"
// @Success 200 {object} dto.PersonResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /picuinhas/persons/{id} [put]
func (h *PicuinhaHandler) UpdatePerson(c echo.Context) error {
	idStr := c.Param("id")
	var id int32
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id format"})
	}

	var req dto.UpdatePersonRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid payload"})
	}

	person, err := h.service.UpdatePerson(c.Request().Context(), id, req.Name, req.Notes)
	if err != nil {
		if errors.Is(err, picuinha.ErrPersonNameRequired) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		}
		if errors.Is(err, picuinha.ErrPersonNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to update person"})
	}

	return c.JSON(http.StatusOK, toPersonResponse(person))
}

// DeletePerson removes a person if no cases exist.
// @Summary Excluir Pessoa
// @Description Deletes a person (only if no cases exist).
// @Tags Picuinhas
// @Accept json
// @Produce json
// @Param id path int true "Person ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /picuinhas/persons/{id} [delete]
func (h *PicuinhaHandler) DeletePerson(c echo.Context) error {
	idStr := c.Param("id")
	var id int32
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id format"})
	}

	err := h.service.DeletePerson(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, picuinha.ErrCaseNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		}
		if errors.Is(err, picuinha.ErrPersonNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		}
		if errors.Is(err, picuinha.ErrPersonHasEntries) {
			return c.JSON(http.StatusConflict, dto.ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to delete person"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "deleted"})
}

// ListPersons returns a list of all registered persons with their balances.
// @Summary Listar Pessoas e Saldos
// @Description Returns a list of persons tracking loans/debts.
// @Tags Picuinhas
// @Accept json
// @Produce json
// @Success 200 {array} dto.PersonResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /picuinhas/persons [get]
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

// CreateCase registers a new picuinha case for a person.
// @Summary Criar Picuinha
// @Description Creates a new picuinha case (one-off, installment, recurring).
// @Tags Picuinhas
// @Accept json
// @Produce json
// @Param payload body dto.CreateCaseRequest true "Case Payload"
// @Success 201 {object} dto.CaseResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /picuinhas/cases [post]
func (h *PicuinhaHandler) CreateCase(c echo.Context) error {
	var req dto.CreateCaseRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid payload"})
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid start_date format, use YYYY-MM-DD"})
	}

	created, err := h.service.CreateCase(c.Request().Context(), picuinha.CreateCaseRequest{
		PersonID:                 req.PersonID,
		Title:                    req.Title,
		CaseType:                 req.CaseType,
		TotalAmount:              req.TotalAmount,
		InstallmentCount:         req.InstallmentCount,
		InstallmentAmount:        req.InstallmentAmount,
		StartDate:                startDate,
		PaymentMethodID:          req.PaymentMethodID,
		InstallmentPlanID:        req.InstallmentPlanID,
		CategoryID:               req.CategoryID,
		InterestRate:             req.InterestRate,
		InterestRateUnit:         req.InterestRateUnit,
		RecurrenceIntervalMonths: req.RecurrenceIntervalMonths,
	})
	if err != nil {
		if errors.Is(err, picuinha.ErrPersonNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		}
		if errors.Is(err, picuinha.ErrCaseTitleRequired) ||
			errors.Is(err, picuinha.ErrCaseTypeInvalid) ||
			errors.Is(err, picuinha.ErrInstallmentCount) ||
			errors.Is(err, picuinha.ErrAmountRequired) ||
			errors.Is(err, picuinha.ErrPaymentMethodRequired) ||
			errors.Is(err, picuinha.ErrInterestRateUnit) ||
			errors.Is(err, picuinha.ErrRecurrenceInterval) ||
			errors.Is(err, picuinha.ErrStartDateRequired) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to create case"})
	}

	return c.JSON(http.StatusCreated, toCaseResponse(created))
}

// ListCases lists picuinha cases for a person.
// @Summary Listar Picuinhas
// @Description Returns a list of cases for a person.
// @Tags Picuinhas
// @Accept json
// @Produce json
// @Param person_id query int true "Person ID"
// @Success 200 {array} dto.CaseResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /picuinhas/cases [get]
func (h *PicuinhaHandler) ListCases(c echo.Context) error {
	personIDParam := c.QueryParam("person_id")
	if personIDParam == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "person_id is required"})
	}
	var personID int32
	if _, err := fmt.Sscanf(personIDParam, "%d", &personID); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid person_id"})
	}

	cases, err := h.service.ListCasesByPerson(c.Request().Context(), personID)
	if err != nil {
		if errors.Is(err, picuinha.ErrPersonNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to list cases"})
	}

	resp := make([]dto.CaseResponse, len(cases))
	for i, picCase := range cases {
		resp[i] = toCaseResponse(&picCase)
	}

	return c.JSON(http.StatusOK, resp)
}

// UpdateCase updates a picuinha case.
// @Summary Atualizar Picuinha
// @Description Updates a picuinha case.
// @Tags Picuinhas
// @Accept json
// @Produce json
// @Param id path int true "Case ID"
// @Param payload body dto.CreateCaseRequest true "Case Payload"
// @Success 200 {object} dto.CaseResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /picuinhas/cases/{id} [put]
func (h *PicuinhaHandler) UpdateCase(c echo.Context) error {
	idStr := c.Param("id")
	var id int32
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id format"})
	}

	var req dto.CreateCaseRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid payload"})
	}
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid start_date format, use YYYY-MM-DD"})
	}

	updated, err := h.service.UpdateCase(c.Request().Context(), id, picuinha.CreateCaseRequest{
		PersonID:                 req.PersonID,
		Title:                    req.Title,
		CaseType:                 req.CaseType,
		TotalAmount:              req.TotalAmount,
		InstallmentCount:         req.InstallmentCount,
		InstallmentAmount:        req.InstallmentAmount,
		StartDate:                startDate,
		PaymentMethodID:          req.PaymentMethodID,
		InstallmentPlanID:        req.InstallmentPlanID,
		CategoryID:               req.CategoryID,
		InterestRate:             req.InterestRate,
		InterestRateUnit:         req.InterestRateUnit,
		RecurrenceIntervalMonths: req.RecurrenceIntervalMonths,
	})
	if err != nil {
		if errors.Is(err, picuinha.ErrCaseNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		}
		if errors.Is(err, picuinha.ErrCaseTitleRequired) ||
			errors.Is(err, picuinha.ErrCaseTypeInvalid) ||
			errors.Is(err, picuinha.ErrInstallmentCount) ||
			errors.Is(err, picuinha.ErrAmountRequired) ||
			errors.Is(err, picuinha.ErrPaymentMethodRequired) ||
			errors.Is(err, picuinha.ErrInterestRateUnit) ||
			errors.Is(err, picuinha.ErrRecurrenceInterval) ||
			errors.Is(err, picuinha.ErrStartDateRequired) {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to update case"})
	}

	return c.JSON(http.StatusOK, toCaseResponse(updated))
}

// DeleteCase removes a picuinha case.
// @Summary Excluir Picuinha
// @Description Deletes a picuinha case.
// @Tags Picuinhas
// @Accept json
// @Produce json
// @Param id path int true "Case ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} dto.ErrorResponse
// @Router /picuinhas/cases/{id} [delete]
func (h *PicuinhaHandler) DeleteCase(c echo.Context) error {
	idStr := c.Param("id")
	var id int32
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id format"})
	}

	if err := h.service.DeleteCase(c.Request().Context(), id); err != nil {
		if errors.Is(err, picuinha.ErrCaseNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to delete case"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "deleted"})
}

// ListCaseInstallments returns installments for a case.
// @Summary Listar Parcelas de Picuinha
// @Description Returns installments for a picuinha case.
// @Tags Picuinhas
// @Accept json
// @Produce json
// @Param id path int true "Case ID"
// @Success 200 {array} dto.CaseInstallmentResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /picuinhas/cases/{id}/installments [get]
func (h *PicuinhaHandler) ListCaseInstallments(c echo.Context) error {
	idStr := c.Param("id")
	var id int32
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id format"})
	}

	items, err := h.service.ListInstallmentsByCase(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, picuinha.ErrCaseNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to list installments"})
	}

	resp := make([]dto.CaseInstallmentResponse, len(items))
	for i, inst := range items {
		resp[i] = toCaseInstallmentResponse(&inst)
	}

	return c.JSON(http.StatusOK, resp)
}

// UpdateCaseInstallment updates a case installment.
// @Summary Atualizar Parcela de Picuinha
// @Description Updates a picuinha installment (paid status/extra).
// @Tags Picuinhas
// @Accept json
// @Produce json
// @Param id path int true "Installment ID"
// @Param payload body dto.UpdateCaseInstallmentRequest true "Installment Payload"
// @Success 200 {object} dto.CaseInstallmentResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /picuinhas/installments/{id} [put]
func (h *PicuinhaHandler) UpdateCaseInstallment(c echo.Context) error {
	idStr := c.Param("id")
	var id int32
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id format"})
	}

	var req dto.UpdateCaseInstallmentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid payload"})
	}

	updated, err := h.service.UpdateInstallment(c.Request().Context(), id, picuinha.UpdateInstallmentRequest{
		IsPaid:      req.IsPaid,
		ExtraAmount: req.ExtraAmount,
	})
	if err != nil {
		if errors.Is(err, picuinha.ErrInstallmentNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to update installment"})
	}

	return c.JSON(http.StatusOK, toCaseInstallmentResponse(updated))
}

func RegisterPicuinhaRoutes(e *echo.Echo, h *PicuinhaHandler) {
	g := e.Group("/picuinhas")
	g.POST("/persons", h.CreatePerson)
	g.GET("/persons", h.ListPersons)
	g.PUT("/persons/:id", h.UpdatePerson)
	g.DELETE("/persons/:id", h.DeletePerson)
	g.GET("/cases", h.ListCases)
	g.POST("/cases", h.CreateCase)
	g.PUT("/cases/:id", h.UpdateCase)
	g.DELETE("/cases/:id", h.DeleteCase)
	g.GET("/cases/:id/installments", h.ListCaseInstallments)
	g.PUT("/installments/:id", h.UpdateCaseInstallment)
}

func toPersonResponse(p *picuinha.Person) dto.PersonResponse {
	return dto.PersonResponse{
		ID:      p.ID,
		Name:    p.Name,
		Notes:   p.Notes,
		Balance: p.Balance,
	}
}

func toCaseResponse(c *picuinha.CaseSummary) dto.CaseResponse {
	return dto.CaseResponse{
		ID:                       c.ID,
		PersonID:                 c.PersonID,
		Title:                    c.Title,
		CaseType:                 c.CaseType,
		TotalAmount:              c.TotalAmount,
		InstallmentCount:         c.InstallmentCount,
		InstallmentAmount:        c.InstallmentAmount,
		StartDate:                c.StartDate.Format("2006-01-02"),
		PaymentMethodID:          c.PaymentMethodID,
		InstallmentPlanID:        c.InstallmentPlanID,
		CategoryID:               c.CategoryID,
		InterestRate:             c.InterestRate,
		InterestRateUnit:         c.InterestRateUnit,
		RecurrenceIntervalMonths: c.RecurrenceIntervalMonths,
		InstallmentsTotal:        c.InstallmentsTotal,
		InstallmentsPaid:         c.InstallmentsPaid,
		AmountPaid:               c.AmountPaid,
		AmountRemaining:          c.AmountRemaining,
		Status:                   c.Status,
	}
}

func toCaseInstallmentResponse(inst *picuinha.CaseInstallment) dto.CaseInstallmentResponse {
	var paidAt *string
	if inst.PaidAt != nil {
		val := inst.PaidAt.Format(time.RFC3339)
		paidAt = &val
	}
	return dto.CaseInstallmentResponse{
		ID:                inst.ID,
		CaseID:            inst.CaseID,
		InstallmentNumber: inst.InstallmentNumber,
		DueDate:           inst.DueDate.Format("2006-01-02"),
		Amount:            inst.Amount,
		ExtraAmount:       inst.ExtraAmount,
		IsPaid:            inst.IsPaid,
		PaidAt:            paidAt,
	}
}
