// Package handlers contains HTTP handler stubs for all API endpoints.
// Each handler is a factory function that takes a *pgxpool.Pool and returns an echo.HandlerFunc.
// TODO: Implement each handler with actual database queries via the services layer.
package handlers

import (
	"net/http"

	"github.com/Monstroxx/eduko-backend/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func stub(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

// ── Auth ────────────────────────────────────────────────────

func Login(db *pgxpool.Pool, cfg *config.Config) echo.HandlerFunc  { return stub }
func Register(db *pgxpool.Pool, cfg *config.Config) echo.HandlerFunc { return stub }

// ── School ──────────────────────────────────────────────────

func GetSchool(db *pgxpool.Pool) echo.HandlerFunc           { return stub }
func UpdateSchool(db *pgxpool.Pool) echo.HandlerFunc        { return stub }
func GetSchoolSettings(db *pgxpool.Pool) echo.HandlerFunc   { return stub }
func UpdateSchoolSettings(db *pgxpool.Pool) echo.HandlerFunc { return stub }

// ── Classes ─────────────────────────────────────────────────

func ListClasses(db *pgxpool.Pool) echo.HandlerFunc       { return stub }
func CreateClass(db *pgxpool.Pool) echo.HandlerFunc       { return stub }
func GetClass(db *pgxpool.Pool) echo.HandlerFunc          { return stub }
func UpdateClass(db *pgxpool.Pool) echo.HandlerFunc       { return stub }
func DeleteClass(db *pgxpool.Pool) echo.HandlerFunc       { return stub }
func ListClassStudents(db *pgxpool.Pool) echo.HandlerFunc { return stub }

// ── Students ────────────────────────────────────────────────

func ListStudents(db *pgxpool.Pool) echo.HandlerFunc      { return stub }
func GetStudent(db *pgxpool.Pool) echo.HandlerFunc        { return stub }
func UpdateStudent(db *pgxpool.Pool) echo.HandlerFunc     { return stub }
func GetStudentAbsences(db *pgxpool.Pool) echo.HandlerFunc { return stub }
func GetStudentExcuses(db *pgxpool.Pool) echo.HandlerFunc { return stub }

// ── Teachers ────────────────────────────────────────────────

func ListTeachers(db *pgxpool.Pool) echo.HandlerFunc { return stub }
func GetTeacher(db *pgxpool.Pool) echo.HandlerFunc   { return stub }

// ── Timetable ───────────────────────────────────────────────

func GetTimetable(db *pgxpool.Pool) echo.HandlerFunc          { return stub }
func CreateTimetableEntry(db *pgxpool.Pool) echo.HandlerFunc   { return stub }
func UpdateTimetableEntry(db *pgxpool.Pool) echo.HandlerFunc   { return stub }
func DeleteTimetableEntry(db *pgxpool.Pool) echo.HandlerFunc   { return stub }

// ── Substitutions ───────────────────────────────────────────

func ListSubstitutions(db *pgxpool.Pool) echo.HandlerFunc   { return stub }
func CreateSubstitution(db *pgxpool.Pool) echo.HandlerFunc  { return stub }
func UpdateSubstitution(db *pgxpool.Pool) echo.HandlerFunc  { return stub }
func DeleteSubstitution(db *pgxpool.Pool) echo.HandlerFunc  { return stub }

// ── Attendance ──────────────────────────────────────────────

func RecordAttendance(db *pgxpool.Pool) echo.HandlerFunc    { return stub }
func UpdateAttendance(db *pgxpool.Pool) echo.HandlerFunc    { return stub }
func GetClassAttendance(db *pgxpool.Pool) echo.HandlerFunc  { return stub }
func GetAttendanceByDate(db *pgxpool.Pool) echo.HandlerFunc { return stub }

// ── Excuses ─────────────────────────────────────────────────

func CreateExcuse(db *pgxpool.Pool) echo.HandlerFunc     { return stub }
func ListExcuses(db *pgxpool.Pool) echo.HandlerFunc      { return stub }
func GetExcuse(db *pgxpool.Pool) echo.HandlerFunc        { return stub }
func ApproveExcuse(db *pgxpool.Pool) echo.HandlerFunc    { return stub }
func RejectExcuse(db *pgxpool.Pool) echo.HandlerFunc     { return stub }
func UploadExcuseForm(db *pgxpool.Pool) echo.HandlerFunc { return stub }
func GenerateExcusePDF(db *pgxpool.Pool) echo.HandlerFunc { return stub }
func ImportExcusesCSV(db *pgxpool.Pool) echo.HandlerFunc { return stub }

// ── Lesson Content ──────────────────────────────────────────

func CreateLessonContent(db *pgxpool.Pool) echo.HandlerFunc { return stub }
func UpdateLessonContent(db *pgxpool.Pool) echo.HandlerFunc { return stub }
func ListLessonContent(db *pgxpool.Pool) echo.HandlerFunc   { return stub }

// ── Appointments ────────────────────────────────────────────

func ListAppointments(db *pgxpool.Pool) echo.HandlerFunc   { return stub }
func CreateAppointment(db *pgxpool.Pool) echo.HandlerFunc  { return stub }
func UpdateAppointment(db *pgxpool.Pool) echo.HandlerFunc  { return stub }
func DeleteAppointment(db *pgxpool.Pool) echo.HandlerFunc  { return stub }

// ── Subjects & Rooms ────────────────────────────────────────

func ListSubjects(db *pgxpool.Pool) echo.HandlerFunc  { return stub }
func CreateSubject(db *pgxpool.Pool) echo.HandlerFunc  { return stub }
func ListRooms(db *pgxpool.Pool) echo.HandlerFunc     { return stub }
func CreateRoom(db *pgxpool.Pool) echo.HandlerFunc    { return stub }

// ── Time Slots ──────────────────────────────────────────────

func ListTimeSlots(db *pgxpool.Pool) echo.HandlerFunc  { return stub }
func CreateTimeSlot(db *pgxpool.Pool) echo.HandlerFunc { return stub }
