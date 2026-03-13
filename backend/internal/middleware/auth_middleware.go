package middleware

import (
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/domain/user"
	"github.com/labstack/echo/v4"
)

func JWTMiddleware(cfg *config.Config, repo user.Repository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(401, "Missing Authorization Header")
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				return echo.NewHTTPError(401, "Invalid Token Format")
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Don't forget to validate the alg is what you expect:
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.NewHTTPError(401, "Unexpected signing method")
				}
				return []byte(cfg.JWT.Secret), nil
			})

			if err != nil || !token.Valid {
				return echo.NewHTTPError(401, "Invalid or Expired Token")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return echo.NewHTTPError(401, "Invalid Token Claims")
			}

			// Validate Session
			userIDStr, ok := claims["user_id"].(string)
			if !ok {
				return echo.NewHTTPError(401, "Invalid User ID in Token")
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				return echo.NewHTTPError(401, "Invalid User ID Format")
			}

			user, err := repo.GetUserByID(userID)
			if err != nil {
				return echo.NewHTTPError(401, "User Not Found")
			}

			sessionID, ok := claims["session_id"].(string)
			// If session_id is missing in token (old tokens) or doesn't match DB, reject.
			// Allowing missing session_id only if you want to support rolling update without logging everyone out immediately.
			// But "One Account One User" implies strict check.
			if !ok || sessionID != user.SessionID {
				return echo.NewHTTPError(401, "Session Expired: Logged in from another device")
			}

			// Store user info in context
			c.Set("user_id", claims["user_id"])
			c.Set("role", claims["role"])
			c.Set("email", claims["email"])

			return next(c)
		}
	}
}

// RoleMiddleware checks if the user has one of the allowed roles
func RoleMiddleware(allowedRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRole := c.Get("role").(string)
			for _, role := range allowedRoles {
				if role == userRole {
					return next(c)
				}
			}
			return echo.NewHTTPError(403, "Forbidden: Insufficient Permissions")
		}
	}
}
