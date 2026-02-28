package handlers

import (
	"net/http"

	"github.com/kemnaker/perjadin-backend/internal/services"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	Service *services.Service
}

func NewAuthHandler(s *services.Service) *AuthHandler {
	return &AuthHandler{Service: s}
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name" validate:"required"`
	Role     string `json:"role"` // Optional, defaults to protokol
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	token, user, err := h.Service.Login(req.Email, req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	user, err := h.Service.Register(req.Email, req.Password, req.Name, req.Role)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user")
	}

	return c.JSON(http.StatusCreated, user)
}

func (h *AuthHandler) GetDemoUsers(c echo.Context) error {
	users, err := h.Service.Repo.GetUsers()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	var demoUsers []map[string]string
	for _, u := range users {
		demoUsers = append(demoUsers, map[string]string{
			"name":     u.Name,
			"email":    u.Email,
			"role":     u.Role,
			"password": "123", // For demo purposes, we expose this
		})
	}

	return c.JSON(http.StatusOK, demoUsers)
}
