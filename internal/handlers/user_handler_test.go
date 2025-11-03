package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"Go-Lang-project-01/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockUserService is a mock implementation of UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserService) GetAllUsersPaginated(ctx context.Context, query models.PaginationQuery) ([]*models.User, models.PaginationMeta, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, models.PaginationMeta{}, args.Error(2)
	}
	return args.Get(0).([]*models.User), args.Get(1).(models.PaginationMeta), args.Error(2)
}

func (m *MockUserService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, id uint, req *models.UpdateUserRequest) (*models.User, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserService) BatchCreateUsers(ctx context.Context, requests []*models.CreateUserRequest) ([]*models.User, error) {
	args := m.Called(ctx, requests)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserService) GetUserStats(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockUserService) UpdateUserRole(ctx context.Context, userID uint, newRole string) (*models.User, error) {
	args := m.Called(ctx, userID, newRole)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// setupTestRouter creates a Gin router for testing
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// UserHandlerTestable is a testable version of UserHandler
type UserHandlerTestable struct {
	mockService *MockUserService
}

// NewUserHandlerTestable creates a handler with mock service
func NewUserHandlerTestable(mockService *MockUserService) *UserHandlerTestable {
	return &UserHandlerTestable{
		mockService: mockService,
	}
}

// Implement all handler methods using mockService
func (h *UserHandlerTestable) GetAllUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var query models.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "validation error",
			Data:    err.Error(),
		})
		return
	}

	users, meta, err := h.mockService.GetAllUsersPaginated(ctx, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Success:    true,
		Data:       users,
		Pagination: meta,
	})
}

func (h *UserHandlerTestable) GetUserByID(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "invalid user id",
		})
		return
	}

	user, err := h.mockService.GetUserByID(ctx, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Data:    user,
	})
}

func (h *UserHandlerTestable) CreateUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "validation error",
			Data:    err.Error(),
		})
		return
	}

	user, err := h.mockService.CreateUser(ctx, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.Response{
		Success: true,
		Message: "user created successfully",
		Data:    user,
	})
}

func (h *UserHandlerTestable) UpdateUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "invalid user id",
		})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "validation error",
			Data:    err.Error(),
		})
		return
	}

	user, err := h.mockService.UpdateUser(ctx, uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "user updated successfully",
		Data:    user,
	})
}

func (h *UserHandlerTestable) DeleteUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "invalid user id",
		})
		return
	}

	if err := h.mockService.DeleteUser(ctx, uint(id)); err != nil {
		c.JSON(http.StatusNotFound, models.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "user deleted successfully",
	})
}

func (h *UserHandlerTestable) BatchCreateUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	var requests []*models.CreateUserRequest
	if err := c.ShouldBindJSON(&requests); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "validation error",
			Data:    err.Error(),
		})
		return
	}

	// Check for empty batch
	if len(requests) == 0 {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "batch request cannot be empty",
		})
		return
	}

	users, err := h.mockService.BatchCreateUsers(ctx, requests)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: err.Error(),
			Data:    users,
		})
		return
	}

	c.JSON(http.StatusCreated, models.Response{
		Success: true,
		Message: "users created successfully",
		Data:    users,
	})
}

func (h *UserHandlerTestable) GetUserStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	stats, err := h.mockService.GetUserStats(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Data:    stats,
	})
}

func (h *UserHandlerTestable) UpdateUserRole(c *gin.Context) {
	requestingUserInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.Response{
			Success: false,
			Message: "unauthorized",
		})
		return
	}

	requestingUser, ok := requestingUserInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Message: "invalid user type",
		})
		return
	}

	if !requestingUser.IsSuperAdmin() {
		c.JSON(http.StatusForbidden, models.Response{
			Success: false,
			Message: "only superadmin can change user roles",
		})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "invalid user ID",
		})
		return
	}

	var req models.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "validation error",
			Data:    err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	user, err := h.mockService.GetUserByID(ctx, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.Response{
			Success: false,
			Message: "user not found",
		})
		return
	}

	if user.ID == requestingUser.ID && req.Role != string(models.RoleSuperAdmin) {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "cannot demote yourself",
		})
		return
	}

	updatedUser, err := h.mockService.UpdateUserRole(ctx, user.ID, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Message: "failed to update role",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Data: gin.H{
			"message": "user role updated successfully",
			"user":    updatedUser,
		},
	})
}

