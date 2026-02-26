package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Monstroxx/eduko-backend/internal/models"
)

type AppointmentService struct {
	db *pgxpool.Pool
}

func NewAppointmentService(db *pgxpool.Pool) *AppointmentService {
	return &AppointmentService{db: db}
}

type CreateAppointmentInput struct {
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	Type        string     `json:"type"`
	Scope       string     `json:"scope"`
	ClassID     *uuid.UUID `json:"class_id,omitempty"`
	SubjectID   *uuid.UUID `json:"subject_id,omitempty"`
	Date        string     `json:"date"`
	TimeSlotID  *uuid.UUID `json:"time_slot_id,omitempty"`
}

func (s *AppointmentService) List(ctx context.Context, schoolID uuid.UUID, apType, classID, from, to string) ([]models.Appointment, error) {
	query := `SELECT id, school_id, title, description, type, scope, class_id, subject_id, date, time_slot_id, created_by, created_at, updated_at
	          FROM appointments WHERE school_id = $1`
	args := []interface{}{schoolID}
	n := 2

	if apType != "" {
		query += fmt.Sprintf(` AND type = $%d`, n)
		args = append(args, apType)
		n++
	}
	if classID != "" {
		query += fmt.Sprintf(` AND (class_id = $%d OR scope = 'school')`, n)
		args = append(args, classID)
		n++
	}
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
	query += ` ORDER BY date`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list appointments: %w", err)
	}
	defer rows.Close()

	list := make([]models.Appointment, 0)
	for rows.Next() {
		var a models.Appointment
		if err := rows.Scan(&a.ID, &a.SchoolID, &a.Title, &a.Description, &a.Type, &a.Scope,
			&a.ClassID, &a.SubjectID, &a.Date, &a.TimeSlotID, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan appointment: %w", err)
		}
		list = append(list, a)
	}
	return list, nil
}

func (s *AppointmentService) Create(ctx context.Context, schoolID, createdBy uuid.UUID, input CreateAppointmentInput) (*models.Appointment, error) {
	var a models.Appointment
	err := s.db.QueryRow(ctx,
		`INSERT INTO appointments (school_id, title, description, type, scope, class_id, subject_id, date, time_slot_id, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 RETURNING id, school_id, title, description, type, scope, class_id, subject_id, date, time_slot_id, created_by, created_at, updated_at`,
		schoolID, input.Title, input.Description, input.Type, input.Scope, input.ClassID, input.SubjectID, input.Date, input.TimeSlotID, createdBy,
	).Scan(&a.ID, &a.SchoolID, &a.Title, &a.Description, &a.Type, &a.Scope,
		&a.ClassID, &a.SubjectID, &a.Date, &a.TimeSlotID, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create appointment: %w", err)
	}
	return &a, nil
}

func (s *AppointmentService) Update(ctx context.Context, schoolID, apID uuid.UUID, input CreateAppointmentInput) (*models.Appointment, error) {
	var a models.Appointment
	err := s.db.QueryRow(ctx,
		`UPDATE appointments SET title=$3, description=$4, type=$5, scope=$6, class_id=$7, subject_id=$8, date=$9, time_slot_id=$10, updated_at=now()
		 WHERE id=$1 AND school_id=$2
		 RETURNING id, school_id, title, description, type, scope, class_id, subject_id, date, time_slot_id, created_by, created_at, updated_at`,
		apID, schoolID, input.Title, input.Description, input.Type, input.Scope, input.ClassID, input.SubjectID, input.Date, input.TimeSlotID,
	).Scan(&a.ID, &a.SchoolID, &a.Title, &a.Description, &a.Type, &a.Scope,
		&a.ClassID, &a.SubjectID, &a.Date, &a.TimeSlotID, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update appointment: %w", err)
	}
	return &a, nil
}

func (s *AppointmentService) Delete(ctx context.Context, schoolID, apID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `DELETE FROM appointments WHERE id=$1 AND school_id=$2`, apID, schoolID)
	if err != nil {
		return fmt.Errorf("delete appointment: %w", err)
	}
	return nil
}
