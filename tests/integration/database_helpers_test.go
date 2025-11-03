package integration

import (
	"testing"

	"Go-Lang-project-01/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCountUsers tests the countUsers helper function
func TestCountUsers(t *testing.T) {
	cleanDatabase()

	// Initially should be 0
	count := countUsers()
	assert.Equal(t, int64(0), count, "Initial user count should be 0")

	// Create first user
	user1, err := seedTestUser("user")
	require.NoError(t, err)
	require.NotNil(t, user1)

	count = countUsers()
	assert.Equal(t, int64(1), count, "After creating 1 user, count should be 1")

	// Create second user
	user2, err := seedTestUser("admin")
	require.NoError(t, err)
	require.NotNil(t, user2)

	count = countUsers()
	assert.Equal(t, int64(2), count, "After creating 2 users, count should be 2")

	// Cleanup
	cleanDatabase()
	count = countUsers()
	assert.Equal(t, int64(0), count, "After cleanup, count should be 0")
}

// TestGetUserByEmail tests the getUserByEmail helper function
func TestGetUserByEmail(t *testing.T) {
	cleanDatabase()

	// Create test user
	createdUser, err := seedTestUser("user")
	require.NoError(t, err)
	require.NotNil(t, createdUser)

	// Test successful retrieval
	t.Run("Get Existing User", func(t *testing.T) {
		user, err := getUserByEmail("user@test.com")
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "user@test.com", user.Email)
		assert.Equal(t, "Test user", user.Name)
		assert.Equal(t, "user", user.Role)
	})

	// Test non-existent user
	t.Run("Get Non-Existent User", func(t *testing.T) {
		user, err := getUserByEmail("nonexistent@test.com")
		assert.Error(t, err)
		assert.NotNil(t, user) // GORM returns empty struct
	})

	cleanDatabase()
}

// TestDeleteUser tests the deleteUser helper function
func TestDeleteUser(t *testing.T) {
	cleanDatabase()

	// Create test user
	user, err := seedTestUser("user")
	require.NoError(t, err)
	require.NotNil(t, user)

	// Verify user exists
	count := countUsers()
	assert.Equal(t, int64(1), count)

	// Delete user
	err = deleteUser(user.ID)
	require.NoError(t, err)

	// Verify user is deleted
	count = countUsers()
	assert.Equal(t, int64(0), count)

	// Try to delete non-existent user (should not error)
	err = deleteUser(9999)
	assert.NoError(t, err) // GORM doesn't error on delete of non-existent record

	cleanDatabase()
}

// TestCleanDatabase tests the cleanDatabase helper function
func TestCleanDatabase(t *testing.T) {
	// Create multiple users
	_, err := seedTestUser("user")
	require.NoError(t, err)
	_, err = seedTestUser("admin")
	require.NoError(t, err)
	_, err = seedTestUser("superadmin")
	require.NoError(t, err)

	// Verify users created
	count := countUsers()
	assert.Equal(t, int64(3), count)

	// Clean database
	cleanDatabase()

	// Verify all users deleted
	count = countUsers()
	assert.Equal(t, int64(0), count)

	// Verify audit logs also cleaned
	auditCount := countAuditLogs()
	assert.Equal(t, int64(0), auditCount)
}

// TestSeedTestUser tests the seedTestUser helper function
func TestSeedTestUser(t *testing.T) {
	cleanDatabase()

	tests := []struct {
		name     string
		role     string
		wantName string
		wantRole string
	}{
		{
			name:     "Create Regular User",
			role:     "user",
			wantName: "Test user",
			wantRole: "user",
		},
		{
			name:     "Create Admin User",
			role:     "admin",
			wantName: "Test admin",
			wantRole: "admin",
		},
		{
			name:     "Create SuperAdmin User",
			role:     "superadmin",
			wantName: "Test superadmin",
			wantRole: "superadmin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := seedTestUser(tt.role)
			require.NoError(t, err)
			require.NotNil(t, user)

			// Verify user properties
			assert.Equal(t, tt.wantName, user.Name)
			assert.Equal(t, tt.role+"@test.com", user.Email)
			assert.Equal(t, tt.wantRole, user.Role)
			assert.Equal(t, 30, user.Age)
			assert.True(t, user.IsActive)
			assert.NotEmpty(t, user.Password)
			assert.NotZero(t, user.ID)

			// Verify user can be retrieved
			foundUser, err := getUserByEmail(user.Email)
			require.NoError(t, err)
			assert.Equal(t, user.ID, foundUser.ID)
		})
	}

	cleanDatabase()
}

