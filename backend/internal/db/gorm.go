package db

import (
	"fmt"
	"log"
	"time"

	"github.com/pet-medical/api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewGORM opens a GORM DB using the same postgres URL format as New (lib/pq style).
// Use for GORM-based handlers; parameterized queries via GORM help prevent SQL injection.
func NewGORM(dbURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("gorm open: %w", err)
	}
	return db, nil
}

// NewGORMWithRetry opens GORM DB with retries so the app can wait for the DB to be ready.
func NewGORMWithRetry(dbURL string) (*gorm.DB, error) {
	var lastErr error
	for i := 0; i < connectRetries; i++ {
		db, err := NewGORM(dbURL)
		if err == nil {
			sqlDB, _ := db.DB()
			if err := sqlDB.Ping(); err != nil {
				sqlDB.Close()
				lastErr = fmt.Errorf("ping: %w", err)
				if i < connectRetries-1 {
					log.Printf("database not ready (attempt %d/%d): %v; retrying in %v", i+1, connectRetries, lastErr, connectRetryDelay)
					time.Sleep(connectRetryDelay)
				}
				continue
			}
			return db, nil
		}
		lastErr = err
		if i < connectRetries-1 {
			log.Printf("database not ready (attempt %d/%d): %v; retrying in %v", i+1, connectRetries, err, connectRetryDelay)
			time.Sleep(connectRetryDelay)
		}
	}
	return nil, fmt.Errorf("gorm connection failed after %d attempts: %w", connectRetries, lastErr)
}

// MigrateUsernameToDisplayName renames the users.username column to display_name for existing databases.
// Safe to run multiple times; only runs if the username column exists.
func MigrateUsernameToDisplayName(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	// Check if users table has column "username" (PostgreSQL)
	var count int
	err = sqlDB.QueryRow(`
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_name = 'users' AND table_schema = 'public' AND column_name = 'username'
	`).Scan(&count)
	if err != nil {
		return fmt.Errorf("check username column: %w", err)
	}
	if count == 0 {
		return nil
	}
	_, err = sqlDB.Exec(`ALTER TABLE users RENAME COLUMN username TO display_name`)
	if err != nil {
		return fmt.Errorf("rename username to display_name: %w", err)
	}
	log.Printf("migrated users.username -> users.display_name")
	return nil
}

// AutoMigrateAll runs GORM AutoMigrate for all models in FK-safe order.
// Creates tables and adds missing columns; does not drop columns or run data migrations.
func AutoMigrateAll(db *gorm.DB) error {
	if err := MigrateUsernameToDisplayName(db); err != nil {
		return err
	}
	return db.AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
		&models.Pet{},
		&models.Vaccination{},
		&models.WeightEntry{},
		&models.Document{},
		&models.PetPhoto{},
		&models.UserCustomOption{},
		&models.DefaultDropdownOption{},
	)
}
