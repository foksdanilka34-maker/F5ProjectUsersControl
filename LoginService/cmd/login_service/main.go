package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/storage"
	auth "github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/app"
	
	"google.golang.org/grpc"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Starting auth service...")


	postgresURL := "postgres://postgres:1234@localhost:5432/postgres?sslmode=disable" 
	redisAddr := "localhost:6379"                                                    
	jwtSecret := "your-super-secret-key-that-is-at-least-32-chars-long"        
	listenAddr := "dg"       

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pgPool, err := storage.NewPostgresPool(ctx, postgresURL)
	if err != nil {
		log.Fatalf("postgres connection error: %v", err)
	}
	defer pgPool.Close()
	log.Println("Postgres connected")

	redisClient, err := storage.NewRedisClient(ctx, redisAddr) 
	if err != nil {
		log.Fatalf("redis connection error: %v", err)
	}
	defer redisClient.Close()
	log.Println("Redis connected")

	postgresStorage := auth.NewStorage(pgPool)
	redisCache := auth.NewRedisCache(redisClient)
	cachedSessionStorage := auth.NewCachedSessionStorage(postgresStorage, redisCache)

	authenticator, err := auth.NewAuthenticator(jwtSecret)
	if err != nil {
		log.Fatalf("failed to create authenticator: %v", err)
	}

	authCore := auth.NewCore(postgresStorage, cachedSessionStorage, authenticator)

	grpcServerImpl := auth.NewAuthServer(authCore)

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