package handlers

import (
	"errors"
	"log"
	"net/http"

	"go-notes-api/internal/services"

	"github.com/gin-gonic/gin"
)

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error apiError `json:"error"`
}

func writeError(
	c *gin.Context,
	status int,
	code string,
	message string,
) {
	c.JSON(status, errorResponse{
		Error: apiError{
			Code:    code,
			Message: message,
		},
	})
}

func writeApplicationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrTitleRequired):
		writeError(
			c,
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"Title is required.",
		)

	default:
		log.Printf(
			"request_id=%s application_error=%v",
			c.GetString("request_id"),
			err,
		)

		writeError(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"An unexpected error occurred.",
		)
	}
}
