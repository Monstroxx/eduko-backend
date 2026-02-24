package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Monstroxx/eduko-backend/internal/models"
)

type SchoolService struct {
	db *pgxpool.Pool
}

func NewSchoolService(db *pgxpool.Pool) *SchoolService {
	return &SchoolService{db: db}
}

func (s *SchoolService) GetByID(ctx context.Context, id uuid.UUID) (*models.School, error) {
	var school models.School
	err := s.db.QueryRow(ctx,
		`SELECT id, name, address, school_type, locale, timezone, created_at, updated_at
		 FROM schools WHERE id = $1`, id,
	).Scan(&school.ID, &school.Name, &school.Address, &school.SchoolType,
		&school.Locale, &school.Timezone, &school.CreatedAt, &school.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get school: %w", err)
	}
	return &school, nil
}

func (s *SchoolService) Update(ctx context.Context, id uuid.UUID, name, address, schoolType, locale, timezone string) (*models.School, error) {
	var school models.School
	err := s.db.QueryRow(ctx,
		`UPDATE schools SET name = $2, address = $3, school_type = $4, locale = $5, timezone = $6, updated_at = now()
		 WHERE id = $1
		 RETURNING id, name, address, school_type, locale, timezone, created_at, updated_at`,
		id, name, address, schoolType, locale, timezone,
	).Scan(&school.ID, &school.Name, &school.Address, &school.SchoolType,
		&school.Locale, &school.Timezone, &school.CreatedAt, &school.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update school: %w", err)
	}
	return &school, nil
}

func (s *SchoolService) GetSettings(ctx context.Context, schoolID uuid.UUID) (map[string]interface{}, error) {
	rows, err := s.db.Query(ctx,
		`SELECT key, value FROM school_settings WHERE school_id = $1`, schoolID)
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	defer rows.Close()

	settings := make(map[string]interface{})
	for rows.Next() {
		var key string
		var value []byte
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("scan setting: %w", err)
		}
		var v interface{}
		if err := json.Unmarshal(value, &v); err != nil {
			settings[key] = string(value)
		} else {
			settings[key] = v
		}
	}
	return settings, nil
}

func (s *SchoolService) UpdateSetting(ctx context.Context, schoolID uuid.UUID, key string, value interface{}) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal value: %w", err)
	}

	_, err = s.db.Exec(ctx,
		`INSERT INTO school_settings (school_id, key, value) VALUES ($1, $2, $3)
		 ON CONFLICT (school_id, key) DO UPDATE SET value = $3`,
		schoolID, key, jsonValue)
	if err != nil {
		return fmt.Errorf("upsert setting: %w", err)
	}
	return nil
}
