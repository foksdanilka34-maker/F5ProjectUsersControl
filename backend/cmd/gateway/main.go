package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	httpAddr := getEnv("HTTP_ADDR", ":8080")
	employeeURLStr := getEnv("EMPLOYEE_SERVICE_URL", "http://localhost:8081")
	businessURLStr := getEnv("BUSINESS_SERVICE_URL", "http://localhost:8082")

	employeeURL, err := url.Parse(employeeURLStr)
	if err != nil {
		log.Fatalf("invalid EMPLOYEE_SERVICE_URL: %v", err)
	}

	businessURL, err := url.Parse(businessURLStr)
	if err != nil {
		log.Fatalf("invalid BUSINESS_SERVICE_URL: %v", err)
	}

	employeeProxy := httputil.NewSingleHostReverseProxy(employeeURL)
	businessProxy := httputil.NewSingleHostReverseProxy(businessURL)

	mux := http.NewServeMux()

	// Route to Employee Service
	employeeHandler := func(w http.ResponseWriter, r *http.Request) {
		employeeProxy.ServeHTTP(w, r)
	}

	// Route to Business Service
	businessHandler := func(w http.ResponseWriter, r *http.Request) {
		businessProxy.ServeHTTP(w, r)
	}

	// Dispatcher
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Health
		if path == "/health" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok","gateway":"active"}`))
			return
		}

		// WebSocket
		if path == "/ws" {
			businessProxy.ServeHTTP(w, r)
			return
		}

		// Employee Service Routes
		if strings.HasPrefix(path, "/api/v1/auth") ||
			strings.HasPrefix(path, "/api/v1/profiles") ||
			strings.HasPrefix(path, "/api/v1/departments") ||
			strings.HasPrefix(path, "/api/v1/positions") ||
			strings.HasPrefix(path, "/api/v1/skills") {
			employeeHandler(w, r)
			return
		}

		// Business Service Routes
		if strings.HasPrefix(path, "/api/v1/projects") ||
			strings.HasPrefix(path, "/api/v1/tasks") ||
			strings.HasPrefix(path, "/api/v1/analytics") {
			businessHandler(w, r)
			return
		}

		http.NotFound(w, r)
	})

	server := &http.Server{
		Addr:         httpAddr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("API Gateway listening on %s (Employee: %s, Business: %s)", httpAddr, employeeURLStr, businessURLStr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("gateway server error: %v", err)
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
