package observability

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInstrumentLogsRequest(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	defer slog.SetDefault(prev)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		SetUserID(r.Context(), 42)
		http.Error(w, `{"error":"task not found"}`, http.StatusNotFound)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/7", nil)
	req.Header.Set("X-Request-ID", "req-123")
	rec := httptest.NewRecorder()
	Instrument(mux).ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); got != "req-123" {
		t.Fatalf("X-Request-ID = %q, want req-123", got)
	}

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("log line is not JSON: %v\n%s", err, buf.String())
	}
	want := map[string]any{
		"level":      "WARN",
		"request_id": "req-123",
		"route":      "/api/v1/tasks/{id}",
		"path":       "/api/v1/tasks/7",
		"status":     float64(404),
		"user_id":    "42",
		"error":      "task not found",
	}
	for k, v := range want {
		if entry[k] != v {
			t.Errorf("%s = %v, want %v", k, entry[k], v)
		}
	}
}

func TestInstrumentGeneratesRequestID(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	defer slog.SetDefault(prev)

	var forwarded string
	h := Instrument(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forwarded = r.Header.Get("X-Request-ID") // шлюз передаёт его дальше в сервис
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/nope", nil))

	id := rec.Header().Get("X-Request-ID")
	if id == "" || id != forwarded {
		t.Fatalf("response id %q, forwarded id %q: want equal and non-empty", id, forwarded)
	}
}
