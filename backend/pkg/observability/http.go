package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Сколько байт тела ответа с ошибкой сохранять для лога.
const maxErrorBody = 512

type ctxKey struct{}

// requestInfo — данные, которые обработчики дописывают по ходу запроса,
// а Instrument в конце пишет в итоговую строку лога.
type requestInfo struct {
	userID string
	err    error
}

// SetUserID добавляет пользователя в итоговую строку лога запроса. Вызывается из Auth middleware.
func SetUserID(ctx context.Context, userID int64) {
	if ri, ok := ctx.Value(ctxKey{}).(*requestInfo); ok {
		ri.userID = strconv.FormatInt(userID, 10)
	}
}

// SetError добавляет внутреннюю ошибку в итоговую строку лога, когда клиенту
// отдаётся общее сообщение (например, ошибка проксирования в шлюзе).
func SetError(ctx context.Context, err error) {
	if ri, ok := ctx.Value(ctxKey{}).(*requestInfo); ok {
		ri.err = err
	}
}

// recorder запоминает код ответа и начало тела, если это ошибка:
// обработчики пишут причину в тело ({"error":"..."}), и так она попадает в лог.
type recorder struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (rec *recorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

func (rec *recorder) Write(b []byte) (int, error) {
	if rec.status >= 400 && rec.body.Len() < maxErrorBody {
		rec.body.Write(b[:min(len(b), maxErrorBody-rec.body.Len())])
	}
	return rec.ResponseWriter.Write(b)
}

// Flush и Unwrap нужны reverse proxy шлюза и http.ResponseController.
func (rec *recorder) Flush() {
	if f, ok := rec.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (rec *recorder) Unwrap() http.ResponseWriter { return rec.ResponseWriter }

// Instrument оборачивает HTTP-обработчик: обеспечивает request_id, считает метрики
// http_requests_total / http_request_duration_seconds и пишет строку access-лога.
//
// Метка route берётся из шаблона ServeMux (r.Pattern), например "/api/v1/tasks/{id}",
// а не из реального пути: иначе каждый id порождал бы новый временной ряд.
func Instrument(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// WebSocket живёт часами и требует Hijack — такие соединения не измеряем.
		if r.Header.Get("Upgrade") != "" {
			next.ServeHTTP(w, r)
			return
		}
		start := time.Now()

		// request_id: от внешнего middleware RequestID, от клиента/шлюза или новый.
		// Кладём его и в заголовок запроса, чтобы шлюз передал его дальше в сервис.
		id := w.Header().Get("X-Request-ID")
		if id == "" {
			id = r.Header.Get("X-Request-ID")
		}
		if id == "" {
			id = uuid.NewString()
		}
		r.Header.Set("X-Request-ID", id)
		w.Header().Set("X-Request-ID", id)

		ri := &requestInfo{}
		r = r.WithContext(context.WithValue(r.Context(), ctxKey{}, ri))
		rec := &recorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		// ServeMux записывает найденный шаблон в r.Pattern: "GET /api/v1/tasks/{id}".
		route := r.Pattern
		if _, path, ok := strings.Cut(route, " "); ok {
			route = path
		}
		if route == "" {
			route = "unmatched"
		}
		duration := time.Since(start)

		httpRequests.WithLabelValues(r.Method, route, strconv.Itoa(rec.status)).Inc()
		httpDuration.WithLabelValues(r.Method, route).Observe(duration.Seconds())

		level := slog.LevelInfo
		switch {
		case rec.status >= 500:
			level = slog.LevelError
		case rec.status >= 400:
			level = slog.LevelWarn
		}
		attrs := []slog.Attr{
			slog.String("request_id", id),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("route", route),
			slog.Int("status", rec.status),
			slog.Float64("duration_ms", float64(duration.Microseconds())/1000),
			slog.String("user_id", ri.userID),
		}
		if msg := errorMessage(ri.err, rec.body.Bytes()); msg != "" {
			attrs = append(attrs, slog.String("error", msg))
		}
		slog.LogAttrs(r.Context(), level, "request completed", attrs...)
	})
}

// errorMessage возвращает текст ошибки: явно переданную через SetError,
// иначе поле "error" из JSON-тела ответа, иначе само тело.
func errorMessage(err error, body []byte) string {
	if err != nil {
		return err.Error()
	}
	var payload struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(body, &payload) == nil && payload.Error != "" {
		return payload.Error
	}
	return strings.TrimSpace(string(body))
}
