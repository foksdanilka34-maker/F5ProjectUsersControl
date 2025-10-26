package storage

import (
	"os"

	"go.uber.org/zap"
)

func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	Logger.Info("env not set, using default", zap.String("key", key), zap.String("default", defaultValue))
	return defaultValue
}
