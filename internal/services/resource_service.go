package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Monstroxx/eduko-backend/internal/models"
)

type ResourceService struct {
	db *pgxpool.Pool
}

func NewResourceService(db *pgxpool.Pool) *ResourceService {
	return &ResourceService{db: db}
}

// Teachers

type TeacherWithUser struct {
	models.Teacher
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Email     *string `json:"email,omitempty"`
}

func (s *ResourceService) ListTeachers(ctx context.Context, schoolID uuid.UUID) ([]TeacherWithUser, error) {
	rows, err := s.db.Query(ctx,
		`SELECT t.id, t.user_id, t.school_id, t.abbreviation, t.created_at, t.updated_at,
		        u.first_name, u.last_name, u.email
		 FROM teachers t JOIN users u ON u.id = t.user_id WHERE t.school_id = $1
		 ORDER BY u.last_name`, schoolID)
	if err != nil {
		return nil, fmt.Errorf("list teachers: %w", err)
	}
	defer rows.Close()

	var list []TeacherWithUser
	for rows.Next() {
		var tw TeacherWithUser
		if err := rows.Scan(&tw.ID, &tw.UserID, &tw.SchoolID, &tw.Abbreviation, &tw.CreatedAt, &tw.UpdatedAt,
			&tw.FirstName, &tw.LastName, &tw.Email); err != nil {
			return nil, fmt.Errorf("scan teacher: %w", err)
		}
		list = append(list, tw)
	}
	return list, nil
}

func (s *ResourceService) GetTeacher(ctx context.Context, schoolID, teacherID uuid.UUID) (*TeacherWithUser, error) {
	var tw TeacherWithUser
	err := s.db.QueryRow(ctx,
		`SELECT t.id, t.user_id, t.school_id, t.abbreviation, t.created_at, t.updated_at,
		        u.first_name, u.last_name, u.email
		 FROM teachers t JOIN users u ON u.id = t.user_id WHERE t.id = $1 AND t.school_id = $2`,
		teacherID, schoolID,
	).Scan(&tw.ID, &tw.UserID, &tw.SchoolID, &tw.Abbreviation, &tw.CreatedAt, &tw.UpdatedAt,
		&tw.FirstName, &tw.LastName, &tw.Email)
	if err != nil {
		return nil, fmt.Errorf("get teacher: %w", err)
	}
	return &tw, nil
}

// Subjects

func (s *ResourceService) ListSubjects(ctx context.Context, schoolID uuid.UUID) ([]models.Subject, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, school_id, name, abbreviation, color, created_at FROM subjects WHERE school_id = $1 ORDER BY name`, schoolID)
	if err != nil {
		return nil, fmt.Errorf("list subjects: %w", err)
	}
	defer rows.Close()

	list := make([]models.Subject, 0)
	for rows.Next() {
		var sub models.Subject
		if err := rows.Scan(&sub.ID, &sub.SchoolID, &sub.Name, &sub.Abbreviation, &sub.Color, &sub.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan subject: %w", err)
		}
		list = append(list, sub)
	}
	return list, nil
}

type CreateSubjectInput struct {
	Name         string  `json:"name"`
	Abbreviation string  `json:"abbreviation"`
	Color        *string `json:"color,omitempty"`
}

func (s *ResourceService) CreateSubject(ctx context.Context, schoolID uuid.UUID, input CreateSubjectInput) (*models.Subject, error) {
	var sub models.Subject
	err := s.db.QueryRow(ctx,
		`INSERT INTO subjects (school_id, name, abbreviation, color) VALUES ($1, $2, $3, $4)
		 RETURNING id, school_id, name, abbreviation, color, created_at`,
		schoolID, input.Name, input.Abbreviation, input.Color,
	).Scan(&sub.ID, &sub.SchoolID, &sub.Name, &sub.Abbreviation, &sub.Color, &sub.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create subject: %w", err)
	}
	return &sub, nil
}

// Rooms

func (s *ResourceService) ListRooms(ctx context.Context, schoolID uuid.UUID) ([]models.Room, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, school_id, name, building, created_at FROM rooms WHERE school_id = $1 ORDER BY name`, schoolID)
	if err != nil {
		return nil, fmt.Errorf("list rooms: %w", err)
	}
	defer rows.Close()

	list := make([]models.Room, 0)
	for rows.Next() {
		var r models.Room
		if err := rows.Scan(&r.ID, &r.SchoolID, &r.Name, &r.Building, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan room: %w", err)
		}
		list = append(list, r)
	}
	return list, nil
}

type CreateRoomInput struct {
	Name     string  `json:"name"`
	Building *string `json:"building,omitempty"`
}

func (s *ResourceService) CreateRoom(ctx context.Context, schoolID uuid.UUID, input CreateRoomInput) (*models.Room, error) {
	var r models.Room
	err := s.db.QueryRow(ctx,
		`INSERT INTO rooms (school_id, name, building) VALUES ($1, $2, $3)
		 RETURNING id, school_id, name, building, created_at`,
		schoolID, input.Name, input.Building,
	).Scan(&r.ID, &r.SchoolID, &r.Name, &r.Building, &r.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create room: %w", err)
	}
	return &r, nil
}

// Time Slots

func (s *ResourceService) ListTimeSlots(ctx context.Context, schoolID uuid.UUID) ([]models.TimeSlot, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, school_id, slot_number, start_time, end_time, label FROM time_slots WHERE school_id = $1 ORDER BY slot_number`, schoolID)
	if err != nil {
		return nil, fmt.Errorf("list time slots: %w", err)
	}
	defer rows.Close()

	list := make([]models.TimeSlot, 0)
	for rows.Next() {
		var ts models.TimeSlot
		if err := rows.Scan(&ts.ID, &ts.SchoolID, &ts.SlotNumber, &ts.StartTime, &ts.EndTime, &ts.Label); err != nil {
			return nil, fmt.Errorf("scan time slot: %w", err)
		}
		list = append(list, ts)
	}
	return list, nil
}

type CreateTimeSlotInput struct {
	SlotNumber int     `json:"slot_number"`
	StartTime  string  `json:"start_time"`
	EndTime    string  `json:"end_time"`
	Label      *string `json:"label,omitempty"`
}

func (s *ResourceService) CreateTimeSlot(ctx context.Context, schoolID uuid.UUID, input CreateTimeSlotInput) (*models.TimeSlot, error) {
	var ts models.TimeSlot
	err := s.db.QueryRow(ctx,
		`INSERT INTO time_slots (school_id, slot_number, start_time, end_time, label) VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, school_id, slot_number, start_time, end_time, label`,
		schoolID, input.SlotNumber, input.StartTime, input.EndTime, input.Label,
	).Scan(&ts.ID, &ts.SchoolID, &ts.SlotNumber, &ts.StartTime, &ts.EndTime, &ts.Label)
	if err != nil {
		return nil, fmt.Errorf("create time slot: %w", err)
	}
	return &ts, nil
}
