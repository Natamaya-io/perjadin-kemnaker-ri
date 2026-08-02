package gup

import (
	"net/http"
	"strconv"

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

func (h *Handler) GetBudgets(c echo.Context) error {
	yearStr := c.QueryParam("year")
	if yearStr == "" {
		yearStr = "2024" // default year or dynamically fetching current year
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
		yearStr = "2024" // default
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
