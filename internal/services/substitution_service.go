package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Monstroxx/eduko-backend/internal/models"
)

type SubstitutionService struct {
	db *pgxpool.Pool
}

func NewSubstitutionService(db *pgxpool.Pool) *SubstitutionService {
	return &SubstitutionService{db: db}
}

type CreateSubstitutionInput struct {
	TimetableEntryID    uuid.UUID `json:"timetable_entry_id"`
	Date                string    `json:"date"`
	Type                string    `json:"type"`
	SubstituteTeacherID *uuid.UUID `json:"substitute_teacher_id,omitempty"`
	SubstituteRoomID    *uuid.UUID `json:"substitute_room_id,omitempty"`
	SubstituteSubjectID *uuid.UUID `json:"substitute_subject_id,omitempty"`
	Note                *string    `json:"note,omitempty"`
}

func (s *SubstitutionService) List(ctx context.Context, schoolID uuid.UUID, date, from, to string) ([]models.Substitution, error) {
	query := `SELECT id, school_id, timetable_entry_id, date, type, substitute_teacher_id,
	                 substitute_room_id, substitute_subject_id, note, created_by, created_at, updated_at
	          FROM substitutions WHERE school_id = $1`
	args := []interface{}{schoolID}
	n := 2

	if date != "" {
		query += fmt.Sprintf(` AND date = $%d`, n)
		args = append(args, date)
		n++
	} else {
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
	}
	query += ` ORDER BY date, created_at`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list substitutions: %w", err)
	}
	defer rows.Close()

	var list []models.Substitution
	for rows.Next() {
		var sub models.Substitution
		if err := rows.Scan(&sub.ID, &sub.SchoolID, &sub.TimetableEntryID, &sub.Date, &sub.Type,
			&sub.SubstituteTeacherID, &sub.SubstituteRoomID, &sub.SubstituteSubjectID,
			&sub.Note, &sub.CreatedBy, &sub.CreatedAt, &sub.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan substitution: %w", err)
		}
		list = append(list, sub)
	}
	return list, nil
}

func (s *SubstitutionService) Create(ctx context.Context, schoolID, createdBy uuid.UUID, input CreateSubstitutionInput) (*models.Substitution, error) {
	var sub models.Substitution
	err := s.db.QueryRow(ctx,
		`INSERT INTO substitutions (school_id, timetable_entry_id, date, type, substitute_teacher_id, substitute_room_id, substitute_subject_id, note, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, school_id, timetable_entry_id, date, type, substitute_teacher_id, substitute_room_id, substitute_subject_id, note, created_by, created_at, updated_at`,
		schoolID, input.TimetableEntryID, input.Date, input.Type, input.SubstituteTeacherID,
		input.SubstituteRoomID, input.SubstituteSubjectID, input.Note, createdBy,
	).Scan(&sub.ID, &sub.SchoolID, &sub.TimetableEntryID, &sub.Date, &sub.Type,
		&sub.SubstituteTeacherID, &sub.SubstituteRoomID, &sub.SubstituteSubjectID,
		&sub.Note, &sub.CreatedBy, &sub.CreatedAt, &sub.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create substitution: %w", err)
	}
	return &sub, nil
}

func (s *SubstitutionService) Update(ctx context.Context, schoolID, subID uuid.UUID, input CreateSubstitutionInput) (*models.Substitution, error) {
	var sub models.Substitution
	err := s.db.QueryRow(ctx,
		`UPDATE substitutions SET timetable_entry_id=$3, date=$4, type=$5, substitute_teacher_id=$6,
		        substitute_room_id=$7, substitute_subject_id=$8, note=$9, updated_at=now()
		 WHERE id=$1 AND school_id=$2
		 RETURNING id, school_id, timetable_entry_id, date, type, substitute_teacher_id, substitute_room_id, substitute_subject_id, note, created_by, created_at, updated_at`,
		subID, schoolID, input.TimetableEntryID, input.Date, input.Type, input.SubstituteTeacherID,
		input.SubstituteRoomID, input.SubstituteSubjectID, input.Note,
	).Scan(&sub.ID, &sub.SchoolID, &sub.TimetableEntryID, &sub.Date, &sub.Type,
		&sub.SubstituteTeacherID, &sub.SubstituteRoomID, &sub.SubstituteSubjectID,
		&sub.Note, &sub.CreatedBy, &sub.CreatedAt, &sub.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update substitution: %w", err)
	}
	return &sub, nil
}

func (s *SubstitutionService) Delete(ctx context.Context, schoolID, subID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `DELETE FROM substitutions WHERE id=$1 AND school_id=$2`, subID, schoolID)
	if err != nil {
		return fmt.Errorf("delete substitution: %w", err)
	}
	return nil
}
