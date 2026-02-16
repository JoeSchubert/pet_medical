package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pet-medical/api/internal/middleware"
	"github.com/pet-medical/api/internal/models"
	"gorm.io/gorm"
)

type mockDocumentUpdateStore struct {
	ownsPet  bool
	docs     map[string]*models.Document
}

func newMockDocumentUpdateStore(ownsPet bool) *mockDocumentUpdateStore {
	return &mockDocumentUpdateStore{ownsPet: ownsPet, docs: make(map[string]*models.Document)}
}

func (m *mockDocumentUpdateStore) OwnsPet(userID, petID uuid.UUID) bool {
	return m.ownsPet
}

func (m *mockDocumentUpdateStore) UpdateName(petID, docID uuid.UUID, name string) (*models.Document, error) {
	key := petID.String() + "/" + docID.String()
	doc, ok := m.docs[key]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	doc.Name = name
	return doc, nil
}

func TestDocuments_Update_EmptyNameReturns400(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	docID := uuid.New()
	mock := newMockDocumentUpdateStore(true)
	mock.docs[petID.String()+"/"+docID.String()] = &models.Document{ID: docID, PetID: petID, Name: "old", FilePath: "/x", CreatedAt: time.Now()}
	h := &DocumentsHandler{DocumentUpdateStore: mock}
	body := bytes.NewBufferString(`{"name":""}`)
	req := httptest.NewRequest(http.MethodPatch, "/pets/"+petID.String()+"/documents/"+docID.String(), body)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.ContextWithUser(req.Context(), &middleware.UserInfo{ID: userID, DisplayName: "u", Role: "user"}))
	req = mux.SetURLVars(req, map[string]string{"petId": petID.String(), "id": docID.String()})
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when name empty, got %d", rec.Code)
	}
	var out map[string]string
	json.NewDecoder(rec.Body).Decode(&out)
	if out["error"] != "name required" {
		t.Errorf("expected error name required, got %v", out)
	}
}

func TestDocuments_Update_ValidNameReturns200(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	docID := uuid.New()
	mock := newMockDocumentUpdateStore(true)
	doc := &models.Document{ID: docID, PetID: petID, Name: "old", FilePath: "/x", CreatedAt: time.Now()}
	mock.docs[petID.String()+"/"+docID.String()] = doc
	h := &DocumentsHandler{DocumentUpdateStore: mock}
	body := bytes.NewBufferString(`{"name":"New Name"}`)
	req := httptest.NewRequest(http.MethodPut, "/pets/"+petID.String()+"/documents/"+docID.String(), body)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.ContextWithUser(req.Context(), &middleware.UserInfo{ID: userID, DisplayName: "u", Role: "user"}))
	req = mux.SetURLVars(req, map[string]string{"petId": petID.String(), "id": docID.String()})
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var out models.Document
	json.NewDecoder(rec.Body).Decode(&out)
	if out.Name != "New Name" {
		t.Errorf("expected name New Name, got %q", out.Name)
	}
}

func TestDocuments_Update_NoOwnershipReturns404(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	docID := uuid.New()
	mock := newMockDocumentUpdateStore(false)
	h := &DocumentsHandler{DocumentUpdateStore: mock}
	body := bytes.NewBufferString(`{"name":"x"}`)
	req := httptest.NewRequest(http.MethodPut, "/pets/"+petID.String()+"/documents/"+docID.String(), body)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.ContextWithUser(req.Context(), &middleware.UserInfo{ID: userID, DisplayName: "u", Role: "user"}))
	req = mux.SetURLVars(req, map[string]string{"petId": petID.String(), "id": docID.String()})
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 when not owner, got %d", rec.Code)
	}
}
