package db

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var ErrEmailTaken = errors.New("email already in use")

// Postgres error code for unique_violation.
const pgUniqueViolation = "23505"

// Name of the unique constraint on users.email (see migration 006).
const usersEmailConstraint = "users_email_key"

type User struct {
	ID              uuid.UUID  `gorm:"column:id;primaryKey"`
	FirstName       *string    `gorm:"column:first_name"`
	LastName        *string    `gorm:"column:last_name"`
	Email           string     `gorm:"column:email"`
	PasswordHash    *string    `gorm:"column:password_hash"`
	EmailVerifiedAt *time.Time `gorm:"column:email_verified_at"`
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var u User
	if err := r.db.WithContext(ctx).First(&u, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	if err := r.db.WithContext(ctx).First(&u, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// Create inserts a new user. Returns ErrEmailTaken if the email is already registered.
func (r *UserRepository) Create(ctx context.Context, u *User) error {
	err := r.db.WithContext(ctx).Create(u).Error
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation && pgErr.ConstraintName == usersEmailConstraint {
		return ErrEmailTaken
	}
	return err
}
