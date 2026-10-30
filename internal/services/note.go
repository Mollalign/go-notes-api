package services

import (
	"context"
	"errors"
	"strings"
	"time"
)

var ErrTitleRequired = errors.New("title is required")

type CreateInput struct {
	Title   string
	Content string
}

type Note struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Repository defines the data operations the service needs
type Repository interface {
	Create(ctx context.Context, input CreateInput) (Note, error)
	List(ctx context.Context) ([]Note, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) ValidateCreateInput(
	input CreateInput,
) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)

	if input.Title == "" {
		return CreateInput{}, ErrTitleRequired
	}

	return input, nil
}

func (s *Service) Create(
	ctx context.Context,
	input CreateInput,
) (Note, error) {
	input, err := s.ValidateCreateInput(input)
	if err != nil {
		return Note{}, err
	}

	note, err := s.repo.Create(ctx, input)
	if err != nil {
		return Note{}, err
	}

	return note, nil
}

func (s *Service) List(
	ctx context.Context,
) ([]Note, error) {
	return s.repo.List(ctx)
}
