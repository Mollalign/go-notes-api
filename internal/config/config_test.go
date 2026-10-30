package config

import (
	"strings"
	"testing"
)

func TestLoadDatabase(t *testing.T) {
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_NAME", "go_notes_api")
	t.Setenv("DB_USER", "go_notes_app")
	t.Setenv("DB_PASSWORD", "test-secret")
	t.Setenv("DB_SSLMODE", "disable")

	cfg, err := LoadDatabase()
	if err != nil {
		t.Fatalf("LoadDatabase() returned an error: %v", err)
	}

	if cfg.Host != "localhost" {
		t.Errorf("expected host localhost, got %q", cfg.Host)
	}

	if cfg.Port != 5432 {
		t.Errorf("expected port 5432, got %d", cfg.Port)
	}

	if cfg.Password != "test-secret" {
		t.Error("password was not preserved correctly")
	}
}

func TestLoadDatabaseRequiresPassword(t *testing.T) {
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_NAME", "go_notes_api")
	t.Setenv("DB_USER", "go_notes_app")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_SSLMODE", "disable")

	_, err := LoadDatabase()
	if err == nil {
		t.Fatal("expected an error when DB_PASSWORD is missing")
	}

	if !strings.Contains(err.Error(), "DB_PASSWORD") {
		t.Errorf("expected a DB_PASSWORD error, got: %v", err)
	}
}
