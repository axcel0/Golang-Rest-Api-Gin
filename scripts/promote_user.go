package main


import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// User model matching the database schema
type User struct {
	ID       uint   `gorm:"primarykey"`
	Name     string
	Email    string `gorm:"unique"`
	Password string
	Age      int
	Role     string `gorm:"default:user"`
	IsActive bool   `gorm:"default:true"`
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run scripts/promote_user.go <email> <role>")
		fmt.Println("")
		fmt.Println("Example:")
		fmt.Println("  go run scripts/promote_user.go admin@test.com admin")
		fmt.Println("  go run scripts/promote_user.go superadmin@test.com superadmin")
		fmt.Println("")
		fmt.Println("Available roles: user, admin, superadmin")
		os.Exit(1)
	}

	email := os.Args[1]
	role := os.Args[2]

	// Validate role
	validRoles := map[string]bool{
		"user":       true,
		"admin":      true,
		"superadmin": true,
	}

	if !validRoles[role] {
		log.Fatalf("Invalid role: %s. Must be one of: user, admin, superadmin", role)
	}

	// Open database
	db, err := gorm.Open(sqlite.Open("goproject.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Find user
	var user User
	result := db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		log.Fatalf("User not found: %s", email)
	}

	// Update role
	oldRole := user.Role
	user.Role = role
	if err := db.Save(&user).Error; err != nil {
		log.Fatalf("Failed to update user role: %v", err)
	}

	fmt.Printf("✅ Successfully updated user:\n")
	fmt.Printf("   Email: %s\n", user.Email)
	fmt.Printf("   Name: %s\n", user.Name)
	fmt.Printf("   Old Role: %s\n", oldRole)
	fmt.Printf("   New Role: %s\n", user.Role)
	fmt.Printf("\n")
	fmt.Printf("🔑 User can now login with updated role permissions!\n")
}
