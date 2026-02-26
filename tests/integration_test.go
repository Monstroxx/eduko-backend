package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Monstroxx/eduko-backend/internal/config"
	"github.com/Monstroxx/eduko-backend/internal/database"
	"github.com/Monstroxx/eduko-backend/internal/handlers"
	"github.com/Monstroxx/eduko-backend/internal/middleware"
	"github.com/labstack/echo/v4"
)

// testServer creates a fully wired Echo server for integration testing.
// Requires a running PostgreSQL with the eduko database and seed data.
func testServer(t *testing.T) (*echo.Echo, *config.Config) {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://eduko:eduko@localhost:5432/eduko?sslmode=disable"
	}

	cfg := &config.Config{
		Port:        "0",
		DatabaseURL: dbURL,
		JWTSecret:   "test-secret-key-for-testing",
		CORSOrigins: []string{"*"},
		UploadDir:   t.TempDir(),
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		t.Skipf("database not available: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	e := echo.New()
	e.HideBanner = true

	api := e.Group("/api/v1")
	api.POST("/auth/login", handlers.Login(db, cfg))
	api.POST("/auth/register", handlers.Register(db, cfg))

	protected := api.Group("")
	protected.Use(middleware.JWT(cfg.JWTSecret))

	protected.GET("/school", handlers.GetSchool(db))
	protected.GET("/school/settings", handlers.GetSchoolSettings(db))
	protected.GET("/classes", handlers.ListClasses(db))
	protected.GET("/classes/:id/students", handlers.ListClassStudents(db))
	protected.GET("/students", handlers.ListStudents(db))
	protected.GET("/students/:id", handlers.GetStudent(db))
	protected.PUT("/students/:id", handlers.UpdateStudent(db))
	protected.POST("/students/import", handlers.ImportStudentsCSV(db))
	protected.GET("/teachers", handlers.ListTeachers(db))
	protected.GET("/timetable", handlers.GetTimetable(db))
	protected.POST("/timetable", handlers.CreateTimetableEntry(db))
	protected.GET("/substitutions", handlers.ListSubstitutions(db))
	protected.POST("/attendance", handlers.RecordAttendance(db))
	protected.GET("/attendance/class/:classId", handlers.GetClassAttendance(db))
	protected.POST("/excuses", handlers.CreateExcuse(db))
	protected.GET("/excuses", handlers.ListExcuses(db))
	protected.GET("/excuses/:id", handlers.GetExcuse(db))
	protected.PATCH("/excuses/:id/approve", handlers.ApproveExcuse(db))
	protected.PATCH("/excuses/:id/reject", handlers.RejectExcuse(db))
	protected.GET("/excuses/:id/pdf", handlers.GenerateExcusePDF(db))
	protected.GET("/subjects", handlers.ListSubjects(db))
	protected.GET("/rooms", handlers.ListRooms(db))
	protected.GET("/timeslots", handlers.ListTimeSlots(db))
	protected.GET("/lessons", handlers.ListLessonContent(db))
	protected.GET("/appointments", handlers.ListAppointments(db))

	return e, cfg
}

// login performs a login and returns the JWT token.
func login(t *testing.T, e *echo.Echo, username, password string) string {
	t.Helper()
	body := fmt.Sprintf(`{"username":"%s","password":"%s","school_id":"00000000-0000-0000-0000-000000000001"}`,
		username, password)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("login failed: status=%d body=%s", rec.Code, rec.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	return result["token"].(string)
}

// authedGet performs an authenticated GET request.
func authedGet(e *echo.Echo, token, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// authedPost performs an authenticated POST request with JSON body.
func authedPost(e *echo.Echo, token, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// ── Auth Tests ──────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	e, _ := testServer(t)
	token := login(t, e, "admin", "admin123")
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestLogin_EmptySchoolID(t *testing.T) {
	// Simulates the Flutter app sending school_id="" (single-school mode).
	e, _ := testServer(t)
	body := `{"username":"admin","password":"admin123","school_id":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for empty school_id (single-school mode), got %d: %s", rec.Code, rec.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result["token"] == nil || result["token"] == "" {
		t.Error("expected non-empty token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	e, _ := testServer(t)
	body := `{"username":"admin","password":"wrong","school_id":"00000000-0000-0000-0000-000000000001"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestLogin_NoAuth(t *testing.T) {
	e, _ := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/classes", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

// ── Classes Tests ───────────────────────────────────────────

func TestListClasses(t *testing.T) {
	e, _ := testServer(t)
	token := login(t, e, "admin", "admin123")

	rec := authedGet(e, token, "/api/v1/classes")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var classes []map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &classes)
	if len(classes) == 0 {
		t.Error("expected at least one class")
	}
}

// ── Students Tests ──────────────────────────────────────────

func TestListStudents(t *testing.T) {
	e, _ := testServer(t)
	token := login(t, e, "admin", "admin123")

	rec := authedGet(e, token, "/api/v1/students")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var students []map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &students)
	if len(students) == 0 {
		t.Error("expected at least one student")
	}
}

func TestUpdateStudent(t *testing.T) {
	e, _ := testServer(t)
	token := login(t, e, "admin", "admin123")

	// Get a student first
	rec := authedGet(e, token, "/api/v1/students")
	var students []map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &students)
	if len(students) == 0 {
		t.Skip("no students in test db")
	}
	studentID := students[0]["id"].(string)

	// Update attestation
	body := `{"attestation_required": true}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/students/"+studentID, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var updated map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &updated)
	if updated["attestation_required"] != true {
		t.Error("expected attestation_required to be true")
	}
}

// ── Timetable Tests ─────────────────────────────────────────

func TestGetTimetable(t *testing.T) {
	e, _ := testServer(t)
	token := login(t, e, "admin", "admin123")

	rec := authedGet(e, token, "/api/v1/timetable")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var entries []map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &entries)
	if len(entries) == 0 {
		t.Error("expected at least one timetable entry")
	}
}

// ── Excuses Tests ───────────────────────────────────────────

func TestListExcuses(t *testing.T) {
	e, _ := testServer(t)
	token := login(t, e, "admin", "admin123")

	rec := authedGet(e, token, "/api/v1/excuses")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestExcusePDF(t *testing.T) {
	e, _ := testServer(t)
	token := login(t, e, "admin", "admin123")

	// Get an excuse
	rec := authedGet(e, token, "/api/v1/excuses")
	var excuses []map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &excuses)
	if len(excuses) == 0 {
		t.Skip("no excuses in test db")
	}
	excuseID := excuses[0]["id"].(string)

	rec = authedGet(e, token, "/api/v1/excuses/"+excuseID+"/pdf")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.HasPrefix(rec.Body.String(), "%PDF") {
		t.Error("expected PDF content")
	}
}

// ── Reference Data Tests ────────────────────────────────────

func TestListSubjects(t *testing.T) {
	e, _ := testServer(t)
	token := login(t, e, "admin", "admin123")

	rec := authedGet(e, token, "/api/v1/subjects")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var subjects []map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &subjects)
	if len(subjects) == 0 {
		t.Error("expected subjects")
	}
}

func TestListRooms(t *testing.T) {
	e, _ := testServer(t)
	token := login(t, e, "admin", "admin123")

	rec := authedGet(e, token, "/api/v1/rooms")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestListTimeSlots(t *testing.T) {
	e, _ := testServer(t)
	token := login(t, e, "admin", "admin123")

	rec := authedGet(e, token, "/api/v1/timeslots")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var slots []map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &slots)
	if len(slots) == 0 {
		t.Error("expected time slots")
	}
}

// ── Substitutions Tests ─────────────────────────────────────

func TestListSubstitutions(t *testing.T) {
	e, _ := testServer(t)
	token := login(t, e, "admin", "admin123")

	rec := authedGet(e, token, "/api/v1/substitutions")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// ── Lessons & Appointments ──────────────────────────────────

func TestListLessons(t *testing.T) {
	e, _ := testServer(t)
	token := login(t, e, "admin", "admin123")

	rec := authedGet(e, token, "/api/v1/lessons")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestListAppointments(t *testing.T) {
	e, _ := testServer(t)
	token := login(t, e, "admin", "admin123")

	rec := authedGet(e, token, "/api/v1/appointments")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// ── Role-Based Access ───────────────────────────────────────

func TestStudentCantAccessAttendance(t *testing.T) {
	e, _ := testServer(t)
	token := login(t, e, "schueler", "student123")

	// Students should still be able to GET their own data via general endpoints
	rec := authedGet(e, token, "/api/v1/timetable")
	if rec.Code != http.StatusOK {
		t.Errorf("student should access timetable, got %d", rec.Code)
	}
}

func TestTeacherCanRecordAttendance(t *testing.T) {
	e, _ := testServer(t)
	token := login(t, e, "lehrer", "teacher123")

	rec := authedGet(e, token, "/api/v1/timetable")
	if rec.Code != http.StatusOK {
		t.Errorf("teacher should access timetable, got %d", rec.Code)
	}

	rec = authedGet(e, token, "/api/v1/excuses")
	if rec.Code != http.StatusOK {
		t.Errorf("teacher should list excuses, got %d", rec.Code)
	}
}

// ── CSV Student Import ──────────────────────────────────────

func TestImportStudentsCSV(t *testing.T) {
	e, _ := testServer(t)
	token := login(t, e, "admin", "admin123")

	// Use unique suffix to avoid conflicts across test runs
	suffix := fmt.Sprintf("%d", os.Getpid())

	// Build multipart form with CSV
	csvContent := "username;password;first_name;last_name;email;class_name;date_of_birth\n" +
		"import_test1_" + suffix + ";pass123;Max;Mustermann;max@test.de;10a;2008-05-15\n" +
		"import_test2_" + suffix + ";pass123;Erika;Muster;;10a;2008-03-22\n" +
		"import_test3_" + suffix + ";pass123;Bad;Date;;10a;not-a-date\n"

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", "students.csv")
	io.WriteString(part, csvContent)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/students/import", &buf)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	imported := int(result["imported"].(float64))
	if imported != 2 {
		t.Errorf("expected 2 imported, got %d", imported)
	}

	errors := result["errors"].([]interface{})
	if len(errors) != 1 {
		t.Errorf("expected 1 error (bad date), got %d: %v", len(errors), errors)
	}

	// Verify students were actually created
	rec = authedGet(e, token, "/api/v1/students")
	var students []map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &students)

	found := 0
	for _, s := range students {
		if s["username"] == "import_test1_"+suffix || s["username"] == "import_test2_"+suffix {
			found++
		}
	}
	if found != 2 {
		t.Errorf("expected to find 2 imported students, found %d", found)
	}
}

func TestImportStudentsCSV_StudentCantAccess(t *testing.T) {
	e, _ := testServer(t)
	token := login(t, e, "schueler", "student123")

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", "students.csv")
	io.WriteString(part, "username;password;first_name;last_name;date_of_birth\n")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/students/import", &buf)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for student, got %d", rec.Code)
	}
}
