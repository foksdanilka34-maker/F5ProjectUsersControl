# Наблюдаемость F5: метрики, логи, алерты

Здесь собран стек, который показывает, что происходит внутри F5:
- **Prometheus** — метрики;
- **Loki + Grafana Alloy** — логи;
- **Alertmanager** — алерты;
- **Grafana** — дашборд и поиск по логам.

Сервисы F5 запускаются как обычно (`make run-*`), стек поднимается рядом одной командой.

## Быстрый запуск

```bash
make docker-up        # Postgres и RabbitMQ (как раньше)
make obs-up           # стек наблюдаемости
make run-employee     # в отдельных терминалах
make run-business
make run-gateway
```

`make run-*` теперь задают `LOG_DIR=../logs`: сервисы пишут JSON-логи и в консоль, и в `logs/<service>.log`, откуда их забирает Alloy.

| Что | Адрес |
|---|---|
| **Grafana**, дашборд F5 Services | http://localhost:3000/d/f5-services (admin / admin) |
| Prometheus | http://localhost:9090 (Status → Targets, Alerts) |
| Alertmanager | http://localhost:9093 |
| Alloy | http://localhost:12345 (граф сбора логов) |
| blackbox exporter | http://localhost:9115 |
| Loki | http://localhost:3100 (только API) |
| Метрики сервисов | http://localhost:8080/metrics, :8081/metrics, :8082/metrics |

```bash
make obs-down                                          # остановить стек
docker compose -f observability/compose.yaml down -v   # ... и удалить накопленные метрики и логи
```

## Как текут данные

```
 браузер / frontend
        │  X-Request-ID
        ▼
   ┌─────────┐  проксирует, передаёт тот же X-Request-ID  ┌──────────┐  ┌──────────┐
   │ gateway │──────────────────────────────────────────►│ employee │  │ business │
   │  :8080  │                                           │  :8081   │  │  :8082   │
   └────┬────┘                                           └────┬─────┘  └────┬─────┘
        │ /metrics, logs/gateway.log                          │             │
        ▼                                                     ▼             ▼
   Prometheus ◄──── pull /metrics каждые 15 с ─────── все три сервиса (host.docker.internal)
      │   └──── blackbox ── GET /health каждого сервиса → probe_success
      │ алерты
      ▼
   Alertmanager                Alloy ── tail logs/*.log + docker-логи Postgres/RabbitMQ ──► Loki
                                                                                           │
   Grafana ◄────────────────────── PromQL / LogQL ──────────────────────────────────────────┘
```

- **Метрики (pull).** Каждый сервис считает метрики в памяти и отдаёт их на `/metrics`. Prometheus опрашивает их раз в 15 секунд. Имя сервиса — метка `job` (`gateway`, `employee`, `business`).
- **Логи (push).** Сервисы пишут JSON в `logs/*.log`, Alloy читает файлы и отправляет строки в Loki. Логи контейнеров Postgres и RabbitMQ Alloy берёт через Docker API. В Loki у каждой строки одна метка — `service_name`.
- **Алерты.** Prometheus вычисляет правила из `prometheus/alerts.yml`. Сработавшие алерты уходят в Alertmanager: critical — получателю `oncall`, warning — `tickets`. У получателей нет интеграций, алерты видны в UI Alertmanager и в Grafana (**Alerting → Active notifications**, Alertmanager: `Alertmanager`).

## Что добавлено в код

Всё собрано в пакете `backend/pkg/observability`, остальной код почти не изменился.

| Файл | Что делает |
|---|---|
| `logging.go` | `SetupLogging(service)` включает JSON-логи `log/slog` с полем `service`. `slog.SetDefault` перенаправляет и стандартный `log.Printf`, поэтому существующие сообщения outbox, воркеров и консьюмеров стали JSON без переписывания |
| `http.go` | `Instrument` — HTTP middleware: request_id, метрики запросов и итоговая строка access-лога |
| `metrics.go` | Объявления метрик, `/metrics`, статистика пула pgxpool |

Точечные изменения:
- `cmd/*/main.go` — вызовы `SetupLogging`, `RegisterPgxPool`, `Instrument` и маршрут `/metrics` вне цепочки middleware;
- `Auth` middleware — `SetUserID`;
- outbox-пулеры, воркер вебхуков GitLab, доставка расширениям, WebSocket-хаб — по одной строке со счётчиком.

