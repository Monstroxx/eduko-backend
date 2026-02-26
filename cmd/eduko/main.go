package main

import (
	"log"

	"github.com/Monstroxx/eduko-backend/internal/config"
	"github.com/Monstroxx/eduko-backend/internal/database"
	"github.com/Monstroxx/eduko-backend/internal/handlers"
	"github.com/Monstroxx/eduko-backend/internal/middleware"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(echomw.Logger())
	e.Use(echomw.Recover())
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins: cfg.CORSOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		AllowCredentials: false,
	}))

	// Health check (no auth)
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok", "version": "0.1.0"})
	})

	// Public routes
	api := e.Group("/api/v1")
	api.POST("/auth/login", handlers.Login(db, cfg))
	api.POST("/auth/register", handlers.Register(db, cfg))

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.JWT(cfg.JWTSecret))

	// School
	protected.GET("/school", handlers.GetSchool(db))
	protected.PUT("/school", handlers.UpdateSchool(db))
	protected.GET("/school/settings", handlers.GetSchoolSettings(db))
	protected.PUT("/school/settings", handlers.UpdateSchoolSettings(db))

	// Classes
	protected.GET("/classes", handlers.ListClasses(db))
	protected.POST("/classes", handlers.CreateClass(db))
	protected.GET("/classes/:id", handlers.GetClass(db))
	protected.PUT("/classes/:id", handlers.UpdateClass(db))
	protected.DELETE("/classes/:id", handlers.DeleteClass(db))
	protected.GET("/classes/:id/students", handlers.ListClassStudents(db))

	// Students
	protected.GET("/students", handlers.ListStudents(db))
	protected.GET("/students/:id", handlers.GetStudent(db))
	protected.PUT("/students/:id", handlers.UpdateStudent(db))
	protected.GET("/students/:id/absences", handlers.GetStudentAbsences(db))
	protected.POST("/students/import", handlers.ImportStudentsCSV(db))
	protected.GET("/students/:id/excuses", handlers.GetStudentExcuses(db))

	// Teachers
	protected.GET("/teachers", handlers.ListTeachers(db))
	protected.GET("/teachers/:id", handlers.GetTeacher(db))

	// Timetable
	protected.GET("/timetable", handlers.GetTimetable(db))
	protected.POST("/timetable", handlers.CreateTimetableEntry(db))
	protected.PUT("/timetable/:id", handlers.UpdateTimetableEntry(db))
	protected.DELETE("/timetable/:id", handlers.DeleteTimetableEntry(db))

	// Substitutions
	protected.GET("/substitutions", handlers.ListSubstitutions(db))
	protected.POST("/substitutions", handlers.CreateSubstitution(db))
	protected.PUT("/substitutions/:id", handlers.UpdateSubstitution(db))
	protected.DELETE("/substitutions/:id", handlers.DeleteSubstitution(db))

	// Attendance
	protected.POST("/attendance", handlers.RecordAttendance(db))
	protected.PUT("/attendance/:id", handlers.UpdateAttendance(db))
	protected.GET("/attendance/class/:classId", handlers.GetClassAttendance(db))
	protected.GET("/attendance/date/:date", handlers.GetAttendanceByDate(db))

	// Excuses
	protected.POST("/excuses", handlers.CreateExcuse(db))
	protected.GET("/excuses", handlers.ListExcuses(db))
	protected.GET("/excuses/:id", handlers.GetExcuse(db))
	protected.PATCH("/excuses/:id/approve", handlers.ApproveExcuse(db))
	protected.PATCH("/excuses/:id/reject", handlers.RejectExcuse(db))
	protected.POST("/excuses/upload", handlers.UploadExcuseForm(db))
	protected.GET("/excuses/:id/pdf", handlers.GenerateExcusePDF(db))
	protected.POST("/excuses/import", handlers.ImportExcusesCSV(db))

	// Lesson Content
	protected.POST("/lessons", handlers.CreateLessonContent(db))
	protected.PUT("/lessons/:id", handlers.UpdateLessonContent(db))
	protected.GET("/lessons", handlers.ListLessonContent(db))

	// Appointments
	protected.GET("/appointments", handlers.ListAppointments(db))
	protected.POST("/appointments", handlers.CreateAppointment(db))
	protected.PUT("/appointments/:id", handlers.UpdateAppointment(db))
	protected.DELETE("/appointments/:id", handlers.DeleteAppointment(db))

	// Subjects & Rooms
	protected.GET("/subjects", handlers.ListSubjects(db))
	protected.POST("/subjects", handlers.CreateSubject(db))
	protected.GET("/rooms", handlers.ListRooms(db))
	protected.POST("/rooms", handlers.CreateRoom(db))

	// Time Slots
	protected.GET("/timeslots", handlers.ListTimeSlots(db))
	protected.POST("/timeslots", handlers.CreateTimeSlot(db))

	port := cfg.Port
	if port == "" {
		port = "8080"
	}
	log.Printf("Eduko backend starting on :%s", port)
	e.Logger.Fatal(e.Start(":" + port))
}
