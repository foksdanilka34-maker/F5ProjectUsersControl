package main

import (
	"context"

	"net"
	"os"
	"os/signal"
	"syscall"

	natsclient "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project/client/nats"
	projectCore "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project/core"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project/client/nats"
	project "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project/repo"
	projectServer "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project/server"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/storage"

	"log"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

	listenAddr := storage.GetEnv("GRPC_PROJECT_LISTEN_ADDR", "0.0.0.0:50053")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pgPool, err := storage.NewPostgresPool(ctx, pgConfig)
	if err != nil {
		log.Fatalf("postgres connection error: %v", err)
	}
	defer pgPool.Close()
	log.Println("Postgres connected")


	projectStorage := project.NewStorage(pgPool)

	natsConfig := &storage.NatsConfig{
		URL: storage.GetEnv("NATS_URL", "nats://localhost:4222"),
	}

	natsClient, err := storage.NewNATSConnection(*natsConfig)
	if err != nil {
		log.Fatalf("nats connection error: %v", err)
	}
	defer natsClient.Close()
	log.Println("NATS connected")

	subscriber := natsclient.NewSubscriber(natsClient, projectStorage)
	if err := subscriber.Start(ctx); err != nil {
		log.Fatalf("failed to start NATS subscriber: %v", err)
	}
	log.Println("NATS subscriber started")

	publisher := nats.NewPublisher(natsClient)
	log.Println("NATS publisher initialized")

	projectCore := projectCore.NewCore(*projectStorage, publisher)
	grpcServerImpl := projectServer.NewProjectServer(projectCore)

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
