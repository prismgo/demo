package repositories

import (
	"context"

	"gorm.io/gorm"

	"prismgo-demo/app/models"
)

// UserRepositoryKey identifies the user repository in the service container.
const UserRepositoryKey = "demo.repository.users"

// UserRepository persists demo customers through public GORM APIs.
type UserRepository struct{ db *gorm.DB }

// NewUserRepository creates a user repository backed by db.
func NewUserRepository(db *gorm.DB) *UserRepository { return &UserRepository{db: db} }

// Create persists a new user.
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// Find returns a user by ID.
func (r *UserRepository) Find(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