Шлюз вместо ручного `strings.HasPrefix` регистрирует префиксы как шаблоны ServeMux. Поведение то же, но у метрик шлюза появилась метка маршрута.

### Логи

На каждый HTTP-запрос пишется одна итоговая строка:

```json
{"time":"2026-09-29T10:15:02.114Z","level":"WARN","msg":"request completed","service":"business",
 "request_id":"5d0c…","method":"GET","path":"/api/v1/tasks/7","route":"/api/v1/tasks/{id}",
 "status":404,"duration_ms":3.2,"user_id":"42","error":"task not found"}
```

- **Уровень:** 2xx → INFO, 4xx → WARN, 5xx → ERROR.
- **error:** хендлеры отвечают `http.Error(w, {"error":"..."})`, а middleware достаёт текст ошибки из тела ответа. Поэтому переписывать хендлеры не пришлось.
- **request_id:** шлюз берёт его из заголовка `X-Request-ID` клиента или генерирует сам и передаёт тот же id в сервис. В логах gateway и сервиса у одного запроса один request_id.
- **user_id:** появляется, когда запрос прошёл `Auth` middleware.
- **Остальные сообщения** (outbox, GitLab, RabbitMQ, WebSocket) — это прежние `log.Printf`, теперь в JSON с полем `service`, но без request_id: они не связаны с HTTP-запросом.

### Метрики

| Метрика | Метки | Что показывает |
|---|---|---|
| `http_requests_total` | method, route, status | RPS, доля ошибок |
| `http_request_duration_seconds` | method, route | Латентность (P50/P95/P99) |
| `pgxpool_acquired_conns`, `pgxpool_idle_conns`, `pgxpool_max_conns` | — | Пул соединений с БД (максимум 10) |
| `pgxpool_empty_acquire_total`, `pgxpool_acquire_duration_seconds_total` | — | Сколько раз и сколько времени запросы ждали соединение |
| `outbox_events_total` | event_type, result | Публикация событий outbox в RabbitMQ: `published` / `failed` |
| `gitlab_webhook_events_total` | event_type, result | Обработка вебхуков GitLab: `processed` / `failed` |
| `extension_deliveries_total` | extension, result | Доставка вебхуков расширениям: `delivered` / `rejected` / `failed` |
| `websocket_clients` | — | Открытые WebSocket-соединения |

`route` — шаблон маршрута (`/api/v1/tasks/{id}`), а не реальный путь. Иначе каждая задача порождала бы новый временной ряд. По той же причине `user_id` и `request_id` есть только в логах.

Фоновые метрики важнее, чем кажется: если outbox не может достучаться до RabbitMQ, пользователь видит 201 Created, а событие никуда не ушло. Расширения не получат вебхук, business не узнает о новом сотруднике. HTTP-графики при этом зелёные.

## Дашборд F5 Services

| Секция | Что смотреть |
|---|---|
| Верхний ряд | RPS, доля 5xx (как в алерте), P95, health каждого сервиса, заполненность пула БД |
| HTTP | RPS и доля 5xx по сервисам; коды ответов (401/403/429/5xx); RPS и 5xx по маршрутам |
| Латентность | P50/P95/P99 для каждого сервиса; P95 по маршрутам |
| База данных | Соединения пула и ожидания свободного соединения |
| Фоновые процессы | Outbox, вебхуки GitLab, доставка расширениям, WebSocket; строки логов с fail/error от воркеров |
| Логи запросов | Логи с фильтрами из переменных вверху: сервис, маршрут, уровень, статус, user_id, request_id |

От графика к логам есть три пути:
1. Выделить участок на любом графике — весь дашборд, включая логи, переключится на этот интервал.
2. Клик по точке на «Доля 5xx по сервисам» или «Ошибки 5xx по маршрутам» — ссылка открывает Explore с логами 5xx этого сервиса или маршрута за тот же интервал.
3. В строке лога раскрыть детали → **Links → Все логи этого запроса**: строки gateway и сервиса с этим request_id.

## Полезные запросы

