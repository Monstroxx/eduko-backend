package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/Monstroxx/eduko-backend/internal/services"
)

func ListClasses(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewClassService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		schoolYear := c.QueryParam("school_year")
		classes, err := svc.List(c.Request().Context(), schoolID, schoolYear)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to list classes")
		}
		return c.JSON(http.StatusOK, classes)
	}
}

func CreateClass(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewClassService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		role := c.Get("role").(string)
		if role != "admin" {
			return echo.NewHTTPError(http.StatusForbidden, "admin only")
		}
		var req services.CreateClassInput
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}
		class, err := svc.Create(c.Request().Context(), schoolID, req)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create class")
		}
		return c.JSON(http.StatusCreated, class)
	}
}

func GetClass(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewClassService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		classID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid class id")
		}
		class, err := svc.GetByID(c.Request().Context(), schoolID, classID)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "class not found")
		}
		return c.JSON(http.StatusOK, class)
	}
}

func UpdateClass(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewClassService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		role := c.Get("role").(string)
		if role != "admin" {
			return echo.NewHTTPError(http.StatusForbidden, "admin only")
		}
		classID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid class id")
		}
		var req services.CreateClassInput
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}
		class, err := svc.Update(c.Request().Context(), schoolID, classID, req)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update class")
		}
		return c.JSON(http.StatusOK, class)
	}
}

func DeleteClass(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewClassService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		role := c.Get("role").(string)
		if role != "admin" {
			return echo.NewHTTPError(http.StatusForbidden, "admin only")
		}
		classID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid class id")
		}
		if err := svc.Delete(c.Request().Context(), schoolID, classID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete class")
		}
		return c.NoContent(http.StatusNoContent)
	}
}

func ListClassStudents(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewClassService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		classID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid class id")
		}
		students, err := svc.ListStudents(c.Request().Context(), schoolID, classID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to list students")
		}
		return c.JSON(http.StatusOK, students)
	}
}
