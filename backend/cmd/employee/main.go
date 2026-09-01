package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/employee/core"
	empHttp "github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/employee/http"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/employee/http/middleware"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/employee/outbox"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/employee/repo"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/postgres"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/rabbitmq"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	httpAddr := getEnv("HTTP_ADDR", ":8081")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "identity")
	dbPassword := getEnv("DB_PASSWORD", "identity")
	dbName := getEnv("DB_NAME", "identity")
	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key")
	accessTTL := getDurationEnv("ACCESS_TTL", 15*time.Minute)
	refreshTTL := getDurationEnv("REFRESH_TTL", 7*24*time.Hour)

	log.Printf("Starting Employee Service on %s...", httpAddr)

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
	authRepo := repo.NewAuthRepo(pool)
	profileRepo := repo.NewProfileRepo(pool)
	orgRepo := repo.NewOrgRepo(pool)

	authService := core.NewAuthService(authRepo, txManager, jwtSecret, accessTTL, refreshTTL)
	profileService := core.NewProfileService(profileRepo, txManager)
	orgService := core.NewOrgService(orgRepo)

	// 4. Start Outbox Poller Worker Pool
	poller := outbox.NewPoller(txManager, rabbit, outbox.PollerConfig{
		PollInterval: 500 * time.Millisecond,
		BatchSize:    20,
		WorkerCount:  4,
	})
	go poller.Start(ctx)

	// 5. Setup HTTP Routes & Middlewares
	mux := http.NewServeMux()

	empHttp.NewAuthHandler(mux, authService)
	empHttp.NewProfileHandler(mux, profileService, authService)
	empHttp.NewOrgHandler(mux, orgService, authService)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"employee"}`))
	})

	rateLimiter := middleware.NewTokenBucketLimiter(100, 200) // 100 req/sec, burst 200

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
		log.Printf("Employee Service HTTP server listening on %s", httpAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down Employee Service gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	log.Println("Employee Service exited cleanly")
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
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
