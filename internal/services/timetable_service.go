package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Monstroxx/eduko-backend/internal/models"
)

type TimetableService struct {
	db *pgxpool.Pool
}

func NewTimetableService(db *pgxpool.Pool) *TimetableService {
	return &TimetableService{db: db}
}

func (s *TimetableService) Get(ctx context.Context, schoolID uuid.UUID, classID, teacherID, date string) ([]models.TimetableEntry, error) {
	query := `SELECT t.id, t.school_id, t.class_id, t.subject_id, t.teacher_id, t.room_id,
	                 t.time_slot_id, t.day_of_week, t.week_type, t.valid_from, t.valid_until,
	                 t.created_at, t.updated_at
	          FROM timetable_entries t WHERE t.school_id = $1`
	args := []interface{}{schoolID}
	n := 2

	if classID != "" {
		query += fmt.Sprintf(` AND t.class_id = $%d`, n)
		args = append(args, classID)
		n++
	}
	if teacherID != "" {
		query += fmt.Sprintf(` AND t.teacher_id = $%d`, n)
		args = append(args, teacherID)
		n++
	}

	query += ` AND (t.valid_until IS NULL OR t.valid_until >= CURRENT_DATE)`
	query += ` ORDER BY t.day_of_week, (SELECT slot_number FROM time_slots WHERE id = t.time_slot_id)`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get timetable: %w", err)
	}
	defer rows.Close()

	var entries []models.TimetableEntry
	for rows.Next() {
		var e models.TimetableEntry
		if err := rows.Scan(&e.ID, &e.SchoolID, &e.ClassID, &e.SubjectID, &e.TeacherID,
			&e.RoomID, &e.TimeSlotID, &e.DayOfWeek, &e.WeekType,
			&e.ValidFrom, &e.ValidUntil, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan timetable: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, nil
}

type CreateTimetableInput struct {
	ClassID    uuid.UUID        `json:"class_id"`
	SubjectID  uuid.UUID        `json:"subject_id"`
	TeacherID  uuid.UUID        `json:"teacher_id"`
	RoomID     *uuid.UUID       `json:"room_id,omitempty"`
	TimeSlotID uuid.UUID        `json:"time_slot_id"`
	DayOfWeek  int              `json:"day_of_week"`
	WeekType   models.WeekType  `json:"week_type"`
	ValidFrom  string           `json:"valid_from"`
	ValidUntil *string          `json:"valid_until,omitempty"`
}

func (s *TimetableService) Create(ctx context.Context, schoolID uuid.UUID, input CreateTimetableInput) (*models.TimetableEntry, error) {
	var e models.TimetableEntry
	err := s.db.QueryRow(ctx,
		`INSERT INTO timetable_entries (school_id, class_id, subject_id, teacher_id, room_id, time_slot_id, day_of_week, week_type, valid_from, valid_until)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 RETURNING id, school_id, class_id, subject_id, teacher_id, room_id, time_slot_id, day_of_week, week_type, valid_from, valid_until, created_at, updated_at`,
		schoolID, input.ClassID, input.SubjectID, input.TeacherID, input.RoomID,
		input.TimeSlotID, input.DayOfWeek, input.WeekType, input.ValidFrom, input.ValidUntil,
	).Scan(&e.ID, &e.SchoolID, &e.ClassID, &e.SubjectID, &e.TeacherID,
		&e.RoomID, &e.TimeSlotID, &e.DayOfWeek, &e.WeekType,
		&e.ValidFrom, &e.ValidUntil, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create timetable entry: %w", err)
	}
	return &e, nil
}

func (s *TimetableService) Update(ctx context.Context, schoolID, entryID uuid.UUID, input CreateTimetableInput) (*models.TimetableEntry, error) {
	var e models.TimetableEntry
	err := s.db.QueryRow(ctx,
		`UPDATE timetable_entries SET class_id=$3, subject_id=$4, teacher_id=$5, room_id=$6,
		        time_slot_id=$7, day_of_week=$8, week_type=$9, valid_from=$10, valid_until=$11, updated_at=now()
		 WHERE id = $1 AND school_id = $2
		 RETURNING id, school_id, class_id, subject_id, teacher_id, room_id, time_slot_id, day_of_week, week_type, valid_from, valid_until, created_at, updated_at`,
		entryID, schoolID, input.ClassID, input.SubjectID, input.TeacherID, input.RoomID,
		input.TimeSlotID, input.DayOfWeek, input.WeekType, input.ValidFrom, input.ValidUntil,
	).Scan(&e.ID, &e.SchoolID, &e.ClassID, &e.SubjectID, &e.TeacherID,
		&e.RoomID, &e.TimeSlotID, &e.DayOfWeek, &e.WeekType,
		&e.ValidFrom, &e.ValidUntil, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update timetable entry: %w", err)
	}
	return &e, nil
}

func (s *TimetableService) Delete(ctx context.Context, schoolID, entryID uuid.UUID) error {
	_, err := s.db.Exec(ctx,
		`DELETE FROM timetable_entries WHERE id = $1 AND school_id = $2`, entryID, schoolID)
	if err != nil {
		return fmt.Errorf("delete timetable entry: %w", err)
	}
	return nil
}
