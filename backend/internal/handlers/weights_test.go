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
)

type mockWeightCreateStore struct {
	petOwner bool
	lastEntry *models.WeightEntry
}

func (m *mockWeightCreateStore) OwnsPet(userID, petID uuid.UUID) bool {
	return m.petOwner
}

func (m *mockWeightCreateStore) Create(entry *models.WeightEntry) error {
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	entry.CreatedAt = time.Now()
	m.lastEntry = entry
	return nil
}

func TestWeights_Create_MeasuredAtRequired(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	mock := &mockWeightCreateStore{petOwner: true}
	h := &WeightsHandler{WeightCreateStore: mock}
	body := bytes.NewBufferString(`{"weight_lbs":10,"entry_unit":"lbs"}`)
	req := httptest.NewRequest(http.MethodPost, "/pets/"+petID.String()+"/weights", body)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.ContextWithUser(req.Context(), &middleware.UserInfo{ID: userID, DisplayName: "u", Role: "user"}))
	req = mux.SetURLVars(req, map[string]string{"petId": petID.String()})
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when measured_at missing, got %d", rec.Code)
	}
	var out map[string]string
	json.NewDecoder(rec.Body).Decode(&out)
	if out["error"] != "measured_at required" {
		t.Errorf("expected error measured_at required, got %v", out)
	}
}

func TestWeights_Create_ApproximateStored(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	mock := &mockWeightCreateStore{petOwner: true}
	h := &WeightsHandler{WeightCreateStore: mock}
	body := bytes.NewBufferString(`{"weight_lbs":12.5,"entry_unit":"lbs","measured_at":"2025-01-15","approximate":true}`)
	req := httptest.NewRequest(http.MethodPost, "/pets/"+petID.String()+"/weights", body)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.ContextWithUser(req.Context(), &middleware.UserInfo{ID: userID, DisplayName: "u", Role: "user"}))
	req = mux.SetURLVars(req, map[string]string{"petId": petID.String()})
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var entry models.WeightEntry
	json.NewDecoder(rec.Body).Decode(&entry)
	if !entry.Approximate {
		t.Error("expected approximate true in response")
	}
	if mock.lastEntry == nil || !mock.lastEntry.Approximate {
		t.Error("expected mock to receive entry with approximate true")
	}
}

func TestWeights_Create_NoOwnershipReturns404(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	mock := &mockWeightCreateStore{petOwner: false}
	h := &WeightsHandler{WeightCreateStore: mock}
	body := bytes.NewBufferString(`{"weight_lbs":10,"entry_unit":"lbs","measured_at":"2025-01-15"}`)
	req := httptest.NewRequest(http.MethodPost, "/pets/"+petID.String()+"/weights", body)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.ContextWithUser(req.Context(), &middleware.UserInfo{ID: userID, DisplayName: "u", Role: "user"}))
	req = mux.SetURLVars(req, map[string]string{"petId": petID.String()})
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 when not owner, got %d", rec.Code)
	}
}
