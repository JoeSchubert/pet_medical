package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pet-medical/api/internal/middleware"
	"github.com/pet-medical/api/internal/models"
	"gorm.io/gorm"
)

type PetsHandler struct {
	DB        *gorm.DB
	UploadDir string // optional; required for Delete to remove document/photo files
}

func (h *PetsHandler) List(w http.ResponseWriter, r *http.Request) {
	u := middleware.GetUser(r.Context())
	if u == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	var pets []models.Pet
	err := h.DB.Where("user_id = ?", u.ID).Order("name").Find(&pets).Error
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	if pets == nil {
		pets = []models.Pet{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pets)
}

func (h *PetsHandler) Get(w http.ResponseWriter, r *http.Request) {
	u := middleware.GetUser(r.Context())
	if u == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	var pet models.Pet
	err = h.DB.Where("id = ? AND user_id = ?", id, u.ID).First(&pet).Error
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pet)
}

func (h *PetsHandler) Create(w http.ResponseWriter, r *http.Request) {
	u := middleware.GetUser(r.Context())
	if u == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	var pet models.Pet
	if err := json.NewDecoder(r.Body).Decode(&pet); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	pet.UserID = u.ID
	pet.ID = uuid.Nil
	if pet.Name == "" {
		http.Error(w, `{"error":"name required"}`, http.StatusBadRequest)
		return
	}
	if err := validatePetInput(&pet); err != nil {
		http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
		return
	}
	if err := h.DB.Create(&pet).Error; err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pet)
}

func (h *PetsHandler) Update(w http.ResponseWriter, r *http.Request) {
	u := middleware.GetUser(r.Context())
	if u == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	var pet models.Pet
	if err := json.NewDecoder(r.Body).Decode(&pet); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	pet.ID = id
	pet.UserID = u.ID
	if pet.Name == "" {
		http.Error(w, `{"error":"name required"}`, http.StatusBadRequest)
		return
	}
	if err := validatePetInput(&pet); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	result := h.DB.Model(&models.Pet{}).Where("id = ? AND user_id = ?", id, u.ID).Updates(map[string]interface{}{
		"name": pet.Name, "species": pet.Species, "breed": pet.Breed, "date_of_birth": pet.DateOfBirth,
		"gender": pet.Gender, "fixed": pet.Fixed, "color": pet.Color, "microchip_id": pet.MicrochipID, "notes": pet.Notes, "photo_url": pet.PhotoURL,
		"updated_at": time.Now(),
	})
	if result.Error != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	if result.RowsAffected == 0 {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	h.DB.Where("id = ?", id).First(&pet)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pet)
}

func (h *PetsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	u := middleware.GetUser(r.Context())
	if u == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	// Verify ownership
	var count int64
	if h.DB.Model(&models.Pet{}).Where("id = ? AND user_id = ?", id, u.ID).Count(&count).Error != nil || count == 0 {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	// Delete uploaded files from disk (documents and photos) when UploadDir is set
	if h.UploadDir != "" {
		var docs []models.Document
		if h.DB.Where("pet_id = ?", id).Find(&docs).Error == nil {
			for _, d := range docs {
				absPath := filepath.Join(h.UploadDir, filepath.FromSlash(d.FilePath))
				_ = os.Remove(absPath)
			}
		}
		var photos []models.PetPhoto
		if h.DB.Where("pet_id = ?", id).Find(&photos).Error == nil {
			for _, p := range photos {
				absPath := filepath.Join(h.UploadDir, filepath.FromSlash(p.FilePath))
				_ = os.Remove(absPath)
			}
		}
	}
	// Delete child rows then pet (no FK cascade in schema)
	h.DB.Where("pet_id = ?", id).Delete(&models.Vaccination{})
	h.DB.Where("pet_id = ?", id).Delete(&models.WeightEntry{})
	h.DB.Where("pet_id = ?", id).Delete(&models.Document{})
	h.DB.Where("pet_id = ?", id).Delete(&models.PetPhoto{})
	result := h.DB.Where("id = ? AND user_id = ?", id, u.ID).Delete(&models.Pet{})
	if result.Error != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

const (
	maxPetNameLen    = 200
	maxPetNotesLen   = 5000
	maxPetStringLen  = 200 // species, breed, color, etc.
	maxPetMicrochipLen = 100
)

func validatePetInput(pet *models.Pet) error {
	if len(pet.Name) > maxPetNameLen {
		return errors.New("name too long")
	}
	if pet.Notes != nil && len(*pet.Notes) > maxPetNotesLen {
		return errors.New("notes too long")
	}
	if pet.Species != nil && len(*pet.Species) > maxPetStringLen {
		return errors.New("species too long")
	}
	if pet.Breed != nil && len(*pet.Breed) > maxPetStringLen {
		return errors.New("breed too long")
	}
	if pet.Color != nil && len(*pet.Color) > maxPetStringLen {
		return errors.New("color too long")
	}
	if pet.MicrochipID != nil && len(*pet.MicrochipID) > maxPetMicrochipLen {
		return errors.New("microchip ID too long")
	}
	if pet.MicrochipCompany != nil && len(*pet.MicrochipCompany) > maxPetStringLen {
		return errors.New("microchip company too long")
	}
	return nil
}
