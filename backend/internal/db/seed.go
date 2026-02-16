package db

import (
	"log"

	"github.com/google/uuid"
	"github.com/pet-medical/api/internal/auth"
	"github.com/pet-medical/api/internal/config"
	"github.com/pet-medical/api/internal/models"
	"gorm.io/gorm"
)

const defaultAdminDisplayName = "admin"
const defaultAdminPassword = "admin123"
const defaultAdminEmail = "admin@example.com"
const defaultAdminRole = "admin"

// SeedDefaultAdmin creates the default admin user if no users exist. Uses cfg for default weight unit, currency, and language.
func SeedDefaultAdmin(db *gorm.DB, cfg *config.Config) error {
	var count int64
	if err := db.Model(&models.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	hash, err := auth.HashPassword(defaultAdminPassword)
	if err != nil {
		return err
	}
	u := models.User{
		ID:           uuid.New(),
		DisplayName:  defaultAdminDisplayName,
		Email:        defaultAdminEmail,
		PasswordHash: hash,
		Role:         defaultAdminRole,
		WeightUnit:   cfg.DefaultWeightUnit,
		Currency:     cfg.DefaultCurrency,
		Language:     cfg.DefaultLanguage,
	}
	if err := db.Create(&u).Error; err != nil {
		return err
	}
	log.Printf("created default admin user: %s", defaultAdminDisplayName)
	return nil
}
