# Eduko API — v1

Base URL: `/api/v1`

## Authentication

All endpoints except `/auth/*` require `Authorization: Bearer <token>`.

### POST /auth/login
Login and receive JWT token.
```json
// Request
{ "username": "string", "password": "string", "school_id": "uuid" }
// Response 200
{ "token": "jwt-string", "user": { ... } }
```

### POST /auth/register
Register a new user (admin only for creating teachers/admin, self for students if enabled).
```json
// Request
{ "username": "string", "password": "string", "email": "string?",
  "first_name": "string", "last_name": "string", "role": "student|teacher|admin",
  "school_id": "uuid" }
// Response 201
{ "user": { ... } }
```

---

## School

### GET /school
Returns the school for the authenticated user.

### PUT /school
Update school details (admin only).

### GET /school/settings
Returns all school settings as key-value pairs.

### PUT /school/settings
Update school settings (admin only).
```json
// Request
{ "key": "excuse_deadline_days", "value": 14 }
```

---

## Classes

### GET /classes
List all classes. Query: `?school_year=2025/2026`

### POST /classes
Create a class (admin only).

### GET /classes/:id
Get class details.

### PUT /classes/:id
Update class (admin only).

### DELETE /classes/:id
Delete class (admin only).

### GET /classes/:id/students
List students in class.

---

## Students

### GET /students
List students. Query: `?class_id=uuid`

### GET /students/:id
Get student details (includes `is_adult` computed field).

### PUT /students/:id
Update student.

### GET /students/:id/absences
Get absence history. Query: `?from=date&to=date&status=unexcused`

### GET /students/:id/excuses
Get excuse history.

---

## Teachers

### GET /teachers
List teachers.

### GET /teachers/:id
Get teacher details.

---

## Timetable

### GET /timetable
Get timetable entries. Query: `?class_id=uuid&teacher_id=uuid&date=date&week_type=A`

### POST /timetable
Create entry (admin only).

### PUT /timetable/:id
Update entry (admin only).

### DELETE /timetable/:id
Delete entry (admin only).

---

## Substitutions

### GET /substitutions
List substitutions. Query: `?date=date&from=date&to=date`

### POST /substitutions
Create substitution (admin only).

### PUT /substitutions/:id
Update substitution (admin only).

### DELETE /substitutions/:id
Delete substitution (admin only).

---

## Attendance

### POST /attendance
Record attendance (teacher only). Supports batch.
```json
// Request — single
{ "student_id": "uuid", "timetable_entry_id": "uuid", "date": "2026-02-24",
  "status": "present|absent|late|excused_leave", "note": "string?" }
// Request — batch
{ "timetable_entry_id": "uuid", "date": "2026-02-24",
  "entries": [{ "student_id": "uuid", "status": "present" }, ...] }
```

### PUT /attendance/:id
Update single attendance record.

### GET /attendance/class/:classId
Get attendance for class. Query: `?date=date`

### GET /attendance/date/:date
Get all attendance for a date.

---

## Excuses ⭐

### POST /excuses
Student creates excuse request.
```json
// Request
{ "date_from": "2026-02-20", "date_to": "2026-02-21",
  "submission_type": "digital|paper", "reason": "string?",
  "attestation_provided": false }
// Response 201 — auto-links to matching attendance records
{ "excuse": { ... }, "linked_absences": 4 }
```

### GET /excuses
List excuses. Query: `?status=pending&student_id=uuid&class_id=uuid`

### GET /excuses/:id
Get excuse details with linked attendance records.

### PATCH /excuses/:id/approve
Approve excuse (class teacher / configured role).
```json
// Request
{ "note": "string?" }
// Side effect: linked attendance records updated
```

### PATCH /excuses/:id/reject
Reject excuse.
```json
{ "reason": "string" }
```

### POST /excuses/upload
Upload signed excuse form (PDF/image).
Multipart form: `file` + `excuse_id`.

### GET /excuses/:id/pdf
Generate downloadable excuse form as PDF.

### POST /excuses/import
Batch import excuses from CSV (admin only).
Multipart form: `file` (CSV).

---

## Lesson Content

### POST /lessons
Record lesson content (teacher only).
```json
{ "timetable_entry_id": "uuid", "date": "2026-02-24",
  "topic": "string", "homework": "string?", "notes": "string?" }
```

### PUT /lessons/:id
Update lesson content.

### GET /lessons
List lesson content. Query: `?class_id=uuid&subject_id=uuid&from=date&to=date`

---

## Appointments

### GET /appointments
List appointments. Query: `?type=exam&class_id=uuid&from=date&to=date`

### POST /appointments
Create appointment (teacher/admin).

### PUT /appointments/:id
Update appointment.

### DELETE /appointments/:id
Delete appointment.

---

## Subjects & Rooms

### GET /subjects
### POST /subjects (admin)
### GET /rooms
### POST /rooms (admin)

---

## Time Slots

### GET /timeslots
### POST /timeslots (admin)

---

## Common Patterns

**Pagination:** `?page=1&per_page=25`
**Sorting:** `?sort=date&order=desc`
**Error Response:**
```json
{ "error": "string", "code": "string", "details": {} }
```

## Data Privacy

All endpoints are scoped to the authenticated user's school (`school_id` from JWT).
Audit log records all write operations for DSGVO compliance.
