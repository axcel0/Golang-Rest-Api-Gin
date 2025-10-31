package services

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"Go-Lang-project-01/internal/models"
	"Go-Lang-project-01/internal/repository"
)

// UserService handles business logic with GORM
type UserService struct {
	repo *repository.UserRepository
}

// NewUserService creates a new GORM user service
func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

// GetAllUsers returns all users
func (s *UserService) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	return s.repo.GetAll(ctx)
}

// GetUserByID returns a user by ID
func (s *UserService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	return s.repo.GetByID(ctx, id)
}

// CreateUser creates a new user with validation
func (s *UserService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	// Validation
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	if req.Email == "" {
		return nil, errors.New("email is required")
	}
	if req.Age <= 0 {
		return nil, errors.New("age must be positive")
	}

	// Check if email already exists
	existingUser, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// Create user
	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Age:      req.Age,
		IsActive: true,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(ctx context.Context, id uint, req *models.UpdateUserRequest) (*models.User, error) {
	// Get existing user
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		// Check if new email already exists
		existingUser, err := s.repo.GetByEmail(ctx, req.Email)
		if err != nil {
			return nil, err
		}
		if existingUser != nil && existingUser.ID != id {
			return nil, errors.New("email already exists")
		}
		user.Email = req.Email
	}
	if req.Age > 0 {
		user.Age = req.Age
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

// BatchCreateUsers creates multiple users concurrently using goroutines
func (s *UserService) BatchCreateUsers(ctx context.Context, requests []*models.CreateUserRequest) ([]*models.User, error) {
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		users   []*models.User
		errList []error
	)

	// Create a channel to limit concurrent goroutines
	semaphore := make(chan struct{}, 5) // Max 5 concurrent operations

	for i, req := range requests {
		wg.Add(1)
		go func(index int, request *models.CreateUserRequest) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Create context with timeout for each goroutine
			goCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			user, err := s.CreateUser(goCtx, request)
			
			mu.Lock()
			defer mu.Unlock()
			
			if err != nil {
				errList = append(errList, fmt.Errorf("user %d: %w", index, err))
			} else {
				users = append(users, user)
			}
		}(i, req)
	}

	wg.Wait()

	if len(errList) > 0 {
		return users, fmt.Errorf("batch create had %d errors: %v", len(errList), errList[0])
	}

	return users, nil
}

// GetUserStats returns user statistics using goroutines
func (s *UserService) GetUserStats(ctx context.Context) (map[string]interface{}, error) {
	var (
		wg          sync.WaitGroup
		mu          sync.Mutex
		totalUsers  int
		activeUsers int
		errors      []error
	)

	// Get total users count
	wg.Add(1)
	go func() {
		defer wg.Done()
		users, err := s.repo.GetAll(ctx)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errors = append(errors, err)
		} else {
			totalUsers = len(users)
		}
	}()

	// Get active users count
	wg.Add(1)
	go func() {
		defer wg.Done()
		users, err := s.repo.GetActiveUsers(ctx)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errors = append(errors, err)
		} else {
			activeUsers = len(users)
		}
	}()

	wg.Wait()

	if len(errors) > 0 {
		return nil, errors[0]
	}

	return map[string]interface{}{
		"total_users":    totalUsers,
		"active_users":   activeUsers,
		"inactive_users": totalUsers - activeUsers,
	}, nil
}
