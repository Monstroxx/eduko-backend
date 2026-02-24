# Eduko — Backend

Modern, open-source school management system. Built with Go + PostgreSQL.

## Features

- Timetable display (manual import, flexible time models)
- Attendance tracking (present, absent, late, excused)
- Excuse management with configurable workflows
- Lesson content & notes
- Schedules, substitutions & cancellations
- Role-based access (students, teachers, class teachers, administration)
- i18n support (German first, extensible)
- Offline-ready API design
- Self-hosted via Docker Compose

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | Go |
| Database | PostgreSQL |
| API | REST (JSON) |
| Auth | JWT |
| Deployment | Docker Compose |
| Config | YAML / ENV |

## Getting Started

```bash
docker compose up
```

## License

MIT
