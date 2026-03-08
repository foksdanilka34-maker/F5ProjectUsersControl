package storage

import (
	"log"
	"os"
)

func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	log.Printf("env not set, using default: key=%s, default=%s", key, defaultValue)
	return defaultValue
}


