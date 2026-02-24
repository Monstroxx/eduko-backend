package handlers

import (
	"errors"
	"net/http"

	"github.com/Monstroxx/eduko-backend/internal/config"
	"github.com/Monstroxx/eduko-backend/internal/services"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func Login(db *pgxpool.Pool, cfg *config.Config) echo.HandlerFunc {
	svc := services.NewAuthService(db)
	return func(c echo.Context) error {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
			SchoolID string `json:"school_id"`
		}
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}
		if req.Username == "" || req.Password == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "username and password required")
		}

		result, err := svc.Login(c.Request().Context(), req.Username, req.Password, req.SchoolID, cfg.JWTSecret)
		if err != nil {
			if errors.Is(err, services.ErrInvalidCredentials) {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "login failed")
		}

		return c.JSON(http.StatusOK, result)
	}
}

func Register(db *pgxpool.Pool, cfg *config.Config) echo.HandlerFunc {
	svc := services.NewAuthService(db)
	return func(c echo.Context) error {
		var req services.RegisterInput
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}
		if req.Username == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "missing required fields")
		}

		user, err := svc.Register(c.Request().Context(), req)
		if err != nil {
			if errors.Is(err, services.ErrUserExists) {
				return echo.NewHTTPError(http.StatusConflict, "user already exists")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "registration failed")
		}

		return c.JSON(http.StatusCreated, map[string]interface{}{"user": user})
	}
}
