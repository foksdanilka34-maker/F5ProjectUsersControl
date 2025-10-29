package storage

import (
	"os"

	"go.uber.org/zap"
)

func GetEnv(logger *zap.Logger, key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	if logger != nil {
		logger.Info("env not set, using default", 
			zap.String("key", key), 
			zap.String("default", defaultValue))
	}
	return defaultValue
}
