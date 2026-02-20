package db

import (
	"log"

	"github.com/pet-medical/api/internal/auth"
	"github.com/pet-medical/api/internal/config"
	"github.com/pet-medical/api/internal/models"
	"gorm.io/gorm"
)

// Free-to-use pet image URLs (Unsplash License). Format: https://images.unsplash.com/photo-{id}?w=400&q=80
const (
	unsplashDog1    = "https://images.unsplash.com/photo-1587300003388-59208cc962cb?w=400&q=80"  // Golden retriever
	unsplashDog2    = "https://images.unsplash.com/photo-1583511655857-d19b40a7a54e?w=400&q=80"  // Dog
	unsplashCat1    = "https://images.unsplash.com/photo-1514888286974-6c03e2ca1dba?w=400&q=80"  // Cat
	unsplashCat2    = "https://images.unsplash.com/photo-1574158622682-e40e69881006?w=400&q=80"  // Cat
	unsplashRabbit  = "https://images.unsplash.com/photo-1657394147242-60a2345f0199?w=400&q=80"  // Rabbit (replaces broken 1585114894510)
)

const demoUserEmail = "jane@example.com"
const demoUserPassword = "demo123"
const demoUserDisplayName = "Jane"

// SeedDemoData creates multiple demo pets (with free public pet images) and an optional second user when no pets exist.
// Safe to run on empty DB; skips if any pets already exist.
func SeedDemoData(gdb *gorm.DB) error {
	var petCount int64
	if err := gdb.Model(&models.Pet{}).Count(&petCount).Error; err != nil {
		return err
	}
	if petCount > 0 {
		return nil
	}

	var admin models.User
	if err := gdb.Where("email = ?", defaultAdminEmail).First(&admin).Error; err != nil {
		return err
	}

	// ---- Admin's pets (4) ----
	pets := []struct {
		name       string
		species    string
		breed      string
		dob        string
		gender     string
		fixed      bool
		color      string
		notes      string
		photoURL   string
		vaxNames   []string
		vaxDates   []string
		nextDue    []string
		costs      []float64
		weights    []float64
		weightDates []string
	}{
		{
			name: "Luna", species: "Dog", breed: "Golden Retriever", dob: "2022-05-15",
			gender: "female", fixed: true, color: "Golden", notes: "Friendly and loves walks.",
			photoURL: unsplashDog1,
			vaxNames: []string{"Rabies", "DHPP (Distemper, Hepatitis, Parvovirus, Parainfluenza)", "Bordetella (Kennel Cough)", "Leptospirosis"},
			vaxDates: []string{"2024-06-01", "2024-08-15", "2024-08-15", "2024-09-01"},
			nextDue:  []string{"2025-06-01", "2025-08-15", "2025-08-15", "2025-09-01"},
			costs:    []float64{45, 0, 35, 42},
			weights:  []float64{58.5, 57.2, 56.8, 55.1, 54.0},
			weightDates: []string{"2025-01-10", "2024-10-01", "2024-07-15", "2024-04-01", "2024-01-10"},
		},
		{
			name: "Max", species: "Dog", breed: "Labrador Retriever", dob: "2020-03-20",
			gender: "male", fixed: true, color: "Black", notes: "Rescue; great with kids.",
			photoURL: unsplashDog2,
			vaxNames: []string{"Rabies", "Bordetella (Kennel Cough)", "DHPP", "Lyme Disease"},
			vaxDates: []string{"2024-01-10", "2024-09-01", "2024-01-10", "2024-05-15"},
			nextDue:  []string{"2025-01-10", "2025-09-01", "2025-01-10", "2025-05-15"},
			costs:    []float64{42, 35, 0, 38},
			weights:  []float64{72.0, 71.2, 70.5, 69.8, 69.0, 68.2},
			weightDates: []string{"2025-02-01", "2024-11-15", "2024-08-01", "2024-05-01", "2024-02-01", "2023-11-01"},
		},
		{
			name: "Whiskers", species: "Cat", breed: "Domestic Shorthair", dob: "2021-08-10",
			gender: "male", fixed: true, color: "Tabby", notes: "Indoor only.",
			photoURL: unsplashCat1,
			vaxNames: []string{"Rabies", "FVRCP (Feline Distemper)", "FeLV (Feline Leukemia)"},
			vaxDates: []string{"2024-04-01", "2024-04-01", "2024-03-15"},
			nextDue:  []string{"2025-04-01", "2025-04-01", "2025-03-15"},
			costs:    []float64{38, 45, 48},
			weights:  []float64{12.5, 12.2, 11.9, 11.5, 11.2},
			weightDates: []string{"2025-01-05", "2024-07-01", "2024-04-01", "2024-01-01", "2023-10-01"},
		},
		{
			name: "Bella", species: "Cat", breed: "Siamese", dob: "2023-01-12",
			gender: "female", fixed: true, color: "Cream", notes: "Vocal and playful.",
			photoURL: unsplashCat2,
			vaxNames: []string{"Rabies", "FVRCP (Feline Distemper)"},
			vaxDates: []string{"2024-11-01", "2024-11-01"},
			nextDue:  []string{"2025-11-01", "2025-11-01"},
			costs:    []float64{40, 45},
			weights:  []float64{8.0, 7.6, 7.2, 6.8},
			weightDates: []string{"2025-01-15", "2024-10-01", "2024-06-01", "2024-02-01"},
		},
		{
			name: "Thumper", species: "Rabbit", breed: "Dutch", dob: "2023-06-01",
			gender: "male", fixed: false, color: "Black and white", notes: "House rabbit.",
			photoURL: unsplashRabbit,
			vaxNames: []string{"RHDV1 / RHDV2 (Rabbit Hemorrhagic Disease)", "Myxomatosis"},
			vaxDates: []string{"2024-05-01", "2024-05-01"},
			nextDue:  []string{"2025-05-01", "2025-05-01"},
			costs:    []float64{55, 45},
			weights:  []float64{4.2, 4.0, 3.8, 3.5},
			weightDates: []string{"2025-01-01", "2024-09-01", "2024-06-01", "2024-03-01"},
		},
	}

	cfg := config.Load()
	weightUnit := cfg.DefaultWeightUnit
	if weightUnit == "" {
		weightUnit = "lbs"
	}
	for _, p := range pets {
		pet := models.Pet{
			UserID:      admin.ID,
			Name:        p.name,
			Species:     &p.species,
			Breed:       &p.breed,
			DateOfBirth: &p.dob,
			Gender:      &p.gender,
			Fixed:       &p.fixed,
			Color:       &p.color,
			Notes:       &p.notes,
			PhotoURL:    &p.photoURL,
		}
		if err := gdb.Create(&pet).Error; err != nil {
			return err
		}
		for i, name := range p.vaxNames {
			vax := models.Vaccination{
				PetID:          pet.ID,
				Name:           name,
				AdministeredAt: p.vaxDates[i],
				NextDue:        ptr(p.nextDue[i]),
			}
			if i < len(p.costs) && p.costs[i] > 0 {
				vax.CostUSD = &p.costs[i]
			}
			if err := gdb.Create(&vax).Error; err != nil {
				return err
			}
		}
		for i, w := range p.weights {
			weightEntry := models.WeightEntry{
				PetID:      pet.ID,
				WeightLbs: w,
				EntryUnit: weightUnit,
				MeasuredAt: p.weightDates[i],
			}
			if err := gdb.Create(&weightEntry).Error; err != nil {
				return err
			}
		}
		log.Printf("seed demo: created pet %s", p.name)
	}

	// ---- Second user (Jane) and her pets ----
	var jane models.User
	err := gdb.Where("email = ?", demoUserEmail).First(&jane).Error
	if err != nil {
		hash, herr := auth.HashPassword(demoUserPassword)
		if herr != nil {
			return herr
		}
		jane = models.User{
			DisplayName:  demoUserDisplayName,
			Email:        demoUserEmail,
			PasswordHash: hash,
			Role:         "user",
			WeightUnit:   cfg.DefaultWeightUnit,
			Currency:     cfg.DefaultCurrency,
			Language:     cfg.DefaultLanguage,
		}
		if err := gdb.Create(&jane).Error; err != nil {
			return err
		}
		log.Printf("seed demo: created user %s (%s)", demoUserDisplayName, demoUserEmail)
	}

	janePets := []struct {
		name        string
		species     string
		breed       string
		dob         string
		photoURL    string
		vaxNames    []string
		vaxDates    []string
		nextDue     []string
		costs       []float64
		weights     []float64
		weightDates []string
	}{
		{
			name: "Buddy", species: "Dog", breed: "Beagle", dob: "2019-11-05", photoURL: unsplashDog1,
			vaxNames:    []string{"Rabies", "DHPP", "Bordetella (Kennel Cough)"},
			vaxDates:    []string{"2024-06-01", "2024-06-01", "2024-08-01"},
			nextDue:     []string{"2025-06-01", "2025-06-01", "2025-08-01"},
			costs:       []float64{42, 0, 32},
			weights:     []float64{28.5, 28.0, 27.5, 27.0},
			weightDates: []string{"2025-01-10", "2024-08-01", "2024-04-01", "2024-01-01"},
		},
		{
			name: "Mittens", species: "Cat", breed: "Mixed breed", dob: "2022-02-14", photoURL: unsplashCat2,
			vaxNames:    []string{"Rabies", "FVRCP (Feline Distemper)"},
			vaxDates:    []string{"2024-06-01", "2024-06-01"},
			nextDue:     []string{"2025-06-01", "2025-06-01"},
			costs:       []float64{38, 44},
			weights:     []float64{10.2, 10.0, 9.8, 9.5},
			weightDates: []string{"2025-01-05", "2024-09-01", "2024-05-01", "2024-01-15"},
		},
	}
	for _, p := range janePets {
		pet := models.Pet{
			UserID:      jane.ID,
			Name:        p.name,
			Species:     &p.species,
			Breed:       &p.breed,
			DateOfBirth: &p.dob,
			PhotoURL:    &p.photoURL,
		}
		if err := gdb.Create(&pet).Error; err != nil {
			return err
		}
		for i, name := range p.vaxNames {
			vax := models.Vaccination{
				PetID:          pet.ID,
				Name:           name,
				AdministeredAt: p.vaxDates[i],
				NextDue:        ptr(p.nextDue[i]),
			}
			if i < len(p.costs) && p.costs[i] > 0 {
				vax.CostUSD = &p.costs[i]
			}
			if err := gdb.Create(&vax).Error; err != nil {
				return err
			}
		}
		for i, w := range p.weights {
			weightEntry := models.WeightEntry{
				PetID:      pet.ID,
				WeightLbs: w,
				EntryUnit: weightUnit,
				MeasuredAt: p.weightDates[i],
			}
			if err := gdb.Create(&weightEntry).Error; err != nil {
				return err
			}
		}
		log.Printf("seed demo: created pet %s for %s", p.name, demoUserDisplayName)
	}

	return nil
}

func ptr[T any](v T) *T { return &v }
