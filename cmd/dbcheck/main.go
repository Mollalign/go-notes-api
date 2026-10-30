package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-notes-api/internal/config"
	"go-notes-api/internal/database"
)

func main() {
	cfg, err := config.LoadDatabase()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	db, err := database.Open(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("PostgreSQL connection successful")
}
