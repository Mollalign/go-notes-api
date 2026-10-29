package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
}

func Load() (Config, error) {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return Config{}, fmt.Errorf(
			"invalid PORT %q: must be a number between 1 and 65535",
			port,
		)
	}

	return Config{
		Port: port,
	}, nil
}

func LoadDatabase() (DatabaseConfig, error) {
	host, err := requiredValue("DB_HOST")
	if err != nil {
		return DatabaseConfig{}, err
	}

	portText := strings.TrimSpace(os.Getenv("DB_PORT"))
	if portText == "" {
		portText = "5432"
	}

	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 {
		return DatabaseConfig{}, fmt.Errorf(
			"invalid DB_PORT: must be a number between 1 and 65535",
		)
	}

	name, err := requiredValue("DB_NAME")
	if err != nil {
		return DatabaseConfig{}, err
	}

	user, err := requiredValue("DB_USER")
	if err != nil {
		return DatabaseConfig{}, err
	}

	password := os.Getenv("DB_PASSWORD")
	if strings.TrimSpace(password) == "" {
		return DatabaseConfig{}, fmt.Errorf(
			"DB_PASSWORD is required",
		)
	}

	sslMode, err := requiredValue("DB_SSLMODE")
	if err != nil {
		return DatabaseConfig{}, err
	}

	return DatabaseConfig{
		Host:     host,
		Port:     port,
		Name:     name,
		User:     user,
		Password: password,
		SSLMode:  sslMode,
	}, nil
}

func requiredValue(key string) (string, error) {
	value, ok := os.LookupEnv(key)

	if !ok || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%s is required", key)
	}

	return strings.TrimSpace(value), nil
}
