package main

import (
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL required")
	}

	m, err := migrate.New("file://migrations", databaseURL)
	if err != nil {
		log.Fatalf("migrate init: %v", err)
	}

	cmd := "up"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	var migErr error
	switch cmd {
	case "up":
		migErr = m.Up()
	case "down":
		migErr = m.Down()
	default:
		log.Fatalf("usage: migrate [up|down]")
	}

	if migErr != nil && migErr != migrate.ErrNoChange {
		log.Fatalf("migrate %s: %v", cmd, migErr)
	}
	log.Printf("migrate %s ok", cmd)
}
