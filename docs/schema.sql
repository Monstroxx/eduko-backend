-- Eduko Database Schema
-- PostgreSQL 15+

-- ============================================================
-- SCHOOL & CONFIG
-- ============================================================

CREATE TABLE schools (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255) NOT NULL,
    address         TEXT,
    school_type     VARCHAR(50) NOT NULL,  -- gymnasium, realschule, grundschule, berufsschule, etc.
    locale          VARCHAR(10) NOT NULL DEFAULT 'de',
    timezone        VARCHAR(50) NOT NULL DEFAULT 'Europe/Berlin',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Flexible key-value config per school (excuse rules, attestation, etc.)
CREATE TABLE school_settings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    key             VARCHAR(100) NOT NULL,
    value           JSONB NOT NULL,
    UNIQUE(school_id, key)
);

-- Default settings:
-- excuse_deadline_days          = 14
-- excuse_granularity            = "day" | "period"
-- attestation_required_days     = 14
-- attestation_required_exam     = true
-- attestation_per_student       = false
-- approval_role                 = "class_teacher"
-- max_exams_per_week            = 3

-- ============================================================
-- USERS & AUTH
-- ============================================================

CREATE TYPE user_role AS ENUM ('student', 'teacher', 'admin');

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    email           VARCHAR(255),
    username        VARCHAR(100) NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    role            user_role NOT NULL,
    first_name      VARCHAR(100) NOT NULL,
    last_name       VARCHAR(100) NOT NULL,
    locale          VARCHAR(10),           -- override school locale
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(school_id, username)
);

-- ============================================================
-- STUDENTS
-- ============================================================

CREATE TABLE students (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    school_id       UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    class_id        UUID,                  -- FK added after classes table
    date_of_birth   DATE NOT NULL,
    is_adult        BOOLEAN GENERATED ALWAYS AS (date_of_birth <= CURRENT_DATE - INTERVAL '18 years') STORED,
    attestation_required BOOLEAN NOT NULL DEFAULT false,  -- per-student override
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- TEACHERS
-- ============================================================

CREATE TABLE teachers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    school_id       UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    abbreviation    VARCHAR(10) NOT NULL,  -- Kürzel, e.g. "MÜL"
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(school_id, abbreviation)
);

-- ============================================================
-- CLASSES & SUBJECTS & ROOMS
-- ============================================================

CREATE TABLE classes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    name            VARCHAR(50) NOT NULL,   -- e.g. "10a", "Q1"
    grade_level     INT,                    -- e.g. 10, 11, 12
    class_teacher_id UUID REFERENCES teachers(id),
    school_year     VARCHAR(20) NOT NULL,   -- e.g. "2025/2026"
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(school_id, name, school_year)
);

-- Add FK for students.class_id
ALTER TABLE students ADD CONSTRAINT fk_students_class FOREIGN KEY (class_id) REFERENCES classes(id);

CREATE TABLE subjects (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,  -- e.g. "Mathematik"
    abbreviation    VARCHAR(10) NOT NULL,   -- e.g. "MA"
    color           VARCHAR(7),             -- hex color for UI, e.g. "#3B82F6"
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(school_id, abbreviation)
);

CREATE TABLE rooms (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    name            VARCHAR(50) NOT NULL,   -- e.g. "A204"
    building        VARCHAR(100),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(school_id, name)
);

-- ============================================================
-- TIME SLOTS (configurable per school)
-- ============================================================

CREATE TABLE time_slots (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    slot_number     INT NOT NULL,           -- 1st period, 2nd period, etc.
    start_time      TIME NOT NULL,
    end_time        TIME NOT NULL,
    label           VARCHAR(50),            -- e.g. "1. Stunde"
    UNIQUE(school_id, slot_number)
);

-- ============================================================
-- TIMETABLE
-- ============================================================

CREATE TYPE week_type AS ENUM ('all', 'A', 'B');

CREATE TABLE timetable_entries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    class_id        UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    subject_id      UUID NOT NULL REFERENCES subjects(id),
    teacher_id      UUID NOT NULL REFERENCES teachers(id),
    room_id         UUID REFERENCES rooms(id),
    time_slot_id    UUID NOT NULL REFERENCES time_slots(id),
    day_of_week     INT NOT NULL CHECK (day_of_week BETWEEN 1 AND 7),  -- 1=Monday
    week_type       week_type NOT NULL DEFAULT 'all',  -- A/B week support
    valid_from      DATE NOT NULL,
    valid_until     DATE,                   -- NULL = indefinite
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_timetable_class ON timetable_entries(class_id, day_of_week);
CREATE INDEX idx_timetable_teacher ON timetable_entries(teacher_id, day_of_week);

