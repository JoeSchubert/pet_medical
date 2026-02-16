package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Pet struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	Name        string     `gorm:"not null" json:"name"`
	Species     *string    `json:"species,omitempty"`
	Breed       *string    `json:"breed,omitempty"`
	DateOfBirth *string    `gorm:"column:date_of_birth" json:"date_of_birth,omitempty"`
	Gender      *string    `json:"gender,omitempty"`
	Fixed       *bool      `gorm:"column:fixed" json:"fixed,omitempty"` // spayed or neutered
	Color       *string    `json:"color,omitempty"`
	MicrochipID     *string `gorm:"column:microchip_id" json:"microchip_id,omitempty"`
	MicrochipCompany *string `gorm:"column:microchip_company" json:"microchip_company,omitempty"`
	Notes       *string    `json:"notes,omitempty"`
	PhotoURL    *string    `gorm:"column:photo_url" json:"photo_url,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (Pet) TableName() string { return "pets" }

func (p *Pet) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type Vaccination struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	PetID          uuid.UUID  `gorm:"type:uuid;not null;column:pet_id" json:"pet_id"`
	Name           string     `gorm:"not null" json:"name"`
	AdministeredAt string     `gorm:"column:administered_at;not null" json:"administered_at"`
	NextDue        *string    `gorm:"column:next_due" json:"next_due,omitempty"`
	CostUSD        *float64   `gorm:"column:cost_usd" json:"cost_usd,omitempty"`
	Veterinarian   *string    `json:"veterinarian,omitempty"`
	BatchNumber    *string    `gorm:"column:batch_number" json:"batch_number,omitempty"`
	Notes          *string    `json:"notes,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (Vaccination) TableName() string { return "vaccinations" }

func (v *Vaccination) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

type WeightEntry struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	PetID       uuid.UUID `gorm:"type:uuid;not null;column:pet_id" json:"pet_id"`
	WeightLbs   float64   `gorm:"column:weight_lbs;not null" json:"weight_lbs"`
	EntryUnit   string    `gorm:"column:entry_unit;not null" json:"entry_unit"`
	MeasuredAt  string    `gorm:"column:measured_at;not null" json:"measured_at"`
	Approximate bool      `gorm:"column:approximate;not null;default:false" json:"approximate"`
	Notes       *string   `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func (WeightEntry) TableName() string { return "weight_entries" }

func (w *WeightEntry) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}

type PetPhoto struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	PetID        uuid.UUID `gorm:"type:uuid;not null;column:pet_id" json:"pet_id"`
	FilePath     string    `gorm:"column:file_path;not null" json:"file_path"`
	DisplayOrder int       `gorm:"column:display_order;not null" json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
}

func (PetPhoto) TableName() string { return "pet_photos" }

func (p *PetPhoto) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type Document struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PetID         uuid.UUID `gorm:"type:uuid;not null;column:pet_id" json:"pet_id"`
	Name          string    `gorm:"not null" json:"name"`
	DocType       *string   `gorm:"column:doc_type" json:"doc_type,omitempty"`
	FilePath      string    `gorm:"column:file_path;not null" json:"file_path"`
	FileSize      *int64    `gorm:"column:file_size" json:"file_size,omitempty"`
	MimeType      *string   `gorm:"column:mime_type" json:"mime_type,omitempty"`
	Notes         *string   `json:"notes,omitempty"`
	ExtractedText *string   `gorm:"column:extracted_text" json:"-"` // OCR/text extraction for search; not exposed in API
	CreatedAt     time.Time `json:"created_at"`
}

func (Document) TableName() string { return "documents" }

func (d *Document) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
