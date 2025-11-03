package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"Go-Lang-project-01/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuthFlow tests the complete authentication flow
func TestAuthFlow(t *testing.T) {
	cleanDatabase()

	t.Run("Complete Auth Flow - Register, Login, Access Protected Route", func(t *testing.T) {
		// Step 1: Register a new user
		registerReq := map[string]interface{}{
			"name":     "John Doe",
			"email":    "john@example.com",
			"password": "securePassword123",
			"age":      25,
		}
		body, _ := json.Marshal(registerReq)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, "Register should succeed")

		var registerResp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &registerResp)
		require.NoError(t, err)
		assert.True(t, registerResp["success"].(bool))

		// Step 2: Login with the registered user
		loginReq := map[string]interface{}{
			"email":    "john@example.com",
			"password": "securePassword123",
		}
		body, _ = json.Marshal(loginReq)

		w = httptest.NewRecorder()
		req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Login should succeed")

		var loginResp map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &loginResp)
		require.NoError(t, err)
		assert.True(t, loginResp["success"].(bool))

		data := loginResp["data"].(map[string]interface{})
		token := data["access_token"].(string)
		assert.NotEmpty(t, token, "Should receive JWT token")

		// Step 3: Access protected route with token
		w = httptest.NewRecorder()
		req = httptest.NewRequest("GET", "/api/v1/users", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Should access protected route with valid token")

		var usersResp map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &usersResp)
		require.NoError(t, err)
		assert.True(t, usersResp["success"].(bool))
	})

	t.Run("Login with invalid credentials", func(t *testing.T) {
		loginReq := map[string]interface{}{
			"email":    "wrong@example.com",
			"password": "wrongpassword",
		}
		body, _ := json.Marshal(loginReq)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code, "Should fail with invalid credentials")
	})

	t.Run("Access protected route without token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/users", nil)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code, "Should fail without token")
	})

	t.Run("Register with duplicate email", func(t *testing.T) {
		// First registration
		registerReq := map[string]interface{}{
			"name":     "Alice",
			"email":    "alice@example.com",
			"password": "password123",
			"age":      28,
		}
		body, _ := json.Marshal(registerReq)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		// Second registration with same email
		w = httptest.NewRecorder()
		req = httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code, "Should fail with duplicate email (409 Conflict)")
	})
}

// TestUserCRUDFlow tests the complete user CRUD operations
func TestUserCRUDFlow(t *testing.T) {
	cleanDatabase()

	t.Run("Complete CRUD Flow as Admin", func(t *testing.T) {
		// Create admin user
		adminUser, err := seedTestUser("admin")
		require.NoError(t, err)
		adminToken, err := getAuthToken(adminUser)
		require.NoError(t, err)

		// Step 1: Create a new user (POST /users)
		createReq := map[string]interface{}{
			"name":     "Bob Smith",
			"email":    "bob@example.com",
			"password": "password123",
			"age":      35,
		}
		body, _ := json.Marshal(createReq)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, "Admin should create user")

		var createResp map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &createResp)
		require.NoError(t, err)

		data := createResp["data"].(map[string]interface{})
		userID := int(data["id"].(float64))
		assert.NotZero(t, userID)

		// Step 2: Get user by ID (GET /users/:id)
		w = httptest.NewRecorder()
		req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/users/%d", userID), nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Should retrieve user by ID")

		var getResp map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &getResp)
		require.NoError(t, err)
		userData := getResp["data"].(map[string]interface{})
		assert.Equal(t, "Bob Smith", userData["name"])
		assert.Equal(t, "bob@example.com", userData["email"])

		// Step 3: Update user (PUT /users/:id)
		updateReq := map[string]interface{}{
			"name": "Bob Smith Updated",
			"age":  36,
		}
		body, _ = json.Marshal(updateReq)

		w = httptest.NewRecorder()
		req = httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%d", userID), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Admin should update user")

		var updateResp map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &updateResp)
		require.NoError(t, err)
		updatedData := updateResp["data"].(map[string]interface{})
		assert.Equal(t, "Bob Smith Updated", updatedData["name"])
		assert.Equal(t, float64(36), updatedData["age"])

		// Step 4: Get all users (GET /users)
		w = httptest.NewRecorder()
		req = httptest.NewRequest("GET", "/api/v1/users", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Should list all users")

		// Step 5: Delete user (DELETE /users/:id)
		w = httptest.NewRecorder()
		req = httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/users/%d", userID), nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Admin should delete user")

		// Verify user is deleted
		w = httptest.NewRecorder()
		req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/users/%d", userID), nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code, "Deleted user should not be found")
	})

	t.Run("Regular user cannot create users", func(t *testing.T) {
		// Create regular user
		regularUser, err := seedTestUser("user")
		require.NoError(t, err)
		userToken, err := getAuthToken(regularUser)
		require.NoError(t, err)

		createReq := map[string]interface{}{
			"name":     "Test User",
			"email":    "test@example.com",
			"password": "password123",
			"age":      25,
		}
		body, _ := json.Marshal(createReq)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+userToken)
		req.Header.Set("Content-Type", "application/json")
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code, "Regular user should not create users")
	})
}

