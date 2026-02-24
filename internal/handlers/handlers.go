package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/Monstroxx/eduko-backend/internal/services"
)

func stub(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

// ── Students ────────────────────────────────────────────────

func ListStudents(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewStudentService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		list, err := svc.List(c.Request().Context(), schoolID, c.QueryParam("class_id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to list students")
		}
		return c.JSON(http.StatusOK, list)
	}
}

func GetStudent(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewStudentService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
		}
		s, err := svc.GetByID(c.Request().Context(), schoolID, id)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "student not found")
		}
		return c.JSON(http.StatusOK, s)
	}
}

func UpdateStudent(db *pgxpool.Pool) echo.HandlerFunc { return stub }

func GetStudentAbsences(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewStudentService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
		}
		list, err := svc.GetAbsences(c.Request().Context(), schoolID, id, c.QueryParam("from"), c.QueryParam("to"))
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get absences")
		}
		return c.JSON(http.StatusOK, list)
	}
}

func GetStudentExcuses(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewExcuseService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		studentID := c.Param("id")
		list, err := svc.List(c.Request().Context(), schoolID, "", studentID, "")
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get excuses")
		}
		return c.JSON(http.StatusOK, list)
	}
}

// ── Teachers ────────────────────────────────────────────────

func ListTeachers(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewResourceService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		list, err := svc.ListTeachers(c.Request().Context(), schoolID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to list teachers")
		}
		return c.JSON(http.StatusOK, list)
	}
}

func GetTeacher(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewResourceService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
		}
		t, err := svc.GetTeacher(c.Request().Context(), schoolID, id)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "teacher not found")
		}
		return c.JSON(http.StatusOK, t)
	}
}

// ── Substitutions ───────────────────────────────────────────

func ListSubstitutions(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewSubstitutionService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		list, err := svc.List(c.Request().Context(), schoolID, c.QueryParam("date"), c.QueryParam("from"), c.QueryParam("to"))
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to list substitutions")
		}
		return c.JSON(http.StatusOK, list)
	}
}

func CreateSubstitution(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewSubstitutionService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		userID := c.Get("user_id").(uuid.UUID)
		role := c.Get("role").(string)
		if role != "admin" {
			return echo.NewHTTPError(http.StatusForbidden, "admin only")
		}
		var req services.CreateSubstitutionInput
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}
		sub, err := svc.Create(c.Request().Context(), schoolID, userID, req)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create substitution")
		}
		return c.JSON(http.StatusCreated, sub)
	}
}

func UpdateSubstitution(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewSubstitutionService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		role := c.Get("role").(string)
		if role != "admin" {
			return echo.NewHTTPError(http.StatusForbidden, "admin only")
		}
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
		}
		var req services.CreateSubstitutionInput
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}
		sub, err := svc.Update(c.Request().Context(), schoolID, id, req)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update substitution")
		}
		return c.JSON(http.StatusOK, sub)
	}
}

func DeleteSubstitution(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewSubstitutionService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		role := c.Get("role").(string)
		if role != "admin" {
			return echo.NewHTTPError(http.StatusForbidden, "admin only")
		}
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
		}
		if err := svc.Delete(c.Request().Context(), schoolID, id); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete substitution")
		}
		return c.NoContent(http.StatusNoContent)
	}
}

// ── Lesson Content ──────────────────────────────────────────

func CreateLessonContent(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewLessonService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		userID := c.Get("user_id").(uuid.UUID)
		var req services.CreateLessonInput
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}
		l, err := svc.Create(c.Request().Context(), schoolID, userID, req)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create lesson")
		}
		return c.JSON(http.StatusCreated, l)
	}
}

func UpdateLessonContent(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewLessonService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
		}
		var req services.CreateLessonInput
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}
		l, err := svc.Update(c.Request().Context(), schoolID, id, req)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update lesson")
		}
		return c.JSON(http.StatusOK, l)
	}
}

func ListLessonContent(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewLessonService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		list, err := svc.List(c.Request().Context(), schoolID, c.QueryParam("class_id"), c.QueryParam("subject_id"), c.QueryParam("from"), c.QueryParam("to"))
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to list lessons")
		}
		return c.JSON(http.StatusOK, list)
	}
}

// ── Appointments ────────────────────────────────────────────

func ListAppointments(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewAppointmentService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		list, err := svc.List(c.Request().Context(), schoolID, c.QueryParam("type"), c.QueryParam("class_id"), c.QueryParam("from"), c.QueryParam("to"))
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to list appointments")
		}
		return c.JSON(http.StatusOK, list)
	}
}

func CreateAppointment(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewAppointmentService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		userID := c.Get("user_id").(uuid.UUID)
		role := c.Get("role").(string)
		if role != "teacher" && role != "admin" {
			return echo.NewHTTPError(http.StatusForbidden, "teachers/admin only")
		}
		var req services.CreateAppointmentInput
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}
		a, err := svc.Create(c.Request().Context(), schoolID, userID, req)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create appointment")
		}
		return c.JSON(http.StatusCreated, a)
	}
}

func UpdateAppointment(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewAppointmentService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		role := c.Get("role").(string)
		if role != "teacher" && role != "admin" {
			return echo.NewHTTPError(http.StatusForbidden, "teachers/admin only")
		}
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
		}
		var req services.CreateAppointmentInput
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}
		a, err := svc.Update(c.Request().Context(), schoolID, id, req)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update appointment")
		}
		return c.JSON(http.StatusOK, a)
	}
}

func DeleteAppointment(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewAppointmentService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		role := c.Get("role").(string)
		if role != "teacher" && role != "admin" {
			return echo.NewHTTPError(http.StatusForbidden, "teachers/admin only")
		}
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
		}
		if err := svc.Delete(c.Request().Context(), schoolID, id); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete appointment")
		}
		return c.NoContent(http.StatusNoContent)
	}
}

// ── Subjects, Rooms, TimeSlots ──────────────────────────────

func ListSubjects(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewResourceService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		list, err := svc.ListSubjects(c.Request().Context(), schoolID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed")
		}
		return c.JSON(http.StatusOK, list)
	}
}

func CreateSubject(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewResourceService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		role := c.Get("role").(string)
		if role != "admin" {
			return echo.NewHTTPError(http.StatusForbidden, "admin only")
		}
		var req services.CreateSubjectInput
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}
		s, err := svc.CreateSubject(c.Request().Context(), schoolID, req)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed")
		}
		return c.JSON(http.StatusCreated, s)
	}
}

func ListRooms(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewResourceService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		list, err := svc.ListRooms(c.Request().Context(), schoolID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed")
		}
		return c.JSON(http.StatusOK, list)
	}
}

func CreateRoom(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewResourceService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		role := c.Get("role").(string)
		if role != "admin" {
			return echo.NewHTTPError(http.StatusForbidden, "admin only")
		}
		var req services.CreateRoomInput
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}
		r, err := svc.CreateRoom(c.Request().Context(), schoolID, req)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed")
		}
		return c.JSON(http.StatusCreated, r)
	}
}

func ListTimeSlots(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewResourceService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		list, err := svc.ListTimeSlots(c.Request().Context(), schoolID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed")
		}
		return c.JSON(http.StatusOK, list)
	}
}

func CreateTimeSlot(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewResourceService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		role := c.Get("role").(string)
		if role != "admin" {
			return echo.NewHTTPError(http.StatusForbidden, "admin only")
		}
		var req services.CreateTimeSlotInput
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}
		ts, err := svc.CreateTimeSlot(c.Request().Context(), schoolID, req)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed")
		}
		return c.JSON(http.StatusCreated, ts)
	}
}
