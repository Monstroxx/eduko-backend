// Package handlers contains HTTP handler factories for all API endpoints.
package handlers

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func stub(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

// ── Students (stubs) ────────────────────────────────────────

func ListStudents(db *pgxpool.Pool) echo.HandlerFunc      { return stub }
func GetStudent(db *pgxpool.Pool) echo.HandlerFunc         { return stub }
func UpdateStudent(db *pgxpool.Pool) echo.HandlerFunc      { return stub }
func GetStudentAbsences(db *pgxpool.Pool) echo.HandlerFunc { return stub }
func GetStudentExcuses(db *pgxpool.Pool) echo.HandlerFunc  { return stub }

// ── Teachers (stubs) ────────────────────────────────────────

func ListTeachers(db *pgxpool.Pool) echo.HandlerFunc { return stub }
func GetTeacher(db *pgxpool.Pool) echo.HandlerFunc   { return stub }

// ── Substitutions (stubs) ───────────────────────────────────

func ListSubstitutions(db *pgxpool.Pool) echo.HandlerFunc  { return stub }
func CreateSubstitution(db *pgxpool.Pool) echo.HandlerFunc { return stub }
func UpdateSubstitution(db *pgxpool.Pool) echo.HandlerFunc { return stub }
func DeleteSubstitution(db *pgxpool.Pool) echo.HandlerFunc { return stub }

// ── Lesson Content (stubs) ──────────────────────────────────

func CreateLessonContent(db *pgxpool.Pool) echo.HandlerFunc { return stub }
func UpdateLessonContent(db *pgxpool.Pool) echo.HandlerFunc { return stub }
func ListLessonContent(db *pgxpool.Pool) echo.HandlerFunc   { return stub }

// ── Appointments (stubs) ────────────────────────────────────

func ListAppointments(db *pgxpool.Pool) echo.HandlerFunc  { return stub }
func CreateAppointment(db *pgxpool.Pool) echo.HandlerFunc { return stub }
func UpdateAppointment(db *pgxpool.Pool) echo.HandlerFunc { return stub }
func DeleteAppointment(db *pgxpool.Pool) echo.HandlerFunc { return stub }

// ── Subjects & Rooms (stubs) ────────────────────────────────

func ListSubjects(db *pgxpool.Pool) echo.HandlerFunc { return stub }
func CreateSubject(db *pgxpool.Pool) echo.HandlerFunc { return stub }
func ListRooms(db *pgxpool.Pool) echo.HandlerFunc     { return stub }
func CreateRoom(db *pgxpool.Pool) echo.HandlerFunc    { return stub }

// ── Time Slots (stubs) ─────────────────────────────────────

func ListTimeSlots(db *pgxpool.Pool) echo.HandlerFunc  { return stub }
func CreateTimeSlot(db *pgxpool.Pool) echo.HandlerFunc { return stub }
