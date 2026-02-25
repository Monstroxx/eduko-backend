package handlers

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		excuseID, err := uuid.Parse(c.FormValue("excuse_id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid excuse_id")
		}

		file, err := c.FormFile("file")
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "file required")
		}

		// Max 10MB
		if file.Size > 10<<20 {
			return echo.NewHTTPError(http.StatusBadRequest, "file too large (max 10MB)")
		}

		uploadDir := os.Getenv("UPLOAD_DIR")
		if uploadDir == "" {
			uploadDir = "./uploads"
		}

		// Ensure directory exists.
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "upload dir error")
		}

		ext := filepath.Ext(file.Filename)
		filename := fmt.Sprintf("%s_%s%s", excuseID.String(), uuid.New().String()[:8], ext)
		dst := filepath.Join(uploadDir, filename)

		src, err := file.Open()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to read file")
		}
		defer src.Close()

		out, err := os.Create(dst)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to save file")
		}
		defer out.Close()

		if _, err := io.Copy(out, src); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to write file")
		}

		// Update excuse with file path.
		_, err = db.Exec(c.Request().Context(),
			`UPDATE excuses SET file_path = $1, updated_at = NOW() WHERE id = $2 AND school_id = $3`,
			filename, excuseID, schoolID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update excuse")
		}

		return c.JSON(http.StatusOK, map[string]string{
			"message":   "file uploaded",
			"file_path": filename,
		})
	}
}

func GenerateExcusePDF(db *pgxpool.Pool) echo.HandlerFunc {
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

		// Get student + school info for the PDF.
		var studentName, schoolName string
		_ = db.QueryRow(c.Request().Context(),
			`SELECT u.first_name || ' ' || u.last_name FROM students s JOIN users u ON u.id = s.user_id WHERE s.id = $1`,
			excuse.StudentID).Scan(&studentName)
		_ = db.QueryRow(c.Request().Context(),
			`SELECT name FROM schools WHERE id = $1`, schoolID).Scan(&schoolName)

		pdf := generateExcusePDFContent(excuse.DateFrom, excuse.DateTo, excuse.Status, excuse.Reason, excuse.SubmittedAt, excuse.SubmissionType, studentName, schoolName)

		c.Response().Header().Set("Content-Type", "application/pdf")
		c.Response().Header().Set("Content-Disposition",
			fmt.Sprintf("attachment; filename=entschuldigung_%s.pdf", excuseID.String()[:8]))
		return c.Blob(http.StatusOK, "application/pdf", pdf)
	}
}

// generateExcusePDFContent creates a minimal PDF document.
// Uses raw PDF syntax to avoid external dependencies.
func generateExcusePDFContent(dateFrom, dateTo time.Time, status interface{}, reason *string, submittedAt time.Time, submissionType interface{}, studentName, schoolName string) []byte {
	reasonStr := "–"
	if reason != nil {
		reasonStr = *reason
	}

	content := fmt.Sprintf(
		"Entschuldigung\n\nSchule: %s\nSchueler: %s\nZeitraum: %s - %s\nStatus: %v\nGrund: %s\nEingereicht: %s\nTyp: %v",
		schoolName, studentName, dateFrom.Format("02.01.2006"), dateTo.Format("02.01.2006"),
		status, reasonStr, submittedAt.Format("02.01.2006 15:04"), submissionType,
	)

	// Minimal valid PDF with text content.
	stream := fmt.Sprintf("BT /F1 12 Tf 50 750 Td (%s) Tj ET", escapePDF(content))
	streamLen := len(stream)

	pdf := fmt.Sprintf(`%%PDF-1.4
1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj
2 0 obj<</Type/Pages/Kids[3 0 R]/Count 1>>endobj
3 0 obj<</Type/Page/Parent 2 0 R/MediaBox[0 0 595 842]/Contents 4 0 R/Resources<</Font<</F1 5 0 R>>>>>>endobj
4 0 obj<</Length %d>>
stream
%s
endstream
endobj
5 0 obj<</Type/Font/Subtype/Type1/BaseFont/Helvetica>>endobj
xref
0 6
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
0000000115 00000 n 
trailer<</Size 6/Root 1 0 R>>
startxref
0
%%%%EOF`, streamLen, stream)

	return []byte(pdf)
}

func escapePDF(s string) string {
	r := strings.NewReplacer(
		"\\", "\\\\",
		"(", "\\(",
		")", "\\)",
		"\n", ") Tj 0 -16 Td (",
	)
	return r.Replace(s)
}

func ImportExcusesCSV(db *pgxpool.Pool) echo.HandlerFunc {
	svc := services.NewExcuseService(db)
	return func(c echo.Context) error {
		schoolID := c.Get("school_id").(uuid.UUID)
		role := c.Get("role").(string)
		if role != "admin" {
			return echo.NewHTTPError(http.StatusForbidden, "admin only")
		}

		file, err := c.FormFile("file")
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "CSV file required")
		}

		src, err := file.Open()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to read file")
		}
		defer src.Close()

		reader := csv.NewReader(src)
		reader.Comma = ';'
		reader.LazyQuotes = true

		// Skip header
		if _, err := reader.Read(); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid CSV")
		}

		var imported int
		var errors []string

		for {
			record, err := reader.Read()
			if err != nil {
				break
			}

			// Expected columns: student_id;date_from;date_to;submission_type;reason
			if len(record) < 4 {
				errors = append(errors, fmt.Sprintf("row %d: not enough columns", imported+2))
				continue
			}

			studentID, err := uuid.Parse(strings.TrimSpace(record[0]))
			if err != nil {
				errors = append(errors, fmt.Sprintf("row %d: invalid student_id", imported+2))
				continue
			}

			dateFrom := strings.TrimSpace(record[1])
			if _, err := time.Parse("2006-01-02", dateFrom); err != nil {
				errors = append(errors, fmt.Sprintf("row %d: invalid date_from", imported+2))
				continue
			}

			dateTo := strings.TrimSpace(record[2])
			if _, err := time.Parse("2006-01-02", dateTo); err != nil {
				errors = append(errors, fmt.Sprintf("row %d: invalid date_to", imported+2))
				continue
			}

			submissionType := strings.TrimSpace(record[3])
			var reason *string
			if len(record) > 4 && strings.TrimSpace(record[4]) != "" {
				r := strings.TrimSpace(record[4])
				reason = &r
			}

			input := services.CreateExcuseInput{
				DateFrom:       dateFrom,
				DateTo:         dateTo,
				SubmissionType: submissionType,
				Reason:         reason,
			}

			_, err = svc.Create(c.Request().Context(), schoolID, studentID, input)
			if err != nil {
				errors = append(errors, fmt.Sprintf("row %d: %v", imported+2, err))
				continue
			}
			imported++
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"imported": imported,
			"errors":   errors,
		})
	}
}
