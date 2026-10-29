package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"go-notes-api/internal/config"
	"go-notes-api/internal/handlers"
	"go-notes-api/internal/services"
)

type NoteResponse struct {
	ID           int    `json:"id"`
	Title        string `json:"title" binding:"required, min=3, max=100"`
	Content      string `json:"content" binding:"required, max=1000"`
	Priority     int    `json:"priority" binding:"required, gte=3, lte=5"`
	Completed    bool   `json:"completed"`
	InternalNote string `json:"-"`
}

type CreateNoteRequest struct {
	Title    string `json:"title" binding:"required,min=3,max=100"`
	Content  string `json:"content" binding:"required,max=1000"`
	Priority int    `json:"priority" binding:"gte=1,lte=5"`
}

// type CreateNoteRequest struct {
// 	Title   string `json:"title"`
// 	Content string `json:"content"`
// }

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error APIError `json:"error"`
}

func respondError(
	c *gin.Context,
	status int,
	code string,
	message string,
) {
	c.JSON(status, ErrorResponse{
		Error: APIError{
			Code:    code,
			Message: message,
		},
	})
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		requestID := c.GetString("request_id")

		log.Printf(
			"request_id=%s method=%s path=%s status=%d duration=%s",
			requestID,
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
		)
	}
}

func RequireDemoHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-Demo-Access") != "allow" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Access denied",
			})
			return
		}

		c.Next()
	}
}

func LifecycleMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		log.Println("Middleware: before")
		ctx.Next()
		log.Println("Middleware: after")
	}
}

func generateRequestID() (string, error) {
	var randomByte [16]byte

	if _, err := rand.Read(randomByte[:]); err != nil {
		return "", err
	}

	return hex.EncodeToString(randomByte[:]), nil
}

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID, err := generateRequestID()

		if err != nil {
			log.Printf("event=request_id_generation_failed error=%v", err)

			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Could not process request",
			})
			return
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

func AuthenticationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		var userID string
		var role string

		switch authHeader {
		case "Bearer demo-user-token":
			userID = "user-123"
			role = "user"
		case "Bearer demo-admin-token":
			userID = "admin-123"
			role = "admin"
		default:
			c.Header("WWW-Authenticate", "Bearer")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Missing or invalid credentials",
			})
			return
		}

		c.Set("user_id", userID)
		c.Set("role", role)

		c.Next()
	}
}

func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")

		if role != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "You do not have permission to access this resource",
			})
			return
		}

		c.Next()
	}
}

func CustomRecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(
		func(c *gin.Context, recovered any) {
			log.Printf(
				"panic_recovered request_id=%s panic=%v",
				c.GetString("request_id"),
				recovered,
			)

			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				ErrorResponse{
					Error: APIError{
						Code:    "INTERNAL_ERROR",
						Message: "An unexpected error occurred",
					},
				},
			)
		},
	)
}

type AppError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *AppError) Error() string {
	return e.Message
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		if c.Writer.Written() {
			log.Printf(
				"request_id=%s error_after_response_written=%v",
				c.GetString("request_id"),
				err,
			)
			return
		}

		if appErr, ok := errors.AsType[*AppError](err); ok {
			c.JSON(appErr.StatusCode, ErrorResponse{
				Error: APIError{
					Code:    appErr.Code,
					Message: appErr.Message,
				},
			})
			return
		}

		log.Printf(
			"request_id=%s unexpected_error=%v",
			c.GetString("request_id"),
			err,
		)

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: APIError{
				Code:    "INTERNAL_ERROR",
				Message: "An unexpected error occurred.",
			},
		})
	}
}

