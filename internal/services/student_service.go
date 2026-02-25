package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Monstroxx/eduko-backend/internal/models"
)

type StudentService struct {
	db *pgxpool.Pool
}

func NewStudentService(db *pgxpool.Pool) *StudentService {
	return &StudentService{db: db}
}

type StudentWithUser struct {
	models.Student
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     *string `json:"email,omitempty"`
	Username  string `json:"username"`
}

func (s *StudentService) List(ctx context.Context, schoolID uuid.UUID, classID string) ([]StudentWithUser, error) {
	query := `SELECT s.id, s.user_id, s.school_id, s.class_id, s.date_of_birth,
	                 (s.date_of_birth <= CURRENT_DATE - INTERVAL '18 years') AS is_adult,
	                 s.attestation_required, s.created_at, s.updated_at,
	                 u.first_name, u.last_name, u.email, u.username
	          FROM students s JOIN users u ON u.id = s.user_id
	          WHERE s.school_id = $1`
	args := []interface{}{schoolID}
	if classID != "" {
		query += ` AND s.class_id = $2`
		args = append(args, classID)
	}
	query += ` ORDER BY u.last_name, u.first_name`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list students: %w", err)
	}
	defer rows.Close()

	var list []StudentWithUser
	for rows.Next() {
		var sw StudentWithUser
		if err := rows.Scan(&sw.ID, &sw.UserID, &sw.SchoolID, &sw.ClassID, &sw.DateOfBirth,
			&sw.IsAdult, &sw.AttestationRequired, &sw.CreatedAt, &sw.UpdatedAt,
			&sw.FirstName, &sw.LastName, &sw.Email, &sw.Username); err != nil {
			return nil, fmt.Errorf("scan student: %w", err)
		}
		list = append(list, sw)
	}
	return list, nil
}

func (s *StudentService) GetByID(ctx context.Context, schoolID, studentID uuid.UUID) (*StudentWithUser, error) {
	var sw StudentWithUser
	err := s.db.QueryRow(ctx,
		`SELECT s.id, s.user_id, s.school_id, s.class_id, s.date_of_birth,
		        (s.date_of_birth <= CURRENT_DATE - INTERVAL '18 years') AS is_adult,
		        s.attestation_required, s.created_at, s.updated_at,
		        u.first_name, u.last_name, u.email, u.username
		 FROM students s JOIN users u ON u.id = s.user_id
		 WHERE s.id = $1 AND s.school_id = $2`, studentID, schoolID,
	).Scan(&sw.ID, &sw.UserID, &sw.SchoolID, &sw.ClassID, &sw.DateOfBirth,
		&sw.IsAdult, &sw.AttestationRequired, &sw.CreatedAt, &sw.UpdatedAt,
		&sw.FirstName, &sw.LastName, &sw.Email, &sw.Username)
	if err != nil {
		return nil, fmt.Errorf("get student: %w", err)
	}
	return &sw, nil
}

type UpdateStudentInput struct {
	ClassID              *string `json:"class_id"`
	AttestationRequired  *bool   `json:"attestation_required"`
}

func (s *StudentService) Update(ctx context.Context, schoolID, studentID uuid.UUID, input UpdateStudentInput) (*StudentWithUser, error) {
	query := `UPDATE students SET updated_at = NOW()`
	args := []interface{}{}
	n := 1

	if input.ClassID != nil {
		query += fmt.Sprintf(`, class_id = $%d`, n)
		args = append(args, *input.ClassID)
		n++
	}
	if input.AttestationRequired != nil {
		query += fmt.Sprintf(`, attestation_required = $%d`, n)
		args = append(args, *input.AttestationRequired)
		n++
	}

	query += fmt.Sprintf(` WHERE id = $%d AND school_id = $%d`, n, n+1)
	args = append(args, studentID, schoolID)

	_, err := s.db.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("update student: %w", err)
	}

	return s.GetByID(ctx, schoolID, studentID)
}

func (s *StudentService) GetAbsences(ctx context.Context, schoolID, studentID uuid.UUID, from, to string) ([]models.Attendance, error) {
	query := `SELECT id, school_id, student_id, timetable_entry_id, date, status, recorded_by, note, created_at, updated_at
	          FROM attendance WHERE school_id = $1 AND student_id = $2 AND status != 'present'`
	args := []interface{}{schoolID, studentID}
	n := 3
	if from != "" {
		query += fmt.Sprintf(` AND date >= $%d`, n)
		args = append(args, from)
		n++
	}
	if to != "" {
		query += fmt.Sprintf(` AND date <= $%d`, n)
		args = append(args, to)
		n++
	}
	query += ` ORDER BY date DESC`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get absences: %w", err)
	}
	defer rows.Close()

	var list []models.Attendance
	for rows.Next() {
		var a models.Attendance
		if err := rows.Scan(&a.ID, &a.SchoolID, &a.StudentID, &a.TimetableEntryID, &a.Date,
			&a.Status, &a.RecordedBy, &a.Note, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan absence: %w", err)
		}
		list = append(list, a)
	}
	return list, nil
}
