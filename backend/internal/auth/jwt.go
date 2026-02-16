package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	jwt.RegisteredClaims
	UserID      string `json:"uid"`
	DisplayName string `json:"display_name"`
	Username    string `json:"username"` // legacy; prefer DisplayName
	Email       string `json:"email"`
	Role        string `json:"role"`
}

func NewJWT(secret string, accessTTLMin, refreshTTLDays int) *JWT {
	return &JWT{
		secret:         secret,
		accessTTL:     time.Duration(accessTTLMin) * time.Minute,
		refreshTTL:    time.Duration(refreshTTLDays) * 24 * time.Hour,
	}
}

type JWT struct {
	secret      string
	accessTTL   time.Duration
	refreshTTL  time.Duration
}

func (j *JWT) NewAccessToken(userID uuid.UUID, displayName, email, role string) (string, error) {
	now := time.Now()
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
		UserID:      userID.String(),
		DisplayName: displayName,
		Email:       email,
		Role:        role,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(j.secret))
}

func (j *JWT) ParseAccessToken(tokenString string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(j.secret), nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}
	claims, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func (j *JWT) RefreshTokenDuration() time.Duration { return j.refreshTTL }
