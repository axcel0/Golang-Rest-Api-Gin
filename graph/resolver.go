package graph

import (
	"Go-Lang-project-01/internal/auth"
	"Go-Lang-project-01/internal/repository"
	"Go-Lang-project-01/internal/services"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	UserService *services.UserService
	UserRepo    *repository.UserRepository
	JWTManager  *auth.JWTManager
}
