package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"go-notes-api/internal/handlers"
)

func TestHelloRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/hello", handlers.Hello)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/hello",
		nil,
	)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	expectedBody := "Hello from Gin!"
	actualBody := recorder.Body.String()

	if actualBody != expectedBody {
		t.Fatalf(
			"expected body %q, got %q",
			expectedBody,
			actualBody,
		)
	}
}
