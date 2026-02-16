package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	DisplayName  string    `gorm:"column:display_name;not null;uniqueIndex" json:"display_name"`
	Email        string    `gorm:"not null;uniqueIndex" json:"email"`
	PasswordHash string    `gorm:"column:password_hash;not null" json:"-"`
	Role         string    `gorm:"not null" json:"role"`
	WeightUnit   string    `gorm:"column:weight_unit" json:"weight_unit"`
	Currency     string    `gorm:"not null" json:"currency"`
	Language     string    `gorm:"not null" json:"language"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (User) TableName() string { return "users" }

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;column:user_id" json:"user_id"`
	TokenHash string    `gorm:"column:token_hash;not null;uniqueIndex" json:"-"`
	ExpiresAt time.Time `gorm:"column:expires_at;not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }

func (r *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// UserCustomOption stores per-user custom dropdown values (species, breed, vaccination).
type UserCustomOption struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;column:user_id" json:"user_id"`
	OptionType string    `gorm:"column:option_type;not null" json:"option_type"`
	Value      string    `gorm:"not null" json:"value"`
	Context    string    `gorm:"not null;default:''" json:"context"`
	CreatedAt  time.Time `json:"created_at"`
}

func (UserCustomOption) TableName() string { return "user_custom_options" }

func (u *UserCustomOption) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
