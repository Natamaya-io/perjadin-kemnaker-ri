package dalkot

import (
	"fmt"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/kemnaker/perjadin-backend/internal/utils/document"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc    Service
	docGen *document.Generator
}

func NewHandler(svc Service) *Handler {
	gotenbergURL := os.Getenv("GOTENBERG_URL")
	if gotenbergURL == "" {
		gotenbergURL = "http://gotenberg:3000"
	}
	return &Handler{
		svc:    svc,
		docGen: document.NewGenerator(gotenbergURL, "templates"),
	}
}

func (h *Handler) CreateRecord(c echo.Context) error {
	var req struct {
		Record      models.DalkotRecord       `json:"record"`
		Assignments []models.DalkotAssignment `json:"assignments"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := h.svc.CreateRecord(&req.Record, req.Assignments); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, req.Record)
}

func (h *Handler) GetRecords(c echo.Context) error {
	role, _ := c.Get("role").(string)
	userIDStr, _ := c.Get("user_id").(string)

	records, err := h.svc.GetRecords()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if role == "protokol" {
		uid, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		}
		filtered := []models.DalkotRecord{}
		for _, r := range records {
			hasAccess := false
			for _, a := range r.Assignments {
				if a.UserID == uid {
					hasAccess = true
					break
				}
			}
			if hasAccess {
				filtered = append(filtered, r)
			}
		}
		return c.JSON(http.StatusOK, filtered)
	}

	return c.JSON(http.StatusOK, records)
}

func (h *Handler) GetRecordByID(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid record id"})
	}

	record, err := h.svc.GetRecordByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "record not found"})
	}

	return c.JSON(http.StatusOK, record)
}

func (h *Handler) UpdateRecord(c echo.Context) error {
	var req models.DalkotRecord
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid record id"})
	}
	req.ID = id

	if err := h.svc.UpdateRecord(&req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, req)
}

func (h *Handler) DeleteRecord(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid record id"})
	}

	if err := h.svc.DeleteRecord(id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "deleted successfully"})
}

func (h *Handler) GetLocations(c echo.Context) error {
	locations, err := h.svc.GetLocations()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, locations)
}

func (h *Handler) GetRates(c echo.Context) error {
	rates, err := h.svc.GetRates()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, rates)
}

func (h *Handler) AddAssignment(c echo.Context) error {
	var req models.DalkotAssignment
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := h.svc.AddAssignment(&req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, req)
}

func (h *Handler) RemoveAssignment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("assignmentId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid assignment id"})
	}

	if err := h.svc.RemoveAssignment(id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "assignment deleted successfully"})
}

func (h *Handler) ExportLaporanPDF(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid record ID")
	}

	record, err := h.svc.GetRecordByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Record not found")
	}

	htmlContent, err := generateLaporanHTML(record)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate HTML template")
	}

	pdfBytes, err := h.docGen.GenerateHTMLToPDF(c.Request().Context(), htmlContent)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to generate PDF: %v", err))
	}

	filename := fmt.Sprintf("Laporan_Dalkot_%s.pdf", record.SPDNumber)
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=\"%s\"", filename))
	return c.Blob(http.StatusOK, "application/pdf", pdfBytes)
}
