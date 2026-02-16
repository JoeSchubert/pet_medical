package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pet-medical/api/internal/middleware"
	"github.com/pet-medical/api/internal/models"
	"gorm.io/gorm"
)

type SettingsHandler struct {
	DB                *gorm.DB
	DefaultWeightUnit string
	DefaultCurrency   string
	DefaultLanguage   string
}

type SettingsDTO struct {
	WeightUnit  string `json:"weight_unit"`
	Currency    string `json:"currency"`
	Language    string `json:"language"`
	Email       string `json:"email,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Role        string `json:"role,omitempty"`
	IsOnlyAdmin bool   `json:"is_only_admin,omitempty"`
}

func (h *SettingsHandler) GetMine(w http.ResponseWriter, r *http.Request) {
	u := middleware.GetUser(r.Context())
	if u == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	var dbUser models.User
	if err := h.DB.Where("id = ?", u.ID).First(&dbUser).Error; err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	weightUnit, currency, language := dbUser.WeightUnit, dbUser.Currency, dbUser.Language
	if weightUnit != "lbs" && weightUnit != "kg" {
		weightUnit = h.DefaultWeightUnit
	}
	if currency == "" {
		currency = h.DefaultCurrency
	}
	if language == "" {
		language = h.DefaultLanguage
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SettingsDTO{WeightUnit: weightUnit, Currency: currency, Language: language})
}

func (h *SettingsHandler) UpdateMine(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	u := middleware.GetUser(r.Context())
	if u == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	var body SettingsDTO
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	body.WeightUnit = strings.TrimSpace(strings.ToLower(body.WeightUnit))
	if body.WeightUnit != "lbs" && body.WeightUnit != "kg" {
		body.WeightUnit = h.DefaultWeightUnit
	}
	body.Currency = strings.TrimSpace(strings.ToUpper(body.Currency))
	if body.Currency == "" {
		body.Currency = h.DefaultCurrency
	}
	body.Language = strings.TrimSpace(strings.ToLower(body.Language))
	if body.Language == "" {
		body.Language = h.DefaultLanguage
	}
	result := h.DB.Model(&models.User{}).Where("id = ?", u.ID).Updates(map[string]interface{}{
		"weight_unit": body.WeightUnit, "currency": body.Currency, "language": body.Language,
	})
	if result.Error != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SettingsDTO{WeightUnit: body.WeightUnit, Currency: body.Currency, Language: body.Language})
}

func (h *SettingsHandler) GetForUser(w http.ResponseWriter, r *http.Request) {
	admin := middleware.GetUser(r.Context())
	if admin == nil || admin.Role != "admin" {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	}
	vars := mux.Vars(r)
	userID := vars["id"]
	if userID == "" {
		http.Error(w, `{"error":"id required"}`, http.StatusBadRequest)
		return
	}
	var dbUser models.User
	if err := h.DB.Where("id = ?", userID).First(&dbUser).Error; err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	weightUnit, currency, language := dbUser.WeightUnit, dbUser.Currency, dbUser.Language
	if weightUnit != "lbs" && weightUnit != "kg" {
		weightUnit = h.DefaultWeightUnit
	}
	if currency == "" {
		currency = h.DefaultCurrency
	}
	if language == "" {
		language = h.DefaultLanguage
	}
	email, role, displayName := dbUser.Email, dbUser.Role, dbUser.DisplayName
	if role != "admin" && role != "user" {
		role = "user"
	}
	var adminCount int64
	h.DB.Model(&models.User{}).Where("role = ?", "admin").Count(&adminCount)
	isOnlyAdmin := role == "admin" && adminCount <= 1
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SettingsDTO{WeightUnit: weightUnit, Currency: currency, Language: language, Email: email, DisplayName: displayName, Role: role, IsOnlyAdmin: isOnlyAdmin})
}

func (h *SettingsHandler) UpdateForUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	admin := middleware.GetUser(r.Context())
	if admin == nil || admin.Role != "admin" {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	}
	vars := mux.Vars(r)
	userID := vars["id"]
	if userID == "" {
		http.Error(w, `{"error":"id required"}`, http.StatusBadRequest)
		return
	}
	var body SettingsDTO
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	body.WeightUnit = strings.TrimSpace(strings.ToLower(body.WeightUnit))
	if body.WeightUnit != "lbs" && body.WeightUnit != "kg" {
		body.WeightUnit = h.DefaultWeightUnit
	}
	body.Currency = strings.TrimSpace(strings.ToUpper(body.Currency))
	if body.Currency == "" {
		body.Currency = h.DefaultCurrency
	}
	body.Language = strings.TrimSpace(strings.ToLower(body.Language))
	if body.Language == "" {
		body.Language = h.DefaultLanguage
	}
	body.Email = strings.TrimSpace(body.Email)
	if body.Email == "" {
		http.Error(w, `{"error":"email required"}`, http.StatusBadRequest)
		return
	}
	body.Role = strings.TrimSpace(strings.ToLower(body.Role))
	if body.Role != "admin" && body.Role != "user" {
		body.Role = "user"
	}
	var existing models.User
	if err := h.DB.Where("id = ?", userID).First(&existing).Error; err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	if body.Email != existing.Email {
		var count int64
		h.DB.Model(&models.User{}).Where("email = ? AND id != ?", body.Email, userID).Count(&count)
		if count > 0 {
			http.Error(w, `{"error":"email already in use"}`, http.StatusConflict)
			return
		}
	}
	result := h.DB.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"weight_unit": body.WeightUnit, "currency": body.Currency, "language": body.Language,
		"email": body.Email, "role": body.Role,
	})
	if result.Error != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	if result.RowsAffected == 0 {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SettingsDTO{WeightUnit: body.WeightUnit, Currency: body.Currency, Language: body.Language, Email: body.Email, Role: body.Role})
}
