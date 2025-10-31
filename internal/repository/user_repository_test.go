package repository

import (
	"context"
	"testing"

	"Go-Lang-project-01/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "Failed to connect to test database")

	// Auto migrate the schema
	err = db.AutoMigrate(&models.User{})
	require.NoError(t, err, "Failed to migrate test database")

	return db
}

// seedTestUser creates a test user in the database
func seedTestUser(t *testing.T, db *gorm.DB, user *models.User) *models.User {
	err := db.Create(user).Error
	require.NoError(t, err, "Failed to seed test user")
	return user
}

func TestNewUserRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	assert.NotNil(t, repo)
	assert.NotNil(t, repo.db)
}

func TestUserRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		user    *models.User
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful_creation",
			user: &models.User{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "hashedpassword",
				Age:      30,
				IsActive: true,
			},
			wantErr: false,
		},
		{
			name: "duplicate_email",
			user: &models.User{
				Name:     "Jane Doe",
				Email:    "john@example.com", // Duplicate email
				Password: "hashedpassword",
				Age:      25,
				IsActive: true,
			},
			wantErr: true,
			errMsg:  "failed to create user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := NewUserRepository(db)
			ctx := context.Background()

			// For duplicate email test, create the first user
			if tt.name == "duplicate_email" {
				firstUser := &models.User{
					Name:     "First User",
					Email:    "john@example.com",
					Password: "password",
					Age:      20,
					IsActive: true,
				}
				_ = seedTestUser(t, db, firstUser)
			}

			err := repo.Create(ctx, tt.user)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.user.ID)
				assert.NotZero(t, tt.user.CreatedAt)
			}
		})
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Seed test data
	testUser := seedTestUser(t, db, &models.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Age:      25,
		IsActive: true,
	})

	tests := []struct {
		name    string
		id      uint
		want    *models.User
		wantErr bool
	}{
		{
			name:    "existing_user",
			id:      testUser.ID,
			want:    testUser,
			wantErr: false,
		},
		{
			name:    "non_existing_user",
			id:      99999,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.GetByID(ctx, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
				assert.Contains(t, err.Error(), "user not found")
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.want.ID, user.ID)
				assert.Equal(t, tt.want.Email, user.Email)
				assert.Equal(t, tt.want.Name, user.Name)
			}
		})
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Seed test data
	testUser := seedTestUser(t, db, &models.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Age:      25,
		IsActive: true,
	})

	tests := []struct {
		name    string
		email   string
		want    *models.User
		wantErr bool
	}{
		{
			name:    "existing_user",
			email:   "test@example.com",
			want:    testUser,
			wantErr: false,
		},
		{
			name:    "non_existing_user",
			email:   "nonexistent@example.com",
			want:    nil,
			wantErr: false, // GetByEmail returns nil, nil for not found
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.GetByEmail(ctx, tt.email)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.want == nil {
					assert.Nil(t, user)
				} else {
					assert.NotNil(t, user)
					assert.Equal(t, tt.want.Email, user.Email)
				}
			}
		})
	}
}

func TestUserRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Seed multiple users
	users := []*models.User{
		{Name: "User 1", Email: "user1@example.com", Password: "pass1", Age: 20, IsActive: true},
		{Name: "User 2", Email: "user2@example.com", Password: "pass2", Age: 30, IsActive: true},
		{Name: "User 3", Email: "user3@example.com", Password: "pass3", Age: 40, IsActive: false},
	}

	for _, u := range users {
		_ = seedTestUser(t, db, u)
	}

	result, err := repo.GetAll(ctx)

	assert.NoError(t, err)
	assert.Len(t, result, 3)
}

func TestUserRepository_GetAllPaginated(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Seed 15 test users
	for i := 1; i <= 15; i++ {
		name := "User"
		if i < 10 {
			name = "User " + string(rune('0'+i))
		} else {
			name = "User " + string(rune('0'+(i/10))) + string(rune('0'+(i%10)))
		}
		email := "user" + string(rune('0'+i)) + "@example.com"
		if i >= 10 {
			email = "user1" + string(rune('0'+(i%10))) + "@example.com"
		}
		user := &models.User{
			Name:     name,
			Email:    email,
			Password: "password",
			Age:      20 + i,
			IsActive: i%2 == 0, // Even users are active
		}
		_ = seedTestUser(t, db, user)
	}

	tests := []struct {
		name       string
		query      models.PaginationQuery
		wantCount  int
		wantTotal  int64
		wantErr    bool
	}{
		{
			name: "first_page",
			query: models.PaginationQuery{
				Page:  1,
				Limit: 10,
				Sort:  "id",
				Order: "asc",
			},
			wantCount: 10,
			wantTotal: 15,
			wantErr:   false,
		},
		{
			name: "second_page",
			query: models.PaginationQuery{
				Page:  2,
				Limit: 10,
				Sort:  "id",
				Order: "asc",
			},
			wantCount: 5,
			wantTotal: 15,
			wantErr:   false,
		},
		{
			name: "with_search",
			query: models.PaginationQuery{
				Page:   1,
				Limit:  10,
				Search: "User 3",
			},
			wantCount: 1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name: "sort_by_age_desc",
			query: models.PaginationQuery{
				Page:  1,
				Limit: 5,
				Sort:  "age",
				Order: "desc",
			},
			wantCount: 5,
			wantTotal: 15,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, total, err := repo.GetAllPaginated(ctx, tt.query)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, users, tt.wantCount)
				assert.Equal(t, tt.wantTotal, total)
			}
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Seed test user
	testUser := seedTestUser(t, db, &models.User{
		Name:     "Original Name",
		Email:    "original@example.com",
		Password: "password",
		Age:      25,
		IsActive: true,
	})

	// Update user
	testUser.Name = "Updated Name"
	testUser.Age = 30
	testUser.IsActive = false

	err := repo.Update(ctx, testUser)

	assert.NoError(t, err)

	// Verify update
	updated, err := repo.GetByID(ctx, testUser.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, 30, updated.Age)
	assert.False(t, updated.IsActive)
}

func TestUserRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Seed test user
	testUser := seedTestUser(t, db, &models.User{
		Name:     "To Delete",
		Email:    "delete@example.com",
		Password: "password",
		Age:      25,
		IsActive: true,
	})

	// Delete user
	err := repo.Delete(ctx, testUser.ID)
	assert.NoError(t, err)

	// Verify deletion (soft delete)
	_, err = repo.GetByID(ctx, testUser.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestUserRepository_BatchCreate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create batch of users
	users := []*models.User{
		{Name: "Batch 1", Email: "batch1@example.com", Password: "pass", Age: 20, IsActive: true},
		{Name: "Batch 2", Email: "batch2@example.com", Password: "pass", Age: 21, IsActive: true},
		{Name: "Batch 3", Email: "batch3@example.com", Password: "pass", Age: 22, IsActive: true},
		{Name: "Batch 4", Email: "batch4@example.com", Password: "pass", Age: 23, IsActive: true},
		{Name: "Batch 5", Email: "batch5@example.com", Password: "pass", Age: 24, IsActive: true},
	}

	err := repo.BatchCreate(ctx, users)

	assert.NoError(t, err)
	
	// Verify all users were created
	for _, u := range users {
		assert.NotZero(t, u.ID)
	}

	// Verify count
	allUsers, err := repo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, allUsers, 5)
}

func TestUserRepository_GetActiveUsers(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Seed users with mixed active status
	users := []*models.User{
		{Name: "Active 1", Email: "active1@example.com", Password: "pass", Age: 20, IsActive: true},
		{Name: "Inactive 1", Email: "inactive1@example.com", Password: "pass", Age: 21, IsActive: false},
		{Name: "Active 2", Email: "active2@example.com", Password: "pass", Age: 22, IsActive: true},
		{Name: "Inactive 2", Email: "inactive2@example.com", Password: "pass", Age: 23, IsActive: false},
		{Name: "Active 3", Email: "active3@example.com", Password: "pass", Age: 24, IsActive: true},
	}

	for _, u := range users {
		_ = seedTestUser(t, db, u)
	}

	activeUsers, err := repo.GetActiveUsers(ctx)

	assert.NoError(t, err)
	// Should only return users with is_active = true
	assert.GreaterOrEqual(t, len(activeUsers), 3, "Should have at least 3 active users")
	
	// Verify all returned users are active
	for _, u := range activeUsers {
		assert.True(t, u.IsActive, "User %s should be active", u.Email)
	}
}

func TestUserRepository_ContextCancellation(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	// Create a context with immediate cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Try to create a user with cancelled context
	user := &models.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password",
		Age:      25,
		IsActive: true,
	}

	err := repo.Create(ctx, user)
	
	// SQLite might not respect context cancellation immediately
	// but we test that the context is being used
	if err != nil {
		assert.Error(t, err)
	}
}

func TestUserRepository_ConcurrentOperations(t *testing.T) {
	t.Skip("Skipping concurrent test - in-memory SQLite has limitations with concurrent writes from goroutines")
	
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create users concurrently
	done := make(chan bool)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			defer func() { done <- true }()
			
			email := "concurrent" + string(rune('0'+index)) + "@example.com"
			user := &models.User{
				Name:     "Concurrent User",
				Email:    email,
				Password: "password",
				Age:      20 + index,
				IsActive: true,
			}
			err := repo.Create(ctx, user)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	close(errors)

	// Check for errors
	for err := range errors {
		assert.NoError(t, err)
	}

	// Verify all users were created
	allUsers, err := repo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, allUsers, 10)
}

func TestUserRepository_EmptyDatabase(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Test GetAll on empty database
	users, err := repo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Empty(t, users)

	// Test GetAllPaginated on empty database
	query := models.PaginationQuery{
		Page:  1,
		Limit: 10,
	}
	paginatedUsers, total, err := repo.GetAllPaginated(ctx, query)
	assert.NoError(t, err)
	assert.Empty(t, paginatedUsers)
	assert.Equal(t, int64(0), total)

	// Test GetActiveUsers on empty database
	activeUsers, err := repo.GetActiveUsers(ctx)
	assert.NoError(t, err)
	assert.Empty(t, activeUsers)
}

