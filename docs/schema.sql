-- Eduko Database Schema
-- PostgreSQL 16+

-- ============================================================
-- ENUMS (all up front)
-- ============================================================

CREATE TYPE user_role AS ENUM ('student', 'teacher', 'admin');
CREATE TYPE week_type AS ENUM ('all', 'A', 'B');
CREATE TYPE substitution_type AS ENUM ('substitution', 'cancellation', 'room_change', 'extra_lesson');
CREATE TYPE attendance_status AS ENUM ('present', 'absent', 'late', 'excused_leave');
CREATE TYPE excuse_status AS ENUM ('pending', 'approved', 'rejected');
CREATE TYPE excuse_submission AS ENUM ('digital', 'paper');
CREATE TYPE appointment_type AS ENUM ('exam', 'test', 'event', 'other');
CREATE TYPE appointment_scope AS ENUM ('school', 'class', 'subject');

-- ============================================================
-- SCHOOL & CONFIG
-- ============================================================

CREATE TABLE schools (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255) NOT NULL,
    address         TEXT,
    school_type     VARCHAR(50) NOT NULL,
    locale          VARCHAR(10) NOT NULL DEFAULT 'de',
    timezone        VARCHAR(50) NOT NULL DEFAULT 'Europe/Berlin',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE school_settings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    key             VARCHAR(100) NOT NULL,
    value           JSONB NOT NULL,
    UNIQUE(school_id, key)
);

-- ============================================================
-- USERS
-- ============================================================

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    email           VARCHAR(255),
    username        VARCHAR(100) NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    role            user_role NOT NULL,
    first_name      VARCHAR(100) NOT NULL,
    last_name       VARCHAR(100) NOT NULL,
    locale          VARCHAR(10),
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(school_id, username)
);

-- ============================================================
-- SUBJECTS & ROOMS (no deps on classes/students)
-- ============================================================

CREATE TABLE subjects (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,
    abbreviation    VARCHAR(10) NOT NULL,
    color           VARCHAR(7),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(school_id, abbreviation)
);

CREATE TABLE rooms (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    name            VARCHAR(50) NOT NULL,
    building        VARCHAR(100),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(school_id, name)
);

CREATE TABLE time_slots (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    slot_number     INT NOT NULL,
    start_time      TIME NOT NULL,
    end_time        TIME NOT NULL,
    label           VARCHAR(50),
    UNIQUE(school_id, slot_number)
);

-- ============================================================
-- TEACHERS
-- ============================================================

CREATE TABLE teachers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    school_id       UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    abbreviation    VARCHAR(10) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(school_id, abbreviation)
);

-- ============================================================
-- CLASSES
-- ============================================================

CREATE TABLE classes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    name            VARCHAR(50) NOT NULL,
    grade_level     INT,
    class_teacher_id UUID REFERENCES teachers(id),
    school_year     VARCHAR(20) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(school_id, name, school_year)
);

-- ============================================================
-- STUDENTS (after classes)
-- ============================================================

