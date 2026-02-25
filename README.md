# Eduko Backend

Modern, open-source school management system. A self-hostable alternative to Untis.

## Features

- **Timetable** — Display with A/B weeks, blocks, epochs support
- **Attendance** — Record per student per lesson (present, absent, late, excused_leave)
- **Excuses** — Auto-links to absences, approval workflow, PDF generation, CSV bulk import
- **Substitutions** — Cancellations, room changes, teacher substitutions, extra lessons
- **Lesson Content** — Topic logging with homework and notes
- **Appointments** — Exams, tests, events with scope (school/class/subject)
- **Student Import** — CSV bulk import with class resolution
- **Role-Based Access** — Student, Teacher, Admin with per-endpoint authorization
- **Multi-Tenant** — school_id scoping on all tables
- **i18n** — German and English locales
- **Audit Log** — DSGVO-compliant change tracking

## Tech Stack

- **Language:** Go 1.22+
- **Framework:** [Echo](https://echo.labstack.com/) v4
- **Database:** PostgreSQL 16 with [pgx](https://github.com/jackc/pgx) v5
- **Auth:** JWT (bcrypt password hashing)
- **Deployment:** Docker Compose or single binary

## Quick Start

### Docker Compose

```bash
docker compose up -d
```

This starts PostgreSQL + the backend on port 8080.

### Manual

```bash
# Prerequisites: Go 1.22+, PostgreSQL 16+

# Create database
createdb eduko
psql -d eduko -f docs/schema.sql
psql -d eduko -f docs/seed.sql  # optional test data

# Build & run
export JWT_SECRET="your-secret-here"
export DATABASE_URL="postgres://user:pass@localhost:5432/eduko?sslmode=disable"
go build -o eduko ./cmd/eduko
./eduko
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://eduko:eduko@localhost:5432/eduko?sslmode=disable` | PostgreSQL connection string |
| `JWT_SECRET` | *(required)* | Secret for JWT signing |
| `PORT` | `8080` | Server port |
| `CORS_ORIGINS` | `*` | Comma-separated allowed origins |
| `UPLOAD_DIR` | `./uploads` | Directory for file uploads |

## API

Full endpoint documentation: [docs/API.md](docs/API.md)

### Key Endpoints

```
POST   /api/v1/auth/login          # Login → JWT token
POST   /api/v1/auth/register       # Register user

GET    /api/v1/timetable           # Timetable entries
GET    /api/v1/substitutions       # Substitution plan
POST   /api/v1/attendance          # Record attendance (batch)
GET    /api/v1/excuses             # List excuses (filterable)
PATCH  /api/v1/excuses/:id/approve # Approve excuse → updates attendance
GET    /api/v1/excuses/:id/pdf     # Generate excuse PDF
POST   /api/v1/students/import     # CSV student bulk import
POST   /api/v1/excuses/import      # CSV excuse bulk import

GET    /health                     # Health check
```

### Test Accounts (seed data)

| Username | Password | Role |
|----------|----------|------|
| `admin` | `admin123` | admin |
| `lehrer` | `teacher123` | teacher |
| `schueler` | `student123` | student |

School ID: `00000000-0000-0000-0000-000000000001`

## Project Structure

```
cmd/eduko/              # Application entrypoint
internal/
  config/               # Environment-based configuration
  database/             # PostgreSQL connection pool
  handlers/             # HTTP handlers (Echo)
  middleware/            # JWT auth middleware
  models/               # Domain models
  services/             # Business logic layer
docs/
  schema.sql            # Database schema (17 tables)
  seed.sql              # Test data
  API.md                # Full API documentation
docker-compose.yml      # PostgreSQL + backend
```

## Testing

```bash
# Requires running PostgreSQL with schema + seed data
go test ./tests/ -v
```

19 integration tests covering auth, CRUD, role-based access, CSV import, and PDF generation.

## License

MIT — see [LICENSE](LICENSE).
