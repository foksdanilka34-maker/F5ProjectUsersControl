package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	appAuth "github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/app/auth"
	natsAuth "github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/app/client/nats"
	appCore "github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/app/core"
	appServer "github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/app/server"
	appSession "github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/app/session"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/logger"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/storage"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var log *zap.Logger

func main() {
	var err error
	log, err = logger.New("login-service")
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	log.Info("Starting auth service")

	_ = godotenv.Load()

	pgConfig := storage.PostgresConfig{
		User:     storage.GetEnv(log, "AUTH_DB_USER", "postgres"),
		Password: storage.GetEnv(log, "AUTH_DB_PASSWORD", ""),

		Host:   storage.GetEnv(log, "AUTH_DB_HOST", "localhost"),
		Port:   storage.GetEnv(log, "AUTH_DB_PORT", "5432"),
		DBName: storage.GetEnv(log, "AUTH_DB", ""),
	}

	redisAddr := storage.GetEnv(log, "REDIS_ADDR", "localhost:6379")
	jwtSecret := storage.GetEnv(log, "JWT_SECRET", "")
	listenAddr := storage.GetEnv(log, "GRPC_LISTEN_ADDR", "0.0.0.0:50051")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pgPool, err := storage.NewPostgresPool(ctx, pgConfig)
	if err != nil {
		log.Fatal("postgres connection error", zap.Error(err))
	}
	defer pgPool.Close()
	log.Info("Postgres connected")

	redisPassword := storage.GetEnv(log, "REDIS_PASSWORD", "")
	redisDB := storage.GetEnv(log, "REDIS_DB", "0")
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
		log.Fatal("redis connection error", zap.Error(err))
	}
	defer redisClient.Close()
	log.Info("Redis connected")

	credentialStorage := appAuth.NewStorage(pgPool)
	sessionStorage := appSession.NewStorage(pgPool)
	redisCache := appSession.NewRedisCache(redisClient)
	cachedSessionStorage := appSession.NewCachedSessionStorage(sessionStorage, redisCache)

	authenticator, err := appSession.NewAuthenticator(jwtSecret)
	if err != nil {
		log.Fatal("failed to create authenticator", zap.Error(err))
	}

	authCore := appCore.NewCore(credentialStorage, cachedSessionStorage, authenticator)

	natsConfig := &storage.NatsConfig{
		URL: storage.GetEnv(log, "NATS_URL", "nats://localhost:4222"),
	}

	natsClient, err := storage.NewNATSConnection(*natsConfig)
	if err != nil {
		log.Fatal("nats connection error", zap.Error(err))
	}
	natsConnection := natsAuth.NewNatsConn(natsClient, authCore)
	go natsConnection.Start()

	grpcServerImpl := appServer.NewAuthServer(authCore)

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	grpcServerImpl.Register(grpcServer)

	go func() {
		log.Info("gRPC server listening", zap.String("addr", lis.Addr().String()))
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("failed to serve gRPC", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	log.Info("Shutting down gRPC server...")
	grpcServer.GracefulStop()

	log.Info("Server gracefully stopped")
}