func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(
			c.Request.Context(),
			timeout,
		)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	router := gin.New()

	router.Use(RequestIDMiddleware())
	router.Use(RequestLogger())
	router.Use(CustomRecoveryMiddleware())
	router.Use(ErrorHandler())

	router.GET("/hello", handlers.Hello)

	noteService := services.NewService()
	noteHandler := handlers.NewNoteHandler(noteService)

	router.POST("/notes/validate", noteHandler.ValidateNoteInput)

	router.GET("/welcome", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to my Go Notes API!")
	})

	router.GET(
		"/demo/slow",
		TimeoutMiddleware(2*time.Second),
		func(c *gin.Context) {
			ctx := c.Request.Context()

			timer := time.NewTimer(5 * time.Second)
			defer timer.Stop()

			select {
			case <-timer.C:
				c.JSON(http.StatusOK, gin.H{
					"message": "Work completed",
				})

			case <-ctx.Done():
				if ctx.Err() == context.DeadlineExceeded {
					c.JSON(http.StatusGatewayTimeout, gin.H{
						"error": "Request timed out",
					})
				}
				return
			}
		},
	)

	// Group 1: /api
	api := router.Group("/api")

	// Group 2: /api/v1
	v1 := api.Group("/v1")

	// Group 3: /api/v1/notes
	notes := v1.Group("/notes")

	// Any authenticated user can reach this endpoint
	router.GET("/profile", AuthenticationMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id": c.GetString("user_id"),
			"role":    c.GetString("role"),
		})
	})

	// Only authenticated admins can reach this endpoint
	router.GET("/admin/report",
		AuthenticationMiddleware(),
		RequireRole("admin"),
		func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"report": "Confidential admin report",
			})
		},
	)

	// Retrieve all notes
	notes.GET("/", func(c *gin.Context) {
		page := c.DefaultQuery("page", "1")
		limit := c.DefaultQuery("limit", "10")
		sort := c.DefaultQuery("sort", "created_at")

		c.String(
			http.StatusOK,
			"Note page=%s, limit=%s, sort=%s",
			page,
			limit,
			sort,
		)
	})

	notes.POST("/bind-demo", func(c *gin.Context) {
		var input CreateNoteRequest

		if err := c.ShouldBindJSON(&input); err != nil {
			respondError(
				c,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				"The request body is invalid or fails validation.",
			)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Request received successfully",
			"title":   input.Title,
			"content": input.Content,
		})
	})

	router.POST("/demo/bind", func(c *gin.Context) {
		var input CreateNoteRequest

		if err := c.Bind(&input); err != nil {
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Input is valid",
		})
	})

	router.GET("/lifecycle", LifecycleMiddleware(), func(c *gin.Context) {
		log.Println("Handler: running")
		c.String(http.StatusOK, "Request completed")
	})

	router.GET("/protected-demo", RequireDemoHeader(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "You reached the handler",
		})
	})

	router.POST("/demo/should-bind", func(c *gin.Context) {
		var input CreateNoteRequest

		if err := c.ShouldBind(&input); err != nil {
			respondError(
				c,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				"The request data is invalid.",
			)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Input is valid",
		})
	})

	// Create a note.
	notes.POST("/", func(c *gin.Context) {
		c.String(http.StatusCreated, "Create note")
	})

	// Retrieve one note.
	notes.GET("/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.String(http.StatusOK, "Get note %s", id)
	})

	// Replace a note.
	notes.PUT("/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.String(http.StatusOK, "Replace note %s", id)
	})

	// Partially update a note.
	notes.PATCH("/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.String(http.StatusOK, "Update note %s", id)
	})

	// Delete a note.
	notes.DELETE("/:id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	router.GET("/note-tags-demo", func(c *gin.Context) {
		response := NoteResponse{
			ID:           42,
			Title:        "Learn Gin",
			Completed:    false,
			InternalNote: "For internal use only",
		}

		c.JSON(http.StatusOK, response)
	})

	router.GET("/request-info", func(c *gin.Context) {
		userAgent := c.GetHeader("User-Agent")
		accept := c.GetHeader("Accept")
		clientVersion := c.GetHeader("X-Client-Version")
		acceptLanguage := c.GetHeader("Accept-Language")

		method := c.Request.Method
		path := c.Request.URL.Path

		c.String(
			http.StatusOK,
			"Method: %s\nPath: %s\nUser-Agent: %s\nAccept: %s\nClient Version: %s Accept-Language %s",
			method,
			path,
			userAgent,
			accept,
			clientVersion,
			acceptLanguage,
		)
	})

	router.GET("/context-demo", func(c *gin.Context) {
		// Read the original HTTP request
		request := c.Request

		method := request.Method
		path := request.URL.Path
		page := c.Query("page")
		userAgent := c.GetHeader("User-Agent")

		// Set a response header
		c.Header("X-API-Version", "1")

		c.String(
			http.StatusOK,
			"Method: %s\nPath: %s\nPage: %s\nUser-Agent: %s",
			method,
			path,
			page,
			userAgent,
		)
	})

	router.GET("/json-demo", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":   "hello from Gin",
			"framework": "Gin",
			"status":    "success",
		})
	})

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
