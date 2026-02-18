package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/pet-medical/api/internal/auth"
	"github.com/pet-medical/api/internal/config"
	"github.com/pet-medical/api/internal/debuglog"
	"github.com/pet-medical/api/internal/i18n"
	"github.com/pet-medical/api/internal/middleware"
	"github.com/pet-medical/api/internal/models"
	"gorm.io/gorm"
)

const (
	refreshCookieName = "refresh_token"
	accessCookieName  = "access_token"
)

type AuthHandler struct {
	DB                *gorm.DB
	JWT               *auth.JWT
	RefreshStore      *auth.RefreshStore
	Config            *config.Config
	DefaultWeightUnit string
	DefaultCurrency   string
	DefaultLanguage   string
	SameSiteCookie    int // http.SameSite value (Lax default; set SAME_SITE_COOKIE=none only if needed)
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string  `json:"access_token"`
	ExpiresIn   int     `json:"expires_in"`
	User        UserDTO `json:"user"`
}

type UserDTO struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Role       string `json:"role"`
	WeightUnit string `json:"weight_unit,omitempty"`
	Currency   string `json:"currency,omitempty"`
	Language   string `json:"language,omitempty"`
}

type RefreshResponse struct {
	AccessToken string  `json:"access_token"`
	ExpiresIn   int     `json:"expires_in"`
	User        UserDTO `json:"user"`
}

