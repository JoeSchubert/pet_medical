package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DefaultDropdownOption is an admin-managed default for species, breed, or vaccination.
// OptionType: "species", "breed", "vaccination". Context: "" for species; species name for breed/vaccination.
// DurationMonths: used only for vaccination (nullable).
type DefaultDropdownOption struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	OptionType     string    `gorm:"type:varchar(50);not null" json:"option_type"`
	Value          string    `gorm:"type:varchar(500);not null" json:"value"`
	Context        string    `gorm:"type:varchar(255);not null;default:''" json:"context"`
	SortOrder      int       `gorm:"not null;default:0" json:"sort_order"`
	DurationMonths *int      `json:"duration_months,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (DefaultDropdownOption) TableName() string {
	return "default_dropdown_options"
}

func (d *DefaultDropdownOption) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
