package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pet-medical/api/internal/auth"
	"github.com/pet-medical/api/internal/middleware"
	"github.com/pet-medical/api/internal/models"
	"gorm.io/gorm"
)

// UserRoleStore abstracts user lookup and role updates for UpdateRole (testable with a mock).
type UserRoleStore interface {
	GetUserByID(id uuid.UUID) (*models.User, error)
	CountAdmins() (int64, error)
	UpdateRole(id uuid.UUID, role string) error
}

type UsersHandler struct {
	DB                *gorm.DB
	UserRoleStore     UserRoleStore // if nil, uses DB via default impl
	DefaultWeightUnit string
	DefaultCurrency   string
	DefaultLanguage   string
}

func (h *UsersHandler) userRoleStore() UserRoleStore {
	if h.UserRoleStore != nil {
		return h.UserRoleStore
	}
	return &gormUserRoleStore{db: h.DB}
}

type gormUserRoleStore struct{ db *gorm.DB }

func (s *gormUserRoleStore) GetUserByID(id uuid.UUID) (*models.User, error) {
	var u models.User
	err := s.db.Where("id = ?", id).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *gormUserRoleStore) CountAdmins() (int64, error) {
	var n int64
	err := s.db.Model(&models.User{}).Where("role = ?", "admin").Count(&n).Error
	return n, err
}

func (s *gormUserRoleStore) UpdateRole(id uuid.UUID, role string) error {
	result := s.db.Model(&models.User{}).Where("id = ?", id).Update("role", role)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

type UserListDTO struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	CreatedAt   string `json:"created_at"`
}

func (h *UsersHandler) List(w http.ResponseWriter, r *http.Request) {
	_ = middleware.GetUser(r.Context())
	var users []models.User
	err := h.DB.Order("display_name").Find(&users).Error
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	if users == nil {
		users = []models.User{}
	}
	out := make([]UserListDTO, len(users))
	for i, u := range users {
		out[i] = UserListDTO{
			ID:          u.ID.String(),
			DisplayName: u.DisplayName,
			Email:       u.Email,
			Role:        u.Role,
			CreatedAt:   u.CreatedAt.Format(time.RFC3339),
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (h *UsersHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	if body.Role != "admin" && body.Role != "user" {
		http.Error(w, `{"error":"role must be admin or user"}`, http.StatusBadRequest)
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, `{"error":"id required"}`, http.StatusBadRequest)
		return
	}
	userID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	store := h.userRoleStore()
	target, err := store.GetUserByID(userID)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	if target.Role == "admin" && body.Role != "admin" {
		adminCount, err := store.CountAdmins()
		if err != nil {
			http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
			return
		}
		if adminCount <= 1 {
			http.Error(w, `{"error":"cannot remove the only admin"}`, http.StatusBadRequest)
			return
		}
	}
	if err := store.UpdateRole(userID, body.Role); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "updated"})
}

func (h *UsersHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		DisplayName string `json:"display_name"`
		Email      string `json:"email"`
		Password   string `json:"password"`
		Role       string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	if body.DisplayName == "" || body.Password == "" {
		http.Error(w, `{"error":"display name and password required"}`, http.StatusBadRequest)
		return
	}
	body.Email = strings.TrimSpace(body.Email)
	if body.Email == "" {
		http.Error(w, `{"error":"email required"}`, http.StatusBadRequest)
		return
	}
	if body.Role != "admin" && body.Role != "user" {
		body.Role = "user"
	}
	hash, err := auth.HashPassword(body.Password)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	u := models.User{
		DisplayName:  body.DisplayName,
		Email:        body.Email,
		PasswordHash: hash,
		Role:         body.Role,
		WeightUnit:   h.DefaultWeightUnit,
		Currency:     h.DefaultCurrency,
		Language:     h.DefaultLanguage,
	}
	if err := h.DB.Create(&u).Error; err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			if strings.Contains(err.Error(), "email") {
				http.Error(w, `{"error":"email already in use"}`, http.StatusConflict)
			} else {
				http.Error(w, `{"error":"display name already exists"}`, http.StatusConflict)
			}
			return
		}
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":           u.ID.String(),
		"display_name": body.DisplayName,
		"email":        body.Email,
		"role":         body.Role,
	})
}
