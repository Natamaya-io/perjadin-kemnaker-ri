package gup

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetTransactions(c echo.Context) error {
	ctx := c.Request().Context()
	trxs, err := h.svc.GetTransactions(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, trxs)
}

func (h *Handler) GetDashboardSummary(c echo.Context) error {
	yearStr := c.QueryParam("year")
	if yearStr == "" {
		yearStr = strconv.Itoa(time.Now().Year())
	}
	year, err := strconv.ParseInt(yearStr, 10, 16)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid year"})
	}

	ctx := c.Request().Context()
	summary, err := h.svc.GetDashboardSummary(ctx, int16(year))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, summary)
}

func (h *Handler) GetNextBusinessID(c echo.Context) error {
	ctx := c.Request().Context()
	nextID, err := h.svc.GetNextBusinessID(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"nextId": nextID})
}

func (h *Handler) GetTransactionByID(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid transaction id"})
	}

	ctx := c.Request().Context()
	trx, err := h.svc.GetTransactionByID(ctx, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "transaction not found"})
	}

	return c.JSON(http.StatusOK, trx)
}

func (h *Handler) CreateTransaction(c echo.Context) error {
	var trx models.GUPTransaction
	if err := c.Bind(&trx); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	ctx := c.Request().Context()
	if err := h.svc.CreateTransaction(ctx, &trx); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, trx)
}

func (h *Handler) UpdateTransaction(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid transaction id"})
	}

	var trx models.GUPTransaction
	if err := c.Bind(&trx); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	trx.ID = id

	ctx := c.Request().Context()
	if err := h.svc.UpdateTransaction(ctx, &trx); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, trx)
}

func (h *Handler) DeleteTransaction(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid transaction id"})
	}

	ctx := c.Request().Context()
	if err := h.svc.DeleteTransaction(ctx, id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) GetBudgets(c echo.Context) error {
	yearStr := c.QueryParam("year")
	if yearStr == "" {
		yearStr = "2026" // default year or dynamically fetching current year
	}
	year, err := strconv.ParseInt(yearStr, 10, 16)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid year"})
	}

	ctx := c.Request().Context()
	budgets, err := h.svc.GetBudgets(ctx, int16(year))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, budgets)
}

func (h *Handler) GetMonthlyLS(c echo.Context) error {
	yearStr := c.QueryParam("year")
	if yearStr == "" {
		yearStr = "2026" // default
	}
	year, err := strconv.ParseInt(yearStr, 10, 16)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid year"})
	}

	ctx := c.Request().Context()
	ls, err := h.svc.GetMonthlyLS(ctx, int16(year))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, ls)
}

func (h *Handler) SaveMonthlyLS(c echo.Context) error {
	ctx := c.Request().Context()
	var payload []models.MonthlyLS
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}
	err := h.svc.SaveMonthlyLS(ctx, payload)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Monthly LS saved successfully"})
}

func (h *Handler) GetMasterData(c echo.Context) error {
	yearStr := c.QueryParam("year")
	if yearStr == "" {
		yearStr = "2026" // default
	}
	year, err := strconv.ParseInt(yearStr, 10, 16)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid year"})
	}

	ctx := c.Request().Context()
	masterData, err := h.svc.GetMasterData(ctx, int16(year))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, masterData)
}

func (h *Handler) CreateProcurementType(c echo.Context) error {
	var pt models.ProcurementType
	if err := c.Bind(&pt); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	ctx := c.Request().Context()
	if err := h.svc.CreateProcurementType(ctx, &pt); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, pt)
}

func (h *Handler) UpdateProcurementType(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	var pt models.ProcurementType
	if err := c.Bind(&pt); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	pt.ID = id

	ctx := c.Request().Context()
	if err := h.svc.UpdateProcurementType(ctx, &pt); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, pt)
}

func (h *Handler) DeleteProcurementType(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	ctx := c.Request().Context()
	if err := h.svc.DeleteProcurementType(ctx, id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "deleted"})
}

func (h *Handler) CreateAccountCode(c echo.Context) error {
	var ac models.AccountCode
	if err := c.Bind(&ac); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	ctx := c.Request().Context()
	if err := h.svc.CreateAccountCode(ctx, &ac); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, ac)
}

func (h *Handler) UpdateAccountCode(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	var ac models.AccountCode
	if err := c.Bind(&ac); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	ac.ID = id

	ctx := c.Request().Context()
	if err := h.svc.UpdateAccountCode(ctx, &ac); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, ac)
}

func (h *Handler) DeleteAccountCode(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	ctx := c.Request().Context()
	if err := h.svc.DeleteAccountCode(ctx, id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "deleted"})
}

// GetLaporan mengembalikan rekapitulasi serapan anggaran per Jenis Pengadaan
// beserta ringkasan total (summary). Endpoint: GET /gup/laporan?year=2025
func (h *Handler) GetLaporan(c echo.Context) error {
	yearStr := c.QueryParam("year")
	if yearStr == "" {
		yearStr = strconv.Itoa(time.Now().Year())
	}
	year, err := strconv.ParseInt(yearStr, 10, 16)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid year"})
	}

	ctx := c.Request().Context()
	resp, err := h.svc.GetLaporan(ctx, int16(year))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}

// SaveBudget menyimpan atau memperbarui anggaran untuk satu Jenis Pengadaan.
// Endpoint: POST /gup/laporan/budget
func (h *Handler) SaveBudget(c echo.Context) error {
	var b models.Budget
	if err := c.Bind(&b); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if b.Year == 0 {
		b.Year = int16(time.Now().Year())
	}

	ctx := c.Request().Context()
	if err := h.svc.SaveBudget(ctx, &b); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, b)
}

