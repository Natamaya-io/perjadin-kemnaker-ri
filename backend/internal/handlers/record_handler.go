package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/kemnaker/perjadin-backend/internal/services"
	"github.com/kemnaker/perjadin-backend/internal/utils"
	"github.com/labstack/echo/v4"
)

type RecordHandler struct {
	Service *services.Service
}

func NewRecordHandler(s *services.Service) *RecordHandler {
	return &RecordHandler{Service: s}
}

func (h *RecordHandler) notifyEmployee(record *models.TravelRecord) {
	if h.Service.Config.Fonnte.Token == "" {
		return
	}

	user, err := h.Service.GetUserByID(record.EmployeeID)
	if err != nil || user.NomorHP == "" {
		return
	}

	message := fmt.Sprintf("Halo *%s*,\n\nAnda telah ditugaskan untuk melaksanakan perjalanan dinas dengan rincian sebagai berikut:\n\n📍 *Tujuan:* %s, %s\n📅 *Tanggal:* %s s/d %s\n🎯 *Kegiatan:* %s\n\nHarap persiapkan diri Anda dan cek aplikasi untuk detail lebih lanjut.\n\n_Pesan ini dikirim otomatis oleh Sistem Perjadin Protokol Kemnaker RI_", 
		user.Name, record.Location, record.Province, record.StartDate.Format("02 Jan 2006"), record.EndDate.Format("02 Jan 2006"), record.Purpose)

	err = utils.SendWhatsAppMessage(h.Service.Config, user.NomorHP, message)
	if err != nil {
		fmt.Printf("Failed to send WhatsApp message to %s: %v\n", user.NomorHP, err)
	} else {
		fmt.Printf("Successfully sent WhatsApp notification to %s\n", user.NomorHP)
	}
}

func (h *RecordHandler) UploadFile(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "File not found in request")
	}

	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer src.Close()

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := uuid.New().String() + "_" + time.Now().Format("20060102150405") + ext

	// Ensure uploads directory exists
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create upload directory")
	}

	// Create destination file
	dstPath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer dst.Close()

	// Copy content
	if _, err = io.Copy(dst, src); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"path": filename,
	})
}

func (h *RecordHandler) CreateRecord(c echo.Context) error {
	var record models.TravelRecord
	if err := c.Bind(&record); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	// Set creator from context
	var creatorID uuid.UUID
	creatorIDInterface := c.Get("user_id")
	if creatorIDInterface != nil {
		if creatorIDStr, ok := creatorIDInterface.(string); ok {
			if id, err := uuid.Parse(creatorIDStr); err == nil {
				creatorID = id
			}
		}
	}

	// Bulk Creation
	if len(record.EmployeeIDs) > 0 {
		var createdRecords []models.TravelRecord
		for _, empID := range record.EmployeeIDs {
			newRecord := record // Copy struct
			newRecord.ID = uuid.Nil // Ensure new ID generation
			newRecord.EmployeeID = empID
			newRecord.CreatorID = creatorID
			// Clear the bulk field to avoid confusion, though GORM ignores it
			newRecord.EmployeeIDs = nil 

			if err := h.Service.CreateRecord(&newRecord); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}
			createdRecords = append(createdRecords, newRecord)
			
			go h.notifyEmployee(&newRecord)
		}
		return c.JSON(http.StatusCreated, createdRecords)
	}

	// Single Creation
	record.CreatorID = creatorID
	if err := h.Service.CreateRecord(&record); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	go h.notifyEmployee(&record)

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

func (h *RecordHandler) UpdateRecord(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid UUID format")
	}

	// Fetch existing
	record, err := h.Service.GetRecordByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Record not found")
	}

	// Bind updates
	if err := c.Bind(&record); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	if err := h.Service.UpdateRecord(record); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, record)
}

func (h *RecordHandler) DeleteRecord(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid UUID format")
	}

	if err := h.Service.DeleteRecord(id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}
