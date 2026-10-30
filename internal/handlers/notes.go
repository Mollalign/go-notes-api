package handlers

import (
	"context"
	"go-notes-api/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type noteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type noteService interface {
	ValidateCreateInput(
		input services.CreateInput,
	) (services.CreateInput, error)

	Create(
		ctx context.Context,
		input services.CreateInput,
	) (services.Note, error)

	List(ctx context.Context) ([]services.Note, error)
}

type NoteHandler struct {
	service noteService
}

func NewNoteHandler(service noteService) *NoteHandler {
	return &NoteHandler{
		service: service,
	}
}

func (h *NoteHandler) ValidateNoteInput(c *gin.Context) {
	var request noteRequest

	if err := c.ShouldBindBodyWithJSON(&request); err != nil {
		writeError(
			c,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Request body is invalid.",
		)
		return
	}

	input, err := h.service.ValidateCreateInput(services.CreateInput{
		Title:   request.Title,
		Content: request.Content,
	})

	if err != nil {
		writeApplicationError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Input passed business rules",
		"title":   input.Title,
		"content": input.Content,
	})
}

func (h *NoteHandler) CreateNote(c *gin.Context) {
	var request noteRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(
			c,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid JSON request body",
		)
		return
	}

	note, err := h.service.Create(
		c.Request.Context(),
		services.CreateInput{
			Title:   request.Title,
			Content: request.Content,
		},
	)

	if err != nil {
		writeApplicationError(c, err)
		return
	}
	c.JSON(http.StatusCreated, note)
}
