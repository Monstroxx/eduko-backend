package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/Monstroxx/eduko-backend/internal/services"
)

func GetTimetable(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewTimetableService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		entries, err := svc.GetEnriched(c.Request().Context(), schoolID,
			c.QueryParam("class_id"), c.QueryParam("teacher_id"), c.QueryParam("date"))
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get timetable")
		}
		return c.JSON(http.StatusOK, entries)
	}
}

func CreateTimetableEntry(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewTimetableService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		role := c.Get("role").(string)
		if role != "admin" {
			return echo.NewHTTPError(http.StatusForbidden, "admin only")
		}
		var req services.CreateTimetableInput
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}
		entry, err := svc.Create(c.Request().Context(), schoolID, req)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create entry")
		}
		return c.JSON(http.StatusCreated, entry)
	}
}

func UpdateTimetableEntry(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewTimetableService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		role := c.Get("role").(string)
		if role != "admin" {
			return echo.NewHTTPError(http.StatusForbidden, "admin only")
		}
		entryID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
		}
		var req services.CreateTimetableInput
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}
		entry, err := svc.Update(c.Request().Context(), schoolID, entryID, req)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update entry")
		}
		return c.JSON(http.StatusOK, entry)
	}
}

func DeleteTimetableEntry(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewTimetableService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		role := c.Get("role").(string)
		if role != "admin" {
			return echo.NewHTTPError(http.StatusForbidden, "admin only")
		}
		entryID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
		}
		if err := svc.Delete(c.Request().Context(), schoolID, entryID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete entry")
		}
		return c.NoContent(http.StatusNoContent)
	}
}
