package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Monstroxx/eduko-backend/internal/models"
)

type LessonService struct {
	db *pgxpool.Pool
}

func NewLessonService(db *pgxpool.Pool) *LessonService {
	return &LessonService{db: db}
}

type CreateLessonInput struct {
	TimetableEntryID uuid.UUID `json:"timetable_entry_id"`
	Date             string    `json:"date"`
	Topic            string    `json:"topic"`
	Homework         *string   `json:"homework,omitempty"`
	Notes            *string   `json:"notes,omitempty"`
}

func (s *LessonService) Create(ctx context.Context, schoolID, recordedBy uuid.UUID, input CreateLessonInput) (*models.LessonContent, error) {
	var l models.LessonContent
	err := s.db.QueryRow(ctx,
		`INSERT INTO lesson_content (school_id, timetable_entry_id, date, topic, homework, notes, recorded_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (timetable_entry_id, date) DO UPDATE SET topic=$4, homework=$5, notes=$6, recorded_by=$7, updated_at=now()
		 RETURNING id, school_id, timetable_entry_id, date, topic, homework, notes, recorded_by, created_at, updated_at`,
		schoolID, input.TimetableEntryID, input.Date, input.Topic, input.Homework, input.Notes, recordedBy,
	).Scan(&l.ID, &l.SchoolID, &l.TimetableEntryID, &l.Date, &l.Topic, &l.Homework, &l.Notes, &l.RecordedBy, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create lesson: %w", err)
	}
	return &l, nil
}

func (s *LessonService) Update(ctx context.Context, schoolID, lessonID uuid.UUID, input CreateLessonInput) (*models.LessonContent, error) {
	var l models.LessonContent
	err := s.db.QueryRow(ctx,
		`UPDATE lesson_content SET topic=$3, homework=$4, notes=$5, updated_at=now()
		 WHERE id=$1 AND school_id=$2
		 RETURNING id, school_id, timetable_entry_id, date, topic, homework, notes, recorded_by, created_at, updated_at`,
		lessonID, schoolID, input.Topic, input.Homework, input.Notes,
	).Scan(&l.ID, &l.SchoolID, &l.TimetableEntryID, &l.Date, &l.Topic, &l.Homework, &l.Notes, &l.RecordedBy, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update lesson: %w", err)
	}
	return &l, nil
}

func (s *LessonService) List(ctx context.Context, schoolID uuid.UUID, classID, subjectID, from, to string) ([]models.LessonContent, error) {
	query := `SELECT l.id, l.school_id, l.timetable_entry_id, l.date, l.topic, l.homework, l.notes, l.recorded_by, l.created_at, l.updated_at
	          FROM lesson_content l
	          JOIN timetable_entries t ON t.id = l.timetable_entry_id
	          WHERE l.school_id = $1`
	args := []interface{}{schoolID}
	n := 2

	if classID != "" {
		query += fmt.Sprintf(` AND t.class_id = $%d`, n)
		args = append(args, classID)
		n++
	}
	if subjectID != "" {
		query += fmt.Sprintf(` AND t.subject_id = $%d`, n)
		args = append(args, subjectID)
		n++
	}
	if from != "" {
		query += fmt.Sprintf(` AND l.date >= $%d`, n)
		args = append(args, from)
		n++
	}
	if to != "" {
		query += fmt.Sprintf(` AND l.date <= $%d`, n)
		args = append(args, to)
		n++
	}
	query += ` ORDER BY l.date DESC`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list lessons: %w", err)
	}
	defer rows.Close()

	var list []models.LessonContent
	for rows.Next() {
		var l models.LessonContent
		if err := rows.Scan(&l.ID, &l.SchoolID, &l.TimetableEntryID, &l.Date, &l.Topic, &l.Homework, &l.Notes, &l.RecordedBy, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan lesson: %w", err)
		}
		list = append(list, l)
	}
	return list, nil
}
