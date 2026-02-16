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
	"github.com/pet-medical/api/internal/auth"
	"github.com/pet-medical/api/internal/middleware"
	"github.com/pet-medical/api/internal/models"
	"gorm.io/gorm"
)

// mockUserRoleStore is an in-memory mock that mimics PostgreSQL-backed user/role behavior.
type mockUserRoleStore struct {
	users map[uuid.UUID]*models.User
}

func newMockUserRoleStore(users []models.User) *mockUserRoleStore {
	m := &mockUserRoleStore{users: make(map[uuid.UUID]*models.User)}
	for i := range users {
		u := &users[i]
		m.users[u.ID] = u
	}
	return m
}

func (m *mockUserRoleStore) GetUserByID(id uuid.UUID) (*models.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *u
	return &cp, nil
}

func (m *mockUserRoleStore) CountAdmins() (int64, error) {
	var n int64
	for _, u := range m.users {
		if u.Role == "admin" {
			n++
		}
	}
	return n, nil
}

func (m *mockUserRoleStore) UpdateRole(id uuid.UUID, role string) error {
	u, ok := m.users[id]
	if !ok {
		return nil
	}
	u.Role = role
	return nil
}

func TestUpdateRole_LastAdminCannotBeDemoted(t *testing.T) {
	hash, _ := auth.HashPassword("pass")
	adminID := uuid.New()
	users := []models.User{{
		ID: adminID, DisplayName: "admin", Email: "admin@test.com", PasswordHash: hash, Role: "admin",
		WeightUnit: "lbs", Currency: "USD", Language: "en",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}}
	mock := newMockUserRoleStore(users)

	h := &UsersHandler{UserRoleStore: mock}
	body := bytes.NewBufferString(`{"role":"user"}`)
	req := httptest.NewRequest(http.MethodPut, "/users/"+adminID.String()+"/role", body)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.ContextWithUser(req.Context(), &middleware.UserInfo{ID: adminID, DisplayName: "admin", Role: "admin"}))
	req = mux.SetURLVars(req, map[string]string{"id": adminID.String()})
	rec := httptest.NewRecorder()
	h.UpdateRole(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when demoting only admin, got %d", rec.Code)
	}
	var out map[string]string
	json.NewDecoder(rec.Body).Decode(&out)
	if out["error"] != "cannot remove the only admin" {
		t.Errorf("expected error message, got %v", out)
	}
}

func TestUpdateRole_SecondAdminCanBeDemoted(t *testing.T) {
	hash, _ := auth.HashPassword("pass")
	admin1 := uuid.New()
	admin2 := uuid.New()
	users := []models.User{
		{ID: admin1, DisplayName: "a1", Email: "a1@test.com", PasswordHash: hash, Role: "admin", WeightUnit: "lbs", Currency: "USD", Language: "en", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: admin2, DisplayName: "a2", Email: "a2@test.com", PasswordHash: hash, Role: "admin", WeightUnit: "lbs", Currency: "USD", Language: "en", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	mock := newMockUserRoleStore(users)

	h := &UsersHandler{UserRoleStore: mock}
	body := bytes.NewBufferString(`{"role":"user"}`)
	req := httptest.NewRequest(http.MethodPut, "/users/"+admin2.String()+"/role", body)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.ContextWithUser(req.Context(), &middleware.UserInfo{ID: admin1, DisplayName: "a1", Role: "admin"}))
	req = mux.SetURLVars(req, map[string]string{"id": admin2.String()})
	rec := httptest.NewRecorder()
	h.UpdateRole(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 when demoting second admin, got %d: %s", rec.Code, rec.Body.String())
	}
	// Mock should have updated in memory
	u, _ := mock.GetUserByID(admin2)
	if u != nil && u.Role != "user" {
		t.Errorf("expected role to be user, got %q", u.Role)
	}
}

func TestUpdateRole_InvalidRoleReturns400(t *testing.T) {
	hash, _ := auth.HashPassword("pass")
	adminID := uuid.New()
	users := []models.User{{
		ID: adminID, DisplayName: "admin", Email: "admin@test.com", PasswordHash: hash, Role: "admin",
		WeightUnit: "lbs", Currency: "USD", Language: "en",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}}
	mock := newMockUserRoleStore(users)

	h := &UsersHandler{UserRoleStore: mock}
	body := bytes.NewBufferString(`{"role":"superuser"}`)
	req := httptest.NewRequest(http.MethodPut, "/users/"+adminID.String()+"/role", body)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.ContextWithUser(req.Context(), &middleware.UserInfo{ID: adminID, DisplayName: "admin", Role: "admin"}))
	req = mux.SetURLVars(req, map[string]string{"id": adminID.String()})
	rec := httptest.NewRecorder()
	h.UpdateRole(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid role, got %d", rec.Code)
	}
}

func TestCreate_EmailRequired(t *testing.T) {
	h := &UsersHandler{DB: nil}
	body := bytes.NewBufferString(`{"display_name":"u1","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", body)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.ContextWithUser(req.Context(), &middleware.UserInfo{ID: uuid.New(), DisplayName: "admin", Role: "admin"}))
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when email missing, got %d", rec.Code)
	}
	var out map[string]string
	json.NewDecoder(rec.Body).Decode(&out)
	if out["error"] != "email required" {
		t.Errorf("expected error email required, got %v", out)
	}
}

func TestCreate_DisplayNameAndPasswordRequired(t *testing.T) {
	h := &UsersHandler{DB: nil}
	body := bytes.NewBufferString(`{"display_name":"","password":"","email":"e@e.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", body)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.ContextWithUser(req.Context(), &middleware.UserInfo{ID: uuid.New(), DisplayName: "admin", Role: "admin"}))
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when display name/password empty, got %d", rec.Code)
	}
}