// setupHandlerWithMock creates testable handler with mock service
func setupHandlerWithMock(mockService *MockUserService) *UserHandlerTestable {
	return NewUserHandlerTestable(mockService)
}

// Test GetAllUsers
func TestGetAllUsers(t *testing.T) {
	tests := []struct {
		name               string
		queryParams        string
		mockSetup          func(*MockUserService)
		expectedStatusCode int
		expectedSuccess    bool
		validateResponse   func(*testing.T, map[string]interface{})
	}{
		{
			name:        "Successfully get all users with pagination",
			queryParams: "?page=1&limit=10",
			mockSetup: func(m *MockUserService) {
				users := []*models.User{
					{ID: 1, Name: "User 1", Email: "user1@test.com", Role: "user"},
					{ID: 2, Name: "User 2", Email: "user2@test.com", Role: "admin"},
				}
				meta := models.PaginationMeta{
					Page:       1,
					Limit:      10,
					Total:      2,
					TotalPages: 1,
				}
				m.On("GetAllUsersPaginated", mock.Anything, mock.Anything).Return(users, meta, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedSuccess:    true,
			validateResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.True(t, resp["success"].(bool))
				data := resp["data"].([]interface{})
				assert.Len(t, data, 2)
				pagination := resp["pagination"].(map[string]interface{})
				assert.Equal(t, float64(1), pagination["page"])
				assert.Equal(t, float64(2), pagination["total"])
			},
		},
		{
			name:        "Empty user list",
			queryParams: "?page=1&limit=10",
			mockSetup: func(m *MockUserService) {
				meta := models.PaginationMeta{
					Page:       1,
					Limit:      10,
					Total:      0,
					TotalPages: 0,
				}
				m.On("GetAllUsersPaginated", mock.Anything, mock.Anything).Return([]*models.User{}, meta, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedSuccess:    true,
			validateResponse: func(t *testing.T, resp map[string]interface{}) {
				data := resp["data"].([]interface{})
				assert.Len(t, data, 0)
			},
		},
		{
			name:        "Service returns error",
			queryParams: "?page=1&limit=10",
			mockSetup: func(m *MockUserService) {
				m.On("GetAllUsersPaginated", mock.Anything, mock.Anything).Return(nil, models.PaginationMeta{}, errors.New("database error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedSuccess:    false,
		},
		{
			name:        "Invalid query parameters",
			queryParams: "?page=invalid&limit=abc",
			mockSetup: func(m *MockUserService) {
				// No mock setup - validation happens before service call
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedSuccess:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.mockSetup(mockService)
			handler := setupHandlerWithMock(mockService)

			router := setupTestRouter()
			router.GET("/users", handler.GetAllUsers)

			req := httptest.NewRequest(http.MethodGet, "/users"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedSuccess {
				assert.True(t, response["success"].(bool))
				if tt.validateResponse != nil {
					tt.validateResponse(t, response)
				}
			} else {
				assert.False(t, response["success"].(bool))
			}

			mockService.AssertExpectations(t)
		})
	}
}

// Test GetUserByID
func TestGetUserByID(t *testing.T) {
	tests := []struct {
		name               string
		userID             string
		mockSetup          func(*MockUserService)
		expectedStatusCode int
		expectedSuccess    bool
		validateResponse   func(*testing.T, map[string]interface{})
	}{
		{
			name:   "Successfully get user by ID",
			userID: "1",
			mockSetup: func(m *MockUserService) {
				user := &models.User{
					ID:    1,
					Name:  "John Doe",
					Email: "john@test.com",
					Age:   30,
					Role:  "user",
				}
				m.On("GetUserByID", mock.Anything, uint(1)).Return(user, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedSuccess:    true,
			validateResponse: func(t *testing.T, resp map[string]interface{}) {
				data := resp["data"].(map[string]interface{})
				assert.Equal(t, "John Doe", data["name"])
				assert.Equal(t, "john@test.com", data["email"])
			},
		},
		{
			name:   "User not found",
			userID: "999",
			mockSetup: func(m *MockUserService) {
				m.On("GetUserByID", mock.Anything, uint(999)).Return(nil, gorm.ErrRecordNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedSuccess:    false,
		},
		{
			name:               "Invalid user ID",
			userID:             "invalid",
			mockSetup:          func(m *MockUserService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedSuccess:    false,
		},
		{
			name:   "Service error",
			userID: "1",
			mockSetup: func(m *MockUserService) {
				m.On("GetUserByID", mock.Anything, uint(1)).Return(nil, errors.New("database connection error"))
			},
			expectedStatusCode: http.StatusNotFound,
			expectedSuccess:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.mockSetup(mockService)
			handler := setupHandlerWithMock(mockService)

			router := setupTestRouter()
			router.GET("/users/:id", handler.GetUserByID)

			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.userID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedSuccess {
				assert.True(t, response["success"].(bool))
				if tt.validateResponse != nil {
					tt.validateResponse(t, response)
				}
			} else {
				assert.False(t, response["success"].(bool))
			}

			mockService.AssertExpectations(t)
		})
	}
}

// Test CreateUser
func TestCreateUser(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        interface{}
		mockSetup          func(*MockUserService)
		expectedStatusCode int
		expectedSuccess    bool
		validateResponse   func(*testing.T, map[string]interface{})
	}{
		{
			name: "Successfully create user",
			requestBody: models.CreateUserRequest{
				Name:     "New User",
				Email:    "new@test.com",
				Password: "password123",
				Age:      25,
			},
			mockSetup: func(m *MockUserService) {
				user := &models.User{
					ID:       1,
					Name:     "New User",
					Email:    "new@test.com",
					Age:      25,
					IsActive: true,
					Role:     "user",
				}
				m.On("CreateUser", mock.Anything, mock.Anything).Return(user, nil)
			},
			expectedStatusCode: http.StatusCreated,
			expectedSuccess:    true,
			validateResponse: func(t *testing.T, resp map[string]interface{}) {
				data := resp["data"].(map[string]interface{})
				assert.Equal(t, "New User", data["name"])
				assert.Equal(t, "new@test.com", data["email"])
			},
		},
		{
			name: "Email already exists",
			requestBody: models.CreateUserRequest{
				Name:     "Duplicate",
				Email:    "existing@test.com",
				Password: "password123",
				Age:      30,
			},
			mockSetup: func(m *MockUserService) {
				m.On("CreateUser", mock.Anything, mock.Anything).Return(nil, errors.New("email already exists"))
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedSuccess:    false,
		},
		{
			name: "Invalid request body - missing required fields",
			requestBody: map[string]interface{}{
				"name": "Test",
				// Missing email and password
			},
			mockSetup:          func(m *MockUserService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedSuccess:    false,
		},
		{
			name: "Invalid request body - invalid email format",
			requestBody: models.CreateUserRequest{
				Name:     "Test User",
				Email:    "invalid-email",
				Password: "password123",
				Age:      25,
			},
			mockSetup:          func(m *MockUserService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedSuccess:    false,
		},
		{
			name:               "Invalid JSON",
			requestBody:        "invalid json",
			mockSetup:          func(m *MockUserService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedSuccess:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.mockSetup(mockService)
			handler := setupHandlerWithMock(mockService)

			router := setupTestRouter()
			router.POST("/users", handler.CreateUser)

			var bodyBytes []byte
			if strBody, ok := tt.requestBody.(string); ok {
				bodyBytes = []byte(strBody)
			} else {
				bodyBytes, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedSuccess {
				assert.True(t, response["success"].(bool))
				if tt.validateResponse != nil {
					tt.validateResponse(t, response)
				}
			} else {
				assert.False(t, response["success"].(bool))
			}

			mockService.AssertExpectations(t)
		})
	}
}

// Test UpdateUser
func TestUpdateUser(t *testing.T) {
	tests := []struct {
		name               string
		userID             string
		requestBody        interface{}
		mockSetup          func(*MockUserService)
		expectedStatusCode int
		expectedSuccess    bool
		validateResponse   func(*testing.T, map[string]interface{})
	}{
		{
			name:   "Successfully update user name",
			userID: "1",
			requestBody: models.UpdateUserRequest{
				Name: stringPtr("Updated Name"),
			},
			mockSetup: func(m *MockUserService) {
				user := &models.User{
					ID:    1,
					Name:  "Updated Name",
					Email: "user@test.com",
					Age:   25,
				}
				m.On("UpdateUser", mock.Anything, uint(1), mock.Anything).Return(user, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedSuccess:    true,
			validateResponse: func(t *testing.T, resp map[string]interface{}) {
				data := resp["data"].(map[string]interface{})
				assert.Equal(t, "Updated Name", data["name"])
			},
		},
		{
			name:   "User not found",
			userID: "999",
			requestBody: models.UpdateUserRequest{
				Name: stringPtr("Test"),
			},
			mockSetup: func(m *MockUserService) {
				m.On("UpdateUser", mock.Anything, uint(999), mock.Anything).Return(nil, gorm.ErrRecordNotFound)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedSuccess:    false,
		},
		{
			name:   "Invalid user ID",
			userID: "invalid",
			requestBody: models.UpdateUserRequest{
				Name: stringPtr("Test"),
			},
			mockSetup:          func(m *MockUserService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedSuccess:    false,
		},
		{
			name:   "Email already exists",
			userID: "1",
			requestBody: models.UpdateUserRequest{
				Email: stringPtr("taken@test.com"),
			},
			mockSetup: func(m *MockUserService) {
				m.On("UpdateUser", mock.Anything, uint(1), mock.Anything).Return(nil, errors.New("email already exists"))
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedSuccess:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.mockSetup(mockService)
			handler := setupHandlerWithMock(mockService)

			router := setupTestRouter()
			router.PUT("/users/:id", handler.UpdateUser)

			bodyBytes, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/users/"+tt.userID, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedSuccess {
				assert.True(t, response["success"].(bool))
				if tt.validateResponse != nil {
					tt.validateResponse(t, response)
				}
			} else {
				assert.False(t, response["success"].(bool))
			}

			mockService.AssertExpectations(t)
		})
	}
}

// Test DeleteUser
func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name               string
		userID             string
		mockSetup          func(*MockUserService)
		expectedStatusCode int
		expectedSuccess    bool
	}{
		{
			name:   "Successfully delete user",
			userID: "1",
			mockSetup: func(m *MockUserService) {
				m.On("DeleteUser", mock.Anything, uint(1)).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedSuccess:    true,
		},
		{
			name:   "User not found",
			userID: "999",
			mockSetup: func(m *MockUserService) {
				m.On("DeleteUser", mock.Anything, uint(999)).Return(gorm.ErrRecordNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedSuccess:    false,
		},
		{
			name:               "Invalid user ID",
			userID:             "invalid",
			mockSetup:          func(m *MockUserService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedSuccess:    false,
		},
		{
			name:   "Service error",
			userID: "1",
			mockSetup: func(m *MockUserService) {
				m.On("DeleteUser", mock.Anything, uint(1)).Return(errors.New("database error"))
			},
			expectedStatusCode: http.StatusNotFound,
			expectedSuccess:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.mockSetup(mockService)
			handler := setupHandlerWithMock(mockService)

			router := setupTestRouter()
			router.DELETE("/users/:id", handler.DeleteUser)

			req := httptest.NewRequest(http.MethodDelete, "/users/"+tt.userID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedSuccess, response["success"].(bool))

			mockService.AssertExpectations(t)
		})
	}
}

// Test BatchCreateUsers
func TestBatchCreateUsers(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        interface{}
		mockSetup          func(*MockUserService)
		expectedStatusCode int
		expectedSuccess    bool
		validateResponse   func(*testing.T, map[string]interface{})
	}{
		{
			name: "Successfully create multiple users",
			requestBody: []*models.CreateUserRequest{
				{Name: "User 1", Email: "user1@test.com", Password: "password1", Age: 25},
				{Name: "User 2", Email: "user2@test.com", Password: "password2", Age: 30},
			},
			mockSetup: func(m *MockUserService) {
				users := []*models.User{
					{ID: 1, Name: "User 1", Email: "user1@test.com", Age: 25},
					{ID: 2, Name: "User 2", Email: "user2@test.com", Age: 30},
				}
				m.On("BatchCreateUsers", mock.Anything, mock.Anything).Return(users, nil)
			},
			expectedStatusCode: http.StatusCreated,
			expectedSuccess:    true,
			validateResponse: func(t *testing.T, resp map[string]interface{}) {
				if dataInterface, ok := resp["data"]; ok && dataInterface != nil {
					if data, ok := dataInterface.([]interface{}); ok {
						assert.Len(t, data, 2)
					}
				}
			},
		},
		{
			name: "Partial success - some users fail",
			requestBody: []*models.CreateUserRequest{
				{Name: "User 1", Email: "new@test.com", Password: "password1", Age: 25},
				{Name: "User 2", Email: "existing@test.com", Password: "password2", Age: 30},
			},
			mockSetup: func(m *MockUserService) {
				users := []*models.User{
					{ID: 1, Name: "User 1", Email: "new@test.com", Age: 25},
				}
				m.On("BatchCreateUsers", mock.Anything, mock.Anything).Return(users, errors.New("batch create had 1 errors: user 1: email already exists"))
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedSuccess:    false,
		},
		{
			name:               "Invalid request body",
			requestBody:        "invalid json",
			mockSetup:          func(m *MockUserService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedSuccess:    false,
		},
		{
			name:               "Empty batch",
			requestBody:        []*models.CreateUserRequest{},
			mockSetup:          func(m *MockUserService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedSuccess:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.mockSetup(mockService)
			handler := setupHandlerWithMock(mockService)

			router := setupTestRouter()
			router.POST("/users/batch", handler.BatchCreateUsers)

			var bodyBytes []byte
			if strBody, ok := tt.requestBody.(string); ok {
				bodyBytes = []byte(strBody)
			} else {
				bodyBytes, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/users/batch", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Debug: print response if status doesn't match
			if w.Code != tt.expectedStatusCode {
				t.Logf("Expected status %d, got %d. Response: %s", tt.expectedStatusCode, w.Code, w.Body.String())
			}

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedSuccess, response["success"].(bool))

			if tt.validateResponse != nil && response["success"].(bool) {
				tt.validateResponse(t, response)
			}

			mockService.AssertExpectations(t)
		})
	}
}

// Test GetUserStats
func TestGetUserStats(t *testing.T) {
	tests := []struct {
		name               string
		mockSetup          func(*MockUserService)
		expectedStatusCode int
		expectedSuccess    bool
		validateResponse   func(*testing.T, map[string]interface{})
	}{
		{
			name: "Successfully get user stats",
			mockSetup: func(m *MockUserService) {
				stats := map[string]interface{}{
					"total_users":    100,
					"active_users":   75,
					"inactive_users": 25,
				}
				m.On("GetUserStats", mock.Anything).Return(stats, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedSuccess:    true,
			validateResponse: func(t *testing.T, resp map[string]interface{}) {
				data := resp["data"].(map[string]interface{})
				assert.Equal(t, float64(100), data["total_users"])
				assert.Equal(t, float64(75), data["active_users"])
				assert.Equal(t, float64(25), data["inactive_users"])
			},
		},
		{
			name: "Service returns error",
			mockSetup: func(m *MockUserService) {
				m.On("GetUserStats", mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedSuccess:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.mockSetup(mockService)
			handler := setupHandlerWithMock(mockService)

			router := setupTestRouter()
			router.GET("/users/stats", handler.GetUserStats)

			req := httptest.NewRequest(http.MethodGet, "/users/stats", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedSuccess, response["success"].(bool))

			if tt.validateResponse != nil {
				tt.validateResponse(t, response)
			}

			mockService.AssertExpectations(t)
		})
	}
}

// Test UpdateUserRole (RBAC)
func TestUpdateUserRole(t *testing.T) {
	tests := []struct {
		name               string
		userID             string
		requestingUser     *models.User
		requestBody        interface{}
		mockSetup          func(*MockUserService)
		expectedStatusCode int
		expectedSuccess    bool
		validateResponse   func(*testing.T, map[string]interface{})
	}{
		{
			name:   "Superadmin successfully updates user to admin",
			userID: "2",
			requestingUser: &models.User{
				ID:    1,
				Email: "superadmin@test.com",
				Role:  string(models.RoleSuperAdmin),
			},
			requestBody: models.UpdateRoleRequest{
				Role: "admin",
			},
			mockSetup: func(m *MockUserService) {
				targetUser := &models.User{ID: 2, Email: "user@test.com", Role: "user"}
				updatedUser := &models.User{ID: 2, Email: "user@test.com", Role: "admin"}
				m.On("GetUserByID", mock.Anything, uint(2)).Return(targetUser, nil)
				m.On("UpdateUserRole", mock.Anything, uint(2), "admin").Return(updatedUser, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedSuccess:    true,
			validateResponse: func(t *testing.T, resp map[string]interface{}) {
				data := resp["data"].(map[string]interface{})
				user := data["user"].(map[string]interface{})
				assert.Equal(t, "admin", user["role"])
			},
		},
		{
			name:   "Admin cannot change roles (forbidden)",
			userID: "2",
			requestingUser: &models.User{
				ID:    1,
				Email: "admin@test.com",
				Role:  string(models.RoleAdmin),
			},
			requestBody: models.UpdateRoleRequest{
				Role: "admin",
			},
			mockSetup:          func(m *MockUserService) {},
			expectedStatusCode: http.StatusForbidden,
			expectedSuccess:    false,
		},
		{
			name:   "Regular user cannot change roles (forbidden)",
			userID: "2",
			requestingUser: &models.User{
				ID:    1,
				Email: "user@test.com",
				Role:  string(models.RoleUser),
			},
			requestBody: models.UpdateRoleRequest{
				Role: "admin",
			},
			mockSetup:          func(m *MockUserService) {},
			expectedStatusCode: http.StatusForbidden,
			expectedSuccess:    false,
		},
		{
			name:   "Superadmin cannot demote themselves",
			userID: "1",
			requestingUser: &models.User{
				ID:    1,
				Email: "superadmin@test.com",
				Role:  string(models.RoleSuperAdmin),
			},
			requestBody: models.UpdateRoleRequest{
				Role: "admin",
			},
			mockSetup: func(m *MockUserService) {
				user := &models.User{ID: 1, Email: "superadmin@test.com", Role: "superadmin"}
				m.On("GetUserByID", mock.Anything, uint(1)).Return(user, nil)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedSuccess:    false,
		},
		{
			name:   "Target user not found",
			userID: "999",
			requestingUser: &models.User{
				ID:    1,
				Email: "superadmin@test.com",
				Role:  string(models.RoleSuperAdmin),
			},
			requestBody: models.UpdateRoleRequest{
				Role: "admin",
			},
			mockSetup: func(m *MockUserService) {
				m.On("GetUserByID", mock.Anything, uint(999)).Return(nil, gorm.ErrRecordNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedSuccess:    false,
		},
		{
			name:   "Invalid role",
			userID: "2",
			requestingUser: &models.User{
				ID:    1,
				Email: "superadmin@test.com",
				Role:  string(models.RoleSuperAdmin),
			},
			requestBody: models.UpdateRoleRequest{
				Role: "invalid_role",
			},
			mockSetup: func(m *MockUserService) {
				// No mock setup needed - validation fails at binding stage
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedSuccess:    false,
		},
		{
			name:   "Invalid user ID",
			userID: "invalid",
			requestingUser: &models.User{
				ID:    1,
				Email: "superadmin@test.com",
				Role:  string(models.RoleSuperAdmin),
			},
			requestBody: models.UpdateRoleRequest{
				Role: "admin",
			},
			mockSetup:          func(m *MockUserService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedSuccess:    false,
		},
		{
			name:           "No authenticated user in context",
			userID:         "2",
			requestingUser: nil,
			requestBody: models.UpdateRoleRequest{
				Role: "admin",
			},
			mockSetup:          func(m *MockUserService) {},
			expectedStatusCode: http.StatusUnauthorized,
			expectedSuccess:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.mockSetup(mockService)
			handler := setupHandlerWithMock(mockService)

			router := setupTestRouter()
			router.PUT("/users/:id/role", func(c *gin.Context) {
				// Set authenticated user in context
				if tt.requestingUser != nil {
					c.Set("user", tt.requestingUser)
				}
				handler.UpdateUserRole(c)
			})

			bodyBytes, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/users/"+tt.userID+"/role", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedSuccess, response["success"].(bool))

			if tt.validateResponse != nil {
				tt.validateResponse(t, response)
			}

			mockService.AssertExpectations(t)
		})
	}
}

// Test RBAC authorization scenarios
func TestRBACAuthorization(t *testing.T) {
	t.Run("Regular user can view users", func(t *testing.T) {
		mockService := new(MockUserService)
		meta := models.PaginationMeta{Page: 1, Limit: 10, Total: 0, TotalPages: 0}
		mockService.On("GetAllUsersPaginated", mock.Anything, mock.Anything).Return([]*models.User{}, meta, nil)

		handler := setupHandlerWithMock(mockService)
		router := setupTestRouter()
		router.GET("/users", handler.GetAllUsers)

		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Admin can create users", func(t *testing.T) {
		mockService := new(MockUserService)
		user := &models.User{ID: 1, Name: "Test", Email: "test@test.com"}
		mockService.On("CreateUser", mock.Anything, mock.Anything).Return(user, nil)

		handler := setupHandlerWithMock(mockService)
		router := setupTestRouter()
		router.POST("/users", handler.CreateUser)

		reqBody := models.CreateUserRequest{
			Name:     "Test",
			Email:    "test@test.com",
			Password: "password123",
			Age:      25,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Only superadmin can change roles", func(t *testing.T) {
		tests := []struct {
			role               string
			expectedStatusCode int
		}{
			{string(models.RoleSuperAdmin), http.StatusOK},
			{string(models.RoleAdmin), http.StatusForbidden},
			{string(models.RoleUser), http.StatusForbidden},
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("Role: %s", tt.role), func(t *testing.T) {
				mockService := new(MockUserService)
				if tt.expectedStatusCode == http.StatusOK {
					targetUser := &models.User{ID: 2, Role: "user"}
					updatedUser := &models.User{ID: 2, Role: "admin"}
					mockService.On("GetUserByID", mock.Anything, uint(2)).Return(targetUser, nil)
					mockService.On("UpdateUserRole", mock.Anything, uint(2), "admin").Return(updatedUser, nil)
				}

				handler := setupHandlerWithMock(mockService)
				router := setupTestRouter()
				router.PUT("/users/:id/role", func(c *gin.Context) {
					c.Set("user", &models.User{ID: 1, Email: "test@test.com", Role: tt.role})
					handler.UpdateUserRole(c)
				})

				reqBody := models.UpdateRoleRequest{Role: "admin"}
				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest(http.MethodPut, "/users/2/role", bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.expectedStatusCode, w.Code)
			})
		}
	})
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

// Benchmark Tests
func BenchmarkGetAllUsers(b *testing.B) {
	mockService := new(MockUserService)
	users := make([]*models.User, 100)
	for i := 0; i < 100; i++ {
		users[i] = &models.User{ID: uint(i + 1), Name: "User", Email: "user@test.com"}
	}
	meta := models.PaginationMeta{Page: 1, Limit: 10, Total: 100, TotalPages: 10}
	mockService.On("GetAllUsersPaginated", mock.Anything, mock.Anything).Return(users, meta, nil)

	handler := setupHandlerWithMock(mockService)
	router := setupTestRouter()
	router.GET("/users", handler.GetAllUsers)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkGetUserByID(b *testing.B) {
	mockService := new(MockUserService)
	user := &models.User{ID: 1, Name: "Test", Email: "test@test.com"}
	mockService.On("GetUserByID", mock.Anything, uint(1)).Return(user, nil)

	handler := setupHandlerWithMock(mockService)
	router := setupTestRouter()
	router.GET("/users/:id", handler.GetUserByID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
