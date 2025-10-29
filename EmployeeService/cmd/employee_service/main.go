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

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/logger"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/storage"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var log *zap.Logger

func main() {
	var err error
	log, err = logger.New("employee-service")
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	log.Info("Starting employee service")

	_ = godotenv.Load()

	pgConfig := storage.PostgresConfig{
		User:     storage.GetEnv(log, "EMPL_DB_USER", "postgres"),
		Password: storage.GetEnv(log, "EMPL_DB_PASSWORD", ""),

		Host:   storage.GetEnv(log, "EMPL_DB_HOST", "localhost"),
		Port:   storage.GetEnv(log, "EMPL_DB_PORT", "5433"),
		DBName: storage.GetEnv(log, "EMPL_DB", ""),
	}

	redisConfig := storage.RedisConfig{
		Addr:     storage.GetEnv(log, "EMPL_REDIS_ADDR", "localhost:6380"),
		Password: storage.GetEnv(log, "EMPL_REDIS_PASSWORD", ""),
		DB:       1,
	}

	listenAddr := storage.GetEnv(log, "GRPC_EMPL_LISTEN_ADDR", "0.0.0.0:50052")

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
	log.Info("Redis connected for employee service")

	employeeStorage := employee.NewStorage(pgPool)
	referenceCache := employee.NewReferenceCache(redisClient)
	cachedEmployeeStorage := employee.NewCachedEmployeeStorage(employeeStorage, referenceCache)
	natsConfig := &storage.NatsConfig{
		URL: storage.GetEnv(log, "NATS_URL", "nats://localhost:4222"),
	}
	natsClient, err := storage.NewNATSConnection(*natsConfig)
	if err != nil {
		log.Fatal("nats connection error", zap.Error(err))
	}
	publisher := natsclient.NewPublisher(natsClient)

	loginServiceConn, err := grpc.NewClient(
		storage.GetEnv(log, "LOGIN_SERVICE_GRPC_ADDR", "localhost:50051"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatal("Login service client creation error", zap.Error(err))
	}
	defer loginServiceConn.Close()

	employeeClien, err := employeeClient.NewAuthClient(loginServiceConn)
	if err != nil {
		log.Fatal("failed to create login service client", zap.Error(err))
	}

	employeeCpre := employeeCore.NewCore(cachedEmployeeStorage, employeeClien, publisher)
	grpcServerImpl := employeeServer.NewEmployeeServer(employeeCpre)

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
