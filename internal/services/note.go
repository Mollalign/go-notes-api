package services

import (
	"errors"
	"strings"
)

var ErrTitleRequired = errors.New("title is required")

type CreateInput struct {
	Title   string
	Content string
}

type Service struct{}

func NewService() *Service {
	return &Service{}
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
