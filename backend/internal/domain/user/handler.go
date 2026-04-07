package user

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc Service
}

func NewHandler(s Service) *Handler {
	return &Handler{svc: s}
}

func (h *Handler) GetUsers(c echo.Context) error {
	role := c.Get("role").(string)
	uidStr := c.Get("user_id").(string)

	users, err := h.svc.GetUsers()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if role != "super_admin" {
		var filtered []models.User
		for _, u := range users {
			if u.ID.String() == uidStr {
				filtered = append(filtered, u)
				break
			}
		}
		return c.JSON(http.StatusOK, filtered)
	}

	return c.JSON(http.StatusOK, users)
}

type CreateUserRequest struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	Password     string `json:"password"`
	NIP          string `json:"nip"`
	NomorHP      string `json:"nomorHp"`
	Pangkat      string `json:"pangkat"`
	Golongan     string `json:"golongan"`
	Jabatan      string `json:"jabatan"`
	TingkatBiaya string `json:"tingkatBiaya"`
}

func (h *Handler) CreateUser(c echo.Context) error {
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	if req.NomorHP == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Nomor HP is required")
	}

	user := models.User{
		Name:         req.Name,
		Email:        req.Email,
		Role:         req.Role,
		Password:     req.Password,
		NIP:          req.NIP,
		NomorHP:      req.NomorHP,
		Pangkat:      req.Pangkat,
		Golongan:     req.Golongan,
		Jabatan:      req.Jabatan,
		TingkatBiaya: req.TingkatBiaya,
	}

	if err := h.svc.CreateUser(&user); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, user)
}

type UpdateUserRequest struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	Password     string `json:"password"`
	NIP          string `json:"nip"`
	NomorHP      string `json:"nomorHp"`
	Pangkat      string `json:"pangkat"`
	Golongan     string `json:"golongan"`
	Jabatan      string `json:"jabatan"`
	TingkatBiaya string `json:"tingkatBiaya"`
}

func (h *Handler) UpdateUser(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid UUID format")
	}

	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	if req.NomorHP == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Nomor HP is required")
	}

	// Get existing user
	existingUser, err := h.svc.GetUserByID(id)
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
	if err := h.svc.UpdateUser(existingUser, req.Password); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, existingUser)
}

func (h *Handler) DeleteUser(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid UUID format")
	}

	if err := h.svc.DeleteUser(id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}
