package models

import (
	"time"

	"github.com/google/uuid"
)

// ── School ──────────────────────────────────────────────────

type School struct {
	ID         uuid.UUID `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	Address    *string   `json:"address,omitempty" db:"address"`
	SchoolType string    `json:"school_type" db:"school_type"`
	Locale     string    `json:"locale" db:"locale"`
	Timezone   string    `json:"timezone" db:"timezone"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type SchoolSetting struct {
	ID       uuid.UUID   `json:"id" db:"id"`
	SchoolID uuid.UUID   `json:"school_id" db:"school_id"`
	Key      string      `json:"key" db:"key"`
	Value    interface{} `json:"value" db:"value"`
}

// ── User ────────────────────────────────────────────────────

type UserRole string

const (
	RoleStudent UserRole = "student"
	RoleTeacher UserRole = "teacher"
	RoleAdmin   UserRole = "admin"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	SchoolID     uuid.UUID `json:"school_id" db:"school_id"`
	Email        *string   `json:"email,omitempty" db:"email"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         UserRole  `json:"role" db:"role"`
	FirstName    string    `json:"first_name" db:"first_name"`
	LastName     string    `json:"last_name" db:"last_name"`
	Locale       *string   `json:"locale,omitempty" db:"locale"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// ── Student ─────────────────────────────────────────────────

type Student struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	UserID               uuid.UUID  `json:"user_id" db:"user_id"`
	SchoolID             uuid.UUID  `json:"school_id" db:"school_id"`
	ClassID              *uuid.UUID `json:"class_id,omitempty" db:"class_id"`
	DateOfBirth          time.Time  `json:"date_of_birth" db:"date_of_birth"`
	IsAdult              bool       `json:"is_adult" db:"is_adult"`
	AttestationRequired  bool       `json:"attestation_required" db:"attestation_required"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
}

// ── Teacher ─────────────────────────────────────────────────

type Teacher struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	SchoolID     uuid.UUID `json:"school_id" db:"school_id"`
	Abbreviation string    `json:"abbreviation" db:"abbreviation"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// ── Class ───────────────────────────────────────────────────

type Class struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	SchoolID       uuid.UUID  `json:"school_id" db:"school_id"`
	Name           string     `json:"name" db:"name"`
	GradeLevel     *int       `json:"grade_level,omitempty" db:"grade_level"`
	ClassTeacherID *uuid.UUID `json:"class_teacher_id,omitempty" db:"class_teacher_id"`
	SchoolYear     string     `json:"school_year" db:"school_year"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// ── Subject ─────────────────────────────────────────────────

type Subject struct {
	ID           uuid.UUID `json:"id" db:"id"`
	SchoolID     uuid.UUID `json:"school_id" db:"school_id"`
	Name         string    `json:"name" db:"name"`
	Abbreviation string    `json:"abbreviation" db:"abbreviation"`
	Color        *string   `json:"color,omitempty" db:"color"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// ── Room ────────────────────────────────────────────────────

type Room struct {
	ID        uuid.UUID `json:"id" db:"id"`
	SchoolID  uuid.UUID `json:"school_id" db:"school_id"`
	Name      string    `json:"name" db:"name"`
	Building  *string   `json:"building,omitempty" db:"building"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ── TimeSlot ────────────────────────────────────────────────

type TimeSlot struct {
	ID         uuid.UUID `json:"id" db:"id"`
	SchoolID   uuid.UUID `json:"school_id" db:"school_id"`
	SlotNumber int       `json:"slot_number" db:"slot_number"`
	StartTime  string    `json:"start_time" db:"start_time"`
	EndTime    string    `json:"end_time" db:"end_time"`
	Label      *string   `json:"label,omitempty" db:"label"`
}

// ── Timetable ───────────────────────────────────────────────

type WeekType string

const (
	WeekAll WeekType = "all"
	WeekA   WeekType = "A"
	WeekB   WeekType = "B"
)

type TimetableEntry struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	SchoolID   uuid.UUID  `json:"school_id" db:"school_id"`
	ClassID    uuid.UUID  `json:"class_id" db:"class_id"`
	SubjectID  uuid.UUID  `json:"subject_id" db:"subject_id"`
	TeacherID  uuid.UUID  `json:"teacher_id" db:"teacher_id"`
	RoomID     *uuid.UUID `json:"room_id,omitempty" db:"room_id"`
	TimeSlotID uuid.UUID  `json:"time_slot_id" db:"time_slot_id"`
	DayOfWeek  int        `json:"day_of_week" db:"day_of_week"`
	WeekType   WeekType   `json:"week_type" db:"week_type"`
	ValidFrom  time.Time  `json:"valid_from" db:"valid_from"`
	ValidUntil *time.Time `json:"valid_until,omitempty" db:"valid_until"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

// ── Substitution ────────────────────────────────────────────

type SubstitutionType string

const (
	SubTypeSubstitution SubstitutionType = "substitution"
	SubTypeCancellation SubstitutionType = "cancellation"
	SubTypeRoomChange   SubstitutionType = "room_change"
	SubTypeExtraLesson  SubstitutionType = "extra_lesson"
)

type Substitution struct {
	ID                  uuid.UUID        `json:"id" db:"id"`
	SchoolID            uuid.UUID        `json:"school_id" db:"school_id"`
	TimetableEntryID    uuid.UUID        `json:"timetable_entry_id" db:"timetable_entry_id"`
	Date                time.Time        `json:"date" db:"date"`
	Type                SubstitutionType `json:"type" db:"type"`
	SubstituteTeacherID *uuid.UUID       `json:"substitute_teacher_id,omitempty" db:"substitute_teacher_id"`
	SubstituteRoomID    *uuid.UUID       `json:"substitute_room_id,omitempty" db:"substitute_room_id"`
	SubstituteSubjectID *uuid.UUID       `json:"substitute_subject_id,omitempty" db:"substitute_subject_id"`
	Note                *string          `json:"note,omitempty" db:"note"`
	CreatedBy           uuid.UUID        `json:"created_by" db:"created_by"`
	CreatedAt           time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at" db:"updated_at"`
}

// ── Attendance ──────────────────────────────────────────────

type AttendanceStatus string

const (
	StatusPresent      AttendanceStatus = "present"
	StatusAbsent       AttendanceStatus = "absent"
	StatusLate         AttendanceStatus = "late"
	StatusExcusedLeave AttendanceStatus = "excused_leave"
)

type Attendance struct {
	ID               uuid.UUID        `json:"id" db:"id"`
	SchoolID         uuid.UUID        `json:"school_id" db:"school_id"`
	StudentID        uuid.UUID        `json:"student_id" db:"student_id"`
	TimetableEntryID uuid.UUID        `json:"timetable_entry_id" db:"timetable_entry_id"`
	Date             time.Time        `json:"date" db:"date"`
	Status           AttendanceStatus `json:"status" db:"status"`
	RecordedBy       uuid.UUID        `json:"recorded_by" db:"recorded_by"`
	Note             *string          `json:"note,omitempty" db:"note"`
	CreatedAt        time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at" db:"updated_at"`
}

// ── Excuse ──────────────────────────────────────────────────

type ExcuseStatus string

const (
	ExcusePending  ExcuseStatus = "pending"
	ExcuseApproved ExcuseStatus = "approved"
	ExcuseRejected ExcuseStatus = "rejected"
)

type ExcuseSubmission string

const (
	SubmissionDigital ExcuseSubmission = "digital"
	SubmissionPaper   ExcuseSubmission = "paper"
)

type Excuse struct {
	ID                   uuid.UUID        `json:"id" db:"id"`
	SchoolID             uuid.UUID        `json:"school_id" db:"school_id"`
	StudentID            uuid.UUID        `json:"student_id" db:"student_id"`
	DateFrom             time.Time        `json:"date_from" db:"date_from"`
	DateTo               time.Time        `json:"date_to" db:"date_to"`
	SubmissionType       ExcuseSubmission `json:"submission_type" db:"submission_type"`
	Status               ExcuseStatus     `json:"status" db:"status"`
	Reason               *string          `json:"reason,omitempty" db:"reason"`
	AttestationProvided  bool             `json:"attestation_provided" db:"attestation_provided"`
	FilePath             *string          `json:"file_path,omitempty" db:"file_path"`
	SubmittedAt          time.Time        `json:"submitted_at" db:"submitted_at"`
	ApprovedBy           *uuid.UUID       `json:"approved_by,omitempty" db:"approved_by"`
	ApprovedAt           *time.Time       `json:"approved_at,omitempty" db:"approved_at"`
	CreatedAt            time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at" db:"updated_at"`
}

// ── Lesson Content ──────────────────────────────────────────

type LessonContent struct {
	ID               uuid.UUID `json:"id" db:"id"`
	SchoolID         uuid.UUID `json:"school_id" db:"school_id"`
	TimetableEntryID uuid.UUID `json:"timetable_entry_id" db:"timetable_entry_id"`
	Date             time.Time `json:"date" db:"date"`
	Topic            string    `json:"topic" db:"topic"`
	Homework         *string   `json:"homework,omitempty" db:"homework"`
	Notes            *string   `json:"notes,omitempty" db:"notes"`
	RecordedBy       uuid.UUID `json:"recorded_by" db:"recorded_by"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// ── Appointment ─────────────────────────────────────────────

type AppointmentType string

const (
	AppointExam  AppointmentType = "exam"
	AppointTest  AppointmentType = "test"
	AppointEvent AppointmentType = "event"
	AppointOther AppointmentType = "other"
)

type AppointmentScope string

const (
	ScopeSchool  AppointmentScope = "school"
	ScopeClass   AppointmentScope = "class"
	ScopeSubject AppointmentScope = "subject"
)

type Appointment struct {
	ID          uuid.UUID        `json:"id" db:"id"`
	SchoolID    uuid.UUID        `json:"school_id" db:"school_id"`
	Title       string           `json:"title" db:"title"`
	Description *string          `json:"description,omitempty" db:"description"`
	Type        AppointmentType  `json:"type" db:"type"`
	Scope       AppointmentScope `json:"scope" db:"scope"`
	ClassID     *uuid.UUID       `json:"class_id,omitempty" db:"class_id"`
	SubjectID   *uuid.UUID       `json:"subject_id,omitempty" db:"subject_id"`
	Date        time.Time        `json:"date" db:"date"`
	TimeSlotID  *uuid.UUID       `json:"time_slot_id,omitempty" db:"time_slot_id"`
	CreatedBy   uuid.UUID        `json:"created_by" db:"created_by"`
	CreatedAt   time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at" db:"updated_at"`
}
