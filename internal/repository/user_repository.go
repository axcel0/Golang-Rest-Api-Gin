package repository

import (
	"context"
	"fmt"
	"strings"

	"Go-Lang-project-01/internal/models"
	"gorm.io/gorm"
)

// UserRepository handles data persistence with GORM
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new GORM user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// GetAll returns all users (with goroutine support via context)
func (r *UserRepository) GetAll(ctx context.Context) ([]*models.User, error) {
	var users []*models.User
	
	// Using context for cancellation support
	if err := r.db.WithContext(ctx).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	
	return users, nil
}

// GetAllPaginated returns paginated users with search and filtering
func (r *UserRepository) GetAllPaginated(ctx context.Context, query models.PaginationQuery) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64
	
	// Base query
	db := r.db.WithContext(ctx).Model(&models.User{})
	
	// Apply search filter
	if query.Search != "" {
		searchPattern := "%" + strings.ToLower(query.Search) + "%"
		db = db.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ?", searchPattern, searchPattern)
	}
	
	// Count total records
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}
	
	// Apply sorting
	sortField := "created_at"
	sortOrder := "desc"
	if query.Sort != "" {
		sortField = query.Sort
	}
	if query.Order != "" {
		sortOrder = query.Order
	}
	db = db.Order(fmt.Sprintf("%s %s", sortField, sortOrder))
	
	// Apply pagination
	offset := (query.Page - 1) * query.Limit
	if err := db.Offset(offset).Limit(query.Limit).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get users: %w", err)
	}
	
	return users, total, nil
}

// GetByID returns a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return &user, nil
}

// GetByEmail returns a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not an error, just not found
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	
	return &user, nil
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// Update updates an existing user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// Delete soft deletes a user
func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&models.User{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// BatchCreate creates multiple users in a transaction (Goroutine example)
func (r *UserRepository) BatchCreate(ctx context.Context, users []*models.User) error {
	// Using transaction for batch insert
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.CreateInBatches(users, 100).Error; err != nil {
			return fmt.Errorf("failed to batch create users: %w", err)
		}
		return nil
	})
}

// GetActiveUsers returns only active users
func (r *UserRepository) GetActiveUsers(ctx context.Context) ([]*models.User, error) {
	var users []*models.User
	
	if err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}
	
	return users, nil
}
