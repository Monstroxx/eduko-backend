package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Monstroxx/eduko-backend/internal/models"
)

type ClassService struct {
	db *pgxpool.Pool
}

func NewClassService(db *pgxpool.Pool) *ClassService {
	return &ClassService{db: db}
}

func (s *ClassService) List(ctx context.Context, schoolID uuid.UUID, schoolYear string) ([]models.Class, error) {
	query := `SELECT id, school_id, name, grade_level, class_teacher_id, school_year, created_at, updated_at
		 FROM classes WHERE school_id = $1`
	args := []interface{}{schoolID}

	if schoolYear != "" {
		query += ` AND school_year = $2`
		args = append(args, schoolYear)
	}
	query += ` ORDER BY name`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list classes: %w", err)
	}
	defer rows.Close()

	var classes []models.Class
	for rows.Next() {
		var c models.Class
		if err := rows.Scan(&c.ID, &c.SchoolID, &c.Name, &c.GradeLevel,
			&c.ClassTeacherID, &c.SchoolYear, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan class: %w", err)
		}
		classes = append(classes, c)
	}
	return classes, nil
}

type CreateClassInput struct {
	Name           string     `json:"name"`
	GradeLevel     *int       `json:"grade_level,omitempty"`
	ClassTeacherID *uuid.UUID `json:"class_teacher_id,omitempty"`
	SchoolYear     string     `json:"school_year"`
}

func (s *ClassService) Create(ctx context.Context, schoolID uuid.UUID, input CreateClassInput) (*models.Class, error) {
	var c models.Class
	err := s.db.QueryRow(ctx,
		`INSERT INTO classes (school_id, name, grade_level, class_teacher_id, school_year)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, school_id, name, grade_level, class_teacher_id, school_year, created_at, updated_at`,
		schoolID, input.Name, input.GradeLevel, input.ClassTeacherID, input.SchoolYear,
	).Scan(&c.ID, &c.SchoolID, &c.Name, &c.GradeLevel, &c.ClassTeacherID,
		&c.SchoolYear, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create class: %w", err)
	}
	return &c, nil
}

func (s *ClassService) GetByID(ctx context.Context, schoolID, classID uuid.UUID) (*models.Class, error) {
	var c models.Class
	err := s.db.QueryRow(ctx,
		`SELECT id, school_id, name, grade_level, class_teacher_id, school_year, created_at, updated_at
		 FROM classes WHERE id = $1 AND school_id = $2`, classID, schoolID,
	).Scan(&c.ID, &c.SchoolID, &c.Name, &c.GradeLevel, &c.ClassTeacherID,
		&c.SchoolYear, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get class: %w", err)
	}
	return &c, nil
}

func (s *ClassService) Update(ctx context.Context, schoolID, classID uuid.UUID, input CreateClassInput) (*models.Class, error) {
	var c models.Class
	err := s.db.QueryRow(ctx,
		`UPDATE classes SET name = $3, grade_level = $4, class_teacher_id = $5, school_year = $6, updated_at = now()
		 WHERE id = $1 AND school_id = $2
		 RETURNING id, school_id, name, grade_level, class_teacher_id, school_year, created_at, updated_at`,
		classID, schoolID, input.Name, input.GradeLevel, input.ClassTeacherID, input.SchoolYear,
	).Scan(&c.ID, &c.SchoolID, &c.Name, &c.GradeLevel, &c.ClassTeacherID,
		&c.SchoolYear, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update class: %w", err)
	}
	return &c, nil
}

func (s *ClassService) Delete(ctx context.Context, schoolID, classID uuid.UUID) error {
	_, err := s.db.Exec(ctx,
		`DELETE FROM classes WHERE id = $1 AND school_id = $2`, classID, schoolID)
	if err != nil {
		return fmt.Errorf("delete class: %w", err)
	}
	return nil
}

func (s *ClassService) ListStudents(ctx context.Context, schoolID, classID uuid.UUID) ([]models.Student, error) {
	rows, err := s.db.Query(ctx,
		`SELECT s.id, s.user_id, s.school_id, s.class_id, s.date_of_birth,
		        (s.date_of_birth <= CURRENT_DATE - INTERVAL '18 years') AS is_adult,
		        s.attestation_required, s.created_at, s.updated_at
		 FROM students s WHERE s.school_id = $1 AND s.class_id = $2
		 ORDER BY (SELECT last_name FROM users WHERE id = s.user_id)`,
		schoolID, classID)
	if err != nil {
		return nil, fmt.Errorf("list class students: %w", err)
	}
	defer rows.Close()

	var students []models.Student
	for rows.Next() {
		var st models.Student
		if err := rows.Scan(&st.ID, &st.UserID, &st.SchoolID, &st.ClassID,
			&st.DateOfBirth, &st.IsAdult, &st.AttestationRequired,
			&st.CreatedAt, &st.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan student: %w", err)
		}
		students = append(students, st)
	}
	return students, nil
}
