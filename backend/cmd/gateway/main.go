package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/gateway/docs"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/gateway/handlers"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/gateway/middleware"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/gateway/service"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/gateway/websocket"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

func main() {

	httpAddr := getEnv("HTTP_ADDR", ":8080")
	identityAddr := getEnv("IDENTITY_SERVICE_ADDR", "localhost:50051")
	businessAddr := getEnv("BUSINESS_SERVICE_ADDR", "localhost:50052")
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key")

	wsHub := websocket.NewHub()
	go wsHub.Run()

	grpcOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             5 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(10 * 1024 * 1024)),
	}

	identityConn, err := grpc.NewClient(identityAddr, grpcOpts...)
	if err != nil {
		log.Fatalf("failed to connect to identity service: %v", err)
	}
	defer identityConn.Close()

	businessConn, err := grpc.NewClient(businessAddr, grpcOpts...)
	if err != nil {
		log.Fatalf("failed to connect to business service: %v", err)
	}
	defer businessConn.Close()

	clients := service.NewClients(identityConn, businessConn)

	authHandler := handlers.NewAuthHTTPHandler(clients.Identity)
	profileHandler := handlers.NewProfileHTTPHandler(clients.Identity)
	projectHandler := handlers.NewProjectHTTPHandler(clients.Business)
	taskHandler := handlers.NewTaskHTTPHandler(clients.Business, wsHub)
	analyticsHandler := handlers.NewAnalyticsHTTPHandler(clients.Business)

	mux := http.NewServeMux()

	authHandler.RegisterRoutes(mux)

	profileHandler.RegisterRoutes(mux)
	projectHandler.RegisterRoutes(mux)
	taskHandler.RegisterRoutes(mux)
	analyticsHandler.RegisterRoutes(mux)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.Handle("/swagger/", docs.SwaggerHandler())
	mux.Handle("/swagger", docs.SwaggerHandler())

	mux.HandleFunc("GET /ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.HandleWebSocket(wsHub, w, r)
	})

	handler := middleware.Chain(
		mux,
		middleware.CORS,
		middleware.RequestID,
		middleware.Auth(jwtSecret),
	)

	publicMux := http.NewServeMux()
	authHandler.RegisterRoutes(publicMux)
	publicMux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	publicMux.Handle("/swagger/", docs.SwaggerHandler())
	publicMux.Handle("/swagger", docs.SwaggerHandler())

	publicMux.HandleFunc("GET /ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.HandleWebSocket(wsHub, w, r)
	})

	finalHandler := &routerWithPublic{
		publicPaths: []string{
			"/api/v1/auth/",
			"/health",
			"/swagger",
			"/ws",
		},
		publicHandler:    middleware.Chain(publicMux, middleware.CORS, middleware.RequestID),
		protectedHandler: handler,
	}

	server := &http.Server{
		Addr:         httpAddr,
		Handler:      finalHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	shutdown := make(chan struct{})
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	log.Println("Gateway started:", httpAddr)
	<-shutdown
}

type routerWithPublic struct {
	publicPaths      []string
	publicHandler    http.Handler
	protectedHandler http.Handler
}

func (r *routerWithPublic) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, path := range r.publicPaths {
		if len(req.URL.Path) >= len(path) && req.URL.Path[:len(path)] == path {
			r.publicHandler.ServeHTTP(w, req)
			return
		}
	}
	r.protectedHandler.ServeHTTP(w, req)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
