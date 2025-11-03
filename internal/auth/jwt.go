// Package auth provides authentication and authorization utilities including
// JWT token management, password hashing, and token validation.
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Common JWT-related errors returned by the JWT manager.
var (
	// ErrInvalidToken is returned when a token cannot be parsed or validated.
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken is returned when a token's expiration time has passed.
	ErrExpiredToken = errors.New("token has expired")
)

// JWTClaims represents the custom claims embedded in JWT tokens.
// It extends the standard JWT registered claims with user-specific information.
type JWTClaims struct {
	UserID uint   `json:"user_id"` // User's unique identifier
	Email  string `json:"email"`   // User's email address
	Role   string `json:"role"`    // User's role for RBAC (user, admin, superadmin)
	jwt.RegisteredClaims
}

// JWTManager manages JWT token operations including generation and validation.
// It handles both access tokens (short-lived) and refresh tokens (long-lived).
type JWTManager struct {
	secretKey            string        // Secret key for signing tokens
	accessTokenDuration  time.Duration // Lifetime of access tokens
	refreshTokenDuration time.Duration // Lifetime of refresh tokens
}

// NewJWTManager creates a new JWT manager with the specified configuration.
// The secret key should be a strong, randomly generated string.
// Access tokens are typically short-lived (minutes to hours).
// Refresh tokens are long-lived (days to weeks).
func NewJWTManager(secretKey string, accessDuration, refreshDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:            secretKey,
		accessTokenDuration:  accessDuration,
		refreshTokenDuration: refreshDuration,
	}
}

// GenerateAccessToken generates a new JWT access token for the given user.
// Access tokens are short-lived and used for API authentication.
// Returns the signed token string or an error if generation fails.
func (m *JWTManager) GenerateAccessToken(userID uint, email, role string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

// GenerateRefreshToken generates a new refresh token
func (m *JWTManager) GenerateRefreshToken(userID uint, email, role string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.refreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

// ValidateToken validates a JWT token and returns the claims
func (m *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return []byte(m.secretKey), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Check if token is expired
	if claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrExpiredToken
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token from a valid refresh token
func (m *JWTManager) RefreshAccessToken(refreshToken string) (string, error) {
	claims, err := m.ValidateToken(refreshToken)
	if err != nil {
		return "", err
	}

	// Generate new access token with same user info
	return m.GenerateAccessToken(claims.UserID, claims.Email, claims.Role)
}