CREATE TABLE students (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    school_id       UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    class_id        UUID REFERENCES classes(id),
    date_of_birth   DATE NOT NULL,
    -- is_adult computed in application layer (CURRENT_DATE not immutable for GENERATED columns)
    attestation_required BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- TIMETABLE
-- ============================================================

CREATE TABLE timetable_entries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    class_id        UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    subject_id      UUID NOT NULL REFERENCES subjects(id),
    teacher_id      UUID NOT NULL REFERENCES teachers(id),
    room_id         UUID REFERENCES rooms(id),
    time_slot_id    UUID NOT NULL REFERENCES time_slots(id),
    day_of_week     INT NOT NULL CHECK (day_of_week BETWEEN 1 AND 7),
    week_type       week_type NOT NULL DEFAULT 'all',
    valid_from      DATE NOT NULL,
    valid_until     DATE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_timetable_class ON timetable_entries(class_id, day_of_week);
CREATE INDEX idx_timetable_teacher ON timetable_entries(teacher_id, day_of_week);

-- ============================================================
-- SUBSTITUTIONS
-- ============================================================

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

CREATE TABLE attendance (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    student_id      UUID NOT NULL REFERENCES students(id),
    timetable_entry_id UUID NOT NULL REFERENCES timetable_entries(id),
    date            DATE NOT NULL,
    status          attendance_status NOT NULL,
    recorded_by     UUID NOT NULL REFERENCES users(id),
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
    file_path       VARCHAR(500),
    submitted_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    approved_by     UUID REFERENCES users(id),
    approved_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

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

CREATE TABLE appointments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    type            appointment_type NOT NULL,
    scope           appointment_scope NOT NULL,
    class_id        UUID REFERENCES classes(id),
    subject_id      UUID REFERENCES subjects(id),
    date            DATE NOT NULL,
    time_slot_id    UUID REFERENCES time_slots(id),
    created_by      UUID NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_appointments_date ON appointments(school_id, date);
CREATE INDEX idx_appointments_class ON appointments(class_id, date);

-- ============================================================
-- AUDIT LOG
-- ============================================================

CREATE TABLE audit_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    user_id         UUID REFERENCES users(id),
    action          VARCHAR(50) NOT NULL,
    entity_type     VARCHAR(50) NOT NULL,
    entity_id       UUID NOT NULL,
    old_value       JSONB,
    new_value       JSONB,
    ip_address      INET,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_log_entity ON audit_log(entity_type, entity_id);
CREATE INDEX idx_audit_log_user ON audit_log(user_id, created_at);

-- ============================================================
-- SEED: Default school for development
-- ============================================================

INSERT INTO schools (id, name, address, school_type, locale, timezone)
VALUES ('00000000-0000-0000-0000-000000000001', 'Eduko Testschule', 'Musterstraße 1, 10115 Berlin', 'gymnasium', 'de', 'Europe/Berlin');

INSERT INTO school_settings (school_id, key, value) VALUES
('00000000-0000-0000-0000-000000000001', 'excuse_deadline_days', '14'),
('00000000-0000-0000-0000-000000000001', 'excuse_granularity', '"day"'),
('00000000-0000-0000-0000-000000000001', 'attestation_required_days', '14'),
('00000000-0000-0000-0000-000000000001', 'attestation_required_exam', 'true'),
('00000000-0000-0000-0000-000000000001', 'approval_role', '"class_teacher"'),
('00000000-0000-0000-0000-000000000001', 'max_exams_per_week', '3');

-- Admin user (password: admin123)
INSERT INTO users (id, school_id, username, password_hash, role, first_name, last_name, email)
VALUES ('00000000-0000-0000-0000-000000000010', '00000000-0000-0000-0000-000000000001',
        'admin', '$2a$10$ePNUgU7ucQrgffBdjAiyn.mkW6ErfT2TYhzjlGCtu.FSTHEEu.KgG',
        'admin', 'Admin', 'User', 'admin@eduko.dev');

-- Teacher user (password: teacher123)
INSERT INTO users (id, school_id, username, password_hash, role, first_name, last_name, email)
VALUES ('00000000-0000-0000-0000-000000000020', '00000000-0000-0000-0000-000000000001',
        'lehrer', '$2a$10$SlzkUhdG97aEdQ2j/XP2Veuzep7q.IrEjmyAO6U8Juwq7tkhjTe8y',
        'teacher', 'Max', 'Mustermann', 'lehrer@eduko.dev');

INSERT INTO teachers (id, user_id, school_id, abbreviation)
VALUES ('00000000-0000-0000-0000-000000000021', '00000000-0000-0000-0000-000000000020',
        '00000000-0000-0000-0000-000000000001', 'MUS');

-- Student user (password: student123)
INSERT INTO users (id, school_id, username, password_hash, role, first_name, last_name, email)
VALUES ('00000000-0000-0000-0000-000000000030', '00000000-0000-0000-0000-000000000001',
        'schueler', '$2a$10$M/M4twLsLo87vWs3z5GCK.pCZGKJS2NRPN4iSYGFqqYopBNemf5pa',
        'student', 'Lisa', 'Schmidt', 'schueler@eduko.dev');

-- Class
INSERT INTO classes (id, school_id, name, grade_level, class_teacher_id, school_year)
VALUES ('00000000-0000-0000-0000-000000000100', '00000000-0000-0000-0000-000000000001',
        '10a', 10, '00000000-0000-0000-0000-000000000021', '2025/2026');

INSERT INTO students (id, user_id, school_id, class_id, date_of_birth)
VALUES ('00000000-0000-0000-0000-000000000031', '00000000-0000-0000-0000-000000000030',
        '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000100', '2008-05-15');

-- Subjects
INSERT INTO subjects (id, school_id, name, abbreviation, color) VALUES
('00000000-0000-0000-0000-000000000200', '00000000-0000-0000-0000-000000000001', 'Mathematik', 'MA', '#3B82F6'),
('00000000-0000-0000-0000-000000000201', '00000000-0000-0000-0000-000000000001', 'Deutsch', 'DE', '#EF4444'),
('00000000-0000-0000-0000-000000000202', '00000000-0000-0000-0000-000000000001', 'Englisch', 'EN', '#22C55E');

-- Rooms
INSERT INTO rooms (id, school_id, name, building) VALUES
('00000000-0000-0000-0000-000000000300', '00000000-0000-0000-0000-000000000001', 'A101', 'Hauptgebäude'),
('00000000-0000-0000-0000-000000000301', '00000000-0000-0000-0000-000000000001', 'B204', 'Neubau');

-- Time slots
INSERT INTO time_slots (id, school_id, slot_number, start_time, end_time, label) VALUES
('00000000-0000-0000-0000-000000000400', '00000000-0000-0000-0000-000000000001', 1, '08:00', '08:45', '1. Stunde'),
('00000000-0000-0000-0000-000000000401', '00000000-0000-0000-0000-000000000001', 2, '08:50', '09:35', '2. Stunde'),
('00000000-0000-0000-0000-000000000402', '00000000-0000-0000-0000-000000000001', 3, '09:50', '10:35', '3. Stunde'),
('00000000-0000-0000-0000-000000000403', '00000000-0000-0000-0000-000000000001', 4, '10:40', '11:25', '4. Stunde'),
('00000000-0000-0000-0000-000000000404', '00000000-0000-0000-0000-000000000001', 5, '11:40', '12:25', '5. Stunde'),
('00000000-0000-0000-0000-000000000405', '00000000-0000-0000-0000-000000000001', 6, '12:30', '13:15', '6. Stunde');

-- Timetable entries (Monday)
INSERT INTO timetable_entries (school_id, class_id, subject_id, teacher_id, room_id, time_slot_id, day_of_week, valid_from) VALUES
('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000100', '00000000-0000-0000-0000-000000000200', '00000000-0000-0000-0000-000000000021', '00000000-0000-0000-0000-000000000300', '00000000-0000-0000-0000-000000000400', 1, '2025-09-01'),
('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000100', '00000000-0000-0000-0000-000000000200', '00000000-0000-0000-0000-000000000021', '00000000-0000-0000-0000-000000000300', '00000000-0000-0000-0000-000000000401', 1, '2025-09-01'),
('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000100', '00000000-0000-0000-0000-000000000201', '00000000-0000-0000-0000-000000000021', '00000000-0000-0000-0000-000000000301', '00000000-0000-0000-0000-000000000402', 1, '2025-09-01'),
('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000100', '00000000-0000-0000-0000-000000000202', '00000000-0000-0000-0000-000000000021', '00000000-0000-0000-0000-000000000301', '00000000-0000-0000-0000-000000000403', 1, '2025-09-01');
