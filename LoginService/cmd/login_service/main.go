package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	appAuth "github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/app/auth"
	appCore "github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/app/core"
	appServer "github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/app/server"
	appSession "github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/app/session"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/storage"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Starting auth service...")

	_ = godotenv.Load()

	pgConfig := storage.PostgresConfig{
		User:     storage.GetEnv("AUTH_DB_USER", "postgres"),
		Password: storage.GetEnv("AUTH_DB_PASSWORD", ""),

		Host:   storage.GetEnv("AUTH_DB_HOST", "localhost"),
		Port:   storage.GetEnv("AUTH_DB_PORT", "5432"),
		DBName: storage.GetEnv("AUTH_DB", ""),
	}

	redisAddr := storage.GetEnv("REDIS_ADDR", "localhost:6379")
	jwtSecret := storage.GetEnv("JWT_SECRET", "")
	listenAddr := storage.GetEnv("GRPC_LISTEN_ADDR", "0.0.0.0:50051")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pgPool, err := storage.NewPostgresPool(ctx, pgConfig)
	if err != nil {
		log.Fatalf("postgres connection error: %v", err)
	}
	defer pgPool.Close()
	log.Println("Postgres connected")

	redisPassword := storage.GetEnv("REDIS_PASSWORD", "")
	redisDB := storage.GetEnv("REDIS_DB", "0")
	redisDBInt := 0
	if redisDB != "" {
		fmt.Sscanf(redisDB, "%d", &redisDBInt)
	}

	redisConfig := storage.RedisConfig{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDBInt,
	}
	redisClient, err := storage.NewRedisClient(ctx, redisConfig)
	if err != nil {
		log.Fatalf("redis connection error: %v", err)
	}
	defer redisClient.Close()
	log.Println("Redis connected")

	credentialStorage := appAuth.NewStorage(pgPool)
	sessionStorage := appSession.NewStorage(pgPool)
	redisCache := appSession.NewRedisCache(redisClient)
	cachedSessionStorage := appSession.NewCachedSessionStorage(sessionStorage, redisCache)

	authenticator, err := appSession.NewAuthenticator(jwtSecret)
	if err != nil {
		log.Fatalf("failed to create authenticator: %v", err)
	}

	authCore := appCore.NewCore(credentialStorage, cachedSessionStorage, authenticator)

	grpcServerImpl := appServer.NewAuthServer(authCore)

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	grpcServerImpl.Register(grpcServer)

	go func() {
		log.Printf("gRPC server listening at %v", lis.Addr())
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
