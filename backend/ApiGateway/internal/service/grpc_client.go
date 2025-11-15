package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// dialGRPC centralizes gRPC connection bootstrapping so service clients share
// consistent timeouts, transport security, and logging behavior.
func dialGRPC(serviceName, host, port string, timeout time.Duration, extraOpts ...grpc.DialOption) *grpc.ClientConn {
	if host == "" || port == "" {
		log.Fatalf("%s host or port is empty", serviceName)
	}

	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	address := fmt.Sprintf("%s:%s", host, port)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}
	opts = append(opts, extraOpts...)

	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		log.Fatalf("failed to connect to %s at %s: %v", serviceName, address, err)
	}

	log.Printf("Successfully connected to %s at %s", serviceName, address)
	return conn
}
