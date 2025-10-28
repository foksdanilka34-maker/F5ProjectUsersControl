package main

import (
	"context"

	"net"
	"os"
	"os/signal"
	"syscall"

	employeeClient "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee/client"
	natsclient "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee/client/nats"
	employeeCore "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee/core"
	employee "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee/repo"
	employeeServer "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee/server"

	"google.golang.org/grpc/credentials/insecure"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/storage"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	storage.Logger.Info("Starting employee service")

	_ = godotenv.Load()

	pgConfig := storage.PostgresConfig{
		User:     storage.GetEnv("EMPL_DB_USER", "postgres"),
		Password: storage.GetEnv("EMPL_DB_PASSWORD", ""),

		Host:   storage.GetEnv("EMPL_DB_HOST", "localhost"),
		Port:   storage.GetEnv("EMPL_DB_PORT", "5433"),
		DBName: storage.GetEnv("EMPL_DB", ""),
	}

	redisConfig := storage.RedisConfig{
		Addr:     storage.GetEnv("EMPL_REDIS_ADDR", "localhost:6380"),
		Password: storage.GetEnv("EMPL_REDIS_PASSWORD", ""),
		DB:       1, 
	}

	listenAddr := storage.GetEnv("GRPC_EMPL_LISTEN_ADDR", "0.0.0.0:50052")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pgPool, err := storage.NewPostgresPool(ctx, pgConfig)
	if err != nil {
		storage.Logger.Fatal("postgres connection error", zap.Error(err))
	}
	defer pgPool.Close()
	storage.Logger.Info("Postgres connected")

	redisClient, err := storage.NewRedisClient(ctx, redisConfig)
	if err != nil {
		storage.Logger.Fatal("redis connection error", zap.Error(err))
	}
	defer redisClient.Close()
	storage.Logger.Info("Redis connected for employee service")

	employeeStorage := employee.NewStorage(pgPool)
	referenceCache := employee.NewReferenceCache(redisClient)
	cachedEmployeeStorage := employee.NewCachedEmployeeStorage(employeeStorage, referenceCache)
	natsConfig := &storage.NatsConfig{
		URL: storage.GetEnv("NATS_URL", "nats://localhost:4222"),
	}
	natsClient, err := storage.NewNATSConnection(*natsConfig)
	if err != nil {
		storage.Logger.Fatal("nats connection error", zap.Error(err))
	}
	publisher := natsclient.NewPublisher(natsClient)

	loginServiceConn, err := grpc.NewClient(
		storage.GetEnv("LOGIN_SERVICE_GRPC_ADDR", "localhost:50051"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		storage.Logger.Fatal("Login service client creation error", zap.Error(err))
	}
	defer loginServiceConn.Close()

	employeeClien, err := employeeClient.NewAuthClient(loginServiceConn)
	if err != nil {
		storage.Logger.Fatal("failed to create login service client", zap.Error(err))
	}

	employeeCpre := employeeCore.NewCore(cachedEmployeeStorage, employeeClien, publisher)
	grpcServerImpl := employeeServer.NewEmployeeServer(employeeCpre)

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

	storage.Logger.Info("Shutting down gRPC employee server...")
	grpcServer.GracefulStop()

	storage.Logger.Info("Server gracefully stopped")
}
