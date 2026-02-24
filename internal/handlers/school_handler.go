package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/Monstroxx/eduko-backend/internal/services"
)

func GetSchool(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewSchoolService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		school, err := svc.GetByID(c.Request().Context(), schoolID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get school")
		}
		return c.JSON(http.StatusOK, school)
	}
}

func UpdateSchool(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewSchoolService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		role := c.Get("role").(string)
		if role != "admin" {
			return echo.NewHTTPError(http.StatusForbidden, "admin only")
		}

		var req struct {
			Name       string `json:"name"`
			Address    string `json:"address"`
			SchoolType string `json:"school_type"`
			Locale     string `json:"locale"`
			Timezone   string `json:"timezone"`
		}
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}

		school, err := svc.Update(c.Request().Context(), schoolID, req.Name, req.Address, req.SchoolType, req.Locale, req.Timezone)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update school")
		}
		return c.JSON(http.StatusOK, school)
	}
}

func GetSchoolSettings(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewSchoolService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		settings, err := svc.GetSettings(c.Request().Context(), schoolID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get settings")
		}
		return c.JSON(http.StatusOK, settings)
	}
}

func UpdateSchoolSettings(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewSchoolService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		role := c.Get("role").(string)
		if role != "admin" {
			return echo.NewHTTPError(http.StatusForbidden, "admin only")
		}

		var req struct {
			Key   string      `json:"key"`
			Value interface{} `json:"value"`
		}
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}

		if err := svc.UpdateSetting(c.Request().Context(), schoolID, req.Key, req.Value); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update setting")
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}
}
