package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"github.com/pet-medical/api/internal/models"
	"gorm.io/gorm"
)

// DefaultOptionsHandler provides admin CRUD for default dropdown options (species, breeds, vaccinations).
// All operations use GORM for parameterized queries.
type DefaultOptionsHandler struct {
	GORM *gorm.DB
}

// DefaultOptionItem is one row for list/create/update.
type DefaultOptionItem struct {
	ID             string  `json:"id"`
	OptionType     string  `json:"option_type"`
	Value          string  `json:"value"`
	Context        string  `json:"context"`
	SortOrder      int     `json:"sort_order"`
	DurationMonths *int    `json:"duration_months,omitempty"`
}

func (h *DefaultOptionsHandler) List(w http.ResponseWriter, r *http.Request) {
	var list []models.DefaultDropdownOption
	if err := h.GORM.Order("option_type, context, sort_order, value").Find(&list).Error; err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	out := make([]DefaultOptionItem, 0, len(list))
	for _, row := range list {
		item := DefaultOptionItem{
			ID:         row.ID.String(),
			OptionType: row.OptionType,
			Value:      row.Value,
			Context:    row.Context,
			SortOrder:  row.SortOrder,
		}
		if row.DurationMonths != nil {
			item.DurationMonths = row.DurationMonths
		}
		out = append(out, item)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (h *DefaultOptionsHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var body DefaultOptionItem
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	body.OptionType = strings.TrimSpace(strings.ToLower(body.OptionType))
	body.Value = strings.TrimSpace(body.Value)
	body.Context = strings.TrimSpace(body.Context)
	if body.Value == "" {
		http.Error(w, `{"error":"value required"}`, http.StatusBadRequest)
		return
	}
	if body.OptionType != "species" && body.OptionType != "breed" && body.OptionType != "vaccination" {
		http.Error(w, `{"error":"invalid option_type"}`, http.StatusBadRequest)
		return
	}
	opt := models.DefaultDropdownOption{
		OptionType:     body.OptionType,
		Value:          body.Value,
		Context:        body.Context,
		SortOrder:      body.SortOrder,
		DurationMonths: body.DurationMonths,
	}
	if err := h.GORM.Create(&opt).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			http.Error(w, `{"error":"already exists"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(DefaultOptionItem{
		ID:             opt.ID.String(),
		OptionType:     opt.OptionType,
		Value:          opt.Value,
		Context:        opt.Context,
		SortOrder:      opt.SortOrder,
		DurationMonths: opt.DurationMonths,
	})
}

func (h *DefaultOptionsHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	var body DefaultOptionItem
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	body.Value = strings.TrimSpace(body.Value)
	body.Context = strings.TrimSpace(body.Context)
	if body.Value == "" {
		http.Error(w, `{"error":"value required"}`, http.StatusBadRequest)
		return
	}
	var opt models.DefaultDropdownOption
	if err := h.GORM.First(&opt, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	opt.Value = body.Value
	opt.Context = body.Context
	opt.SortOrder = body.SortOrder
	opt.DurationMonths = body.DurationMonths
	if err := h.GORM.Save(&opt).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			http.Error(w, `{"error":"already exists"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(DefaultOptionItem{
		ID:             opt.ID.String(),
		OptionType:     opt.OptionType,
		Value:          opt.Value,
		Context:        opt.Context,
		SortOrder:      opt.SortOrder,
		DurationMonths: opt.DurationMonths,
	})
}

func (h *DefaultOptionsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	res := h.GORM.Delete(&models.DefaultDropdownOption{}, "id = ?", id)
	if res.Error != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	if res.RowsAffected == 0 {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