// TestPaginationAndFiltering tests pagination features
func TestPaginationAndFiltering(t *testing.T) {
	cleanDatabase()

	// Create admin user for auth
	adminUser, err := seedTestUser("admin")
	require.NoError(t, err)
	adminToken, err := getAuthToken(adminUser)
	require.NoError(t, err)

	// Create multiple users for pagination testing
	for i := 1; i <= 15; i++ {
		createReq := map[string]interface{}{
			"name":     fmt.Sprintf("User %d", i),
			"email":    fmt.Sprintf("user%d@example.com", i),
			"password": "password123",
			"age":      20 + i,
		}
		body, _ := json.Marshal(createReq)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		testRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)
	}

	t.Run("Get first page with default limit", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/users?page=1&limit=5", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		data := resp["data"].([]interface{})
		assert.Len(t, data, 5, "Should return 5 users")

		pagination := resp["pagination"].(map[string]interface{})
		assert.Equal(t, float64(1), pagination["page"])
		assert.Equal(t, float64(5), pagination["limit"])
		assert.Greater(t, int(pagination["total"].(float64)), 10)
	})

	t.Run("Get second page", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/users?page=2&limit=5", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		pagination := resp["pagination"].(map[string]interface{})
		assert.Equal(t, float64(2), pagination["page"])
	})

	t.Run("Search users by name", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/users?search=User%201", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		data := resp["data"].([]interface{})
		assert.Greater(t, len(data), 0, "Should find users matching search")
	})
}

// TestBatchOperations tests batch creation
func TestBatchOperations(t *testing.T) {
	cleanDatabase()

	adminUser, err := seedTestUser("admin")
	require.NoError(t, err)
	adminToken, err := getAuthToken(adminUser)
	require.NoError(t, err)

	t.Run("Batch create multiple users", func(t *testing.T) {
		batchReq := []map[string]interface{}{
			{
				"name":     "Batch User 1",
				"email":    "batch1@example.com",
				"password": "password123",
				"age":      25,
			},
			{
				"name":     "Batch User 2",
				"email":    "batch2@example.com",
				"password": "password123",
				"age":      30,
			},
			{
				"name":     "Batch User 3",
				"email":    "batch3@example.com",
				"password": "password123",
				"age":      35,
			},
		}
		body, _ := json.Marshal(batchReq)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/users/batch", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, "Batch create should succeed")

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		data := resp["data"].([]interface{})
		assert.Len(t, data, 3, "Should create 3 users")
	})

	t.Run("Batch create with duplicate email", func(t *testing.T) {
		batchReq := []map[string]interface{}{
			{
				"name":     "Unique User",
				"email":    "unique@example.com",
				"password": "password123",
				"age":      25,
			},
			{
				"name":     "Duplicate Email",
				"email":    "batch1@example.com", // Already exists
				"password": "password123",
				"age":      30,
			},
		}
		body, _ := json.Marshal(batchReq)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/users/batch", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Should fail with duplicate email")
	})
}

// TestUserStats tests statistics endpoint
func TestUserStats(t *testing.T) {
	cleanDatabase()

	// Create users with different active states
	for i := 1; i <= 5; i++ {
		user := &models.User{
			Name:     fmt.Sprintf("User %d", i),
			Email:    fmt.Sprintf("user%d@stats.com", i),
			Password: "hashed",
			Age:      25,
			Role:     "user",
			IsActive: i <= 3, // First 3 active, last 2 inactive
		}
		testDB.Create(user)
	}

	regularUser, err := seedTestUser("user")
	require.NoError(t, err)
	userToken, err := getAuthToken(regularUser)
	require.NoError(t, err)

	t.Run("Get user statistics", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/users/stats", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		data := resp["data"].(map[string]interface{})
		assert.Greater(t, int(data["total_users"].(float64)), 4)
		assert.Greater(t, int(data["active_users"].(float64)), 2)
	})
}
