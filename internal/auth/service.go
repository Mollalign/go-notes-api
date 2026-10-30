package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrWeakPassword       = errors.New("password must contain at least 12 characters")
	ErrEmailTaken         = errors.New("email already registered")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Repository interface {
	Create(ctx context.Context, email, passwordHash string) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func normalizeEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return "", ErrInvalidEmail
	}

	return email, nil
}

func validatePassword(password string) error {
	if utf8.RuneCountInString(password) < 12 {
		return ErrWeakPassword
	}

	if len(password) > 1024 {
		return fmt.Errorf("password is too long")
	}

	return nil
}

func (s *Service) Register(
	ctx context.Context,
	email string,
	password string,
) (User, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return User{}, err
	}

	if err := validatePassword(password); err != nil {
		return User{}, err
	}

	passwordHash, err := HashPassword(password)
	if err != nil {
		return User{}, err
	}

	return s.repo.Create(ctx, email, passwordHash)
}

func (s *Service) Authenticate(
	ctx context.Context,
	email string,
	password string,
) (User, error) {
	email, err := normalizeEmail(email)
	if err != nil || password == "" || len(password) > 1024 {
		return User{}, ErrInvalidCredentials
	}

	user, err := s.repo.GetByEmail(ctx, email)
	if errors.Is(err, ErrUserNotFound) {
		return User{}, ErrInvalidCredentials
	}
	if err != nil {
		return User{}, fmt.Errorf("find user: %w", err)
	}

	matches, err := VerifyPassword(user.PasswordHash, password)
	if err != nil {
		return User{}, fmt.Errorf("verify stored password: %w", err)
	}

	if !matches {
		return User{}, ErrInvalidCredentials
	}

	return user, nil
}
