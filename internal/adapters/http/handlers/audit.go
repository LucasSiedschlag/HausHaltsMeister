package handlers

import (
	"log"

	"github.com/LucasSiedschlag/HausHaltsMeister/internal/domain/audit"
	"github.com/labstack/echo/v4"
)

func recordAudit(c echo.Context, recorder audit.Recorder, event audit.Event) {
	if recorder == nil {
		return
	}
	if err := recorder.Record(c.Request().Context(), event); err != nil {
		log.Printf("audit: %v", err)
	}
}
