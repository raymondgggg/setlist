package db

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrVerificationTokenInvalid = errors.New("verification token is invalid, expired, or already used")

type EmailVerificationToken struct {
	ID        uuid.UUID  `gorm:"column:id;primaryKey"`
	UserID    uuid.UUID  `gorm:"column:user_id"`
	TokenHash string     `gorm:"column:token_hash"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	ExpiresAt time.Time  `gorm:"column:expires_at"`
	UsedAt    *time.Time `gorm:"column:used_at"`
}

type EmailVerificationRepository struct {
	db *gorm.DB
}

func NewEmailVerificationRepository(db *gorm.DB) *EmailVerificationRepository {
	return &EmailVerificationRepository{db: db}
}

func (r *EmailVerificationRepository) Create(ctx context.Context, t *EmailVerificationToken) error {
	return r.db.WithContext(ctx).Create(t).Error
}

// Consume marks a valid token as used and sets email_verified_at on its user,
// returning that user's ID. Returns ErrVerificationTokenInvalid if no valid token matches.
func (r *EmailVerificationRepository) Consume(ctx context.Context, tokenHash string, now time.Time) (uuid.UUID, error) {
	var userID uuid.UUID
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var token EmailVerificationToken
		// The used_at IS NULL condition makes this single-use even when
		// concurrent requests present the same token.
		res := tx.Model(&token).
			Clauses(clause.Returning{}).
			Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", tokenHash, now).
			Update("used_at", now)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrVerificationTokenInvalid
		}
		userID = token.UserID

		// Don't overwrite the timestamp if the user was already verified.
		return tx.Model(&User{}).
			Where("id = ? AND email_verified_at IS NULL", userID).
			Update("email_verified_at", now).Error
	})
	return userID, err
}
