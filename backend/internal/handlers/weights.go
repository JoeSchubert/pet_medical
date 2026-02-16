package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pet-medical/api/internal/middleware"
	"github.com/pet-medical/api/internal/models"
	"gorm.io/gorm"
)

// WeightCreateStore abstracts pet ownership and weight creation for tests (mocked instead of DB).
type WeightCreateStore interface {
	OwnsPet(userID, petID uuid.UUID) bool
	Create(entry *models.WeightEntry) error
}

type WeightsHandler struct {
	DB               *gorm.DB
	WeightCreateStore WeightCreateStore // when non-nil, Create uses this instead of DB
}

func (h *WeightsHandler) ensurePetOwnership(r *http.Request, petID uuid.UUID) bool {
	u := middleware.GetUser(r.Context())
	if u == nil {
		return false
	}
	var count int64
	h.DB.Model(&models.Pet{}).Where("id = ? AND user_id = ?", petID, u.ID).Count(&count)
	return count > 0
}

func (h *WeightsHandler) List(w http.ResponseWriter, r *http.Request) {
	u := middleware.GetUser(r.Context())
	if u == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	petID, err := uuid.Parse(vars["petId"])
	if err != nil {
		http.Error(w, `{"error":"invalid pet id"}`, http.StatusBadRequest)
		return
	}
	if !h.ensurePetOwnership(r, petID) {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	var list []models.WeightEntry
	err = h.DB.Where("pet_id = ?", petID).Order("measured_at DESC").Find(&list).Error
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []models.WeightEntry{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *WeightsHandler) createOwnsPet(r *http.Request, petID uuid.UUID) bool {
	if h.WeightCreateStore != nil {
		u := middleware.GetUser(r.Context())
		if u == nil {
			return false
		}
		return h.WeightCreateStore.OwnsPet(u.ID, petID)
	}
	return h.ensurePetOwnership(r, petID)
}

func (h *WeightsHandler) Create(w http.ResponseWriter, r *http.Request) {
	u := middleware.GetUser(r.Context())
	if u == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	petID, err := uuid.Parse(vars["petId"])
	if err != nil {
		http.Error(w, `{"error":"invalid pet id"}`, http.StatusBadRequest)
		return
	}
	if !h.createOwnsPet(r, petID) {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	var body struct {
		WeightLbs   float64 `json:"weight_lbs"`
		WeightKg    *float64 `json:"weight_kg"`
		EntryUnit   string   `json:"entry_unit"`
		MeasuredAt  string   `json:"measured_at"`
		Approximate bool     `json:"approximate"`
		Notes       *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	if body.MeasuredAt == "" {
		http.Error(w, `{"error":"measured_at required"}`, http.StatusBadRequest)
		return
	}
	entryUnit := body.EntryUnit
	if entryUnit != "kg" && entryUnit != "lbs" {
		entryUnit = "lbs"
	}
	weightLbs := body.WeightLbs
	if body.WeightKg != nil {
		weightLbs = *body.WeightKg * 2.20462
		entryUnit = "kg"
	}
	entry := models.WeightEntry{
		PetID:       petID,
		WeightLbs:   weightLbs,
		EntryUnit:   entryUnit,
		MeasuredAt:  body.MeasuredAt,
		Approximate: body.Approximate,
		Notes:       body.Notes,
	}
	if h.WeightCreateStore != nil {
		if err := h.WeightCreateStore.Create(&entry); err != nil {
			http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
			return
		}
	} else if err := h.DB.Create(&entry).Error; err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(entry)
}

func (h *WeightsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	u := middleware.GetUser(r.Context())
	if u == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	petID, _ := uuid.Parse(vars["petId"])
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	if !h.ensurePetOwnership(r, petID) {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	result := h.DB.Where("id = ? AND pet_id = ?", id, petID).Delete(&models.WeightEntry{})
	if result.Error != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	if result.RowsAffected == 0 {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
