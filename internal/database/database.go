package database

import (
	"fmt"
	"log"

	"github.com/bookwise/api/config"
	"github.com/bookwise/api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase initializes the database connection
func InitDatabase(cfg *config.Config) error {
	dsn := cfg.Database.GetDSN()
	
	// Set GORM logger level based on GIN_MODE
	logLevel := logger.Info
	if cfg.Server.GinMode == "release" {
		logLevel = logger.Silent
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB for connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	DB = db
	log.Println("✅ Database connection established")

	return nil
}

// AutoMigrate runs automatic migration for models
func AutoMigrate() error {
	log.Println("Running database migrations...")
	
	err := DB.AutoMigrate(
		&models.Book{},
		&models.Quiz{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create indexes
	if err := createIndexes(); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	log.Println("✅ Database migrations completed")
	return nil
}

// createIndexes creates additional indexes for better performance
func createIndexes() error {
	// Index for ISBN lookup
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_books_isbn ON books(isbn)").Error; err != nil {
		return err
	}

	// Index for quiz book_id lookup
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_quizzes_book_id ON quizzes(book_id)").Error; err != nil {
		return err
	}

	// Index for quiz status
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_books_quiz_status ON books(quiz_status)").Error; err != nil {
		return err
	}

	log.Println("✅ Database indexes created")
	return nil
}

// CloseDatabase closes the database connection
func CloseDatabase() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

