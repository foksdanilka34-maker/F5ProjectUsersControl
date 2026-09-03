package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/consumer"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/core"
	bizHttp "github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/http"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/http/middleware"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/outbox"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/repo"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/crypto"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/postgres"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/rabbitmq"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	httpAddr := getEnv("HTTP_ADDR", ":8082")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5435")
	dbUser := getEnv("DB_USER", "business")
	dbPassword := getEnv("DB_PASSWORD", "business")
	dbName := getEnv("DB_NAME", "business")
	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key")
	gitlabTokenKey := getEnv("GITLAB_TOKEN_KEY", jwtSecret)
	publicURL := getEnv("PUBLIC_BASE_URL", "http://localhost:8080")

	log.Printf("Starting Business Service on %s...", httpAddr)

	// 1. Connect Postgres
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

	// 2. Connect RabbitMQ
	rabbit, err := rabbitmq.Connect(rabbitURL)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}
	defer rabbit.Close()

	// 3. Initialize Repositories & Services
	txManager := repo.NewTxManager(pool)
	projectRepo := repo.NewProjectRepo(pool)
	taskRepo := repo.NewTaskRepo(pool)
	analyticsRepo := repo.NewAnalyticsRepo(pool)
	gitlabRepo := repo.NewGitLabRepo(pool)
	extensionRepo := repo.NewExtensionRepo(pool)

	sealer, err := crypto.NewSealer(gitlabTokenKey)
	if err != nil {
		log.Fatalf("failed to init token sealer: %v", err)
	}

	wsHub := bizHttp.NewWSHub()
	go wsHub.Run(ctx)

	projectService := core.NewProjectService(projectRepo, txManager)
	taskService := core.NewTaskService(taskRepo, projectRepo, txManager, wsHub)
	analyticsService := core.NewAnalyticsService(analyticsRepo)
	gitlabService := core.NewGitLabService(gitlabRepo, taskRepo, txManager, sealer, wsHub, publicURL)
	gitlabWebhooks := core.NewGitLabWebhookService(gitlabRepo, taskService, txManager, wsHub, gitlabService)
	extensionService := core.NewExtensionService(extensionRepo, taskRepo, txManager, sealer)

	// 4. Start Business Outbox Poller
	bizPoller := outbox.NewPoller(txManager, rabbit, wsHub, outbox.PollerConfig{
		PollInterval: 500 * time.Millisecond,
		BatchSize:    20,
		WorkerCount:  4,
	})
	go bizPoller.Start(ctx)

	// 5. Start GitLab Webhook Worker
	go gitlabWebhooks.RunWorker(ctx, time.Second, 20)

	// 6. Start RabbitMQ Consumer Pools
	empConsumer := consumer.NewEmployeeConsumer(rabbit, txManager)
	if err := empConsumer.Start(ctx, 4); err != nil {
		log.Fatalf("failed to start rabbitmq consumer pool: %v", err)
	}

	extConsumer := consumer.NewExtensionConsumer(rabbit, txManager, sealer)
	if err := extConsumer.Start(ctx, 4); err != nil {
		log.Fatalf("failed to start extension dispatch consumer: %v", err)
	}

	// 7. Setup HTTP Routes & Middlewares
	jwtValidator := middleware.NewJWTValidator(jwtSecret)
	mux := http.NewServeMux()

	bizHttp.NewProjectHandler(mux, projectService, jwtValidator)
	bizHttp.NewTaskHandler(mux, taskService, jwtValidator)
	bizHttp.NewAnalyticsHandler(mux, analyticsService, jwtValidator)
	bizHttp.NewGitLabHandler(mux, gitlabService, gitlabWebhooks, jwtValidator)
	bizHttp.NewExtensionHandler(mux, extensionService, jwtValidator)

	// WebSocket Route
	mux.HandleFunc("GET /ws", wsHub.HandleWS)

	startTime := time.Now()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"service": "business",
			"uptime":  time.Since(startTime).Truncate(time.Second).String(),
		})
	})

	rateLimiter := middleware.NewTokenBucketLimiter(150, 300)

	handler := middleware.RequestID(
		middleware.CORS(
			middleware.RateLimiter(rateLimiter)(mux),
		),
	)

	server := &http.Server{
		Addr:         httpAddr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("Business Service HTTP server listening on %s", httpAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down Business Service gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	log.Println("Business Service exited cleanly")
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
