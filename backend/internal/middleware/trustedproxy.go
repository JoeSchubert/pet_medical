package middleware

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pet-medical/api/internal/auth"
	"github.com/pet-medical/api/internal/config"
	"github.com/pet-medical/api/internal/models"
	"gorm.io/gorm"
)

// TrustedProxyAuth runs before AuthRequired. When the request comes from a trusted proxy
// and has the configured forwarded-email header (e.g. X-Forwarded-Email from oauth2-proxy),
// it finds or creates a user by email, issues JWT + refresh, sets cookies, and injects the user
// into context so the current request is authenticated.
func TrustedProxyAuth(cfg *config.Config, db *gorm.DB, jwt *auth.JWT, refreshStore *auth.RefreshStore, defaultWeightUnit, defaultCurrency, defaultLanguage string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.IsTrustedProxy(r.RemoteAddr) {
				next.ServeHTTP(w, r)
				return
			}
			email := strings.TrimSpace(strings.ToLower(r.Header.Get(cfg.ForwardedEmailHeader)))
			if email == "" {
				next.ServeHTTP(w, r)
				return
			}
			// If we already have a valid token, let normal auth handle it
			if token := extractAccessToken(r); token != "" {
				if _, err := jwt.ParseAccessToken(token); err == nil {
					next.ServeHTTP(w, r)
					return
				}
			}
			// Find or create user by email
			var u models.User
			err := db.Where("LOWER(TRIM(email)) = ?", email).First(&u).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					u = createUserFromProxy(db, email, r.Header.Get(cfg.ForwardedUserHeader), defaultWeightUnit, defaultCurrency, defaultLanguage)
					if u.ID == uuid.Nil {
						log.Printf("[TRUSTED_PROXY] failed to create user for email %q", email)
						next.ServeHTTP(w, r)
						return
					}
					log.Printf("[TRUSTED_PROXY] created user %s for email %s", u.DisplayName, email)
				} else {
					log.Printf("[TRUSTED_PROXY] lookup user by email: %v", err)
					next.ServeHTTP(w, r)
					return
				}
			}
			applyDefaults(&u, defaultWeightUnit, defaultCurrency, defaultLanguage)
			accessToken, err := jwt.NewAccessToken(u.ID, u.DisplayName, u.Email, u.Role)
			if err != nil {
				log.Printf("[TRUSTED_PROXY] new access token: %v", err)
				next.ServeHTTP(w, r)
				return
			}
			expiresAt := time.Now().Add(jwt.RefreshTokenDuration())
			refreshToken, err := refreshStore.Create(u.ID, expiresAt)
			if err != nil {
				log.Printf("[TRUSTED_PROXY] refresh create: %v", err)
				next.ServeHTTP(w, r)
				return
			}
			secure := !cfg.Development && cfg.IsRequestHTTPS(r)
			setRefreshCookie(w, refreshToken, int(jwt.RefreshTokenDuration().Seconds()), secure, int(cfg.SameSiteCookie))
			setAccessCookie(w, accessToken, 15*60, secure, int(cfg.SameSiteCookie))
			ctx := context.WithValue(r.Context(), UserContextKey, &UserInfo{
				ID:          u.ID,
				DisplayName: u.DisplayName,
				Email:       u.Email,
				Role:        u.Role,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractAccessToken(r *http.Request) string {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	if c, err := r.Cookie(accessCookieName); err == nil && c.Value != "" {
		return c.Value
	}
	return ""
}

func applyDefaults(u *models.User, defaultWeightUnit, defaultCurrency, defaultLanguage string) {
	if u.WeightUnit == "" || (u.WeightUnit != "lbs" && u.WeightUnit != "kg") {
		u.WeightUnit = defaultWeightUnit
	}
	if u.Currency == "" {
		u.Currency = defaultCurrency
	}
	if u.Language == "" {
		u.Language = defaultLanguage
	}
}

// createUserFromProxy creates a user with the given email; display name is derived from forwarded-user header or email. PasswordHash is empty (OAuth-only).
func createUserFromProxy(db *gorm.DB, email, forwardedUser, defaultWeightUnit, defaultCurrency, defaultLanguage string) models.User {
	displayName := strings.TrimSpace(forwardedUser)
	if displayName == "" {
		displayName = deriveDisplayNameFromEmail(email)
	}
	displayName = sanitizeDisplayName(displayName)
	// Ensure unique display_name
	base := displayName
	for i := 0; i < 100; i++ {
		if i > 0 {
			displayName = base + strconv.Itoa(i)
		}
		var exists models.User
		if err := db.Where("display_name = ?", displayName).First(&exists).Error; err == gorm.ErrRecordNotFound {
			break
		}
	}
	u := models.User{
		DisplayName:  displayName,
		Email:        email,
		PasswordHash: "", // OAuth-only
		Role:         "user",
		WeightUnit:   defaultWeightUnit,
		Currency:     defaultCurrency,
		Language:     defaultLanguage,
	}
	if err := db.Create(&u).Error; err != nil {
		return models.User{}
	}
	return u
}

func deriveDisplayNameFromEmail(email string) string {
	at := strings.Index(email, "@")
	if at <= 0 {
		return "user"
	}
	return email[:at]
}

func sanitizeDisplayName(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if out == "" {
		return "user"
	}
	if len(out) > 64 {
		out = out[:64]
	}
	return out
}

func sameSiteMode(sameSite int) http.SameSite {
	switch sameSite {
	case int(http.SameSiteNoneMode):
		return http.SameSiteNoneMode
	case int(http.SameSiteStrictMode):
		return http.SameSiteStrictMode
	default:
		return http.SameSiteLaxMode
	}
}

func setRefreshCookie(w http.ResponseWriter, token string, maxAge int, secure bool, sameSite int) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSiteMode(sameSite),
	})
}

func setAccessCookie(w http.ResponseWriter, token string, maxAge int, secure bool, sameSite int) {
	http.SetCookie(w, &http.Cookie{
		Name:     accessCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSiteMode(sameSite),
	})
}
