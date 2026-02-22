package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/kemnaker/perjadin-backend/internal/services"
	"github.com/labstack/echo/v4"
)

type RecordHandler struct {
	Service *services.Service
}

func NewRecordHandler(s *services.Service) *RecordHandler {
	return &RecordHandler{Service: s}
}

func (h *RecordHandler) CreateRecord(c echo.Context) error {
	var record models.TravelRecord
	if err := c.Bind(&record); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	// Set creator from context
	creatorIDStr := c.Get("user_id").(string)
	creatorID, err := uuid.Parse(creatorIDStr)
	if err == nil {
		record.CreatorID = creatorID
	}

	if err := h.Service.CreateRecord(&record); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, record)
}

func (h *RecordHandler) GetRecords(c echo.Context) error {
	filters := make(map[string]interface{})
	status := c.QueryParam("status")
	if status != "" {
		filters["status"] = status
	}

	records, err := h.Service.GetRecords(filters)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, records)
}

func (h *RecordHandler) GetRecordByID(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid UUID format")
	}

	record, err := h.Service.GetRecordByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Record not found")
	}

	return c.JSON(http.StatusOK, record)
}
