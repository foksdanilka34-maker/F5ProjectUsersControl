package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/identity/core"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/identity/repo"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/identity/server"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/nats"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/postgres"
)

func main() {
	// Config
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "identity")
	dbPassword := getEnv("DB_PASSWORD", "identity")
	dbName := getEnv("DB_NAME", "identity")
	grpcPort := getIntEnv("GRPC_PORT", 50051)
	natsURL := getEnv("NATS_URL", "nats://localhost:4222")
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key")
	accessTTL := getDurationEnv("ACCESS_TTL", 15*time.Minute)
	refreshTTL := getDurationEnv("REFRESH_TTL", 7*24*time.Hour)

	// Database
	pool, err := postgres.Connect(context.Background(), &postgres.Config{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		Database: dbName,
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	// NATS
	natsConn, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("failed to connect to NATS: %v", err)
	}
	defer natsConn.Close()

	publisher, err := nats.NewPublisher(natsConn)
	if err != nil {
		log.Fatalf("failed to create NATS publisher: %v", err)
	}

	// Repos
	authRepo := repo.NewAuthRepo(pool)
	profileRepo := repo.NewProfileRepo(pool)

	// Core
	authenticator := core.NewAuthenticator(jwtSecret, accessTTL, refreshTTL)
	authService := core.NewAuthService(authRepo, authenticator, refreshTTL)
	profileService := core.NewProfileService(profileRepo, authService, &publisherAdapter{publisher})

	// Server
	srv := server.NewIdentityServer(authService, profileService)

	shutdown := make(chan struct{})
	go func() {
		if err := srv.Start(fmt.Sprintf(":%d", grpcPort)); err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	log.Println("IdentityService started:", grpcPort)
	<-shutdown
}

// publisherAdapter adapts nats.Publisher to core.EmployeeEventPublisher
type publisherAdapter struct {
	pub *nats.Publisher
}

func (a *publisherAdapter) PublishEmployeeCreated(ctx context.Context, userID int64, fullName string) error {
	return a.pub.PublishEmployeeCreated(ctx, userID, fullName, nil)
}

func (a *publisherAdapter) PublishEmployeeUpdated(ctx context.Context, userID int64, fullName string) error {
	return a.pub.PublishEmployeeUpdated(ctx, userID, fullName, nil)
}

func (a *publisherAdapter) PublishEmployeeDeleted(ctx context.Context, userID int64) error {
	return a.pub.PublishEmployeeDeleted(ctx, userID)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getIntEnv(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		n := 0
		for _, c := range val {
			if c < '0' || c > '9' {
				return defaultVal
			}
			n = n*10 + int(c-'0')
		}
		return n
	}
	return defaultVal
}

func getDurationEnv(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		d, err := time.ParseDuration(val)
		if err == nil {
			return d
		}
	}
	return defaultVal
}
