package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/kemnaker/perjadin-backend/internal/services"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	Service *services.Service
}

func NewUserHandler(s *services.Service) *UserHandler {
	return &UserHandler{Service: s}
}

func (h *UserHandler) GetUsers(c echo.Context) error {
	users, err := h.Service.GetUsers()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, users)
}

type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Password string `json:"password"`
	NIP      string `json:"nip"`
	NomorHP  string `json:"nomorHp"`
	Pangkat  string `json:"pangkat"`
	Golongan string `json:"golongan"`
	Jabatan  string `json:"jabatan"`
	TingkatBiaya string `json:"tingkatBiaya"`
}

func (h *UserHandler) CreateUser(c echo.Context) error {
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Role:     req.Role,
		Password: req.Password,
		NIP:      req.NIP,
		NomorHP:  req.NomorHP,
		Pangkat:  req.Pangkat,
		Golongan: req.Golongan,
		Jabatan:  req.Jabatan,
		TingkatBiaya: req.TingkatBiaya,
	}

	if err := h.Service.CreateUser(&user); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, user)
}

type UpdateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Password string `json:"password"`
	NIP      string `json:"nip"`
	NomorHP  string `json:"nomorHp"`
	Pangkat  string `json:"pangkat"`
	Golongan string `json:"golongan"`
	Jabatan  string `json:"jabatan"`
	TingkatBiaya string `json:"tingkatBiaya"`
}

func (h *UserHandler) UpdateUser(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid UUID format")
	}

	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}
	
	// Get existing user
	existingUser, err := h.Service.Repo.GetUserByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	// Update fields
	existingUser.Name = req.Name
	existingUser.Email = req.Email
	existingUser.Role = req.Role
	existingUser.NIP = req.NIP
	existingUser.NomorHP = req.NomorHP
	existingUser.Pangkat = req.Pangkat
	existingUser.Golongan = req.Golongan
	existingUser.Jabatan = req.Jabatan
	existingUser.TingkatBiaya = req.TingkatBiaya
	
	// Pass new password (if any) separately
	if err := h.Service.UpdateUser(existingUser, req.Password); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, existingUser)
}

func (h *UserHandler) DeleteUser(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid UUID format")
	}

	if err := h.Service.DeleteUser(id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}
