package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Monstroxx/eduko-backend/internal/models"
)

type AttendanceService struct {
	db *pgxpool.Pool
}

func NewAttendanceService(db *pgxpool.Pool) *AttendanceService {
	return &AttendanceService{db: db}
}

type RecordAttendanceInput struct {
	StudentID        uuid.UUID               `json:"student_id"`
	TimetableEntryID uuid.UUID               `json:"timetable_entry_id"`
	Date             string                  `json:"date"`
	Status           models.AttendanceStatus  `json:"status"`
	Note             *string                  `json:"note,omitempty"`
}

type BatchAttendanceInput struct {
	// Single record fields
	StudentID        uuid.UUID               `json:"student_id"`
	TimetableEntryID uuid.UUID               `json:"timetable_entry_id"`
	Date             string                  `json:"date"`
	Status           models.AttendanceStatus  `json:"status"`
	Note             *string                  `json:"note,omitempty"`
	// Batch fields
	Entries          []struct {
		StudentID uuid.UUID               `json:"student_id"`
		Status    models.AttendanceStatus  `json:"status"`
		Note      *string                  `json:"note,omitempty"`
	} `json:"entries"`
}

func (s *AttendanceService) Record(ctx context.Context, schoolID, recordedBy uuid.UUID, input RecordAttendanceInput) (*models.Attendance, error) {
	var a models.Attendance
	err := s.db.QueryRow(ctx,
		`INSERT INTO attendance (school_id, student_id, timetable_entry_id, date, status, recorded_by, note)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (student_id, timetable_entry_id, date)
		 DO UPDATE SET status = $5, note = $7, recorded_by = $6, updated_at = now()
		 RETURNING id, school_id, student_id, timetable_entry_id, date, status, recorded_by, note, created_at, updated_at`,
		schoolID, input.StudentID, input.TimetableEntryID, input.Date, input.Status, recordedBy, input.Note,
	).Scan(&a.ID, &a.SchoolID, &a.StudentID, &a.TimetableEntryID, &a.Date,
		&a.Status, &a.RecordedBy, &a.Note, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("record attendance: %w", err)
	}
	return &a, nil
}

func (s *AttendanceService) RecordBatch(ctx context.Context, schoolID, recordedBy uuid.UUID, input BatchAttendanceInput) (int, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	count := 0
	for _, entry := range input.Entries {
		_, err := tx.Exec(ctx,
			`INSERT INTO attendance (school_id, student_id, timetable_entry_id, date, status, recorded_by, note)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)
			 ON CONFLICT (student_id, timetable_entry_id, date)
			 DO UPDATE SET status = $5, note = $7, recorded_by = $6, updated_at = now()`,
			schoolID, entry.StudentID, input.TimetableEntryID, input.Date, entry.Status, recordedBy, entry.Note,
		)
		if err != nil {
			return 0, fmt.Errorf("record attendance entry: %w", err)
		}
		count++
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}
	return count, nil
}

func (s *AttendanceService) Update(ctx context.Context, schoolID, attendanceID uuid.UUID, status models.AttendanceStatus, note *string) (*models.Attendance, error) {
	var a models.Attendance
	err := s.db.QueryRow(ctx,
		`UPDATE attendance SET status = $3, note = $4, updated_at = now()
		 WHERE id = $1 AND school_id = $2
		 RETURNING id, school_id, student_id, timetable_entry_id, date, status, recorded_by, note, created_at, updated_at`,
		attendanceID, schoolID, status, note,
	).Scan(&a.ID, &a.SchoolID, &a.StudentID, &a.TimetableEntryID, &a.Date,
		&a.Status, &a.RecordedBy, &a.Note, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update attendance: %w", err)
	}
	return &a, nil
}

func (s *AttendanceService) GetByClass(ctx context.Context, schoolID, classID uuid.UUID, date string) ([]models.Attendance, error) {
	rows, err := s.db.Query(ctx,
		`SELECT a.id, a.school_id, a.student_id, a.timetable_entry_id, a.date, a.status, a.recorded_by, a.note, a.created_at, a.updated_at
		 FROM attendance a
		 JOIN students s ON s.id = a.student_id
		 WHERE a.school_id = $1 AND s.class_id = $2 AND a.date = $3
		 ORDER BY a.date, a.created_at`,
		schoolID, classID, date)
	if err != nil {
		return nil, fmt.Errorf("get class attendance: %w", err)
	}
	defer rows.Close()

	list := make([]models.Attendance, 0)
	for rows.Next() {
		var a models.Attendance
		if err := rows.Scan(&a.ID, &a.SchoolID, &a.StudentID, &a.TimetableEntryID, &a.Date,
			&a.Status, &a.RecordedBy, &a.Note, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan attendance: %w", err)
		}
		list = append(list, a)
	}
	return list, nil
}

func (s *AttendanceService) GetByDate(ctx context.Context, schoolID uuid.UUID, date string) ([]models.Attendance, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, school_id, student_id, timetable_entry_id, date, status, recorded_by, note, created_at, updated_at
		 FROM attendance WHERE school_id = $1 AND date = $2
		 ORDER BY created_at`,
		schoolID, date)
	if err != nil {
		return nil, fmt.Errorf("get attendance by date: %w", err)
	}
	defer rows.Close()

	list := make([]models.Attendance, 0)
	for rows.Next() {
		var a models.Attendance
		if err := rows.Scan(&a.ID, &a.SchoolID, &a.StudentID, &a.TimetableEntryID, &a.Date,
			&a.Status, &a.RecordedBy, &a.Note, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan attendance: %w", err)
		}
		list = append(list, a)
	}
	return list, nil
}
