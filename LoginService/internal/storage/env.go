package storage

import (
	"log"
	"os"
)

func GetEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
	log.Printf("unable to use .env value")
    return defaultValue
}