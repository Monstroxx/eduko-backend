package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/Monstroxx/eduko-backend/internal/middleware"
	"github.com/Monstroxx/eduko-backend/internal/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
)

type AuthService struct {
	db *pgxpool.Pool
}

func NewAuthService(db *pgxpool.Pool) *AuthService {
	return &AuthService{db: db}
}

type LoginResult struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func (s *AuthService) Login(ctx context.Context, username, password, schoolID, jwtSecret string) (*LoginResult, error) {
	var user models.User
	var passwordHash string

	var err error
	if schoolID == "" {
		// Single-school mode: no school_id filter — match by username only.
		err = s.db.QueryRow(ctx,
			`SELECT u.id, u.school_id, u.email, u.username, u.role, u.first_name, u.last_name, u.locale, u.is_active, u.password_hash
			 FROM users u WHERE u.username = $1`,
			username,
		).Scan(
			&user.ID, &user.SchoolID, &user.Email, &user.Username, &user.Role,
			&user.FirstName, &user.LastName, &user.Locale, &user.IsActive, &passwordHash,
		)
	} else {
		err = s.db.QueryRow(ctx,
			`SELECT u.id, u.school_id, u.email, u.username, u.role, u.first_name, u.last_name, u.locale, u.is_active, u.password_hash
			 FROM users u WHERE u.username = $1 AND u.school_id = $2`,
			username, schoolID,
		).Scan(
			&user.ID, &user.SchoolID, &user.Email, &user.Username, &user.Role,
			&user.FirstName, &user.LastName, &user.Locale, &user.IsActive, &passwordHash,
		)
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("query user: %w", err)
	}

	if !user.IsActive {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := generateJWT(user, jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("generate jwt: %w", err)
	}

	return &LoginResult{Token: token, User: user}, nil
}

type RegisterInput struct {
	Username  string          `json:"username"`
	Password  string          `json:"password"`
	Email     *string         `json:"email,omitempty"`
	FirstName string          `json:"first_name"`
	LastName  string          `json:"last_name"`
	Role      models.UserRole `json:"role"`
	SchoolID  uuid.UUID       `json:"school_id"`
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*models.User, error) {
	// Check if user exists
	var exists bool
	err := s.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND school_id = $2)`,
		input.Username, input.SchoolID,
	).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("check user exists: %w", err)
	}
	if exists {
		return nil, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	var user models.User
	err = s.db.QueryRow(ctx,
		`INSERT INTO users (school_id, email, username, password_hash, role, first_name, last_name)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, school_id, email, username, role, first_name, last_name, locale, is_active, created_at, updated_at`,
		input.SchoolID, input.Email, input.Username, string(hash), input.Role,
		input.FirstName, input.LastName,
	).Scan(
		&user.ID, &user.SchoolID, &user.Email, &user.Username, &user.Role,
		&user.FirstName, &user.LastName, &user.Locale, &user.IsActive,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return &user, nil
}

func generateJWT(user models.User, secret string) (string, error) {
	claims := middleware.JWTClaims{
		UserID:   user.ID,
		SchoolID: user.SchoolID,
		Role:     string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
