package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/Monstroxx/eduko-backend/internal/models"
	"github.com/Monstroxx/eduko-backend/internal/services"
)

func RecordAttendance(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewAttendanceService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		userID := c.Get("user_id").(uuid.UUID)
		role := c.Get("role").(string)
		if role != "teacher" && role != "admin" {
			return echo.NewHTTPError(http.StatusForbidden, "teachers only")
		}

		// Bind into batch struct — if entries is populated, do batch; otherwise single
		var batch services.BatchAttendanceInput
		if err := c.Bind(&batch); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}

		if len(batch.Entries) > 0 {
			count, err := svc.RecordBatch(c.Request().Context(), schoolID, userID, batch)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to record attendance")
			}
			return c.JSON(http.StatusOK, map[string]interface{}{"recorded": count})
		}

		// Single record — re-parse from batch fields
		req := services.RecordAttendanceInput{
			StudentID:        batch.StudentID,
			TimetableEntryID: batch.TimetableEntryID,
			Date:             batch.Date,
			Status:           batch.Status,
			Note:             batch.Note,
		}

		attendance, err := svc.Record(c.Request().Context(), schoolID, userID, req)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to record attendance")
		}
		return c.JSON(http.StatusCreated, attendance)
	}
}

func UpdateAttendance(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewAttendanceService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		attendanceID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
		}

		var req struct {
			Status models.AttendanceStatus `json:"status"`
			Note   *string                 `json:"note,omitempty"`
		}
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}

		attendance, err := svc.Update(c.Request().Context(), schoolID, attendanceID, req.Status, req.Note)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update attendance")
		}
		return c.JSON(http.StatusOK, attendance)
	}
}

func GetClassAttendance(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewAttendanceService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		classID, err := uuid.Parse(c.Param("classId"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid class id")
		}
		date := c.QueryParam("date")
		if date == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "date required")
		}

		list, err := svc.GetByClass(c.Request().Context(), schoolID, classID, date)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get attendance")
		}
		return c.JSON(http.StatusOK, list)
	}
}

func GetAttendanceByDate(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewAttendanceService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		date := c.Param("date")

		list, err := svc.GetByDate(c.Request().Context(), schoolID, date)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get attendance")
		}
		return c.JSON(http.StatusOK, list)
	}
}
