package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// demo-extension эмулирует стороннее приложение, подключаемое к F5 Project & Team Control
// через платформу расширений: получает вебхуки о событиях задач, отдаёт свою панель
// в карточку задачи и вкладку в проект, читает/пишет собственные метаданные задачи.

type webhookEvent struct {
	ReceivedAt time.Time       `json:"received_at"`
	Event      string          `json:"event"`
	Data       json.RawMessage `json:"data"`
}

type app struct {
	f5BaseURL      string
	extensionKey   string
	sharedSecret   string
	mu             sync.Mutex
	receivedEvents []webhookEvent
}

func main() {
	a := &app{
		f5BaseURL:    getEnv("F5_BASE_URL", "http://localhost:8080"),
		extensionKey: getEnv("EXTENSION_KEY", "demo-timer"),
		sharedSecret: getEnv("EXTENSION_SECRET", "demo-secret-change-me"),
	}
	port := getEnv("PORT", "8090")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /panel", a.handlePanel)
	mux.HandleFunc("GET /tab", a.handleTab)
	mux.HandleFunc("POST /webhooks", a.handleWebhook)
	mux.HandleFunc("POST /api/checklist", a.handleSetChecklist)
	mux.HandleFunc("GET /api/checklist", a.handleGetChecklist)

	log.Printf("[demo-extension] listening on :%s (F5_BASE_URL=%s, key=%s)", port, a.f5BaseURL, a.extensionKey)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

var panelTemplate = template.Must(template.New("panel").Parse(`
<!DOCTYPE html>
<html><head><meta charset="utf-8"><style>
  body { font-family: system-ui, sans-serif; margin: 0; padding: 12px; background: #fff8f0; }
  h4 { margin: 0 0 8px; font-size: 13px; color: #92400e; }
  label { display: flex; align-items: center; gap: 6px; font-size: 13px; color: #374151; margin-bottom: 4px; }
  button { margin-top: 8px; font-size: 12px; padding: 4px 10px; border-radius: 8px; border: 1px solid #d97706; background: #fff; color: #92400e; cursor: pointer; }
  .saved { color: #059669; font-size: 11px; margin-left: 8px; }
</style></head><body>
  <h4>🧩 Демо-чеклист (task {{.TaskID}})</h4>
  <div id="items">Загрузка...</div>
  <button onclick="save()">Сохранить</button>
  <span class="saved" id="saved"></span>
  <script>
    const taskId = {{.TaskID}};
    const base = {{.Base}};
    const items = ["Написать тесты", "Обновить документацию", "Прогнать линтер"];
    fetch(base + "/api/checklist?task_id=" + taskId).then(r => r.json()).then(data => {
      const checked = data.checked || [];
      document.getElementById("items").innerHTML = items.map((it, i) =>
        '<label><input type="checkbox" data-i="' + i + '" ' + (checked.includes(i) ? "checked" : "") + '>' + it + '</label>'
      ).join("");
    });
    function save() {
      const checked = Array.from(document.querySelectorAll('input[type=checkbox]:checked')).map(el => Number(el.dataset.i));
      fetch(base + "/api/checklist?task_id=" + taskId, {
        method: "POST", headers: {"Content-Type": "application/json"}, body: JSON.stringify({checked})
      }).then(() => {
        document.getElementById("saved").textContent = "Сохранено ✓";
        setTimeout(() => document.getElementById("saved").textContent = "", 2000);
      });
    }
  </script>
</body></html>
`))

func (a *app) handlePanel(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = panelTemplate.Execute(w, map[string]any{"TaskID": template.JS(taskID), "Base": template.JSStr(selfBaseURL(r))})
}

var tabTemplate = template.Must(template.New("tab").Parse(`
<!DOCTYPE html>
<html><head><meta charset="utf-8"><style>
  body { font-family: system-ui, sans-serif; margin: 0; padding: 16px; }
  h3 { font-size: 14px; color: #111827; }
  table { width: 100%; border-collapse: collapse; font-size: 12px; }
  th, td { text-align: left; padding: 6px 8px; border-bottom: 1px solid #f3f4f6; }
  th { color: #6b7280; font-weight: 500; }
  code { background: #f3f4f6; padding: 1px 4px; border-radius: 4px; }
</style></head><body>
  <h3>🧩 Демо-расширение — журнал событий проекта {{.ProjectID}}</h3>
  <table>
    <thead><tr><th>Получено</th><th>Событие</th><th>Задача</th></tr></thead>
    <tbody>
    {{range .Events}}
      <tr><td>{{.ReceivedAt.Format "15:04:05"}}</td><td><code>{{.Event}}</code></td><td>{{.TaskID}}</td></tr>
    {{else}}
      <tr><td colspan="3">Событий пока нет — создайте или переместите задачу в проекте.</td></tr>
    {{end}}
    </tbody>
  </table>
</body></html>
`))

func (a *app) handleTab(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")

	a.mu.Lock()
	events := make([]struct {
		webhookEvent
		TaskID int64
	}, 0, len(a.receivedEvents))
	for _, e := range a.receivedEvents {
		var d struct {
			TaskID int64 `json:"task_id"`
		}
		_ = json.Unmarshal(e.Data, &d)
		events = append(events, struct {
			webhookEvent
			TaskID int64
		}{e, d.TaskID})
	}
	a.mu.Unlock()

	for i, j := 0, len(events)-1; i < j; i, j = i+1, j-1 {
		events[i], events[j] = events[j], events[i]
	}
	if len(events) > 20 {
		events = events[:20]
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tabTemplate.Execute(w, map[string]any{"ProjectID": projectID, "Events": events})
}

func (a *app) handleWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	expected := sign(a.sharedSecret, body)
	got := r.Header.Get("X-F5-Signature")
	if !hmac.Equal([]byte(expected), []byte(got)) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	var envelope struct {
		Event string          `json:"event"`
		Data  json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	a.mu.Lock()
	a.receivedEvents = append(a.receivedEvents, webhookEvent{
		ReceivedAt: time.Now(),
		Event:      envelope.Event,
		Data:       envelope.Data,
	})
	a.mu.Unlock()

	log.Printf("[demo-extension] received %s: %s", envelope.Event, string(envelope.Data))
	w.WriteHeader(http.StatusOK)
}

func (a *app) handleSetChecklist(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<16))
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	if err := a.callF5(r.Context(), "/api/v1/extensions/properties/tasks/"+taskID+"/checklist", body); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *app) handleGetChecklist(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")

	value, err := a.callF5Get(r.Context(), "/api/v1/extensions/properties/tasks/"+taskID+"/checklist")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"checked":[]}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(value)
}

func (a *app) callF5(ctx context.Context, path string, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, a.f5BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	a.setAuthHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("f5 responded with status %d", resp.StatusCode)
	}
	return nil
}

func (a *app) callF5Get(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.f5BaseURL+path, nil)
	if err != nil {
		return nil, err
	}
	a.setAuthHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("f5 responded with status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func (a *app) setAuthHeaders(req *http.Request) {
	req.Header.Set("X-Extension-Key", a.extensionKey)
	req.Header.Set("X-Extension-Secret", a.sharedSecret)
}

func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func selfBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