func (h *AuthHandler) applyUserDefaults(u *models.User) {
	if u.WeightUnit == "" || (u.WeightUnit != "lbs" && u.WeightUnit != "kg") {
		u.WeightUnit = h.DefaultWeightUnit
	}
	if u.Currency == "" {
		u.Currency = h.DefaultCurrency
	}
	if u.Language == "" {
		u.Language = h.DefaultLanguage
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	log.Print(i18n.Tf("log.auth.login_request", r.Method))
	debuglog.Debugf("login: method=%s", r.Method)
	if r.Method != http.MethodPost {
		log.Print(i18n.T("log.auth.login_rejected_method"))
		http.Error(w, `{"error":"error.method_not_allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Print(i18n.Tf("log.auth.login_decode_error", err))
		http.Error(w, `{"error":"error.invalid_request"}`, http.StatusBadRequest)
		return
	}
	log.Print(i18n.Tf("log.auth.login_attempt", req.Email))
	if req.Email == "" || req.Password == "" {
		log.Print(i18n.T("log.auth.login_rejected_missing"))
		http.Error(w, `{"error":"error.email_password_required"}`, http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	var u models.User
	err := h.DB.Where("LOWER(email) = ?", email).First(&u).Error
	if err != nil {
		log.Print(i18n.Tf("log.auth.login_failed_not_found", req.Email, err))
		http.Error(w, `{"error":"error.invalid_credentials"}`, http.StatusUnauthorized)
		return
	}
	h.applyUserDefaults(&u)
	if u.PasswordHash == "" {
		log.Print(i18n.T("log.auth.login_rejected_missing"))
		http.Error(w, `{"error":"error.invalid_credentials"}`, http.StatusUnauthorized)
		return
	}
	if !auth.CheckPassword(u.PasswordHash, req.Password) {
		log.Print(i18n.Tf("log.auth.login_failed_bad_password", req.Email))
		http.Error(w, `{"error":"error.invalid_credentials"}`, http.StatusUnauthorized)
		return
	}

	accessToken, err := h.JWT.NewAccessToken(u.ID, u.DisplayName, u.Email, u.Role)
	if err != nil {
		log.Print(i18n.Tf("log.auth.login_jwt_error", err))
		http.Error(w, `{"error":"error.internal_error"}`, http.StatusInternalServerError)
		return
	}

	expiresAt := time.Now().Add(h.JWT.RefreshTokenDuration())
	refreshToken, err := h.RefreshStore.Create(u.ID, expiresAt)
	if err != nil {
		log.Print(i18n.Tf("log.auth.login_refresh_error", err))
		http.Error(w, `{"error":"error.internal_error"}`, http.StatusInternalServerError)
		return
	}

	h.setRefreshCookie(w, r, refreshToken, int(h.JWT.RefreshTokenDuration().Seconds()))
	h.setAccessCookie(w, r, accessToken, 15*60)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		AccessToken: accessToken,
		ExpiresIn:   int(h.JWT.RefreshTokenDuration().Minutes()),
		User: UserDTO{
			ID:          u.ID.String(),
			DisplayName: u.DisplayName,
			Email:       u.Email,
			Role:       u.Role,
			WeightUnit: u.WeightUnit,
			Currency:   u.Currency,
			Language:   u.Language,
		},
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil || cookie.Value == "" {
		http.Error(w, `{"error":"refresh token required"}`, http.StatusUnauthorized)
		return
	}
	userID, err := h.RefreshStore.Consume(cookie.Value)
	if err != nil {
		h.clearRefreshCookie(w, r)
		http.Error(w, `{"error":"invalid or expired refresh token"}`, http.StatusUnauthorized)
		return
	}

	var u models.User
	err = h.DB.Where("id = ?", userID).First(&u).Error
	if err != nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusUnauthorized)
		return
	}
	h.applyUserDefaults(&u)

	accessToken, err := h.JWT.NewAccessToken(u.ID, u.DisplayName, u.Email, u.Role)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	expiresAt := time.Now().Add(h.JWT.RefreshTokenDuration())
	newRefresh, err := h.RefreshStore.Create(u.ID, expiresAt)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	h.setRefreshCookie(w, r, newRefresh, int(h.JWT.RefreshTokenDuration().Seconds()))
	h.setAccessCookie(w, r, accessToken, 15*60)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(RefreshResponse{
		AccessToken: accessToken,
		ExpiresIn:   int(h.JWT.RefreshTokenDuration().Minutes()),
		User: UserDTO{
			ID:          u.ID.String(),
			DisplayName: u.DisplayName,
			Email:       u.Email,
			Role:       u.Role,
			WeightUnit: u.WeightUnit,
			Currency:   u.Currency,
			Language:   u.Language,
		},
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.clearRefreshCookie(w, r)
	h.clearAccessCookie(w, r)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "logged out"})
}

func (h *AuthHandler) sameSite() http.SameSite {
	switch h.SameSiteCookie {
	case int(http.SameSiteNoneMode):
		return http.SameSiteNoneMode
	case int(http.SameSiteStrictMode):
		return http.SameSiteStrictMode
	default:
		return http.SameSiteLaxMode
	}
}

func (h *AuthHandler) secure(r *http.Request) bool {
	return !h.Config.Development && h.Config.IsRequestHTTPS(r)
}

func (h *AuthHandler) setRefreshCookie(w http.ResponseWriter, r *http.Request, token string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   h.secure(r),
		SameSite: h.sameSite(),
	})
}

func (h *AuthHandler) clearRefreshCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.secure(r),
		SameSite: h.sameSite(),
	})
}

func (h *AuthHandler) setAccessCookie(w http.ResponseWriter, r *http.Request, token string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     accessCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   h.secure(r),
		SameSite: h.sameSite(),
	})
}

func (h *AuthHandler) clearAccessCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     accessCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.secure(r),
		SameSite: h.sameSite(),
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	u := middleware.GetUser(r.Context())
	if u == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	debuglog.Debugf("auth/me: user_id=%s", u.ID.String())
	var dbUser models.User
	if err := h.DB.Where("id = ?", u.ID).First(&dbUser).Error; err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	h.applyUserDefaults(&dbUser)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UserDTO{
		ID:          dbUser.ID.String(),
		DisplayName: dbUser.DisplayName,
		Email:       dbUser.Email,
		Role:        dbUser.Role,
		WeightUnit:  dbUser.WeightUnit,
		Currency:    dbUser.Currency,
		Language:    dbUser.Language,
	})
}

// ChangePasswordRequest is the body for POST /api/auth/change-password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	u := middleware.GetUser(r.Context())
	if u == nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"error.invalid_request"}`, http.StatusBadRequest)
		return
	}
	req.NewPassword = strings.TrimSpace(req.NewPassword)
	if req.CurrentPassword == "" || req.NewPassword == "" {
		http.Error(w, `{"error":"error.password_required"}`, http.StatusBadRequest)
		return
	}
	if len(req.NewPassword) < 8 {
		http.Error(w, `{"error":"error.password_too_short"}`, http.StatusBadRequest)
		return
	}
	var dbUser models.User
	if err := h.DB.Where("id = ?", u.ID).First(&dbUser).Error; err != nil {
		http.Error(w, `{"error":"error.not_found"}`, http.StatusNotFound)
		return
	}
	if dbUser.PasswordHash == "" {
		http.Error(w, `{"error":"error.no_password_account"}`, http.StatusBadRequest)
		return
	}
	if !auth.CheckPassword(dbUser.PasswordHash, req.CurrentPassword) {
		http.Error(w, `{"error":"error.invalid_credentials"}`, http.StatusUnauthorized)
		return
	}
	hash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, `{"error":"error.internal_error"}`, http.StatusInternalServerError)
		return
	}
	if err := h.DB.Model(&models.User{}).Where("id = ?", u.ID).Update("password_hash", hash).Error; err != nil {
		http.Error(w, `{"error":"error.internal_error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "ok"})
}