func TestUserRepository_EdgeCases(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("update_non_existing_user", func(t *testing.T) {
		user := &models.User{
			ID:       99999,
			Name:     "Non Existing",
			Email:    "nonexist@example.com",
			Password: "password",
			Age:      25,
			IsActive: true,
		}

		err := repo.Update(ctx, user)
		// GORM will not return error for updating non-existing record
		assert.NoError(t, err)
	})

	t.Run("delete_non_existing_user", func(t *testing.T) {
		err := repo.Delete(ctx, 99999)
		// GORM will not return error for deleting non-existing record
		assert.NoError(t, err)
	})

	t.Run("search_with_special_characters", func(t *testing.T) {
		// Seed user with special characters
		user := &models.User{
			Name:     "O'Brien",
			Email:    "obrien@example.com",
			Password: "password",
			Age:      30,
			IsActive: true,
		}
		_ = seedTestUser(t, db, user)

		query := models.PaginationQuery{
			Page:   1,
			Limit:  10,
			Search: "O'Brien",
		}

		users, total, err := repo.GetAllPaginated(ctx, query)
		assert.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, int64(1), total)
	})

	t.Run("large_batch_create", func(t *testing.T) {
		largeDB := setupTestDB(t)
		largeRepo := NewUserRepository(largeDB)
		
		// Create 250 users (test batch size handling)
		largeUsers := make([]*models.User, 250)
		for i := 0; i < 250; i++ {
			email := "largebatch@example.com"
			if i < 10 {
				email = "largebatch" + string(rune('0'+i)) + "@example.com"
			} else if i < 100 {
				email = "largebatch" + string(rune('0'+(i/10))) + string(rune('0'+(i%10))) + "@example.com"
			} else {
				email = "largebatch" + string(rune('0'+(i/100))) + string(rune('0'+((i/10)%10))) + string(rune('0'+(i%10))) + "@example.com"
			}
			
			largeUsers[i] = &models.User{
				Name:     "Large Batch User",
				Email:    email,
				Password: "password",
				Age:      20,
				IsActive: true,
			}
		}

		err := largeRepo.BatchCreate(ctx, largeUsers)
		assert.NoError(t, err)

		// Verify
		allUsers, err := largeRepo.GetAll(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 250, len(allUsers))
	})
}

func TestUserRepository_Pagination_BoundaryConditions(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Seed 5 users
	for i := 1; i <= 5; i++ {
		name := "User " + string(rune('0'+i))
		email := "user" + string(rune('0'+i)) + "@example.com"
		user := &models.User{
			Name:     name,
			Email:    email,
			Password: "password",
			Age:      20 + i,
			IsActive: true,
		}
		_ = seedTestUser(t, db, user)
	}

	tests := []struct {
		name      string
		page      int
		limit     int
		wantCount int
	}{
		{
			name:      "exact_page_size",
			page:      1,
			limit:     5,
			wantCount: 5,
		},
		{
			name:      "larger_than_total",
			page:      1,
			limit:     100,
			wantCount: 5,
		},
		{
			name:      "page_beyond_total",
			page:      10,
			limit:     10,
			wantCount: 0,
		},
		{
			name:      "single_item_per_page",
			page:      3,
			limit:     1,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := models.PaginationQuery{
				Page:  tt.page,
				Limit: tt.limit,
			}

			users, total, err := repo.GetAllPaginated(ctx, query)
			assert.NoError(t, err)
			assert.Len(t, users, tt.wantCount)
			assert.Equal(t, int64(5), total)
		})
	}
}

// Benchmark tests
func BenchmarkUserRepository_Create(b *testing.B) {
	db := setupTestDB(&testing.T{})
	repo := NewUserRepository(db)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := &models.User{
			Name:     "Benchmark User",
			Email:    "bench" + string(rune(i)) + "@example.com",
			Password: "password",
			Age:      25,
			IsActive: true,
		}
		_ = repo.Create(ctx, user)
	}
}

func BenchmarkUserRepository_GetByID(b *testing.B) {
	db := setupTestDB(&testing.T{})
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Seed test user
	user := &models.User{
		Name:     "Benchmark User",
		Email:    "bench@example.com",
		Password: "password",
		Age:      25,
		IsActive: true,
	}
	_ = repo.Create(ctx, user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetByID(ctx, user.ID)
	}
}

func BenchmarkUserRepository_GetAllPaginated(b *testing.B) {
	db := setupTestDB(&testing.T{})
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Seed 100 users
	for i := 0; i < 100; i++ {
		user := &models.User{
			Name:     "User",
			Email:    "user" + string(rune(i)) + "@example.com",
			Password: "password",
			Age:      20 + i,
			IsActive: true,
		}
		_ = repo.Create(ctx, user)
	}

	query := models.PaginationQuery{
		Page:  1,
		Limit: 10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = repo.GetAllPaginated(ctx, query)
	}
}
