package chatbot

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Ask(c echo.Context) error {
	var req AskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}

	resp, err := h.svc.Ask(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    resp,
	})
}

func (h *Handler) GetSnapshot(c echo.Context) error {
	year := time.Now().Year()
	gupSnap, err := h.svc.(*service).repo.GetGupSnapshot(c.Request().Context(), year)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}
	dalkotSnap, err := h.svc.(*service).repo.GetDalkotSnapshot(c.Request().Context(), year)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"gup":    gupSnap,
			"dalkot": dalkotSnap,
		},
	})
}
