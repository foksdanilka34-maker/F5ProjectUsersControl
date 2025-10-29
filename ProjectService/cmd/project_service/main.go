package main

import (
	"context"

	"net"
	"os"
	"os/signal"
	"syscall"

	// projectClient "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project/client"
	// natsclient "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project/client/nats"
	projectCore "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project/core"
	project "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project/repo"
	projectServer "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project/server"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/logger"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/storage"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var log *zap.Logger

func main() {
	var err error
	log, err = logger.New("project-service")
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	log.Info("Starting project service")

	_ = godotenv.Load()

	pgConfig := storage.PostgresConfig{
		User:     storage.GetEnv(log, "PROJECT_DB_USER", "postgres"),
		Password: storage.GetEnv(log, "PROJECT_DB_PASSWORD", ""),

		Host:   storage.GetEnv(log, "PROJECT_DB_HOST", "localhost"),
		Port:   storage.GetEnv(log, "PROJECT_DB_PORT", "5433"),
		DBName: storage.GetEnv(log, "PROJECT_DB", ""),
	}

	redisConfig := storage.RedisConfig{
		Addr:     storage.GetEnv(log, "PROJECT_REDIS_ADDR", "localhost:6380"),
		Password: storage.GetEnv(log, "PROJECT_REDIS_PASSWORD", ""),
		DB:       2,
	}

	listenAddr := storage.GetEnv(log, "GRPC_PROJECT_LISTEN_ADDR", "0.0.0.0:50053")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pgPool, err := storage.NewPostgresPool(ctx, pgConfig)
	if err != nil {
		log.Fatal("postgres connection error", zap.Error(err))
	}
	defer pgPool.Close()
	log.Info("Postgres connected")

	redisClient, err := storage.NewRedisClient(ctx, redisConfig)
	if err != nil {
		log.Fatal("redis connection error", zap.Error(err))
	}
	defer redisClient.Close()
	log.Info("Redis connected for project service")

	projectStorage := project.NewStorage(pgPool)
	projectCache := project.NewCacheStorage(redisClient)

	cachedProjectStorage := project.NewCachedProjectStorage(projectStorage, projectCache)
	// natsConfig := &storage.NatsConfig{
	// 	URL: storage.GetEnv(log, "NATS_URL", "nats://localhost:4222"),
	// }
	// // natsClient, err := storage.NewNATSConnection(*natsConfig)
	// // if err != nil {
	// // 	log.Fatal("nats connection error", zap.Error(err))
	// // }
	// // //publisher := natsclient.NewPublisher(natsClient)

	// loginServiceConn, err := grpc.NewClient(
	// 	storage.GetEnv(log, "EMPLOYEE_SERVICE_GRPC_ADDR", "localhost:50052"),
	// 	grpc.WithTransportCredentials(insecure.NewCredentials()),
	// )

	// if err != nil {
	// 	log.Fatal("Employee service client creation error", zap.Error(err))
	// }
	// defer loginServiceConn.Close()

	// projectClient1, err := projectClient.projectClient(loginServiceConn)
	// if err != nil {
	// 	log.Fatal("failed to create login service client", zap.Error(err))
	// }

	projectCore := projectCore.NewCore(cachedProjectStorage)
	grpcServerImpl := projectServer.NewProjectServer(projectCore)

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

	log.Info("Shutting down gRPC employee server...")
	grpcServer.GracefulStop()

	log.Info("Server gracefully stopped")
}
