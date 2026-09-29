package observability

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Метрики регистрируются в стандартном реестре client_golang и отдаются на /metrics.
// Какой сервис их отдал, Prometheus записывает в метку job (см. observability/prometheus).
var (
	httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "HTTP requests processed, by method, route and status code.",
	}, []string{"method", "route", "status"})

	// Граница 0.5 нужна, чтобы алерт «P95 > 500 мс» считался точно.
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds, by method and route.",
		Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	}, []string{"method", "route"})
)

// Фоновые конвейеры. Их сбои не видны в HTTP-ответах: задача сохранилась, а событие
// не ушло в RabbitMQ, вебхук GitLab не обработался, расширение не получило уведомление.
var (
	OutboxEvents = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "outbox_events_total",
		Help: "Outbox records handled by the poller, by event type and result (published|failed).",
	}, []string{"event_type", "result"})

	GitLabWebhookEvents = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gitlab_webhook_events_total",
		Help: "GitLab webhook events handled by the worker, by event type and result (processed|failed).",
	}, []string{"event_type", "result"})

	ExtensionDeliveries = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "extension_deliveries_total",
		Help: "Webhook deliveries to extensions, by extension key and result (delivered|rejected|failed).",
	}, []string{"extension", "result"})

	WebSocketClients = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "websocket_clients",
		Help: "Currently connected WebSocket clients.",
	})
)

// MetricsHandler отдаёт метрики в формате Prometheus.
func MetricsHandler() http.Handler { return promhttp.Handler() }

// RegisterPgxPool экспортирует статистику пула соединений pgxpool.
// Значения читаются из pool.Stat() в момент опроса Prometheus.
func RegisterPgxPool(pool *pgxpool.Pool) {
	stat := func(f func(*pgxpool.Stat) float64) func() float64 {
		return func() float64 { return f(pool.Stat()) }
	}
	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "pgxpool_acquired_conns", Help: "Connections currently in use by the application.",
	}, stat(func(s *pgxpool.Stat) float64 { return float64(s.AcquiredConns()) }))
	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "pgxpool_idle_conns", Help: "Idle connections in the pool.",
	}, stat(func(s *pgxpool.Stat) float64 { return float64(s.IdleConns()) }))
	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "pgxpool_max_conns", Help: "Maximum pool size.",
	}, stat(func(s *pgxpool.Stat) float64 { return float64(s.MaxConns()) }))
	promauto.NewCounterFunc(prometheus.CounterOpts{
		Name: "pgxpool_empty_acquire_total", Help: "Acquires that had to wait because the pool had no free connection.",
	}, stat(func(s *pgxpool.Stat) float64 { return float64(s.EmptyAcquireCount()) }))
	promauto.NewCounterFunc(prometheus.CounterOpts{
		Name: "pgxpool_acquire_duration_seconds_total", Help: "Total time spent acquiring connections.",
	}, stat(func(s *pgxpool.Stat) float64 { return s.AcquireDuration().Seconds() }))
}
