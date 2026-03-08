package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/logs/core"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/logs/repo"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/nats"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/postgres"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5434")
	dbUser := getEnv("DB_USER", "logs")
	dbPassword := getEnv("DB_PASSWORD", "logs")
	dbName := getEnv("DB_NAME", "logs")
	natsURL := getEnv("NATS_URL", "nats://localhost:4222")

	pool, err := postgres.Connect(ctx, &postgres.Config{
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

	natsConn, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("failed to connect to NATS: %v", err)
	}
	defer natsConn.Close()

	subscriber, err := nats.NewSubscriber(natsConn)
	if err != nil {
		log.Fatalf("failed to create NATS subscriber: %v", err)
	}

	logsRepo := repo.NewLogRepo(pool)
	logsService := core.NewLogService(logsRepo)

	handler := func(ctx context.Context, entry *nats.LogEntry) error {
		return logsService.HandleLogEntry(ctx, &core.NATSLogEntry{
			Service:   entry.Service,
			Level:     entry.Level,
			Message:   entry.Message,
			Timestamp: entry.Timestamp,
			Data:      entry.Data,
		})
	}
	if err := subscriber.SubscribeLogs(ctx, handler); err != nil {
		log.Fatalf("failed to subscribe to logs: %v", err)
	}

	log.Println("LogService started, listening for logs...")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("Shutting down LogService...")
	cancel()
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}


