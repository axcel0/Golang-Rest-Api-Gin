// Package utils provides common utility functions including ID generation,
// response helpers, and validation utilities for the application.
package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateID generates a random unique ID
func GenerateID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}
