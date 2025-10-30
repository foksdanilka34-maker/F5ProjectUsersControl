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

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/storage"

	"log"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting project service")

	_ = godotenv.Load()

	pgConfig := storage.PostgresConfig{
		User:     storage.GetEnv("PROJECT_DB_USER", "postgres"),
		Password: storage.GetEnv("PROJECT_DB_PASSWORD", ""),

		Host:   storage.GetEnv("PROJECT_DB_HOST", "localhost"),
		Port:   storage.GetEnv("PROJECT_DB_PORT", "5433"),
		DBName: storage.GetEnv("PROJECT_DB", ""),
	}

	redisConfig := storage.RedisConfig{
		Addr:     storage.GetEnv("PROJECT_REDIS_ADDR", "localhost:6380"),
		Password: storage.GetEnv("PROJECT_REDIS_PASSWORD", ""),
		DB:       2,
	}

	listenAddr := storage.GetEnv("GRPC_PROJECT_LISTEN_ADDR", "0.0.0.0:50053")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pgPool, err := storage.NewPostgresPool(ctx, pgConfig)
	if err != nil {
		log.Fatalf("postgres connection error: %v", err)
	}
	defer pgPool.Close()
	log.Println("Postgres connected")

	redisClient, err := storage.NewRedisClient(ctx, redisConfig)
	if err != nil {
		log.Fatalf("redis connection error: %v", err)
	}
	defer redisClient.Close()
	log.Println("Redis connected for project service")

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
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	grpcServerImpl.Register(grpcServer)

	go func() {
		log.Printf("gRPC server listening on %s", lis.Addr().String())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	log.Println("Shutting down gRPC employee server...")
	grpcServer.GracefulStop()

	log.Println("Server gracefully stopped")
}
