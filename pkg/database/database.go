// Package database provides database connection management and initialization
// using GORM ORM with SQLite driver for persistent data storage.
package database

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect establishes database connection using SQLite
func Connect() error {
	var err error
	DB, err = gorm.Open(sqlite.Open("goproject.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("✅ Database connected successfully! (SQLite)")
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