```logql
# Все логи одного запроса (gateway + сервис)
{service_name=~".+"} |= "<request_id>"

# Ошибки 5xx по сервису, маршруту и тексту ошибки за 15 минут
sum by (service_name, route, error) (count_over_time({service_name=~"gateway|employee|business"} | json | status >= 500 [15m]))

# Всё, что делал пользователь 42
{service_name=~"gateway|employee|business"} | json | user_id="42"

# Ошибки фоновых процессов
{service_name=~"employee|business"} != "request completed" |~ "(?i)fail|error"

# Логи Postgres и RabbitMQ
{service_name=~"identity_db|business_db|rabbitmq"}
```

```promql
# Доля 5xx по сервисам
sum by (job) (rate(http_requests_total{status=~"5.."}[5m])) / sum by (job) (rate(http_requests_total[5m]))

# Самые медленные маршруты
topk(5, histogram_quantile(0.95, sum by (job, route, le) (rate(http_request_duration_seconds_bucket[5m]))))

# Неопубликованные события outbox за час
sum by (job, event_type) (increase(outbox_events_total{result="failed"}[1h]))
```

## Учебные сбои

В проекте уже есть `Chaos` middleware. Он подключён к маршрутам employee и business и включается переменными окружения. Ниже сбои, которые можно устроить без изменения кода.

| Как сломать | Что увидите | Какой алерт |
|---|---|---|
| `ENABLE_CHAOS=true CHAOS_FAILURE_RATE=10 make run-business` | ~10% запросов business → 500 `chaos: simulated server error` | `F5HighErrorRate` (critical) для business и gateway |
| `ENABLE_CHAOS=true CHAOS_MAX_LATENCY_MS=1500 make run-employee` | Задержки до 1.5 с на маршрутах employee | `F5HighLatencyP95` |
| Остановить employee (Ctrl+C) | Шлюз отвечает 502, в логе gateway `proxy to employee service: … connection refused` | `F5ServiceDown` (employee) |
| `docker compose -f backend/docker-compose.yml stop rabbitmq` | После создания или изменения задач `outbox_events_total{result="failed"}` растёт, HTTP при этом зелёный | `F5OutboxPublishFailures` |
| `docker compose -f backend/docker-compose.yml stop business_db` | business отвечает 500, но `/health` остаётся UP: он не проверяет БД | `F5HighErrorRate`, но **не** `F5ServiceDown` |

Два последних случая особенно полезны для тренировки:
- **RabbitMQ.** Посмотрите, прекращаются ли ошибки outbox после `start rabbitmq`. Если нет, клиент не переподключается сам. Это находка для тикета.
- **БД.** Health-check, который не проверяет зависимости, говорит «всё хорошо», когда сервис бесполезен. Тоже хороший повод для тикета.

---

## Runbooks

Первое действие по любому алерту — открыть дашборд **F5 Services** на интервале начала проблемы.

### F5HighErrorRate

Больше 1% (warning) или 5% (critical) ответов сервиса (`job`) за 5 минут — 5xx.

1. **Какой сервис?** Если горит и gateway, и employee/business, первопричина в сервисе, gateway только пробрасывает ошибку. Если горит только gateway, смотрите 502 в его логах: сервис недоступен.
2. **Какой маршрут и какая ошибка?** Панель «Ошибки 5xx по маршрутам» → клик → логи. Или:
   ```logql
   sum by (route, error) (count_over_time({service_name="business"} | json | status >= 500 [15m]))
   ```
3. **Один `error` на всех маршрутах** (например, `failed to connect`, `connection refused`) — проблема с БД или RabbitMQ. Проверьте секцию «База данных» и логи `{service_name=~"identity_db|business_db|rabbitmq"}`.
4. **`chaos: simulated server error`** — включён `ENABLE_CHAOS`.
5. Возьмите 1–2 request_id и откройте «Все логи этого запроса».

**Эскалация.** warning — тикет команде бэкенда: маршрут, текст ошибки, request_id, время начала. critical — дежурному разработчику сразу. Если причина в БД или брокере — ещё и тем, кто отвечает за инфраструктуру.

### F5HighLatencyP95

P95 маршрута (`job`, `route`) выше 500 мс в течение 5 минут.

