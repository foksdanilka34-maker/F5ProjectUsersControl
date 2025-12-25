package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/core"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/repo"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/server"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/nats"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/postgres"
)

func main() {
	// Config
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5433")
	dbUser := getEnv("DB_USER", "business")
	dbPassword := getEnv("DB_PASSWORD", "business")
	dbName := getEnv("DB_NAME", "business")
	grpcPort := getIntEnv("GRPC_PORT", 50052)
	natsURL := getEnv("NATS_URL", "nats://localhost:4222")

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

	subscriber, err := nats.NewSubscriber(natsConn)
	if err != nil {
		log.Fatalf("failed to create NATS subscriber: %v", err)
	}

	// Repos
	projectRepo := repo.NewProjectRepo(pool)
	taskRepo := repo.NewTaskRepo(pool)
	analyticsRepo := repo.NewAnalyticsRepo(pool)

	// Core
	projectService := core.NewProjectService(projectRepo)
	taskService := core.NewTaskService(taskRepo)
	analyticsService := core.NewAnalyticsService(analyticsRepo)

	// Subscribe to employee events
	eventHandler := NewEmployeeEventHandler(pool)
	if err := subscriber.SubscribeEmployeeEvents(context.Background(), eventHandler.HandleEmployeeEvent); err != nil {
		log.Fatalf("failed to subscribe to employee events: %v", err)
	}

	// Server
	srv := server.NewBusinessServer(projectService, taskService, analyticsService)

	shutdown := make(chan struct{})
	go func() {
		if err := srv.Start(fmt.Sprintf(":%d", grpcPort)); err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	log.Println("BusinessService started:", grpcPort)
	<-shutdown
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
