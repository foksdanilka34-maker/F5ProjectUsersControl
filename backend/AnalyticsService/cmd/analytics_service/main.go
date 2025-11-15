package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics/cache"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics/client/nats"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics/core"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics/repo"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/AnalyticsService/internal/app/analytics/server"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/eventbus"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/storage"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting analytics service")

	_ = godotenv.Load()

	pgConfig := storage.PostgresConfig{
		User:     storage.GetEnv("ANALYTICS_DB_USER", "postgres"),
		Password: storage.GetEnv("ANALYTICS_DB_PASSWORD", ""),
		Host:     storage.GetEnv("ANALYTICS_DB_HOST", "localhost"),
		Port:     storage.GetEnv("ANALYTICS_DB_PORT", "5436"),
		DBName:   storage.GetEnv("ANALYTICS_DB", "analytics"),
	}

	redisConfig := storage.RedisConfig{
		Addr:     storage.GetEnv("ANALYTICS_REDIS_ADDR", "localhost:6383"),
		Password: storage.GetEnv("ANALYTICS_REDIS_PASSWORD", ""),
		DB:       0,
	}

	listenAddr := storage.GetEnv("GRPC_ANALYTICS_LISTEN_ADDR", "0.0.0.0:50054")
	natsConfig := storage.NatsConfig{
		URL: storage.GetEnv("NATS_URL", "nats://localhost:4222"),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pgPool, err := storage.NewPostgresPool(ctx, pgConfig)
	if err != nil {
		log.Fatalf("postgres connection error: %v", err)
	}
	defer pgPool.Close()
	log.Println("PostgreSQL connected")

	redisClient, err := storage.NewRedisClient(ctx, redisConfig)
	if err != nil {
		log.Fatalf("redis connection error: %v", err)
	}
	defer redisClient.Close()
	log.Println("Redis connected")

	cacheLayer := cache.NewRedisCache(redisClient)

	storageLayer := repo.NewStorage(pgPool)

	analyticsCore := core.NewCore(storageLayer, cacheLayer)

	natsClient, err := storage.NewNATSConnection(natsConfig)
	if err != nil {
		log.Printf("nats connection error: %v (optional, continuing without NATS)", err)
	} else {
		defer natsClient.Close()
		if err := eventbus.EnsureJetStreamStreams(natsClient.JS); err != nil {
			log.Printf("warning: failed to ensure JetStream streams: %v", err)
		}
		log.Println("NATS connected")
		subscriber := nats.NewSubscriber(natsClient.JS, analyticsCore)
		go func() {
			if err := subscriber.Start(ctx); err != nil {
				log.Printf("failed to start NATS subscriber: %v", err)
			}
		}()
	}
	grpcServerImpl := server.NewAnalyticsServer(analyticsCore)

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	grpcServerImpl.Register(grpcServer)

	reflection.Register(grpcServer)
	log.Println("gRPC reflection enabled")

	go func() {
		log.Printf("gRPC server listening on %s", lis.Addr().String())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	log.Println("Shutting down gRPC server...")
	grpcServer.GracefulStop()

	log.Println("Server gracefully stopped")
}
