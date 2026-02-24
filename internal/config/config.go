package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	CORSOrigins []string
	UploadDir   string
}

func Load() (*Config, error) {
	dbURL := getEnv("DATABASE_URL", "postgres://eduko:eduko@localhost:5432/eduko?sslmode=disable")
	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required")
	}

	corsOrigins := strings.Split(getEnv("CORS_ORIGINS", "*"), ",")

	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: dbURL,
		JWTSecret:   jwtSecret,
		CORSOrigins: corsOrigins,
		UploadDir:   getEnv("UPLOAD_DIR", "./uploads"),
	}, nil
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
