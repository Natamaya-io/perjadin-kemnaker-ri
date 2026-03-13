package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc Service
}

func NewHandler(s Service) *Handler {
	return &Handler{svc: s}
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name" validate:"required"`
	Role     string `json:"role"`
}

func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	token, user, err := h.svc.Login(req.Email, req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

func (h *Handler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	user, err := h.svc.Register(req.Email, req.Password, req.Name, req.Role)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user")
	}

	return c.JSON(http.StatusCreated, user)
}

func (h *Handler) GetDemoUsers(c echo.Context) error {
	demoUsers, err := h.svc.GetDemoUsers()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, demoUsers)
}
