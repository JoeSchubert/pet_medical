package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pet-medical/api/internal/middleware"
	"github.com/pet-medical/api/internal/models"
	"gorm.io/gorm"
)

type VaccinationsHandler struct {
	DB *gorm.DB
}

func (h *VaccinationsHandler) ensurePetOwnership(r *http.Request, petID uuid.UUID) bool {
	u := middleware.GetUser(r.Context())
	if u == nil {
		return false
	}
	var count int64
	h.DB.Model(&models.Pet{}).Where("id = ? AND user_id = ?", petID, u.ID).Count(&count)
	return count > 0
}

func (h *VaccinationsHandler) List(w http.ResponseWriter, r *http.Request) {
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
	var list []models.Vaccination
	err = h.DB.Where("pet_id = ?", petID).Order("administered_at DESC").Find(&list).Error
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []models.Vaccination{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *VaccinationsHandler) Get(w http.ResponseWriter, r *http.Request) {
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
	var v models.Vaccination
	err = h.DB.Where("id = ? AND pet_id = ?", id, petID).First(&v).Error
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func (h *VaccinationsHandler) Create(w http.ResponseWriter, r *http.Request) {
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
	var v models.Vaccination
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	v.PetID = petID
	v.ID = uuid.Nil
	if v.Name == "" || v.AdministeredAt == "" {
		http.Error(w, `{"error":"name and administered_at required"}`, http.StatusBadRequest)
		return
	}
	if err := h.DB.Create(&v).Error; err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(v)
}

func (h *VaccinationsHandler) Update(w http.ResponseWriter, r *http.Request) {
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
	var v models.Vaccination
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	v.ID = id
	v.PetID = petID
	if v.Name == "" || v.AdministeredAt == "" {
		http.Error(w, `{"error":"name and administered_at required"}`, http.StatusBadRequest)
		return
	}
	result := h.DB.Model(&models.Vaccination{}).Where("id = ? AND pet_id = ?", id, petID).Updates(map[string]interface{}{
		"name": v.Name, "administered_at": v.AdministeredAt, "next_due": v.NextDue, "cost_usd": v.CostUSD,
		"veterinarian": v.Veterinarian, "batch_number": v.BatchNumber, "notes": v.Notes,
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
	h.DB.Where("id = ?", id).First(&v)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func (h *VaccinationsHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
	result := h.DB.Where("id = ? AND pet_id = ?", id, petID).Delete(&models.Vaccination{})
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
