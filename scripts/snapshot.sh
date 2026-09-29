#!/usr/bin/env bash
# «Сохранить улики» перед перезапуском: состояние сервисов, БД, брокера, алертов и логов
# в один архив, который можно приложить к тикету.
#
#   ./scripts/snapshot.sh [период логов, по умолчанию 30m]
#
# Результат: incidents/incident-<дата-время>.tar.gz. Каждый шаг необязательный:
# если что-то недоступно, в файле будет ошибка, а снимок всё равно соберётся.
set -uo pipefail
source "$(dirname "$0")/lib.sh"
need curl
need jq

since=${1:-30m}
name="incident-$(date +%Y%m%d-%H%M%S)"
dir="$ROOT_DIR/incidents/$name"
mkdir -p "$dir"

# step <файл> <команда...> — выполнить и сохранить вывод, не прерываясь на ошибках.
step() {
  local file=$1; shift
  printf '  %-28s' "$file"
  if "$@" >"$dir/$file" 2>&1; then echo "${GREEN}ok${RESET}"; else echo "${YELLOW}частично/недоступно${RESET}"; fi
}

prom_query() {
  curl -sfG "$PROM_URL/api/v1/query" --data-urlencode "query=$1" |
    jq -r '.data.result[] | "\(.metric | del(.__name__) | tostring)  \(.value[1])"'
}

key_metrics() {
  echo "# Доля 5xx по сервисам (5m)"
  prom_query 'sum by (job) (rate(http_requests_total{status=~"5.."}[5m])) / sum by (job) (rate(http_requests_total[5m]))'
  echo "# P95 по сервисам (5m), секунды"
  prom_query 'histogram_quantile(0.95, sum by (job, le) (rate(http_request_duration_seconds_bucket[5m])))'
  echo "# Пул БД: занято / максимум"
  prom_query 'pgxpool_acquired_conns'
  prom_query 'pgxpool_max_conns'
  echo "# Outbox failed за 30m"
  prom_query 'sum by (job, event_type) (increase(outbox_events_total{result="failed"}[30m]))'
  echo "# Health-check"
  prom_query 'probe_success{job="health"}'
}

loki_errors() {
  curl -sfG "$LOKI_URL/loki/api/v1/query_range" \
    --data-urlencode 'query={service_name=~"gateway|employee|business"} | json | status >= 500 or level = "ERROR"' \
    --data-urlencode "since=$since" --data-urlencode "limit=2000" |
    jq -r '.data.result[] | .stream.service_name as $s | .values[] | "\($s) \(.[1])"' | sort
}

pg_activity() {
  local db
  for db in identity business; do
    echo "## $db"
    psql_db "$db" -P footer=off -c "
      SELECT pid, state, wait_event_type, now() - state_change AS in_state,
             now() - query_start AS query_age, left(query, 120) AS query
      FROM pg_stat_activity WHERE datname = '$db' AND pid <> pg_backend_pid()
      ORDER BY state_change"
  done
}

tail_logs() {
  local f
  for f in "$ROOT_DIR"/logs/*.log; do
    [ -e "$f" ] || continue
    echo "===== $(basename "$f") ====="
    tail -n 2000 "$f"
  done
}

echo "${BOLD}Снимок $name${RESET}"
step health.txt           "$(dirname "$0")/healthcheck.sh"
step key-metrics.txt      key_metrics
step alerts.json          curl -sf "$ALERTMANAGER_URL/api/v2/alerts"
step errors-$since.log    loki_errors
step logs-tail.log        tail_logs
step pg_stat_activity.txt pg_activity
step outbox.txt           "$(dirname "$0")/outbox-status.sh"
step rabbitmq-queues.txt  docker exec rabbitmq rabbitmqctl -q list_queues name messages messages_unacknowledged consumers
step containers.txt       docker ps -a --format 'table {{.Names}}\t{{.Status}}\t{{.Image}}'

tar -czf "$dir.tar.gz" -C "$ROOT_DIR/incidents" "$name" && rm -rf "$dir"
echo "Готово: ${BOLD}incidents/$name.tar.gz${RESET}"
