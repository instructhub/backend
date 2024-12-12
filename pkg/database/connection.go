package db

import (
	"fmt"
	"os"
	"time"

	"github.com/instructhub/backend/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// GORM DB Client
var DB *gorm.DB

// InitPostgres initializes the PostgreSQL connection
func init() {
	// Set connection options

	host := os.Getenv("DATABASE_HOST")
	user := os.Getenv("DATABASE_USER")
	password := os.Getenv("DATABASE_PASSWORD")
	dbname := os.Getenv("DATABASE_DBNAME")
	port := os.Getenv("DATABASE_PORT")
	sslmode := os.Getenv("DATABASE_SSLMODE")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Log.Sugar().Fatalf("Failed to connect to database: %v", err)
	}

	// Set up connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		logger.Log.Sugar().Fatalf("Failed to get SQL DB instance: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)                  // Set the maximum number of idle connections
	sqlDB.SetMaxOpenConns(100)                 // Set the maximum number of open connections
	sqlDB.SetConnMaxLifetime(30 * time.Minute) // Set the maximum lifetime of a connection

	// Check connection
	if err = sqlDB.Ping(); err != nil {
		logger.Log.Sugar().Fatal("Failed to ping PostgreSQL:", err)
	}

	logger.Log.Info("Successfully connected to PostgreSQL")
}

// GetDB returns the GORM DB instance
func GetDB() *gorm.DB {
	return DB
}

// ClosePostgres closes the PostgreSQL connection
func ClosePostgres() {
	sqlDB, err := DB.DB()
	if err != nil {
		logger.Log.Sugar().Fatal("Failed to get SQL DB instance:", err)
	}

	if err := sqlDB.Close(); err != nil {
		logger.Log.Sugar().Fatal("Failed to close PostgreSQL connection:", err)
	}
	logger.Log.Info("Successfully disconnected to PostgreSQL")
}
