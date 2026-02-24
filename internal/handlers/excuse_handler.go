package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/Monstroxx/eduko-backend/internal/services"
)

func CreateExcuse(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewExcuseService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		userID := c.Get("user_id").(uuid.UUID)

		var req services.CreateExcuseInput
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}

		// Look up student ID from user ID
		var studentID uuid.UUID
		err := db.QueryRow(c.Request().Context(),
			`SELECT id FROM students WHERE user_id = $1 AND school_id = $2`, userID, schoolID,
		).Scan(&studentID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "user is not a student")
		}

		// TODO: validate deadline, attestation rules from school settings

		result, err := svc.Create(c.Request().Context(), schoolID, studentID, req)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create excuse")
		}
		return c.JSON(http.StatusCreated, result)
	}
}

func ListExcuses(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewExcuseService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		status := c.QueryParam("status")
		studentID := c.QueryParam("student_id")
		classID := c.QueryParam("class_id")

		list, err := svc.List(c.Request().Context(), schoolID, status, studentID, classID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to list excuses")
		}
		return c.JSON(http.StatusOK, list)
	}
}

func GetExcuse(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewExcuseService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		excuseID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
		}

		excuse, err := svc.GetByID(c.Request().Context(), schoolID, excuseID)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "excuse not found")
		}
		return c.JSON(http.StatusOK, excuse)
	}
}

func ApproveExcuse(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewExcuseService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		userID := c.Get("user_id").(uuid.UUID)
		role := c.Get("role").(string)
		if role != "teacher" && role != "admin" {
			return echo.NewHTTPError(http.StatusForbidden, "teachers only")
		}

		excuseID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
		}

		excuse, err := svc.Approve(c.Request().Context(), schoolID, excuseID, userID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to approve excuse")
		}
		return c.JSON(http.StatusOK, excuse)
	}
}

func RejectExcuse(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewExcuseService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		role := c.Get("role").(string)
		if role != "teacher" && role != "admin" {
			return echo.NewHTTPError(http.StatusForbidden, "teachers only")
		}

		excuseID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
		}

		var req struct {
			Reason string `json:"reason"`
		}
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}

		excuse, err := svc.Reject(c.Request().Context(), schoolID, excuseID, req.Reason)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to reject excuse")
		}
		return c.JSON(http.StatusOK, excuse)
	}
}

func UploadExcuseForm(db *pgxpool.Pool) echo.HandlerFunc {
	return stub // TODO: file upload handling
}

func GenerateExcusePDF(db *pgxpool.Pool) echo.HandlerFunc {
	return stub // TODO: PDF generation
}

func ImportExcusesCSV(db *pgxpool.Pool) echo.HandlerFunc {
	return stub // TODO: CSV import
}
