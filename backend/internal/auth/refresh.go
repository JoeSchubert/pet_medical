package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/pet-medical/api/internal/models"
	"gorm.io/gorm"
)

const refreshTokenBytes = 32

type RefreshStore struct {
	db *gorm.DB
}

func NewRefreshStore(db *gorm.DB) *RefreshStore {
	return &RefreshStore{db: db}
}

func (s *RefreshStore) Create(userID uuid.UUID, expiresAt time.Time) (token string, err error) {
	b := make([]byte, refreshTokenBytes)
	if _, err = rand.Read(b); err != nil {
		return "", err
	}
	token = hex.EncodeToString(b)
	hash := HashToken(token)
	rec := models.RefreshToken{
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: expiresAt,
	}
	if err = s.db.Create(&rec).Error; err != nil {
		return "", err
	}
	return token, nil
}

func (s *RefreshStore) Consume(token string) (userID uuid.UUID, err error) {
	hash := HashToken(token)
	var rec models.RefreshToken
	err = s.db.Where("token_hash = ? AND expires_at > NOW()", hash).First(&rec).Error
	if err != nil {
		return uuid.Nil, err
	}
	userID = rec.UserID
	_ = s.db.Where("token_hash = ?", hash).Delete(&models.RefreshToken{}).Error
	return userID, nil
}

func (s *RefreshStore) RevokeAllForUser(userID uuid.UUID) error {
	return s.db.Where("user_id = ?", userID).Delete(&models.RefreshToken{}).Error
}

func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
