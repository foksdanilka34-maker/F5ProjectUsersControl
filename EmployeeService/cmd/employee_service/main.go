package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	employeeClient "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee/client"
	natsclient "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee/client/nats"
	employeeCore "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee/core"
	employee "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee/repo"
	employeeServer "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee/server"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/storage"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting employee service")

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
		log.Fatalf("postgres connection error: %v", err)
	}
	defer pgPool.Close()
	log.Println("Postgres connected")

	redisClient, err := storage.NewRedisClient(ctx, redisConfig)
	if err != nil {
		log.Fatalf("redis connection error: %v", err)
	}
	defer redisClient.Close()
	log.Println("Redis connected for employee service")

	employeeStorage := employee.NewStorage(pgPool)
	referenceCache := employee.NewReferenceCache(redisClient)
	cachedEmployeeStorage := employee.NewCachedEmployeeStorage(employeeStorage, referenceCache)
	natsConfig := &storage.NatsConfig{
		URL: storage.GetEnv("NATS_URL", "nats://localhost:4222"),
	}
	natsClient, err := storage.NewNATSConnection(*natsConfig)
	if err != nil {
		log.Fatalf("nats connection error: %v", err)
	}
	publisher := natsclient.NewPublisher(natsClient)

	loginServiceConn, err := grpc.NewClient(
		storage.GetEnv("LOGIN_SERVICE_GRPC_ADDR", "localhost:50051"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatalf("Login service client creation error: %v", err)
	}
	defer loginServiceConn.Close()

	employeeClien, err := employeeClient.NewAuthClient(loginServiceConn)
	if err != nil {
		log.Fatalf("failed to create login service client: %v", err)
	}

	employeeCpre := employeeCore.NewCore(cachedEmployeeStorage, employeeClien, publisher)
	grpcServerImpl := employeeServer.NewEmployeeServer(employeeCpre)

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

	log.Println("Shutting down gRPC employee server...")
	grpcServer.GracefulStop()

	log.Println("Server gracefully stopped")
}