1. **Один маршрут или все маршруты сервиса?** Если все, смотрите пул БД: `pgxpool_acquired_conns` = max и рост «Ожидания соединения» означают, что запросы ждут соединение.
2. **Реальные длительности:**
   ```logql
   {service_name="business"} | json | route="/api/v1/analytics/dashboard" | duration_ms > 500
   ```
3. **Медленные запросы в Postgres:**
   ```bash
   docker exec business_db psql -U business -c "SELECT state, now()-query_start AS age, left(query,80) FROM pg_stat_activity WHERE state <> 'idle' ORDER BY age DESC"
   ```

**Эскалация.** Тикет команде бэкенда: маршрут, P95, примеры `duration_ms`, что показала БД.

### F5ServiceDown

blackbox больше минуты не получает 2xx от `/health` сервиса (`service`).

1. Жив ли процесс? Проверьте терминал с `make run-*` и выполните `curl -i localhost:808X/health`.
2. `connection refused` — процесс не запущен или упал. Последние строки лога: `{service_name="employee"}`.
3. Если процесс работает, а проба падает, проверьте саму пробу: http://localhost:9115/probe?module=http_2xx&target=http://host.docker.internal:8081/health&debug=true

`/health` проверяет только, что процесс отвечает, но не БД и не RabbitMQ. Если сервис жив, но не работает, этот алерт не сработает, зато сработает `F5HighErrorRate`.

**Эскалация.** critical — дежурному разработчику сразу.

### F5OutboxPublishFailures

Outbox-пулер больше 5 минут не может опубликовать события (`event_type`) в RabbitMQ. Данные в БД сохранены, но другие части системы о них не узнают.

1. Доступен ли RabbitMQ? `docker compose -f backend/docker-compose.yml ps rabbitmq`, UI http://localhost:15672.
2. Текст ошибки:
   ```logql
   {service_name=~"employee|business"} |= "Outbox Poller" |= "Failed to publish"
   ```
3. Если RabbitMQ уже поднят, а ошибки продолжаются, проверьте, переподключается ли клиент сам, или нужен перезапуск сервиса.
4. Сколько событий застряло — таблица outbox в БД сервиса: записи с ошибкой.

**Эскалация.** Тикет с высоким приоритетом: какие события не ушли, с какого времени, текст ошибки. Если RabbitMQ недоступен — инфраструктура.

### F5GitLabWebhookFailures

За 15 минут больше 3 вебхуков GitLab (`event_type`) обработаны с ошибкой: задачи не двигаются по доске по событиям push/MR/pipeline.

1. Текст ошибки:
   ```logql
   {service_name="business"} |= "GitLab Webhook Worker" |= "Failed to process"
   ```
2. Типичные причины: интеграция проекта удалена или отключена, неверный токен GitLab, недоступен инстанс GitLab.

**Эскалация.** Тикет команде бэкенда: тип события, проект, текст ошибки.

### F5DBPoolSaturated

Больше 90% соединений пула pgxpool сервиса (`job`) заняты дольше 2 минут. Скоро запросы начнут ждать соединение и падать по таймауту.

1. Секция «База данных»: растут ли ожидания соединения?
2. Кто держит соединения:
   ```bash
   docker exec business_db psql -U business -c "SELECT state, count(*), max(now()-state_change) FROM pg_stat_activity WHERE datname='business' GROUP BY 1"
   ```
   Много `idle in transaction` означает, что транзакции не закрываются (утечка). Много `active` с большим возрастом — медленные запросы.

**Эскалация.** Тикет команде бэкенда с выводом `pg_stat_activity`. Если одновременно растут 5xx — сразу дежурному.

---

## Версии

| Компонент | Образ |
|---|---|
| Prometheus | `prom/prometheus:v3.15.0` |
| blackbox exporter | `prom/blackbox-exporter:v0.28.0` |
| Alertmanager | `prom/alertmanager:v0.34.1` |
| Loki | `grafana/loki:3.7.8` |
| Alloy | `grafana/alloy:v1.20.1` |
| Grafana | `grafana/grafana:13.2.2` |

Go: `github.com/prometheus/client_golang` v1.23.2 (последняя версия, совместимая с `go 1.24` в `go.mod`).
