package storage

import (
	"log"
	"os"
)

func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	log.Printf("env: %s not set, using default", key)
	return defaultValue
}