-- ============================================================
-- SUBSTITUTIONS & CANCELLATIONS
-- ============================================================

CREATE TYPE substitution_type AS ENUM ('substitution', 'cancellation', 'room_change', 'extra_lesson');

CREATE TABLE substitutions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id           UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    timetable_entry_id  UUID NOT NULL REFERENCES timetable_entries(id),
    date                DATE NOT NULL,
    type                substitution_type NOT NULL,
    substitute_teacher_id UUID REFERENCES teachers(id),
    substitute_room_id  UUID REFERENCES rooms(id),
    substitute_subject_id UUID REFERENCES subjects(id),
    note                TEXT,
    created_by          UUID NOT NULL REFERENCES users(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_substitutions_date ON substitutions(school_id, date);

-- ============================================================
-- ATTENDANCE
-- ============================================================

CREATE TYPE attendance_status AS ENUM ('present', 'absent', 'late', 'excused_leave');

CREATE TABLE attendance (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    student_id      UUID NOT NULL REFERENCES students(id),
    timetable_entry_id UUID NOT NULL REFERENCES timetable_entries(id),
    date            DATE NOT NULL,
    status          attendance_status NOT NULL,
    recorded_by     UUID NOT NULL REFERENCES users(id),  -- teacher who recorded
    note            TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(student_id, timetable_entry_id, date)
);

CREATE INDEX idx_attendance_student ON attendance(student_id, date);
CREATE INDEX idx_attendance_date ON attendance(school_id, date);

-- ============================================================
-- EXCUSES
-- ============================================================

CREATE TYPE excuse_status AS ENUM ('pending', 'approved', 'rejected');
CREATE TYPE excuse_submission AS ENUM ('digital', 'paper');

CREATE TABLE excuses (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    student_id      UUID NOT NULL REFERENCES students(id),
    date_from       DATE NOT NULL,
    date_to         DATE NOT NULL,
    submission_type excuse_submission NOT NULL,
    status          excuse_status NOT NULL DEFAULT 'pending',
    reason          TEXT,
    attestation_provided BOOLEAN NOT NULL DEFAULT false,
    file_path       VARCHAR(500),           -- uploaded PDF/image
    submitted_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    approved_by     UUID REFERENCES users(id),
    approved_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Links excuse to specific attendance records
CREATE TABLE excuse_attendance (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    excuse_id       UUID NOT NULL REFERENCES excuses(id) ON DELETE CASCADE,
    attendance_id   UUID NOT NULL REFERENCES attendance(id),
    UNIQUE(excuse_id, attendance_id)
);

CREATE INDEX idx_excuses_student ON excuses(student_id, status);

-- ============================================================
-- LESSON CONTENT
-- ============================================================

CREATE TABLE lesson_content (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    timetable_entry_id UUID NOT NULL REFERENCES timetable_entries(id),
    date            DATE NOT NULL,
    topic           TEXT NOT NULL,
    homework        TEXT,
    notes           TEXT,
    recorded_by     UUID NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(timetable_entry_id, date)
);

CREATE INDEX idx_lesson_content_date ON lesson_content(school_id, date);

-- ============================================================
-- APPOINTMENTS & EXAMS
-- ============================================================

CREATE TYPE appointment_type AS ENUM ('exam', 'test', 'event', 'other');
CREATE TYPE appointment_scope AS ENUM ('school', 'class', 'subject');

CREATE TABLE appointments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    type            appointment_type NOT NULL,
    scope           appointment_scope NOT NULL,
    class_id        UUID REFERENCES classes(id),       -- if scope = class/subject
    subject_id      UUID REFERENCES subjects(id),      -- if scope = subject
    date            DATE NOT NULL,
    time_slot_id    UUID REFERENCES time_slots(id),    -- optional: specific period
    created_by      UUID NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_appointments_date ON appointments(school_id, date);
CREATE INDEX idx_appointments_class ON appointments(class_id, date);

-- ============================================================
-- AUDIT LOG (for DSGVO compliance)
-- ============================================================

CREATE TABLE audit_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    user_id         UUID REFERENCES users(id),
    action          VARCHAR(50) NOT NULL,   -- e.g. "attendance.create", "excuse.approve"
    entity_type     VARCHAR(50) NOT NULL,
    entity_id       UUID NOT NULL,
    old_value       JSONB,
    new_value       JSONB,
    ip_address      INET,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_log_entity ON audit_log(entity_type, entity_id);
CREATE INDEX idx_audit_log_user ON audit_log(user_id, created_at);
