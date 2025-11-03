package handlers

import (
	"context"
	"net/http"
	"time"

	"Go-Lang-project-01/internal/auth"
	"Go-Lang-project-01/internal/models"
	"Go-Lang-project-01/internal/repository"
	"Go-Lang-project-01/internal/services"
	"Go-Lang-project-01/pkg/logger"
	"Go-Lang-project-01/pkg/utils"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	userRepo     *repository.UserRepository
	jwtManager   *auth.JWTManager
	auditService *services.AuditService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userRepo *repository.UserRepository, jwtManager *auth.JWTManager, auditService *services.AuditService) *AuthHandler {
	return &AuthHandler{
		userRepo:     userRepo,
		jwtManager:   jwtManager,
		auditService: auditService,
	}
}

// Register godoc
// @Summary      Register new user
// @Description  Create a new user account with email and password
// @Tags         authentication
// @Accept       json
// @Produce      json
// @Param        request  body      models.RegisterRequest  true  "Register request"
// @Success      201      {object}  map[string]interface{}  "User registered successfully with tokens"
// @Failure      400      {object}  map[string]interface{}  "Invalid request body"
// @Failure      409      {object}  map[string]interface{}  "Email already registered"
// @Failure      500      {object}  map[string]interface{}  "Internal server error"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if email already exists
	existingUser, _ := h.userRepo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		logger.Warn("Registration failed: email already exists", "email", req.Email)
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"message": "email already registered",
		})
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		logger.Error("Failed to hash password", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to process registration",
		})
		return
	}

	// Create user
	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Age:      req.Age,
		Role:     string(models.RoleUser), // Default role for new registrations
		IsActive: true,
	}

	if err := h.userRepo.Create(ctx, &user); err != nil {
		logger.Error("Failed to create user", "error", err, "email", req.Email)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to create user",
		})
		return
	}

	// Generate tokens
	accessToken, err := h.jwtManager.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		logger.Error("Failed to generate access token", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to generate tokens",
		})
		return
	}

	refreshToken, err := h.jwtManager.GenerateRefreshToken(user.ID, user.Email, user.Role)
	if err != nil {
		logger.Error("Failed to generate refresh token", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to generate tokens",
		})
		return
	}

	logger.Info("User registered successfully", "user_id", user.ID, "email", user.Email)

	// Log audit trail
	h.auditService.LogAuthAction(c, &user.ID, models.AuditActionRegister, true, "")

	// Return response
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "user registered successfully",
		"data": models.LoginResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    24 * 60 * 60, // 24 hours in seconds
			User:         user,
		},
	})
}

// Login godoc
// @Summary      User login
// @Description  Authenticate user with email and password
// @Tags         authentication
// @Accept       json
// @Produce      json
// @Param        request  body      models.LoginRequest     true  "Login credentials"
// @Success      200      {object}  map[string]interface{}  "Login successful with tokens"
// @Failure      400      {object}  map[string]interface{}  "Invalid request body"
// @Failure      401      {object}  map[string]interface{}  "Invalid credentials or inactive account"
// @Failure      500      {object}  map[string]interface{}  "Internal server error"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find user by email
	user, err := h.userRepo.GetByEmail(ctx, req.Email)
	if err != nil || user == nil {
		logger.Warn("Login failed: user not found", "email", req.Email)
		// Log failed login attempt
		h.auditService.LogAuthAction(c, nil, models.AuditActionLoginFailed, false, "User not found")
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "invalid email or password",
		})
		return
	}

	// Check if user is active
	if !user.IsActive {
		logger.Warn("Login failed: user inactive", "email", req.Email)
		h.auditService.LogAuthAction(c, &user.ID, models.AuditActionLoginFailed, false, "Account inactive")
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "account is inactive",
		})
		return
	}

	// Verify password
	if err := auth.CheckPassword(req.Password, user.Password); err != nil {
		logger.Warn("Login failed: invalid password", "email", req.Email)
		h.auditService.LogAuthAction(c, &user.ID, models.AuditActionLoginFailed, false, "Invalid password")
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "invalid email or password",
		})
		return
	}

	// Generate tokens
	accessToken, err := h.jwtManager.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		logger.Error("Failed to generate access token", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to generate tokens",
		})
		return
	}

	refreshToken, err := h.jwtManager.GenerateRefreshToken(user.ID, user.Email, user.Role)
	if err != nil {
		logger.Error("Failed to generate refresh token", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to generate tokens",
		})
		return
	}

	logger.Info("User logged in successfully", "user_id", user.ID, "email", user.Email)
	
	// Log successful login
	h.auditService.LogAuthAction(c, &user.ID, models.AuditActionLogin, true, "")

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "login successful",
		"data": models.LoginResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    24 * 60 * 60, // 24 hours in seconds
			User:         *user,
		},
	})
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Get a new access token using refresh token
// @Tags         authentication
// @Accept       json
// @Produce      json
// @Param        request  body      models.RefreshTokenRequest   true  "Refresh token"
// @Success      200      {object}  map[string]interface{}       "Token refreshed successfully"
// @Failure      400      {object}  map[string]interface{}       "Invalid request body"
// @Failure      401      {object}  map[string]interface{}       "Invalid or expired refresh token"
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Generate new access token
	accessToken, err := h.jwtManager.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		logger.Warn("Token refresh failed", "error", err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "invalid or expired refresh token",
		})
		return
	}
	
	// Validate refresh token to get claims for audit logging
	claims, _ := h.jwtManager.ValidateToken(req.RefreshToken)

	logger.Info("Access token refreshed successfully")
	
	// Log token refresh
	if claims != nil {
		h.auditService.LogAuthAction(c, &claims.UserID, models.AuditActionRefreshToken, true, "")
	}

	// Return new access token
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "token refreshed successfully",
		"data": models.RefreshTokenResponse{
			AccessToken: accessToken,
			TokenType:   "Bearer",
			ExpiresIn:   24 * 60 * 60, // 24 hours in seconds
		},
	})
}

// GetProfile godoc
// @Summary      Get user profile
// @Description  Get the authenticated user's profile information
// @Tags         authentication
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {object}  map[string]interface{}  "User profile"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      404  {object}  map[string]interface{}  "User not found"
// @Router       /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "unauthorized",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get user from database
	user, err := h.userRepo.GetByID(ctx, userID.(uint))
	if err != nil {
		logger.Error("Failed to get user profile", "error", err, "user_id", userID)
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "user not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
	})
}
