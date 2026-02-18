package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pet-medical/api/internal/debuglog"
	"github.com/pet-medical/api/internal/middleware"
	"github.com/pet-medical/api/internal/models"
	"github.com/pet-medical/api/internal/upload"
	"gorm.io/gorm"
)

type PhotosHandler struct {
	DB             *gorm.DB
	UploadDir      string
	MaxPhotoBytes  int64 // max upload size; 0 = use default 10MB
}

func (h *PhotosHandler) ensurePetOwnership(r *http.Request, petID uuid.UUID) bool {
	u := middleware.GetUser(r.Context())
	if u == nil {
		return false
	}
	var count int64
	h.DB.Model(&models.Pet{}).Where("id = ? AND user_id = ?", petID, u.ID).Count(&count)
	return count > 0
}

func (h *PhotosHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
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
	debuglog.Debugf("photos list: pet_id=%s", petID)
	var list []models.PetPhoto
	err = h.DB.Where("pet_id = ?", petID).Order("display_order ASC, created_at ASC").Find(&list).Error
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []models.PetPhoto{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *PhotosHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
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
	maxBytes := h.MaxPhotoBytes
	if maxBytes <= 0 {
		maxBytes = 10 * 1024 * 1024 // 10 MB default
	}
	if err := r.ParseMultipartForm(maxBytes + 1024); err != nil {
		http.Error(w, `{"error":"invalid multipart"}`, http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, `{"error":"file required"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()
	if header.Size > 0 && header.Size > maxBytes {
		http.Error(w, `{"error":"error.upload_too_large"}`, http.StatusRequestEntityTooLarge)
		return
	}

	headerBuf := make([]byte, upload.MaxHeaderBytes)
	n, _ := io.ReadFull(file, headerBuf)
	headerBytes := headerBuf[:n]
	if !upload.AllowedImage(headerBytes) {
		http.Error(w, `{"error":"invalid file type: only JPEG, PNG, GIF, and WebP images are allowed"}`, http.StatusBadRequest)
		return
	}
	ext := strings.ToLower(filepath.Ext(upload.SafeBasename(header.Filename)))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		// keep
	default:
		ext = ".jpg"
	}
	relPath := filepath.Join("photos", petID.String(), uuid.New().String()+ext)
	relPath = filepath.ToSlash(relPath)
	absPath := filepath.Join(h.UploadDir, relPath)
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		http.Error(w, `{"error":"save failed"}`, http.StatusInternalServerError)
		return
	}
	remaining := maxBytes - int64(len(headerBytes))
	if remaining < 0 {
		remaining = 0
	}
	out, err := io.ReadAll(io.MultiReader(bytes.NewReader(headerBytes), io.LimitReader(file, remaining)))
	if err != nil {
		http.Error(w, `{"error":"save failed"}`, http.StatusInternalServerError)
		return
	}
	if err := os.WriteFile(absPath, out, 0644); err != nil {
		http.Error(w, `{"error":"save failed"}`, http.StatusInternalServerError)
		return
	}

	var maxOrder int
	h.DB.Raw("SELECT COALESCE(MAX(display_order), 0) FROM pet_photos WHERE pet_id = ?", petID).Scan(&maxOrder)
	photo := models.PetPhoto{PetID: petID, FilePath: relPath, DisplayOrder: maxOrder + 1}
	if err := h.DB.Create(&photo).Error; err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(photo)
}

func (h *PhotosHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
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
	var photo models.PetPhoto
	err = h.DB.Where("id = ? AND pet_id = ?", id, petID).First(&photo).Error
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	absPath := filepath.Join(h.UploadDir, photo.FilePath)
	_ = os.Remove(absPath)
	result := h.DB.Where("id = ? AND pet_id = ?", id, petID).Delete(&models.PetPhoto{})
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

func (h *PhotosHandler) SetAvatar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
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
	var photo models.PetPhoto
	err = h.DB.Where("id = ? AND pet_id = ?", id, petID).First(&photo).Error
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	avatarURL := "/api/uploads/" + photo.FilePath
	h.DB.Model(&models.Pet{}).Where("id = ?", petID).Update("photo_url", avatarURL)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"photo_url": avatarURL})
}
