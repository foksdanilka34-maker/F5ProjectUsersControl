package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/observability"
)

func main() {
	closeLogs, err := observability.SetupLogging("gateway")
	if err != nil {
		log.Fatalf("failed to set up logging: %v", err)
	}
	defer closeLogs()

	httpAddr := getEnv("HTTP_ADDR", ":8080")
	employeeURLStr := getEnv("EMPLOYEE_SERVICE_URL", "http://localhost:8081")
	businessURLStr := getEnv("BUSINESS_SERVICE_URL", "http://localhost:8082")

	employeeProxy, err := newProxy("employee", employeeURLStr)
	if err != nil {
		log.Fatalf("invalid EMPLOYEE_SERVICE_URL: %v", err)
	}
	businessProxy, err := newProxy("business", businessURLStr)
	if err != nil {
		log.Fatalf("invalid BUSINESS_SERVICE_URL: %v", err)
	}

	startTime := time.Now()
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"gateway": "active",
			"uptime":  time.Since(startTime).Truncate(time.Second).String(),
		})
	})

	// WebSocket
	mux.Handle("/ws", businessProxy)

	// Маршрутизация по префиксу. Каждый префикс зарегистрирован как шаблон ServeMux
	// (сам путь и всё поддерево), поэтому в метриках шлюза метка route — это префикс.
	routes := map[string]http.Handler{
		// Employee Service
		"/api/v1/auth":        employeeProxy,
		"/api/v1/profiles":    employeeProxy,
		"/api/v1/departments": employeeProxy,
		"/api/v1/positions":   employeeProxy,
		"/api/v1/skills":      employeeProxy,
		// Business Service
		"/api/v1/projects":   businessProxy,
		"/api/v1/tasks":      businessProxy,
		"/api/v1/analytics":  businessProxy,
		"/api/v1/gitlab":     businessProxy,
		"/api/v1/extensions": businessProxy,
	}
	for prefix, h := range routes {
		mux.Handle(prefix, h)
		mux.Handle(prefix+"/", h)
	}

	// /metrics — вне Instrument: опрос Prometheus не попадает в метрики и логи запросов.
	root := http.NewServeMux()
	root.Handle("GET /metrics", observability.MetricsHandler())
	root.Handle("/", observability.Instrument(mux))

	server := &http.Server{
		Addr:         httpAddr,
		Handler:      root,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("API Gateway listening on %s (Employee: %s, Business: %s)", httpAddr, employeeURLStr, businessURLStr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("gateway server error: %v", err)
	}
}

// newProxy создаёт reverse proxy к сервису. Если сервис недоступен, клиент получает
// 502 с общим сообщением, а настоящая причина (например, connection refused)
// попадает в строку лога шлюза вместе с request_id.
func newProxy(name, rawURL string) (*httputil.ReverseProxy, error) {
	target, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		observability.SetError(r.Context(), fmt.Errorf("proxy to %s service: %w", name, err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":"` + name + ` service unavailable"}`))
	}
	// Шлюз уже выставил X-Request-ID (тот же, что передан сервису) — убираем дубль из ответа сервиса.
	proxy.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Del("X-Request-ID")
		return nil
	}
	return proxy, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