// TestGetAuthToken tests the getAuthToken helper function
func TestGetAuthToken(t *testing.T) {
	cleanDatabase()

	// Create test user
	user, err := seedTestUser("user")
	require.NoError(t, err)
	require.NotNil(t, user)

	// Generate token
	token, err := getAuthToken(user)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify token format (JWT has 3 parts separated by dots)
	assert.Contains(t, token, ".")
	parts := len(token)
	assert.Greater(t, parts, 50, "JWT token should be reasonably long")

	// Test with different roles
	t.Run("Generate Token for Admin", func(t *testing.T) {
		admin, err := seedTestUser("admin")
		require.NoError(t, err)

		token, err := getAuthToken(admin)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	cleanDatabase()
}

// TestCountAuditLogs tests the countAuditLogs helper function
func TestCountAuditLogs(t *testing.T) {
	cleanDatabase()

	// Initially should be 0
	count := countAuditLogs()
	assert.Equal(t, int64(0), count, "Initial audit log count should be 0")

	// Create audit log manually
	userID1 := uint(1)
	auditLog := &models.AuditLog{
		UserID:    &userID1,
		Action:    "test_action",
		Resource:  "test_resource",
		IPAddress: "127.0.0.1",
		UserAgent: "test-agent",
		Success:   true,
	}
	err := testDB.Create(auditLog).Error
	require.NoError(t, err)

	// Verify count increased
	count = countAuditLogs()
	assert.Equal(t, int64(1), count, "After creating 1 audit log, count should be 1")

	// Create another audit log
	userID2 := uint(2)
	auditLog2 := &models.AuditLog{
		UserID:    &userID2,
		Action:    "test_action_2",
		Resource:  "test_resource_2",
		IPAddress: "127.0.0.2",
		UserAgent: "test-agent-2",
		Success:   false,
		ErrorMsg:  "test error",
	}
	err = testDB.Create(auditLog2).Error
	require.NoError(t, err)

	count = countAuditLogs()
	assert.Equal(t, int64(2), count, "After creating 2 audit logs, count should be 2")

	// Cleanup
	cleanDatabase()
	count = countAuditLogs()
	assert.Equal(t, int64(0), count, "After cleanup, count should be 0")
}

// TestHelperFunctionsIntegration tests all helper functions working together
func TestHelperFunctionsIntegration(t *testing.T) {
	cleanDatabase()

	// Step 1: Verify empty state
	assert.Equal(t, int64(0), countUsers())
	assert.Equal(t, int64(0), countAuditLogs())

	// Step 2: Create users
	user1, err := seedTestUser("user")
	require.NoError(t, err)
	user2, err := seedTestUser("admin")
	require.NoError(t, err)

	// Step 3: Verify counts
	assert.Equal(t, int64(2), countUsers())

	// Step 4: Generate tokens
	token1, err := getAuthToken(user1)
	require.NoError(t, err)
	assert.NotEmpty(t, token1)

	token2, err := getAuthToken(user2)
	require.NoError(t, err)
	assert.NotEmpty(t, token2)
	assert.NotEqual(t, token1, token2, "Tokens should be unique")

	// Step 5: Retrieve users by email
	foundUser1, err := getUserByEmail("user@test.com")
	require.NoError(t, err)
	assert.Equal(t, user1.ID, foundUser1.ID)

	foundUser2, err := getUserByEmail("admin@test.com")
	require.NoError(t, err)
	assert.Equal(t, user2.ID, foundUser2.ID)

	// Step 6: Delete one user
	err = deleteUser(user1.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), countUsers())

	// Step 7: Verify deleted user cannot be found
	_, err = getUserByEmail("user@test.com")
	assert.Error(t, err)

	// Step 8: Clean everything
	cleanDatabase()
	assert.Equal(t, int64(0), countUsers())
	assert.Equal(t, int64(0), countAuditLogs())
}
