package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Monstroxx/eduko-backend/internal/models"
)

type ExcuseService struct {
	db *pgxpool.Pool
}

func NewExcuseService(db *pgxpool.Pool) *ExcuseService {
	return &ExcuseService{db: db}
}

type CreateExcuseInput struct {
	DateFrom             string `json:"date_from"`
	DateTo               string `json:"date_to"`
	SubmissionType       string `json:"submission_type"`
	Reason               *string `json:"reason,omitempty"`
	AttestationProvided  bool   `json:"attestation_provided"`
}

type ExcuseWithLinks struct {
	models.Excuse
	LinkedAbsences int `json:"linked_absences"`
}

func (s *ExcuseService) Create(ctx context.Context, schoolID, studentID uuid.UUID, input CreateExcuseInput) (*ExcuseWithLinks, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var excuse models.Excuse
	err = tx.QueryRow(ctx,
		`INSERT INTO excuses (school_id, student_id, date_from, date_to, submission_type, status, reason, attestation_provided)
		 VALUES ($1, $2, $3, $4, $5, 'pending', $6, $7)
		 RETURNING id, school_id, student_id, date_from, date_to, submission_type, status, reason,
		           attestation_provided, file_path, submitted_at, approved_by, approved_at, created_at, updated_at`,
		schoolID, studentID, input.DateFrom, input.DateTo, input.SubmissionType,
		input.Reason, input.AttestationProvided,
	).Scan(&excuse.ID, &excuse.SchoolID, &excuse.StudentID, &excuse.DateFrom, &excuse.DateTo,
		&excuse.SubmissionType, &excuse.Status, &excuse.Reason, &excuse.AttestationProvided,
		&excuse.FilePath, &excuse.SubmittedAt, &excuse.ApprovedBy, &excuse.ApprovedAt,
		&excuse.CreatedAt, &excuse.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert excuse: %w", err)
	}

	// Auto-link to matching attendance records
	result, err := tx.Exec(ctx,
		`INSERT INTO excuse_attendance (excuse_id, attendance_id)
		 SELECT $1, a.id FROM attendance a
		 WHERE a.student_id = $2 AND a.school_id = $3
		   AND a.date >= $4 AND a.date <= $5
		   AND a.status = 'absent'`,
		excuse.ID, studentID, schoolID, input.DateFrom, input.DateTo)
	if err != nil {
		return nil, fmt.Errorf("link attendance: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &ExcuseWithLinks{
		Excuse:         excuse,
		LinkedAbsences: int(result.RowsAffected()),
	}, nil
}

func (s *ExcuseService) List(ctx context.Context, schoolID uuid.UUID, status, studentID, classID string) ([]models.Excuse, error) {
	query := `SELECT e.id, e.school_id, e.student_id, e.date_from, e.date_to, e.submission_type,
	                 e.status, e.reason, e.attestation_provided, e.file_path, e.submitted_at,
	                 e.approved_by, e.approved_at, e.created_at, e.updated_at
	          FROM excuses e`
	args := []interface{}{schoolID}
	where := ` WHERE e.school_id = $1`
	n := 2

	if status != "" {
		where += fmt.Sprintf(` AND e.status = $%d`, n)
		args = append(args, status)
		n++
	}
	if studentID != "" {
		where += fmt.Sprintf(` AND e.student_id = $%d`, n)
		args = append(args, studentID)
		n++
	}
	if classID != "" {
		query += ` JOIN students s ON s.id = e.student_id`
		where += fmt.Sprintf(` AND s.class_id = $%d`, n)
		args = append(args, classID)
		n++
	}

	rows, err := s.db.Query(ctx, query+where+` ORDER BY e.submitted_at DESC`, args...)
	if err != nil {
		return nil, fmt.Errorf("list excuses: %w", err)
	}
	defer rows.Close()

	var list []models.Excuse
	for rows.Next() {
		var e models.Excuse
		if err := rows.Scan(&e.ID, &e.SchoolID, &e.StudentID, &e.DateFrom, &e.DateTo,
			&e.SubmissionType, &e.Status, &e.Reason, &e.AttestationProvided,
			&e.FilePath, &e.SubmittedAt, &e.ApprovedBy, &e.ApprovedAt,
			&e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan excuse: %w", err)
		}
		list = append(list, e)
	}
	return list, nil
}

func (s *ExcuseService) GetByID(ctx context.Context, schoolID, excuseID uuid.UUID) (*models.Excuse, error) {
	var e models.Excuse
	err := s.db.QueryRow(ctx,
		`SELECT id, school_id, student_id, date_from, date_to, submission_type, status, reason,
		        attestation_provided, file_path, submitted_at, approved_by, approved_at, created_at, updated_at
		 FROM excuses WHERE id = $1 AND school_id = $2`, excuseID, schoolID,
	).Scan(&e.ID, &e.SchoolID, &e.StudentID, &e.DateFrom, &e.DateTo,
		&e.SubmissionType, &e.Status, &e.Reason, &e.AttestationProvided,
		&e.FilePath, &e.SubmittedAt, &e.ApprovedBy, &e.ApprovedAt,
		&e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get excuse: %w", err)
	}
	return &e, nil
}

func (s *ExcuseService) Approve(ctx context.Context, schoolID, excuseID, approvedBy uuid.UUID) (*models.Excuse, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var e models.Excuse
	now := time.Now()
	err = tx.QueryRow(ctx,
		`UPDATE excuses SET status = 'approved', approved_by = $3, approved_at = $4, updated_at = $4
		 WHERE id = $1 AND school_id = $2
		 RETURNING id, school_id, student_id, date_from, date_to, submission_type, status, reason,
		           attestation_provided, file_path, submitted_at, approved_by, approved_at, created_at, updated_at`,
		excuseID, schoolID, approvedBy, now,
	).Scan(&e.ID, &e.SchoolID, &e.StudentID, &e.DateFrom, &e.DateTo,
		&e.SubmissionType, &e.Status, &e.Reason, &e.AttestationProvided,
		&e.FilePath, &e.SubmittedAt, &e.ApprovedBy, &e.ApprovedAt,
		&e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("approve excuse: %w", err)
	}

	// Update linked attendance from absent → excused_leave
	_, err = tx.Exec(ctx,
		`UPDATE attendance SET status = 'excused_leave', updated_at = now()
		 WHERE id IN (SELECT attendance_id FROM excuse_attendance WHERE excuse_id = $1)`,
		excuseID)
	if err != nil {
		return nil, fmt.Errorf("update linked attendance: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return &e, nil
}

func (s *ExcuseService) Reject(ctx context.Context, schoolID, excuseID uuid.UUID, reason string) (*models.Excuse, error) {
	var e models.Excuse
	err := s.db.QueryRow(ctx,
		`UPDATE excuses SET status = 'rejected', updated_at = now()
		 WHERE id = $1 AND school_id = $2
		 RETURNING id, school_id, student_id, date_from, date_to, submission_type, status, reason,
		           attestation_provided, file_path, submitted_at, approved_by, approved_at, created_at, updated_at`,
		excuseID, schoolID,
	).Scan(&e.ID, &e.SchoolID, &e.StudentID, &e.DateFrom, &e.DateTo,
		&e.SubmissionType, &e.Status, &e.Reason, &e.AttestationProvided,
		&e.FilePath, &e.SubmittedAt, &e.ApprovedBy, &e.ApprovedAt,
		&e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("reject excuse: %w", err)
	}
	return &e, nil
}
