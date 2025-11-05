package config

import (
	"log"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/storage"
)

type Config struct {
	Host                 string
	Port                 string
	Environment          string
	LoginServiceHost     string
	LoginServicePort     string
	EmployeeServiceHost  string
	EmployeeServicePort  string
	ProjectServiceHost   string
	ProjectServicePort   string
	AnalyticsServiceHost string
	AnalyticsServicePort string
	JWTSecret            string
}

func NewConfig() *Config {
	return &Config{
		Host:                 storage.GetEnv("GATEWAY_HOST", "0.0.0.0"),
		Port:                 storage.GetEnv("GATEWAY_PORT", "8080"),
		Environment:          storage.GetEnv("ENVIRONMENT", "development"),
		LoginServiceHost:     storage.GetEnv("LOGIN_SERVICE_HOST", "localhost"),
		LoginServicePort:     storage.GetEnv("LOGIN_SERVICE_PORT", "50051"),
		EmployeeServiceHost:  storage.GetEnv("EMPLOYEE_SERVICE_HOST", "localhost"),
		EmployeeServicePort:  storage.GetEnv("EMPLOYEE_SERVICE_PORT", "50052"),
		ProjectServiceHost:   storage.GetEnv("PROJECT_SERVICE_HOST", "localhost"),
		ProjectServicePort:   storage.GetEnv("PROJECT_SERVICE_PORT", "50053"),
		AnalyticsServiceHost: storage.GetEnv("ANALYTICS_SERVICE_HOST", "localhost"),
		AnalyticsServicePort: storage.GetEnv("ANALYTICS_SERVICE_PORT", "50054"),
		JWTSecret:            getJWTSecret(),
	}
}

func getJWTSecret() string {
	secret := storage.GetEnv("JWT_SECRET", "")
	if secret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}
	return secret
}
