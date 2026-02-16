package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/pet-medical/api/internal/auth"
)

type contextKey string

const UserContextKey contextKey = "user"

const accessCookieName = "access_token"
const refreshCookieName = "refresh_token"

type UserInfo struct {
	ID          uuid.UUID
	DisplayName string
	Email       string
	Role        string
}

func AuthRequired(jwt *auth.JWT) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// User may already be set by TrustedProxyAuth (proxy login)
			if u := GetUser(r.Context()); u != nil {
				next.ServeHTTP(w, r)
				return
			}
			token := ""
			if header := r.Header.Get("Authorization"); header != "" && strings.HasPrefix(header, "Bearer ") {
				token = strings.TrimPrefix(header, "Bearer ")
			}
			if token == "" {
				if c, err := r.Cookie(accessCookieName); err == nil && c.Value != "" {
					token = c.Value
				}
			}
			if token == "" {
				log.Printf("[AUTH] protected route %s missing token (header or cookie)", r.URL.Path)
				http.Error(w, `{"error":"missing authorization"}`, http.StatusUnauthorized)
				return
			}
			claims, err := jwt.ParseAccessToken(token)
			if err != nil {
				log.Printf("[AUTH] protected route %s token invalid or expired: %v", r.URL.Path, err)
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}
			userID, err := uuid.Parse(claims.UserID)
			if err != nil {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}
			displayName := claims.DisplayName
			if displayName == "" {
				displayName = claims.Username
			}
			ctx := context.WithValue(r.Context(), UserContextKey, &UserInfo{
				ID:          userID,
				DisplayName: displayName,
				Email:       claims.Email,
				Role:        claims.Role,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUser(ctx context.Context) *UserInfo {
	u, _ := ctx.Value(UserContextKey).(*UserInfo)
	return u
}

// ContextWithUser returns a context with the user set. Used by tests to inject auth.
func ContextWithUser(ctx context.Context, u *UserInfo) context.Context {
	return context.WithValue(ctx, UserContextKey, u)
}
