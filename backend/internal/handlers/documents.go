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
	"github.com/pet-medical/api/internal/extract"
	"github.com/pet-medical/api/internal/middleware"
	"github.com/pet-medical/api/internal/models"
	"github.com/pet-medical/api/internal/upload"
	"gorm.io/gorm"
)

// DocumentUpdateStore abstracts pet ownership and document name update for tests.
type DocumentUpdateStore interface {
	OwnsPet(userID, petID uuid.UUID) bool
	UpdateName(petID, docID uuid.UUID, name string) (*models.Document, error)
}

type DocumentsHandler struct {
	DB                  *gorm.DB
	UploadDir           string
	MaxDocumentBytes    int64 // max upload size; 0 = use default 25MB
	DocumentUpdateStore DocumentUpdateStore // when non-nil, Update uses this instead of DB
}

func (h *DocumentsHandler) ensurePetOwnership(r *http.Request, petID uuid.UUID) bool {
	u := middleware.GetUser(r.Context())
	if u == nil {
		return false
	}
	var count int64
	h.DB.Model(&models.Pet{}).Where("id = ? AND user_id = ?", petID, u.ID).Count(&count)
	return count > 0
}

func (h *DocumentsHandler) List(w http.ResponseWriter, r *http.Request) {
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
	debuglog.Debugf("documents list: pet_id=%s", petID)
	sortBy := strings.ToLower(r.URL.Query().Get("sort"))
	if sortBy != "name" && sortBy != "date" {
		sortBy = "date"
	}
	search := strings.TrimSpace(r.URL.Query().Get("search"))

	q := h.DB.Where("pet_id = ?", petID)
	if search != "" {
		pattern := "%" + search + "%"
		q = q.Where("name ILIKE ? OR (extracted_text IS NOT NULL AND extracted_text ILIKE ?)", pattern, pattern)
	}
	if sortBy == "name" {
		q = q.Order("name ASC")
	} else {
		q = q.Order("created_at DESC")
	}
	var list []models.Document
	err = q.Find(&list).Error
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []models.Document{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *DocumentsHandler) Get(w http.ResponseWriter, r *http.Request) {
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
	var doc models.Document
	err = h.DB.Where("id = ? AND pet_id = ?", id, petID).First(&doc).Error
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

func (h *DocumentsHandler) Create(w http.ResponseWriter, r *http.Request) {
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
	maxBytes := h.MaxDocumentBytes
	if maxBytes <= 0 {
		maxBytes = 25 * 1024 * 1024 // 25 MB default
	}
	if err := r.ParseMultipartForm(maxBytes + 1024); err != nil {
		http.Error(w, `{"error":"invalid multipart"}`, http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
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
	if !upload.AllowedDocument(headerBytes) {
		http.Error(w, `{"error":"invalid file type: only PDF, Word, RTF, and PNG/JPEG images are allowed"}`, http.StatusBadRequest)
		return
	}
	safeName := upload.SafeBasename(header.Filename)
	if name == "" {
		name = safeName
	}
	relPath := filepath.Join("documents", petID.String(), uuid.New().String()+"_"+safeName)
	relPath = filepath.ToSlash(relPath)
	absPath := filepath.Join(h.UploadDir, relPath)
	remaining := maxBytes - int64(len(headerBytes))
	if remaining < 0 {
		remaining = 0
	}
	body := io.LimitReader(file, remaining)
	if err := saveUpload(io.MultiReader(bytes.NewReader(headerBytes), body), absPath); err != nil {
		http.Error(w, `{"error":"save failed"}`, http.StatusInternalServerError)
		return
	}

	doc := models.Document{
		PetID:    petID,
		Name:     name,
		FilePath: relPath,
		FileSize: &header.Size,
	}
	if header.Header.Get("Content-Type") != "" {
		doc.MimeType = &header.Header["Content-Type"][0]
	}
	if err := h.DB.Create(&doc).Error; err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	// Extract text in background for full-text search (OCR for images, text from PDF/DOCX/RTF).
	go func(docID uuid.UUID, absPath string) {
		text, err := extract.ExtractText(absPath)
		if err != nil {
			debuglog.Debugf("document extract text: %v", err)
			return
		}
		if text == "" {
			return
		}
		if err := h.DB.Model(&models.Document{}).Where("id = ?", docID).Update("extracted_text", text).Error; err != nil {
			debuglog.Debugf("document update extracted_text: %v", err)
		}
	}(doc.ID, absPath)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(doc)
}

func saveUpload(r io.Reader, destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}
	out, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return os.WriteFile(destPath, out, 0644)
}

func (h *DocumentsHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
	var doc models.Document
	if h.DB.Where("id = ? AND pet_id = ?", id, petID).First(&doc).Error == nil && h.UploadDir != "" && doc.FilePath != "" {
		absPath := filepath.Join(h.UploadDir, filepath.FromSlash(doc.FilePath))
		_ = os.Remove(absPath)
	}
	result := h.DB.Where("id = ? AND pet_id = ?", id, petID).Delete(&models.Document{})
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

func (h *DocumentsHandler) updateOwnsPet(r *http.Request, petID uuid.UUID) bool {
	if h.DocumentUpdateStore != nil {
		u := middleware.GetUser(r.Context())
		if u == nil {
			return false
		}
		return h.DocumentUpdateStore.OwnsPet(u.ID, petID)
	}
	return h.ensurePetOwnership(r, petID)
}

func (h *DocumentsHandler) Update(w http.ResponseWriter, r *http.Request) {
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
	if !h.updateOwnsPet(r, petID) {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		http.Error(w, `{"error":"name required"}`, http.StatusBadRequest)
		return
	}
	if h.DocumentUpdateStore != nil {
		doc, err := h.DocumentUpdateStore.UpdateName(petID, id, name)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(doc)
		return
	}
	result := h.DB.Model(&models.Document{}).Where("id = ? AND pet_id = ?", id, petID).Update("name", name)
	if result.Error != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	if result.RowsAffected == 0 {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	var doc models.Document
	if err := h.DB.Where("id = ? AND pet_id = ?", id, petID).First(&doc).Error; err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}
