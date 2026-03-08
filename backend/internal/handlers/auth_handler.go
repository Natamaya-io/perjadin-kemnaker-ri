package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

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
	ctx := context.Background()
	cacheKey := "demo_users"

	// 1. Try to get from Redis Cache first
	if h.Service.RedisClient != nil {
		cachedData, err := h.Service.RedisClient.Get(ctx, cacheKey).Result()
		if err == nil && cachedData != "" {
			var demoUsers []map[string]string
			if err := json.Unmarshal([]byte(cachedData), &demoUsers); err == nil {
				c.Logger().Info("Cache Hit for GetDemoUsers")
				return c.JSON(http.StatusOK, demoUsers)
			}
		}
	}

	// 2. Cache Miss, get from Database
	c.Logger().Info("Cache Miss for GetDemoUsers, fetching from DB")
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

	// 3. Save to Redis Cache (set expiration to 1 hour)
	if h.Service.RedisClient != nil {
		cacheBytes, err := json.Marshal(demoUsers)
		if err == nil {
			h.Service.RedisClient.Set(ctx, cacheKey, cacheBytes, time.Hour)
		}
	}

	return c.JSON(http.StatusOK, demoUsers)
}
