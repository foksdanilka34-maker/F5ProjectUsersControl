package main

import (
	"context"
	"fmt"
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
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	storage.Logger.Info("Starting auth service")

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
		storage.Logger.Fatal("postgres connection error", zap.Error(err))
	}
	defer pgPool.Close()
	storage.Logger.Info("Postgres connected")

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
		storage.Logger.Fatal("redis connection error", zap.Error(err))
	}
	defer redisClient.Close()
	storage.Logger.Info("Redis connected")

	credentialStorage := appAuth.NewStorage(pgPool)
	sessionStorage := appSession.NewStorage(pgPool)
	redisCache := appSession.NewRedisCache(redisClient)
	cachedSessionStorage := appSession.NewCachedSessionStorage(sessionStorage, redisCache)

	authenticator, err := appSession.NewAuthenticator(jwtSecret)
	if err != nil {
		storage.Logger.Fatal("failed to create authenticator", zap.Error(err))
	}

	authCore := appCore.NewCore(credentialStorage, cachedSessionStorage, authenticator)

	grpcServerImpl := appServer.NewAuthServer(authCore)

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		storage.Logger.Fatal("failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	grpcServerImpl.Register(grpcServer)

	go func() {
		storage.Logger.Info("gRPC server listening", zap.String("addr", lis.Addr().String()))
		if err := grpcServer.Serve(lis); err != nil {
			storage.Logger.Fatal("failed to serve gRPC", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	storage.Logger.Info("Shutting down gRPC server...")
	grpcServer.GracefulStop()

	storage.Logger.Info("Server gracefully stopped")
}
