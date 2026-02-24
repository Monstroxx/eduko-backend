package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	SchoolID uuid.UUID `json:"school_id"`
	Role     string    `json:"role"`
	jwt.RegisteredClaims
}

func JWT(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if auth == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization format")
			}

			token, err := jwt.ParseWithClaims(parts[1], &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			claims, ok := token.Claims.(*JWTClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid claims")
			}

			c.Set("user_id", claims.UserID)
			c.Set("school_id", claims.SchoolID)
			c.Set("role", claims.Role)

			return next(c)
		}
	}
}

// RequireRole creates middleware that restricts access to specific roles.
func RequireRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, ok := c.Get("role").(string)
			if !ok {
				return echo.NewHTTPError(http.StatusForbidden, "no role found")
			}
			for _, r := range roles {
				if r == role {
					return next(c)
				}
			}
			return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
		}
	}
}
