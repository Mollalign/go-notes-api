package handlers

import (
	"go-notes-api/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type createNoteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type noteInputService interface {
	ValidateCreateInput(
		input services.CreateInput,
	) (services.CreateInput, error)
}

type NoteHandler struct {
	service noteInputService
}

func NewNoteHandler(service noteInputService) *NoteHandler {
	return &NoteHandler{
		service: service,
	}
}

func (h *NoteHandler) ValidateNoteInput(c *gin.Context) {
	var request createNoteRequest

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
