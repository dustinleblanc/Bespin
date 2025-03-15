package database

import (
	"fmt"
	"log"
	"os"

	"github.com/dustinleblanc/go-bespin-api/pkg/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewConnection creates a new database connection
func NewConnection() (*gorm.DB, error) {
	// Get database connection parameters from environment variables
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "postgres"
	}

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "bespin"
	}

	// Create DSN string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Create logger
	logger := log.New(os.Stdout, "[Database] ", log.LstdFlags)
	logger.Printf("Connecting to PostgreSQL at %s:%s", host, port)

	// Open database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto migrate models
	if err := db.AutoMigrate(&models.WebhookReceipt{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	logger.Println("Successfully connected to database")
	return db, nil
}
