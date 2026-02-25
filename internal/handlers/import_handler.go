package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// ImportStudentsCSV handles bulk student import via CSV.
//
// Expected CSV format (semicolon-delimited, UTF-8):
//
//	username;password;first_name;last_name;email;class_name;date_of_birth
//
// - email is optional (can be empty)
// - class_name must match an existing class name
// - date_of_birth format: YYYY-MM-DD
// - admin-only endpoint
func ImportStudentsCSV(db *pgxpool.Pool) echo.HandlerFunc {
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
		reader.TrimLeadingSpace = true

		// Read header
		header, err := reader.Read()
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid CSV: cannot read header")
		}

		// Build column index map
		colIdx := map[string]int{}
		for i, h := range header {
			colIdx[strings.TrimSpace(strings.ToLower(h))] = i
		}

		// Validate required columns
		required := []string{"username", "password", "first_name", "last_name", "date_of_birth"}
		for _, col := range required {
			if _, ok := colIdx[col]; !ok {
				return echo.NewHTTPError(http.StatusBadRequest,
					fmt.Sprintf("missing required column: %s (have: %s)", col, strings.Join(header, ", ")))
			}
		}

		// Pre-load classes for name→id mapping
		classMap := map[string]uuid.UUID{}
		rows, err := db.Query(c.Request().Context(),
			`SELECT id, name FROM classes WHERE school_id = $1`, schoolID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to load classes")
		}
		for rows.Next() {
			var id uuid.UUID
			var name string
			if err := rows.Scan(&id, &name); err == nil {
				classMap[strings.ToLower(strings.TrimSpace(name))] = id
			}
		}
		rows.Close()

		var imported int
		var errors []string
		rowNum := 1 // 1-indexed, header is row 0

		for {
			record, err := reader.Read()
			if err != nil {
				break
			}
			rowNum++

			getCol := func(name string) string {
				if idx, ok := colIdx[name]; ok && idx < len(record) {
					return strings.TrimSpace(record[idx])
				}
				return ""
			}

			username := getCol("username")
			password := getCol("password")
			firstName := getCol("first_name")
			lastName := getCol("last_name")
			email := getCol("email")
			className := getCol("class_name")
			dobStr := getCol("date_of_birth")

			// Validate
			if username == "" || password == "" || firstName == "" || lastName == "" {
				errors = append(errors, fmt.Sprintf("row %d: missing required fields", rowNum))
				continue
			}

			dob, err := time.Parse("2006-01-02", dobStr)
			if err != nil {
				errors = append(errors, fmt.Sprintf("row %d: invalid date_of_birth '%s'", rowNum, dobStr))
				continue
			}

			// Resolve class
			var classID *uuid.UUID
			if className != "" {
				if id, ok := classMap[strings.ToLower(className)]; ok {
					classID = &id
				} else {
					errors = append(errors, fmt.Sprintf("row %d: class '%s' not found", rowNum, className))
					continue
				}
			}

			// Hash password
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				errors = append(errors, fmt.Sprintf("row %d: password hash error", rowNum))
				continue
			}

			// Transaction: create user + student
			tx, err := db.Begin(c.Request().Context())
			if err != nil {
				errors = append(errors, fmt.Sprintf("row %d: tx error", rowNum))
				continue
			}

			var emailPtr *string
			if email != "" {
				emailPtr = &email
			}

			var userID uuid.UUID
			err = tx.QueryRow(c.Request().Context(),
				`INSERT INTO users (school_id, email, username, password_hash, role, first_name, last_name)
				 VALUES ($1, $2, $3, $4, 'student', $5, $6)
				 RETURNING id`,
				schoolID, emailPtr, username, string(hash), firstName, lastName,
			).Scan(&userID)
			if err != nil {
				tx.Rollback(c.Request().Context())
				if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
					errors = append(errors, fmt.Sprintf("row %d: user '%s' already exists", rowNum, username))
				} else {
					errors = append(errors, fmt.Sprintf("row %d: user create error: %v", rowNum, err))
				}
				continue
			}

			_, err = tx.Exec(c.Request().Context(),
				`INSERT INTO students (user_id, school_id, class_id, date_of_birth, attestation_required)
				 VALUES ($1, $2, $3, $4, false)`,
				userID, schoolID, classID, dob,
			)
			if err != nil {
				tx.Rollback(c.Request().Context())
				errors = append(errors, fmt.Sprintf("row %d: student create error: %v", rowNum, err))
				continue
			}

			if err := tx.Commit(c.Request().Context()); err != nil {
				errors = append(errors, fmt.Sprintf("row %d: commit error", rowNum))
				continue
			}

			imported++
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"imported": imported,
			"errors":   errors,
			"total":    rowNum - 1,
		})
	}
}
